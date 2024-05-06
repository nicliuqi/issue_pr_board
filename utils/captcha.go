package utils

import (
	"fmt"
	config2 "issue_pr_board/utils/aj-captcha/config"
	constant "issue_pr_board/utils/aj-captcha/const"
	"issue_pr_board/utils/aj-captcha/service"
)

var blockPuzzleConfig = &config2.BlockPuzzleConfig{Offset: 10}
var captchaConfig = config2.BuildConfig(constant.MemCacheKey, constant.DefaultResourceRoot, blockPuzzleConfig, 2*60)
var Factory = service.NewCaptchaServiceFactory(captchaConfig)

func InitCaptchaFactory() {
	fmt.Println("init captcha factory")
	Factory.RegisterCache(constant.MemCacheKey, service.NewMemCacheService(20))
	Factory.RegisterService(constant.BlockPuzzleCaptcha, service.NewBlockPuzzleCaptchaService(Factory))
}
