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
	rememberMe := c.PostForm("rememberMe") == "true"

	// 验证验证码
	if !utils.VerifyCaptcha(captchaKey, captchaCode) {
		log.Println("验证码错误")
		utils.FailWithData(c, "验证码错误", gin.H{"errorType": "captcha"})
		return
	}
	user := &models.User{Username: username, Password: password}
	err := user.Verify()
	if err != nil {
		log.Println("账号或者密码错误")
		utils.FailWithData(c, "用户名或密码错误", gin.H{"errorType": "credentials"})
		return
	}
	// 生成token
	token, err := GetToken(username)
	if err != nil {
		log.Println("获取token失败")
		utils.FailWithMsg(c, "获取token失败")
		return
	}

	// 如果勾选记住密码，生成 rememberToken (支持多设备)
	var rememberToken string
	if rememberMe {
		userAgent := c.GetHeader("User-Agent")
		rememberToken, err = models.GenerateRememberToken(user.ID, userAgent)
		if err != nil {
			log.Println("生成记住密码令牌失败:", err)
			// 不影响正常登录，只是不返回 rememberToken
		}
	}

	// 登录成功返回token
	utils.OkDetailed(c, "登录成功", gin.H{
		"accessToken":   token,
		"tokenType":     "Bearer",
		"refreshToken":  nil,
		"expires":       nil,
		"rememberToken": rememberToken,
	})
}

// RememberLogin 使用记住密码令牌登录
func RememberLogin(c *gin.Context) {
	rememberToken := c.PostForm("rememberToken")
	if rememberToken == "" {
		utils.FailWithMsg(c, "无效的登录令牌")
		return
	}

	// 验证令牌并获取用户
	user, err := models.VerifyAndGetUserByToken(rememberToken)
	if err != nil {
		log.Println("记住密码令牌验证失败:", err)
		utils.FailWithMsg(c, "登录令牌已过期，请重新登录")
		return
	}

	// 生成新的 JWT token
	token, err := GetToken(user.Username)
	if err != nil {
		log.Println("获取token失败")
		utils.FailWithMsg(c, "获取token失败")
		return
	}

	// 返回登录成功
	utils.OkDetailed(c, "自动登录成功", gin.H{
		"accessToken":  token,
		"tokenType":    "Bearer",
		"refreshToken": nil,
		"expires":      nil,
	})
}

// UserOut 用户退出登录
func UserOut(c *gin.Context) {
	// 拿到jwt中的username
	if _, Is := c.Get("username"); Is {
		// 如果前端传递了 rememberToken，删除它
		rememberToken := c.Query("rememberToken")
		if rememberToken == "" {
			rememberToken = c.PostForm("rememberToken")
		}
		if rememberToken != "" {
			if err := models.DeleteRememberToken(rememberToken); err != nil {
				log.Println("删除记住密码令牌失败:", err)
			}
		}
		utils.OkWithMsg(c, "退出成功")
	}
}
