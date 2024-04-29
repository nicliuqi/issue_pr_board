package controllers

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"io"
	"issue_pr_board/config"
	"issue_pr_board/models"
	"issue_pr_board/utils"
	"math/big"
	"net/http"
	"regexp"
	"time"
)

type GeneralResp struct {
	Code  int         `json:"code"`
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

func genValidateCode(width int) string {
	validateCode := ""
	for i := 0; i < width; i++ {
		randomInt, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			logs.Error("Fail to generate validate code, err:", err)
			return validateCode
		}
		validateCode += randomInt.String()
	}
	return validateCode
}

func verifyEmailFormat(email string) bool {
	pattern := `^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

func searchEmailRecord(addr string) bool {
	o := orm.NewOrm()
	searchSql := fmt.Sprintf("select * from verify where addr='%s'", addr)
	err := o.Raw(searchSql).QueryRow()
	if err == orm.ErrNoRows {
		return false
	}
	return true
}

func checkCode(addr string, code string) bool {
	o := orm.NewOrm()
	searchSql := fmt.Sprintf("select * from verify where addr='%s' and code='%s'", addr, code)
	err := o.Raw(searchSql).QueryRow()
	if err == orm.ErrNoRows {
		return false
	}
	return true
}

func cleanCode(addr string, code string) {
	o := orm.NewOrm()
	sql := fmt.Sprintf("delete from verify where addr='%s' and code='%s'", addr, code)
	_, err := o.Raw(sql).Exec()
	if err != nil {
		logs.Error("Fail to delete verify record, err:", err)
	} else {
		logs.Info("The expired record had been deleted which addr is:", addr)
	}
}

func CleanVerification() error {
	var verifications []models.Verify
	o := orm.NewOrm()
	sql := "select * from verify"
	_, err := o.Raw(sql).QueryRows(&verifications)
	if err != nil {
		logs.Error("Fail to get all verify records")
		return err
	}
	timeUnix := time.Now().Unix()
	expire := config.AppConfig.VerifyExpire
	for _, verification := range verifications {
		addr := verification.Addr
		code := verification.Code
		created := verification.Created
		if (timeUnix - created) > int64(expire) {
			cleanCode(addr, code)
		}
	}
	return nil
}

type GetCaptchaController struct {
	BaseController
}

func (c *GetCaptchaController) Post() {
	params, err := getParams(c.Ctx.Request)
	if err != nil {
		c.CustomJsonReturn(errorRes(err))
	}
	if params.CaptchaType != "blockPuzzle" {
		c.CustomJsonReturn(errorRes(errors.New("参数CaptchaType须为blockPuzzle")))
	}
	ser := utils.Factory.GetService(params.CaptchaType)
	data, err := ser.Get()
	if err != nil {
		c.CustomJsonReturn(errorRes(err))
	}
	res, err := json.Marshal(successRes(data))
	if err != nil {
		c.CustomJsonReturn(errorRes(err))
	}
	c.CustomJsonReturn(utils.JsonToMap(string(res)))
}

type CheckCaptchaController struct {
	BaseController
}

func (c *CheckCaptchaController) Post() {
	params, err := getParams(c.Ctx.Request)
	if params == nil {
		c.CustomJsonReturn(errorRes(errors.New("参数不能为空")))
	}
	if params.CaptchaType != "blockPuzzle" {
		c.CustomJsonReturn(errorRes(errors.New("参数CaptchaType须为blockPuzzle")))
	}
	if params.Token == "" || params.PointJson == "" || params.CaptchaType == "" || params.Email == "" {
		c.CustomJsonReturn(errorRes(errors.New("参数传递不完整")))
	}
	if err != nil {
		c.CustomJsonReturn(errorRes(err))
	}
	ser := utils.Factory.GetService(params.CaptchaType)
	err = ser.Check(params.Token, params.PointJson)
	if err != nil {
		c.CustomJsonReturn(errorRes(err))
	}
	if !verifyEmailFormat(params.Email) {
		logs.Error("Invalid email address:", params.Email)
		c.CustomJsonReturn(errorVerifyRes(errors.New("待验证邮箱地址非法")))
	}
	captchaValue := genValidateCode(6)
	timeUnix := time.Now().Unix()
	interval, _ := beego.AppConfig.Int64("verifyinterval")
	var verify models.Verify
	ep := utils.EmailParams{Receiver: params.Email, Code: captchaValue}
	if searchEmailRecord(params.Email) {
		o := orm.NewOrm()
		qs := o.QueryTable("verify")
		err := qs.Filter("addr", params.Email).One(&verify)
		if err != nil {
			return
		}
		created := verify.Created
		if (timeUnix - created) < interval {
			logs.Error("The interval between two verifications cannot be less than 1 minute, addr:", params.Email)
			c.CustomJsonReturn(errorVerifyRes(errors.New("发送验证码的时间间隔不能低于一分钟")))
		}
		go utils.SendVerifyEmail(ep)
		_, err = qs.Filter("addr", params.Email).Update(orm.Params{
			"Code":    captchaValue,
			"Created": created,
		})
		if err != nil {
			logs.Error("Fail to update verify, err:", err)
		}
	} else {
		verify.Addr = params.Email
		verify.Code = captchaValue
		verify.Created = timeUnix
		go utils.SendVerifyEmail(ep)
		o := orm.NewOrm()
		_, err = o.Insert(&verify)
		if err != nil {
			logs.Error("Insert verify failed, err:", err)
		}
	}
	c.CustomJsonReturn(successRes(nil))
}

type clientParams struct {
	Token       string `json:"token"`
	PointJson   string `json:"pointJson"`
	CaptchaType string `json:"captchaType"`
	Email       string `json:"email"`
}

func getParams(request *http.Request) (*clientParams, error) {
	params := &clientParams{}
	all, _ := io.ReadAll(request.Body)
	if len(all) <= 0 {
		return nil, nil
	}
	err := json.Unmarshal(all, params)
	if err != nil {
		return nil, err
	}
	return params, nil
}

func successRes(data interface{}) map[string]interface{} {
	ret := make(map[string]interface{})
	ret["error"] = false
	ret["repCode"] = "0000"
	ret["repData"] = data
	ret["repMsg"] = nil
	ret["successRes"] = true
	return ret
}

func errorRes(err error) map[string]interface{} {
	ret := make(map[string]interface{})
	ret["error"] = true
	ret["repCode"] = "0001"
	ret["repData"] = nil
	ret["repMsg"] = err.Error()
	ret["successRes"] = false
	return ret
}

func errorVerifyRes(err error) map[string]interface{} {
	ret := make(map[string]interface{})
	ret["error"] = true
	ret["repCode"] = "0002"
	ret["repData"] = nil
	ret["repMsg"] = err.Error()
	ret["successRes"] = false
	return ret
}
