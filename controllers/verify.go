package controllers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/go-playground/validator/v10"
	"io"
	"issue_pr_board/config"
	"issue_pr_board/models"
	"issue_pr_board/utils"
	"math/big"
	"net/http"
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

func (c *GetCaptchaController) Get() {
	captchaId := utils.GetCaptcha()
	data := make(map[string]interface{})
	data["captcha_id"] = captchaId
	data["src"] = fmt.Sprintf("/captcha/%v.png", captchaId)
	c.ApiJsonReturn("success", 200, data)
}

type CheckCaptchaController struct {
	BaseController
}

func (c *CheckCaptchaController) Post() {
	params, err := getParams(c.Ctx.Request)
	validate := validator.New()
	validateErr := validate.Struct(params)
	if validateErr != nil {
		c.ApiJsonReturn("参数错误", 400, validateErr)
	}
	if !utils.VerifyCaptcha(params.CaptchaId, params.Challenge) {
		c.ApiJsonReturn("验证失败", 400, nil)
	}
	captchaValue := genValidateCode(6)
	timeUnix := time.Now().Unix()
	interval, _ := beego.AppConfig.Int64("verifyinterval")
	var verify models.Verify
	ep := utils.EmailParams{Receiver: params.Email, Code: captchaValue}
	if searchEmailRecord(params.Email) {
		o := orm.NewOrm()
		qs := o.QueryTable("verify")
		err = qs.Filter("addr", params.Email).One(&verify)
		if err != nil {
			c.ApiJsonReturn("系统异常，请联系管理", 400, nil)
		}
		created := verify.Created
		if (timeUnix - created) < interval {
			logs.Error("The interval between two verifications cannot be less than 1 minute, addr:", params.Email)
			c.ApiJsonReturn("发送验证码的时间间隔不能低于一分钟", 400, nil)
		}
		go utils.SendVerifyEmail(ep)
		_, err = qs.Filter("addr", params.Email).Update(orm.Params{
			"Code":    captchaValue,
			"Created": created,
		})
		if err != nil {
			logs.Error("Fail to update verify, err:", err)
			c.ApiJsonReturn("系统异常，请联系管理", 400, nil)
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
			c.ApiJsonReturn("系统异常，请联系管理", 400, nil)
		}
	}
	c.ApiJsonReturn("邮箱验证码发送成功", 200, nil)
}

type clientParams struct {
	CaptchaId string `json:"captcha_id" validate:"max=16"`
	Challenge string `json:"challenge" validate:"len=6"`
	Email     string `json:"email" validate:"email"`
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
