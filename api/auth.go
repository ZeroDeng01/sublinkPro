package api

import (
	"fmt"
	"sublink/config"
	"sublink/middlewares"
	"sublink/models"
	"sublink/services/geoip"
	"sublink/services/notifications"
	"sublink/utils"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/gin-gonic/gin"
)

// 获取token
func GetToken(user *models.User) (string, error) {
	credentialSign := models.GenerateCredentialSign(user.Username, user.Password)
	c := &middlewares.JwtClaims{
		Username:       user.Username,
		CredentialSign: credentialSign,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 14)), // 14天后过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),                          // 签发时间
			Subject:   user.Username,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString([]byte(config.GetJwtSecret()))
}

// GetCaptcha 获取验证码配置和图形验证码
func GetCaptcha(c *gin.Context) {
	captchaCfg := config.GetCaptchaConfig()

	response := gin.H{
		"mode":             captchaCfg.Mode,           // 实际使用的模式
		"configuredMode":   captchaCfg.ConfiguredMode, // 用户配置的模式
		"degraded":         captchaCfg.Degraded,       // 是否已降级
		"captchaKey":       "",
		"captchaBase64":    "",
		"turnstileSiteKey": "",
	}

	switch captchaCfg.Mode {
	case config.CaptchaModeDisabled:
		// 关闭验证码，不需要返回验证码数据
		utils.OkDetailed(c, "验证码已关闭", response)
		return

	case config.CaptchaModeTurnstile:
		// Turnstile 模式，返回 site key
		response["turnstileSiteKey"] = captchaCfg.TurnstileSiteKey
		utils.OkDetailed(c, "获取 Turnstile 配置成功", response)
		return

	default:
		// 传统验证码模式
		id, bs4, _, err := utils.GetCaptcha()
		if err != nil {
			utils.Error("获取验证码失败: %v", err)
			utils.FailWithMsg(c, "获取验证码失败")
			return
		}
		response["captchaKey"] = id
		response["captchaBase64"] = bs4
		utils.OkDetailed(c, "获取验证码成功", response)
	}
}

// 用户登录
func UserLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	captchaCode := c.PostForm("captchaCode")
	captchaKey := c.PostForm("captchaKey")
	ip := c.ClientIP()

	// 0. 检查IP是否被封禁
	limiter := GetLoginLimiter()
	if isBanned, banUntil := limiter.CheckBan(ip); isBanned {
		minutes := int(time.Until(banUntil).Minutes()) + 1
		utils.FailWithMsg(c, fmt.Sprintf("由于多次登录失败，IP已被封禁，请 %d 分钟后再试", minutes))
		return
	}

	// 验证验证码（根据配置模式）
	captchaCfg := config.GetCaptchaConfig()
	switch captchaCfg.Mode {
	case config.CaptchaModeDisabled:
		// 关闭验证码，跳过验证
		utils.Debug("验证码已关闭，跳过验证")

	case config.CaptchaModeTurnstile:
		// Turnstile 模式
		turnstileToken := c.PostForm("turnstileToken")
		if turnstileToken == "" {
			utils.Warn("Turnstile 令牌为空")
			utils.FailWithData(c, "请完成人机验证", gin.H{"errorType": "captcha"})
			return
		}
		verified, err := utils.VerifyTurnstile(turnstileToken, config.GetTurnstileSecretKey(), ip, config.GetTurnstileProxyLink())
		if err != nil {
			utils.Error("Turnstile 验证出错: %v", err)
			utils.FailWithData(c, "人机验证失败", gin.H{"errorType": "captcha"})
			return
		}
		if !verified {
			utils.Warn("Turnstile 验证未通过")
			utils.FailWithData(c, "人机验证未通过", gin.H{"errorType": "captcha"})
			return
		}

	default:
		// 传统验证码模式
		if !utils.VerifyCaptcha(captchaKey, captchaCode) {
			utils.Warn("验证码错误")
			utils.FailWithData(c, "验证码错误", gin.H{"errorType": "captcha"})
			return
		}
	}
	user := &models.User{Username: username, Password: password}
	err := user.Verify()
	if err != nil {
		utils.Warn("账号或者密码错误: %v", err)
		limiter.RecordFailure(ip) // 记录失败
		utils.FailWithData(c, "用户名或密码错误", gin.H{"errorType": "credentials"})
		return
	}
	// 登录成功，清除失败记录
	limiter.ClearFailures(ip)
	if user.TOTPEnabled {
		challengeToken, err := issuePendingMFAChallenge(user)
		if err != nil {
			utils.Error("生成 MFA 挑战失败: %v", err)
			utils.FailWithMsg(c, "生成登录验证失败")
			return
		}
		utils.OkDetailed(c, "需要进行二次验证", gin.H{
			"requiresMFA":    true,
			"challengeToken": challengeToken,
			"methods":        []string{"totp", "recovery_code"},
		})
		return
	}

	respondLoginSuccess(c, user, ip)
}

// UserOut 用户退出登录
func UserOut(c *gin.Context) {
	// 拿到jwt中的username
	if _, Is := c.Get("username"); Is {
		utils.OkWithMsg(c, "退出成功")
	}
}

func notifyUserLogin(username, ip string) {
	location, err := geoip.GetLocation(ip)
	if err != nil {
		location = "未知位置"
	}
	if location == "" {
		location = "未知位置"
	}
	timeStr := time.Now().Format("2006-01-02 15:04:05")

	payload := notifications.Payload{
		Title:   "用户登录通知",
		Message: fmt.Sprintf("用户 %s 已登录\nIP: %s (%s)\n时间: %s", username, ip, location, timeStr),
		Data: map[string]interface{}{
			"username": username,
			"ip":       ip,
			"location": location,
			"time":     timeStr,
		},
		Time: timeStr,
	}

	notifications.Publish("security.user_login", payload)
}
