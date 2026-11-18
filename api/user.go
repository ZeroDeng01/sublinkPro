package api

import (
	"encoding/json"
	"strings"
	"sublink/database"
	"sublink/models"
	"sublink/services/ai"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID       int
	Username string
	Nickname string
	Avatar   string
	Mobile   string
	Email    string
}

// 新增用户
func UserAdd(c *gin.Context) {
	user := &models.User{
		Username: "test",
		Password: "test",
	}
	err := user.Create()
	if err != nil {
		utils.Error("Failed to create user: %v", err)
	}
	utils.OkWithMsg(c, "User created")
}

// 获取用户信息
func UserMe(c *gin.Context) {
	// 获取jwt中的username
	// 返回用户信息
	username, ok := currentUsernameFromContext(c)
	if !ok {
		return
	}
	user := &models.User{Username: username}
	err := user.Find()
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	aiSettings, aiSettingsErr := models.GetSystemAISettings()
	aiEnabled := false
	aiConfigured := false
	if aiSettingsErr == nil {
		aiEnabled = aiSettings.Enabled
		aiConfigured = aiSettings.Configured
	}
	utils.OkDetailed(c, "User info fetched successfully", gin.H{
		"avatar": "",
		"ai": gin.H{
			"enabled":    aiEnabled,
			"configured": aiConfigured,
		},
		"nickname": user.Nickname,
		"userId":   user.ID,
		"username": user.Username,
		"roles":    []string{"ADMIN"},
		"mfa": gin.H{
			"enabled":                user.TOTPEnabled,
			"pendingEnrollment":      buildMFAStatus(user).PendingEnrollment,
			"recoveryCodesRemaining": user.CountRecoveryCodes(),
		},
	})
}

// 获取所有用户
func UserPages(c *gin.Context) {
	// 获取jwt中的username
	// 返回用户信息
	username, ok := currentUsernameFromContext(c)
	if !ok {
		return
	}
	user := &models.User{Username: username}
	users, err := user.All()
	if err != nil {
		utils.Error("Failed to get user info: %v", err)
	}
	list := []*User{}
	for i := range users {
		list = append(list, &User{
			ID:       users[i].ID,
			Username: users[i].Username,
			Nickname: users[i].Nickname,
			Avatar:   "",
		})
	}
	utils.OkDetailed(c, "User info fetched successfully", gin.H{
		"list": list,
	})
}

// 更新用户信息

func UserSet(c *gin.Context) {
	NewUsername := c.PostForm("username")
	NewPassword := c.PostForm("password")
	if NewUsername == "" || NewPassword == "" {
		utils.FailWithMsg(c, "Username and password are required")
		return
	}
	username, ok := currentUsernameFromContext(c)
	if !ok {
		return
	}
	user := &models.User{Username: username}

	// 先查找用户获取ID
	if err := user.Find(); err != nil {
		utils.FailWithMsg(c, "User not found")
		return
	}

	err := user.Set(&models.User{
		Username: NewUsername,
		Password: NewPassword,
	})
	if err != nil {
		utils.Error("Failed to change password: %v", err)
		utils.FailWithMsg(c, err.Error())
		return
	}

	// 修改成功
	utils.OkWithMsg(c, "Modified successfully")
}

// 修改密码
func UserChangePassword(c *gin.Context) {
	type ChangePasswordRequest struct {
		OldPassword     string `json:"oldPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required"`
		ConfirmPassword string `json:"confirmPassword" binding:"required"`
		Code            string `json:"code"`
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid request parameters: "+err.Error())
		return
	}

	// 验证两次密码是否一致
	if req.NewPassword != req.ConfirmPassword {
		utils.FailWithMsg(c, "Passwords do not match")
		return
	}

	// 验证密码长度
	if len(req.NewPassword) < 6 {
		utils.FailWithMsg(c, "Password must be at least 6 characters")
		return
	}

	username, ok := currentUsernameFromContext(c)
	if !ok {
		return
	}
	user := &models.User{Username: username}
	if err := user.Find(); err != nil {
		utils.FailWithMsg(c, "User not found")
		return
	}

	if err := requireMFAReauth(user, req.OldPassword, req.Code); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	// 更新密码
	updateUser := &models.User{Password: req.NewPassword}
	if err := user.Set(updateUser); err != nil {
		utils.Error("Password change failed: %v", err)
		utils.FailWithMsg(c, "Failed to change password")
		return
	}

	utils.OkWithMsg(c, "Password changed")
}

// 更新个人资料（用户名、昵称）
func UserUpdateProfile(c *gin.Context) {
	type UpdateProfileRequest struct {
		Username string `json:"username"`
		Nickname string `json:"nickname"`
		Password string `json:"password" binding:"required"`
		Code     string `json:"code"`
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid request parameters: "+err.Error())
		return
	}

	// 验证用户名不能为空
	if req.Username == "" {
		utils.FailWithMsg(c, "Username is required")
		return
	}

	// 获取当前用户
	username, ok := currentUsernameFromContext(c)
	if !ok {
		return
	}
	user := &models.User{Username: username}

	// 查找用户获取ID
	if err := user.Find(); err != nil {
		utils.FailWithMsg(c, "User not found")
		return
	}

	if err := requireMFAReauth(user, req.Password, req.Code); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	// 使用 map 更新字段，避免 GORM 忽略零值
	updates := map[string]any{
		"username": req.Username,
		"nickname": req.Nickname,
	}

	if err := database.DB.Where("username = ?", user.Username).Model(&models.User{}).Updates(updates).Error; err != nil {
		utils.Error("Failed to update profile: %v", err)
		utils.FailWithMsg(c, "Failed to update profile: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "Profile updated")
}

func requireSessionAuthForAISettings(c *gin.Context) bool {
	if strings.TrimSpace(c.GetHeader("X-API-Key")) == "" {
		return true
	}
	utils.Forbidden(c, "AI assistant settings are only accessible from a login session")
	return false
}

func resolveAIRequestBaseURL(requestBaseURL, savedBaseURL string) (string, bool, error) {
	trimmedRequestBaseURL := strings.TrimSpace(requestBaseURL)
	if trimmedRequestBaseURL == "" {
		return strings.TrimSpace(savedBaseURL), true, nil
	}
	normalizedRequestBaseURL, err := ai.NormalizeBaseURL(trimmedRequestBaseURL)
	if err != nil {
		return "", false, err
	}
	normalizedSavedBaseURL, err := ai.NormalizeBaseURL(savedBaseURL)
	if err != nil {
		normalizedSavedBaseURL = strings.TrimSpace(savedBaseURL)
	}
	return normalizedRequestBaseURL, normalizedRequestBaseURL == strings.TrimSpace(normalizedSavedBaseURL), nil
}

func UserGetAISettings(c *gin.Context) {
	if !requireSessionAuthForAISettings(c) {
		return
	}

	settings, err := models.GetSystemAISettings()
	if err != nil {
		utils.FailWithMsg(c, "Failed to fetch AI settings: "+err.Error())
		return
	}

	utils.OkDetailed(c, "AI settings fetched", settings)
}

func UserListAIModels(c *gin.Context) {
	if !requireSessionAuthForAISettings(c) {
		return
	}

	type listAIModelsRequest struct {
		BaseURL      string            `json:"baseUrl"`
		APIKey       string            `json:"apiKey"`
		ExtraHeaders map[string]string `json:"extraHeaders"`
	}

	var req listAIModelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid request parameters: "+err.Error())
		return
	}

	current, err := models.GetSystemAISettings()
	if err != nil {
		utils.FailWithMsg(c, "Failed to read current AI settings: "+err.Error())
		return
	}

	baseURL, usingSavedBaseURL, err := resolveAIRequestBaseURL(req.BaseURL, current.BaseURL)
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = current.BaseURL
	}
	apiKey := strings.TrimSpace(req.APIKey)
	if apiKey == "" && usingSavedBaseURL {
		apiKey = current.RawAPIKey
	}
	extraHeaders := req.ExtraHeaders
	if len(extraHeaders) == 0 && usingSavedBaseURL {
		extraHeaders = current.ExtraHeaders
	}

	modelsList, err := ai.DiscoverModels(c.Request.Context(), ai.ClientConfig{
		BaseURL:      baseURL,
		APIKey:       apiKey,
		ExtraHeaders: extraHeaders,
	})
	if err != nil {
		utils.FailWithMsg(c, "Failed to fetch model list: "+err.Error())
		return
	}

	utils.OkDetailed(c, "Model list fetched", gin.H{"models": modelsList})
}

func UserUpdateAISettings(c *gin.Context) {
	if !requireSessionAuthForAISettings(c) {
		return
	}

	type updateAISettingsRequest struct {
		Enabled      bool              `json:"enabled"`
		BaseURL      string            `json:"baseUrl"`
		Model        string            `json:"model"`
		RequestType  string            `json:"requestType"`
		APIKey       string            `json:"apiKey"`
		Temperature  float64           `json:"temperature"`
		MaxTokens    int               `json:"maxTokens"`
		ExtraHeaders map[string]string `json:"extraHeaders"`
	}

	var req updateAISettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid request parameters: "+err.Error())
		return
	}

	current, err := models.GetSystemAISettings()
	if err != nil {
		utils.FailWithMsg(c, "Failed to read current AI settings: "+err.Error())
		return
	}

	validatedBaseURL, err := ai.NormalizeBaseURL(req.BaseURL)
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	if req.Enabled && strings.TrimSpace(validatedBaseURL) == "" {
		utils.FailWithMsg(c, "Base URL is required when AI is enabled")
		return
	}
	if req.Enabled && strings.TrimSpace(req.Model) == "" {
		utils.FailWithMsg(c, "Model is required when AI is enabled")
		return
	}
	if req.Enabled && strings.TrimSpace(req.APIKey) == "" && !current.HasKey {
		utils.FailWithMsg(c, "API Key is required when AI is enabled")
		return
	}
	if req.Temperature < 0 || req.Temperature > 2 {
		utils.FailWithMsg(c, "temperature must be between 0 and 2")
		return
	}
	if req.MaxTokens < 0 {
		utils.FailWithMsg(c, "maxTokens cannot be negative")
		return
	}

	headersJSON := ""
	if len(req.ExtraHeaders) > 0 {
		payload, err := json.Marshal(req.ExtraHeaders)
		if err != nil {
			utils.FailWithMsg(c, "Invalid extra request headers format")
			return
		}
		headersJSON = string(payload)
	}

	if err := models.UpdateSystemAISettings(models.UserAISettings{
		Enabled:         req.Enabled,
		BaseURL:         validatedBaseURL,
		Model:           strings.TrimSpace(req.Model),
		RequestType:     models.NormalizeSystemAIRequestType(req.RequestType),
		RawAPIKey:       strings.TrimSpace(req.APIKey),
		Temperature:     req.Temperature,
		MaxTokens:       req.MaxTokens,
		ExtraHeadersRaw: headersJSON,
	}); err != nil {
		utils.FailWithMsg(c, "Failed to save AI settings: "+err.Error())
		return
	}

	settings, err := models.GetSystemAISettings()
	if err != nil {
		utils.FailWithMsg(c, "Failed to read AI settings: "+err.Error())
		return
	}
	utils.OkDetailed(c, "AI settings saved", settings)
}

func UserTestAISettings(c *gin.Context) {
	if !requireSessionAuthForAISettings(c) {
		return
	}

	type testAISettingsRequest struct {
		BaseURL      string            `json:"baseUrl"`
		Model        string            `json:"model"`
		RequestType  string            `json:"requestType"`
		APIKey       string            `json:"apiKey"`
		Temperature  *float64          `json:"temperature"`
		MaxTokens    *int              `json:"maxTokens"`
		ExtraHeaders map[string]string `json:"extraHeaders"`
	}

	var req testAISettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid request parameters: "+err.Error())
		return
	}

	current, err := models.GetSystemAISettings()
	if err != nil {
		utils.FailWithMsg(c, "Failed to read current AI settings: "+err.Error())
		return
	}

	baseURL, usingSavedBaseURL, err := resolveAIRequestBaseURL(req.BaseURL, current.BaseURL)
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = current.BaseURL
	}
	model := strings.TrimSpace(req.Model)
	if model == "" && usingSavedBaseURL {
		model = current.Model
	}
	requestType := models.NormalizeSystemAIRequestType(req.RequestType)
	if strings.TrimSpace(req.RequestType) == "" && usingSavedBaseURL {
		requestType = current.RequestType
	}
	apiKey := strings.TrimSpace(req.APIKey)
	if apiKey == "" && usingSavedBaseURL {
		apiKey = current.RawAPIKey
	}
	temperature := 0.2
	if usingSavedBaseURL {
		temperature = current.Temperature
	}
	if req.Temperature != nil {
		temperature = *req.Temperature
	}
	maxTokens := 1200
	if usingSavedBaseURL {
		maxTokens = current.MaxTokens
	}
	if req.MaxTokens != nil {
		maxTokens = *req.MaxTokens
	}
	extraHeaders := req.ExtraHeaders
	if len(extraHeaders) == 0 && usingSavedBaseURL {
		extraHeaders = current.ExtraHeaders
	}

	client, err := ai.NewClient(ai.ClientConfig{
		BaseURL:      baseURL,
		APIKey:       apiKey,
		Model:        model,
		RequestType:  requestType,
		Temperature:  temperature,
		MaxTokens:    maxTokens,
		ExtraHeaders: extraHeaders,
	})
	if err != nil {
		utils.FailWithMsg(c, "Invalid AI settings: "+err.Error())
		return
	}

	result, err := client.TestConnection(c.Request.Context())
	if err != nil {
		utils.FailWithMsg(c, "Connection test failed: "+err.Error())
		return
	}

	utils.OkDetailed(c, "Connection test succeeded", result)
}
