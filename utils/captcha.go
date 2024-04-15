package utils

import (
	"fmt"
	config2 "github.com/TestsLing/aj-captcha-go/config"
	constant "github.com/TestsLing/aj-captcha-go/const"
	"github.com/TestsLing/aj-captcha-go/service"
	"image/color"
)

var watermarkConfig = &config2.WatermarkConfig{
	FontSize: 12,
	Color:	  color.RGBA{R: 255, G: 255, B: 255, A: 255},
	Text:	  "",
}
var clickWordConfig = &config2.ClickWordConfig{
	FontSize: 25,
	FontNum:  4,
}
var blockPuzzleConfig = &config2.BlockPuzzleConfig{Offset: 10}
var captchaConfig = config2.BuildConfig(constant.MemCacheKey, constant.DefaultResourceRoot, watermarkConfig,
    clickWordConfig, blockPuzzleConfig, 2*60)
var Factory = service.NewCaptchaServiceFactory(captchaConfig)

func InitCaptchaFactory() {
	fmt.Println("init captcha factory")
	Factory.RegisterCache(constant.MemCacheKey, service.NewMemCacheService(20))
	Factory.RegisterService(constant.BlockPuzzleCaptcha, service.NewBlockPuzzleCaptchaService(Factory))
}
