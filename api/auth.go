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
		utils.OkDetailedI18n(c, "Captcha is disabled", response, "backend.auth.captcha.disabled", nil)
		return

	case config.CaptchaModeTurnstile:
		// Turnstile 模式，返回 site key
		response["turnstileSiteKey"] = captchaCfg.TurnstileSiteKey
		utils.OkDetailedI18n(c, "Turnstile configuration loaded", response, "backend.auth.captcha.turnstileLoaded", nil)
		return

	default:
		// 传统验证码模式
		id, bs4, _, err := utils.GetCaptcha()
		if err != nil {
			utils.Error("Failed to load captcha: %v", err)
			utils.FailWithI18n(c, "Failed to load captcha", "backend.auth.captcha.loadFailed", nil)
			return
		}
		response["captchaKey"] = id
		response["captchaBase64"] = bs4
		utils.OkDetailedI18n(c, "Captcha loaded", response, "backend.auth.captcha.loaded", nil)
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
		utils.FailWithI18n(c, fmt.Sprintf("Too many failed login attempts; this IP is blocked for %d minutes", minutes), "backend.auth.login.ipBanned", map[string]any{"minutes": minutes})
		return
	}

	// 验证验证码（根据配置模式）
	captchaCfg := config.GetCaptchaConfig()
	switch captchaCfg.Mode {
	case config.CaptchaModeDisabled:
		// 关闭验证码，跳过验证
		utils.Debug("Captcha disabled; skipping verification")

	case config.CaptchaModeTurnstile:
		// Turnstile 模式
		turnstileToken := c.PostForm("turnstileToken")
		if turnstileToken == "" {
			utils.Warn("Turnstile token is empty")
			utils.FailWithDataI18n(c, "Please complete the human verification", gin.H{"errorType": "captcha"}, "backend.auth.login.turnstileRequired", nil)
			return
		}
		verified, err := utils.VerifyTurnstile(turnstileToken, config.GetTurnstileSecretKey(), ip, config.GetTurnstileProxyLink())
		if err != nil {
			utils.Error("Turnstile verification error: %v", err)
			utils.FailWithDataI18n(c, "Human verification failed", gin.H{"errorType": "captcha"}, "backend.auth.login.turnstileFailed", nil)
			return
		}
		if !verified {
			utils.Warn("Turnstile verification rejected")
			utils.FailWithDataI18n(c, "Human verification rejected", gin.H{"errorType": "captcha"}, "backend.auth.login.turnstileRejected", nil)
			return
		}

	default:
		// 传统验证码模式
		if !utils.VerifyCaptcha(captchaKey, captchaCode) {
			utils.Warn("Invalid captcha")
			utils.FailWithDataI18n(c, "Invalid captcha", gin.H{"errorType": "captcha"}, "backend.auth.login.captchaInvalid", nil)
			return
		}
	}
	user := &models.User{Username: username, Password: password}
	err := user.Verify()
	if err != nil {
		utils.Warn("Invalid username or password: %v", err)
		limiter.RecordFailure(ip) // 记录失败
		utils.FailWithDataI18n(c, "Incorrect username or password", gin.H{"errorType": "credentials"}, "backend.auth.login.invalidCredentials", nil)
		return
	}
	// 登录成功，清除失败记录
	limiter.ClearFailures(ip)
	if user.TOTPEnabled {
		challengeToken, err := issuePendingMFAChallenge(user)
		if err != nil {
			utils.Error("Failed to create MFA challenge: %v", err)
			utils.FailWithI18n(c, "Failed to create login challenge", "backend.auth.login.mfaChallengeFailed", nil)
			return
		}
		utils.OkDetailedI18n(c, "Two-factor verification required", gin.H{
			"requiresMFA":    true,
			"challengeToken": challengeToken,
			"methods":        []string{"totp", "recovery_code"},
		}, "backend.auth.login.mfaRequired", nil)
		return
	}

	respondLoginSuccess(c, user, ip)
}

// UserOut 用户退出登录
func UserOut(c *gin.Context) {
	// 拿到jwt中的username
	if _, Is := c.Get("username"); Is {
		utils.OkDetailedI18n(c, "Logged out", nil, "backend.auth.logout.success", nil)
	}
}

func notifyUserLogin(username, ip string) {
	location, err := geoip.GetLocation(ip)
	if err != nil {
		location = "Unknown location"
	}
	if location == "" {
		location = "Unknown location"
	}
	timeStr := time.Now().Format("2006-01-02 15:04:05")

	payload := notifications.Payload{
		Title:   "User login notification",
		Message: fmt.Sprintf("User %s signed in\nIP: %s (%s)\nTime: %s", username, ip, location, timeStr),
		Data: map[string]any{
			"username": username,
			"ip":       ip,
			"location": location,
			"time":     timeStr,
		},
		Time: timeStr,
	}

	notifications.Publish("security.user_login", payload)
}
