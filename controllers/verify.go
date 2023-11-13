package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/chenhg5/collection"
	"issue_pr_board/models"
	"issue_pr_board/utils"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

type GeneralResp struct {
	Code  int         `json:"code"`
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

type VerifyController struct {
	BaseController
}

func genValidateCode(width int) string {
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rand.Seed(time.Now().UnixNano())

	var sb strings.Builder
	for i := 0; i < width; i++ {
		_, err := fmt.Fprintf(&sb, "%d", numeric[rand.Intn(r)])
		if err != nil {
			return ""
		}
	}
	return sb.String()
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

func (c *VerifyController) Post() {
	authorization := c.Ctx.Input.Header("Authorization")
	authResult := utils.CheckAuth(authorization)
	if !authResult {
		c.ApiJsonReturn("访问权限限制", 401, "")
	}
	body := c.Ctx.Input.RequestBody
	if body == nil {
		return
	}
	reqBody := collection.Collect(string(body)).ToMap()
	addr, ok := reqBody["email"]
	if !ok {
		logs.Error("Need an email address to send")
		c.ApiJsonReturn("缺少必要参数email", 400, "")
	}
	if !verifyEmailFormat(addr.(string)) {
		logs.Error("Invalid email address:", addr)
		c.ApiJsonReturn("待验证邮箱地址非法", 400, "")
	}
	captchaValue := genValidateCode(6)
	timeUnix := time.Now().Unix()
	interval, _ := beego.AppConfig.Int64("verifyinterval")
	var verify models.Verify
	ep := utils.EmailParams{Receiver: addr.(string), Code: captchaValue}
	if searchEmailRecord(addr.(string)) {
		o := orm.NewOrm()
		qs := o.QueryTable("verify")
		err := qs.Filter("addr", addr).One(&verify)
		if err != nil {
			return
		}
		created := verify.Created
		if (timeUnix - created) < interval {
			logs.Error("The interval between two verifications cannot be less than 1 minute, addr:", addr)
			c.ApiJsonReturn("发送验证码的时间间隔不能低于一分钟", 400, "")
		}
		err = utils.SendVerifyEmail(ep)
		if err != nil {
			logs.Error("Fail to send email, err:", err)
			c.ApiJsonReturn("验证邮件发送失败", 400, "")
		}
		_, err = qs.Filter("addr", addr).Update(orm.Params{
			"Code":    captchaValue,
			"Created": created,
		})
		if err != nil {
			logs.Error("Fail to update verify, err:", err)
		}
	} else {
		verify.Addr = addr.(string)
		verify.Code = captchaValue
		verify.Created = timeUnix
		err := utils.SendVerifyEmail(ep)
		if err != nil {
			logs.Error("Fail to send email, err:", err)
			c.ApiJsonReturn("验证邮件发送失败", 400, "")
		}
		o := orm.NewOrm()
		_, err = o.Insert(&verify)
		if err != nil {
			logs.Error("Insert verify failed, err:", err)
		}
	}
	c.ApiJsonReturn("成功发送验证邮件", 200, "")
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
	expire, _ := beego.AppConfig.Int64("verifyexpire")
	for _, verification := range verifications {
		addr := verification.Addr
		code := verification.Code
		created := verification.Created
		if (timeUnix - created) > expire {
			cleanCode(addr, code)
		}
	}
	return nil
}
