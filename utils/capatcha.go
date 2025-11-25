package utils

import (
	"github.com/mojocn/base64Captcha"
)

var store = base64Captcha.DefaultMemStore

// GetCaptcha 获取验证码
func GetCaptcha() (string, string, string, error) {
	driver := base64Captcha.NewDriverDigit(60, 180, 4, 0.6, 100)
	return base64Captcha.NewCaptcha(driver, store).Generate()
}

// VerifyCaptcha 验证验证码
func VerifyCaptcha(id string, answer string) bool {
	return store.Verify(id, answer, true)
}
