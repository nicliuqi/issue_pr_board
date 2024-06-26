package utils

import (
	"github.com/beego/beego/v2/client/cache"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web/captcha"
)

var cpt *captcha.Captcha

func InitCaptcha() {
	store := cache.NewMemoryCache()
	cpt = captcha.NewWithFilter("/captcha/", store)
}

func GetCaptcha() string {
	value, err := cpt.CreateCaptcha()
	if err != nil {
		logs.Error("Create Captcha Error:", err)
		return ""
	}
	return value
}

func VerifyCaptcha(captchaId, code string) bool {
	return cpt.Verify(captchaId, code)
}
