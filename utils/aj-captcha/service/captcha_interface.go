package service

type CaptchaInterface interface {
	Get() (map[string]interface{}, error)
	Check(token string, pointJson string) error
	Verification(token string, pointJson string) error
}
