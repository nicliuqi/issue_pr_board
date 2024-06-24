package controllers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/go-playground/validator/v10"

	"issue_pr_board/config"
	"issue_pr_board/models"
	"issue_pr_board/utils"
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

func cleanCode(addr string, code string) {
	o := orm.NewOrm()
	qt := o.QueryTable("verify")
	num, err := qt.Filter("addr", addr).Filter("code", code).Delete()
	if err != nil {
		logs.Error("Fail to delete verify record, err:", err)
	}
	annoyAddr := strings.Split(strings.Split(addr, "@")[0], "")[0] + "***@" + strings.Split(addr, "@")[1]
	if num != 0 {
		logs.Info("The expired record had been deleted which addr is:", annoyAddr)
	}
}

func CleanVerification() error {
	var verifications []models.Verify
	o := orm.NewOrm()
	if _, err := o.QueryTable("verify").All(&verifications); err != nil {
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
	c.ApiJsonReturn("Success", http.StatusOK, data)
}

type CheckCaptchaController struct {
	BaseController
}

func (c *CheckCaptchaController) Post() {
	params, err := getParams(c.Ctx.Request)
	validate := validator.New()
	if validateErr := validate.Struct(params); validateErr != nil {
		c.ApiJsonReturn("Invalid params", http.StatusBadRequest, nil)
	}
	if !utils.VerifyCaptcha(params.CaptchaId, params.Challenge) {
		c.ApiJsonReturn("Verification error", http.StatusBadRequest, nil)
	}

	captchaValue := genValidateCode(6)
	timeUnix := time.Now().Unix()
	interval := config.AppConfig.VerifyInterval

	var verify models.Verify
	addr := strings.ToLower(params.Email)
	ep := utils.EmailParams{Receiver: addr, Code: captchaValue}
	o := orm.NewOrm()
	if models.SearchEmailRecord(addr) {
		qs := o.QueryTable("verify")
		if err = qs.Filter("addr", addr).One(&verify); err != nil {
			c.ApiJsonReturn("Server error", http.StatusInternalServerError, nil)
		}
		annoyAddr := strings.Split(strings.Split(addr, "@")[0], "")[0] + "***@" +
			strings.Split(addr, "@")[1]
		created := verify.Created
		if (timeUnix - created) < interval {
			logs.Error("The interval between two verifications cannot be less than 1 minute, addr:", annoyAddr)
			c.ApiJsonReturn("The interval between two verifications cannot be less than 1 minute",
				http.StatusBadRequest, nil)
		}
		go utils.SendVerifyEmail(ep)
		if _, err = qs.Filter("addr", addr).Update(orm.Params{
			"Code":    captchaValue,
			"Created": created,
		}); err != nil {
			logs.Error("Fail to update verify, err:", err)
			c.ApiJsonReturn("Server error", http.StatusInternalServerError, nil)
		}
	} else {
		verify.Addr = addr
		verify.Code = captchaValue
		verify.Created = timeUnix
		go utils.SendVerifyEmail(ep)
		if _, err = o.Insert(&verify); err != nil {
			logs.Error("Insert verify failed, err:", err)
			c.ApiJsonReturn("Server error", http.StatusInternalServerError, nil)
		}
	}
	c.ApiJsonReturn("Success to send verification", http.StatusOK, nil)
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
	if err := json.Unmarshal(all, params); err != nil {
		return nil, err
	}
	return params, nil
}
