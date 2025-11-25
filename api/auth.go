package api

import (
	"log"
	"sublink/middlewares"
	"sublink/models"
	"sublink/utils"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/gin-gonic/gin"
)

// 获取token
func GetToken(username string) (string, error) {
	c := &middlewares.JwtClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 14)), // 14天后过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),                          // 签发时间
			Subject:   username,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(middlewares.Secret)
}

// 获取captcha图形验证码
func GetCaptcha(c *gin.Context) {
	id, bs4, _, err := utils.GetCaptcha()
	if err != nil {
		log.Println("获取验证码失败")
		utils.FailWithMsg(c, "获取验证码失败")
		return
	}
	utils.OkDetailed(c, "获取验证码成功", gin.H{
		"captchaKey":    id,
		"captchaBase64": bs4,
	})
}

// 用户登录
func UserLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	captchaCode := c.PostForm("captchaCode")
	captchaKey := c.PostForm("captchaKey")
	// 验证验证码
	if !utils.VerifyCaptcha(captchaKey, captchaCode) {
		log.Println("验证码错误")
		utils.FailWithMsg(c, "验证码错误")
		return
	}
	user := &models.User{Username: username, Password: password}
	err := user.Verify()
	if err != nil {
		log.Println("账号或者密码错误")
		utils.FailWithMsg(c, "账号或者密码错误")
		return
	}
	// 生成token
	token, err := GetToken(username)
	if err != nil {
		log.Println("获取token失败")
		utils.FailWithMsg(c, "获取token失败")
		return
	}
	// 登录成功返回token
	utils.OkDetailed(c, "登录成功", gin.H{
		"accessToken":  token,
		"tokenType":    "Bearer",
		"refreshToken": nil,
		"expires":      nil,
	})
}
func UserOut(c *gin.Context) {
	// 拿到jwt中的username
	if _, Is := c.Get("username"); Is {
		utils.OkWithMsg(c, "退出成功")
	}
}
