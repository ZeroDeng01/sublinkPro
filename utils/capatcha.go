package utils

import (
	"image/color"

	"github.com/mojocn/base64Captcha"
)

var store = base64Captcha.DefaultMemStore

// GetCaptcha 获取验证码
func GetCaptcha() (string, string, string, error) {
	whiteBg := &color.RGBA{R: 255, G: 255, B: 255, A: 255}
	driver := base64Captcha.NewDriverString(60, 180, 0, 0, 4, base64Captcha.TxtNumbers, whiteBg, nil, nil)
	return base64Captcha.NewCaptcha(driver, store).Generate()
}

// VerifyCaptcha 验证验证码
func VerifyCaptcha(id string, answer string) bool {
	return store.Verify(id, answer, true)
}
