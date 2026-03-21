package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sublink/config"
	"sublink/database"
	"sublink/models"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

const mfaResetHeader = "X-MFA-Reset-Token"

type pendingMFAClaims struct {
	Username    string `json:"username"`
	Purpose     string `json:"purpose"`
	ChallengeID string `json:"challengeId"`
	jwt.RegisteredClaims
}

type MFAStatusResponse struct {
	Enabled             bool `json:"enabled"`
	PendingEnrollment   bool `json:"pendingEnrollment"`
	RecoveryCodesRemain int  `json:"recoveryCodesRemaining"`
}

type verifyMFARequest struct {
	ChallengeToken string `json:"challengeToken" binding:"required"`
	Code           string `json:"code" binding:"required"`
}

type recoveryMFARequest struct {
	ChallengeToken string `json:"challengeToken" binding:"required"`
	RecoveryCode   string `json:"recoveryCode" binding:"required"`
}

type reauthRequest struct {
	Password string `json:"password" binding:"required"`
	Code     string `json:"code"`
}

type beginTOTPEnrollmentRequest struct {
	Password string `json:"password" binding:"required"`
	Code     string `json:"code"`
}

type confirmTOTPEnrollmentRequest struct {
	Code string `json:"code" binding:"required"`
}

type disableTOTPRequest struct {
	Password string `json:"password" binding:"required"`
	Code     string `json:"code"`
}

type regenerateRecoveryCodesRequest struct {
	Password string `json:"password" binding:"required"`
	Code     string `json:"code"`
}

type resetTOTPRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	ResetToken string `json:"resetToken"`
}

type signedMFAResetToken struct {
	Username string `json:"username"`
	Purpose  string `json:"purpose"`
	Exp      int64  `json:"exp"`
}

func buildLoginSuccessData(token string) gin.H {
	return gin.H{
		"accessToken":  token,
		"tokenType":    "Bearer",
		"refreshToken": nil,
		"expires":      nil,
	}
}

func generateChallengeID() (string, error) {
	buf, err := models.GenerateTOTPSecret()
	if err != nil {
		return "", err
	}
	return strings.ToLower(strings.ReplaceAll(buf, "=", "")), nil
}

func issuePendingMFAChallenge(user *models.User) (string, error) {
	challengeID, err := generateChallengeID()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(models.TOTPChallengeTTL)
	challenge := &models.MFALoginChallenge{
		ChallengeID: challengeID,
		Username:    user.Username,
		Purpose:     "mfa_login",
		ExpiresAt:   expiresAt.Unix(),
		MaxAttempts: models.TOTPChallengeMaxAttempts,
	}
	if err := database.DB.Create(challenge).Error; err != nil {
		return "", err
	}
	claims := &pendingMFAClaims{
		Username:    user.Username,
		Purpose:     "mfa_login",
		ChallengeID: challengeID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Username,
			ID:        challengeID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetJwtSecret()))
}

func parsePendingMFAChallenge(tokenString string) (*pendingMFAClaims, *models.MFALoginChallenge, error) {
	token, err := jwt.ParseWithClaims(tokenString, &pendingMFAClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetJwtSecret()), nil
	})
	if err != nil {
		return nil, nil, err
	}
	claims, ok := token.Claims.(*pendingMFAClaims)
	if !ok || !token.Valid || claims.Purpose != "mfa_login" || strings.TrimSpace(claims.ChallengeID) == "" {
		return nil, nil, fmt.Errorf("invalid challenge token")
	}
	var challenge models.MFALoginChallenge
	if err := database.DB.Where("challenge_id = ?", claims.ChallengeID).First(&challenge).Error; err != nil {
		return nil, nil, err
	}
	if challenge.Purpose != "mfa_login" || challenge.Username != claims.Username {
		return nil, nil, fmt.Errorf("invalid challenge token")
	}
	nowUnix := time.Now().Unix()
	if challenge.ConsumedAt > 0 || challenge.ExpiresAt < nowUnix {
		return nil, nil, fmt.Errorf("challenge expired")
	}
	if challenge.AttemptCount >= challenge.MaxAttempts {
		return nil, nil, fmt.Errorf("challenge blocked")
	}
	return claims, &challenge, nil
}

func recordMFAChallengeFailure(challenge *models.MFALoginChallenge) {
	if challenge == nil {
		return
	}
	_ = database.DB.Model(&models.MFALoginChallenge{}).
		Where("id = ? AND consumed_at = 0 AND expires_at >= ? AND attempt_count < max_attempts", challenge.ID, time.Now().Unix()).
		UpdateColumn("attempt_count", gorm.Expr("attempt_count + 1")).Error
}

func consumeMFAChallenge(challenge *models.MFALoginChallenge) error {
	if challenge == nil {
		return fmt.Errorf("challenge missing")
	}
	consumedAt := time.Now().Unix()
	result := database.DB.Model(&models.MFALoginChallenge{}).
		Where("id = ? AND consumed_at = 0 AND expires_at >= ? AND attempt_count < max_attempts", challenge.ID, consumedAt).
		UpdateColumn("consumed_at", consumedAt)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("challenge already consumed")
	}
	challenge.ConsumedAt = consumedAt
	return nil
}

func requireCurrentUser(c *gin.Context) (*models.User, bool) {
	usernameValue, exists := c.Get("username")
	if !exists {
		utils.Forbidden(c, "请求未携带token")
		return nil, false
	}
	user, err := models.FindUserByUsername(usernameValue.(string))
	if err != nil {
		utils.FailWithMsg(c, "用户不存在")
		return nil, false
	}
	return user, true
}

func requirePasswordReauth(user *models.User, password string) error {
	verifyUser := &models.User{Username: user.Username, Password: password}
	return verifyUser.Verify()
}

func requireMFAReauth(user *models.User, password, code string) error {
	if err := requirePasswordReauth(user, password); err != nil {
		return fmt.Errorf("当前密码错误")
	}
	if user.TOTPEnabled {
		if err := user.VerifyTOTPChallenge(code, time.Now()); err != nil {
			return fmt.Errorf("TOTP 验证失败")
		}
	}
	return nil
}

func respondLoginSuccess(c *gin.Context, user *models.User, ip string) {
	token, err := GetToken(user)
	if err != nil {
		utils.Error("获取token失败: %v", err)
		utils.FailWithMsg(c, "获取token失败")
		return
	}
	go notifyUserLogin(user.Username, ip)
	utils.OkDetailed(c, "登录成功", buildLoginSuccessData(token))
}

func buildMFAStatus(user *models.User) MFAStatusResponse {
	return MFAStatusResponse{
		Enabled:             user.TOTPEnabled,
		PendingEnrollment:   strings.TrimSpace(user.TOTPPendingSecret) != "",
		RecoveryCodesRemain: user.CountRecoveryCodes(),
	}
}

func signMFAResetToken(username string, expiresAt time.Time) (string, error) {
	secret := strings.TrimSpace(config.GetMFAResetSecret())
	if secret == "" {
		return "", fmt.Errorf("MFA reset secret 未配置")
	}
	payload := signedMFAResetToken{
		Username: strings.ToLower(strings.TrimSpace(username)),
		Purpose:  "mfa_reset",
		Exp:      expiresAt.Unix(),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write(body)
	sig := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(body) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func validateScopedMFAReset(username, provided string) bool {
	provided = strings.TrimSpace(provided)
	if provided == "" {
		return false
	}
	parts := strings.Split(provided, ".")
	if len(parts) != 2 {
		return false
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	secret := strings.TrimSpace(config.GetMFAResetSecret())
	if secret == "" {
		return false
	}
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write(body)
	if !hmac.Equal(signature, h.Sum(nil)) {
		return false
	}
	var payload signedMFAResetToken
	if err := json.Unmarshal(body, &payload); err != nil {
		return false
	}
	if payload.Purpose != "mfa_reset" || payload.Exp < time.Now().Unix() {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(payload.Username), strings.TrimSpace(username))
}

func GetMFAStatus(c *gin.Context) {
	user, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	utils.OkDetailed(c, "获取 MFA 状态成功", buildMFAStatus(user))
}

func BeginTOTPEnrollment(c *gin.Context) {
	user, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	var req beginTOTPEnrollmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "请求参数错误: "+err.Error())
		return
	}
	if err := requireMFAReauth(user, req.Password, req.Code); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	secret, provisioningURI, recoveryCodes, err := user.BeginTOTPEnrollment()
	if err != nil {
		utils.Error("开始 TOTP 绑定失败: %v", err)
		utils.FailWithMsg(c, "开始 TOTP 绑定失败")
		return
	}
	utils.OkDetailed(c, "TOTP 绑定准备完成", gin.H{
		"secret":            secret,
		"provisioningUri":   provisioningURI,
		"recoveryCodes":     recoveryCodes,
		"recoveryCodeCount": len(recoveryCodes),
		"status":            buildMFAStatus(user),
	})
}

func ConfirmTOTPEnrollment(c *gin.Context) {
	user, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	var req confirmTOTPEnrollmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "请求参数错误: "+err.Error())
		return
	}
	if err := user.ConfirmTOTPEnrollment(req.Code, time.Now()); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkDetailed(c, "TOTP 已启用", buildMFAStatus(user))
}

func VerifyTOTPLogin(c *gin.Context) {
	var req verifyMFARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "请求参数错误: "+err.Error())
		return
	}
	claims, challenge, err := parsePendingMFAChallenge(req.ChallengeToken)
	if err != nil {
		utils.FailWithData(c, "登录验证已过期或不可用，请重新输入用户名和密码", gin.H{"errorType": "mfa"})
		return
	}
	user, err := models.LoadUserForMFAChallenge(claims.Username)
	if err != nil {
		utils.FailWithData(c, "用户不存在", gin.H{"errorType": "credentials"})
		return
	}
	if err := user.VerifyTOTPChallenge(req.Code, time.Now()); err != nil {
		recordMFAChallengeFailure(challenge)
		utils.FailWithData(c, err.Error(), gin.H{"errorType": "mfa"})
		return
	}
	if err := consumeMFAChallenge(challenge); err != nil {
		utils.FailWithData(c, "登录验证已失效，请重新登录", gin.H{"errorType": "mfa"})
		return
	}
	respondLoginSuccess(c, user, c.ClientIP())
}

func VerifyRecoveryCodeLogin(c *gin.Context) {
	var req recoveryMFARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "请求参数错误: "+err.Error())
		return
	}
	claims, challenge, err := parsePendingMFAChallenge(req.ChallengeToken)
	if err != nil {
		utils.FailWithData(c, "登录验证已过期或不可用，请重新输入用户名和密码", gin.H{"errorType": "mfa"})
		return
	}
	user, err := models.LoadUserForMFAChallenge(claims.Username)
	if err != nil {
		utils.FailWithData(c, "用户不存在", gin.H{"errorType": "credentials"})
		return
	}
	if err := user.UseRecoveryCode(req.RecoveryCode); err != nil {
		recordMFAChallengeFailure(challenge)
		utils.FailWithData(c, err.Error(), gin.H{"errorType": "mfa"})
		return
	}
	if err := consumeMFAChallenge(challenge); err != nil {
		utils.FailWithData(c, "登录验证已失效，请重新登录", gin.H{"errorType": "mfa"})
		return
	}
	respondLoginSuccess(c, user, c.ClientIP())
}

func DisableTOTP(c *gin.Context) {
	user, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	var req disableTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "请求参数错误: "+err.Error())
		return
	}
	if err := requireMFAReauth(user, req.Password, req.Code); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	if err := user.DisableTOTP(); err != nil {
		utils.Error("关闭 TOTP 失败: %v", err)
		utils.FailWithMsg(c, "关闭 TOTP 失败")
		return
	}
	utils.OkDetailed(c, "TOTP 已关闭", buildMFAStatus(user))
}

func RegenerateRecoveryCodes(c *gin.Context) {
	user, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	var req regenerateRecoveryCodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "请求参数错误: "+err.Error())
		return
	}
	if err := requireMFAReauth(user, req.Password, req.Code); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	if !user.TOTPEnabled {
		utils.FailWithMsg(c, "用户未启用 TOTP")
		return
	}
	codes, err := user.RegenerateRecoveryCodes()
	if err != nil {
		utils.Error("重置恢复码失败: %v", err)
		utils.FailWithMsg(c, "重置恢复码失败")
		return
	}
	utils.OkDetailed(c, "恢复码已重置", gin.H{
		"recoveryCodes": codes,
		"status":        buildMFAStatus(user),
	})
}

func ResetTOTP(c *gin.Context) {
	var req resetTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "请求参数错误: "+err.Error())
		return
	}
	providedResetToken := strings.TrimSpace(req.ResetToken)
	if providedResetToken == "" {
		providedResetToken = strings.TrimSpace(c.GetHeader(mfaResetHeader))
	}
	if !validateScopedMFAReset(req.Username, providedResetToken) {
		utils.Forbidden(c, "MFA 重置令牌无效")
		return
	}
	user := &models.User{Username: req.Username, Password: req.Password}
	if err := user.Verify(); err != nil {
		utils.FailWithData(c, "用户名或密码错误", gin.H{"errorType": "credentials"})
		return
	}
	if err := user.DisableTOTP(); err != nil {
		utils.Error("应急重置 TOTP 失败: %v", err)
		utils.FailWithMsg(c, "重置 TOTP 失败")
		return
	}
	utils.OkWithMsg(c, "TOTP 已重置，请重新登录后重新绑定")
}

func ReauthMFA(c *gin.Context) {
	user, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	var req reauthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "请求参数错误: "+err.Error())
		return
	}
	if err := requireMFAReauth(user, req.Password, req.Code); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkDetailed(c, "验证成功", buildMFAStatus(user))
}

func GenerateScopedMFAResetToken(username string, expiresAt time.Time) (string, error) {
	return signMFAResetToken(username, expiresAt)
}

func cleanupExpiredMFAChallenges() {
	nowUnix := time.Now().Unix()
	_ = database.DB.Where("expires_at < ? OR consumed_at > 0", nowUnix).Delete(&models.MFALoginChallenge{}).Error
}
