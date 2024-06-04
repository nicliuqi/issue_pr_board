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
	if value, err := cpt.CreateCaptcha(); err != nil {
		logs.Error("[GetCaptcha] Fail to get image captcha, err:", err)
		return ""
	} else {
		return value
	}
}

func VerifyCaptcha(captchaId, code string) bool {
	return cpt.Verify(captchaId, code)
}
