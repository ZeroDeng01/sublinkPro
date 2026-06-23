package api

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sublink/cache"
	"sublink/database"
	"sublink/models"
	"sublink/services/ai"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type Temp struct {
	File             string `json:"file"`
	Text             string `json:"text"`
	Category         string `json:"category"`
	RuleSource       string `json:"ruleSource"`
	UseProxy         bool   `json:"useProxy"`
	ProxyLink        string `json:"proxyLink"`
	EnableIncludeAll bool   `json:"enableIncludeAll"`
	CreateDate       string `json:"create_date"`
}

type TemplateEditSessionAcceptRequest struct {
	CurrentText string `json:"currentText"`
}

// 定义允许操作的基础目录

var baseTemplateDir string

func init() {
	// === 修改点开始 ===
	// 获取当前工作目录 (Current Working Directory)
	// 当您在项目根目录运行 `go run main.go` 时，这将是项目根目录
	cwd, err := os.Getwd()
	if err != nil {
		utils.Fatal("无法获取当前工作目录: %v", err)
	}

	// 将 "template" 路径解析为相对于当前工作目录的绝对路径
	absPath, err := filepath.Abs(filepath.Join(cwd, "template"))
	if err != nil {
		utils.Fatal("无法解析 template 目录的绝对路径: %v", err)
	}
	baseTemplateDir = absPath
	utils.Info("基础模板目录已初始化为: %s (基于当前工作目录)", baseTemplateDir)
	// === 修改点结束 ===

	// 确保基础模板目录存在，如果不存在则创建
	if _, err := os.Stat(baseTemplateDir); os.IsNotExist(err) {
		if err := os.MkdirAll(baseTemplateDir, 0755); err != nil {
			utils.Fatal("无法创建基础模板目录 %s: %v", baseTemplateDir, err)
		}
		utils.Info("已创建基础模板目录: %s", baseTemplateDir)
	}
}

// safeFilename 生成安全的文件路径，防止目录遍历
func safeFilePath(filename string) (string, error) {
	// 1. 清理用户提供的文件名，移除冗余的 "." 和 ".." 等。
	cleanFilename := filepath.Clean(filename)

	// 2. 严格检查文件名是否包含任何路径分隔符。
	// 这强制只允许在 baseTemplateDir 下直接操作文件，不能通过文件名创建子目录。
	if strings.ContainsRune(cleanFilename, os.PathSeparator) ||
		strings.ContainsRune(cleanFilename, '/') ||
		strings.ContainsRune(cleanFilename, '\\') {
		return "", errors.New("文件名不能包含路径分隔符")
	}

	// 3. 禁止使用特殊文件名（如 ".", "..", 或空字符串）。
	if cleanFilename == "." || cleanFilename == ".." || cleanFilename == "" {
		return "", errors.New("文件名无效或指向目录本身")
	}

	// 4. 将基础目录（已是绝对路径）和清理后的文件名安全地连接起来。
	fullPath := filepath.Join(baseTemplateDir, cleanFilename)

	// 5. 再次清理完整路径，确保最终路径是规范化的。
	finalCleanPath := filepath.Clean(fullPath)

	// 6. **核心安全检查**: 验证最终路径是否仍在 `baseTemplateDir` 的范围内。
	// `filepath.Rel` 计算 `finalCleanPath` 相对于 `baseTemplateDir` 的相对路径。
	// 如果 `finalCleanPath` 跳出了 `baseTemplateDir`，那么 `relPath` 会以 ".." 开头。
	relPath, err := filepath.Rel(baseTemplateDir, finalCleanPath)
	if err != nil {
		// `filepath.Rel` 错误通常表示路径不兼容或发生异常，视为不安全。
		return "", errors.New("路径处理错误: " + err.Error())
	}
	if strings.HasPrefix(relPath, "..") {
		// 如果相对路径以 ".." 开头，说明存在目录遍历企图。
		return "", errors.New("检测到目录遍历尝试")
	}

	// 7. 确保最终路径不是 `baseTemplateDir` 本身（例如，如果用户传入 "."）。
	// 这防止了将根目录本身作为“文件”进行操作。
	if finalCleanPath == baseTemplateDir {
		return "", errors.New("文件名无效或指向根目录本身")
	}

	return finalCleanPath, nil
}

func GetTempS(c *gin.Context) {
	// 由于 init() 函数已经确保了 baseTemplateDir 的存在，这里无需再次检查和创建。
	files, err := os.ReadDir(baseTemplateDir)
	if err != nil {
		utils.Error("读取模板目录失败: %v", err)
		utils.FailWithMsg(c, "服务器错误：无法读取模板文件")
		return
	}

	var temps []Temp
	for _, file := range files {
		// 跳过目录，因为我们只处理文件
		if file.IsDir() {
			continue
		}

		// **修复点：对读取的文件名也使用 safeFilePath 进行验证**
		// 这可以防止通过符号链接（symlink）进行的目录遍历，从而避免信息泄露。
		fullPathToRead, err := safeFilePath(file.Name())
		if err != nil {
			utils.Warn("跳过不安全或非法文件 (读取): %s, 错误: %v", file.Name(), err)
			continue // 跳过不安全的文件
		}

		info, err := file.Info()
		if err != nil {
			utils.Warn("获取文件信息失败: %s, 错误: %v", file.Name(), err)
			continue
		}
		modTime := info.ModTime().Format("2006-01-02 15:04:05")

		// 优先从缓存读取模板内容
		var textContent string
		if cached, ok := cache.GetTemplateContent(file.Name()); ok {
			textContent = cached
		} else {
			// 缓存未命中，从文件读取并写入缓存
			textBytes, readErr := os.ReadFile(fullPathToRead)
			if readErr != nil {
				utils.Error("读取文件内容失败: %s, 错误: %v", fullPathToRead, readErr)
				continue // 跳过无法读取的文件
			}
			textContent = string(textBytes)
			// 写入缓存
			cache.SetTemplateContent(file.Name(), textContent)
		}

		// 从数据库获取模板元数据
		var tmplMeta models.Template
		category := models.InferTemplateCategory(file.Name())
		ruleSource := ""
		useProxy := false
		proxyLink := ""
		enableIncludeAll := false
		if err := tmplMeta.FindByName(file.Name()); err == nil {
			if tmplMeta.Category == "clash" || tmplMeta.Category == "surge" {
				category = tmplMeta.Category
			}
			ruleSource = tmplMeta.RuleSource
			useProxy = tmplMeta.UseProxy
			proxyLink = tmplMeta.ProxyLink
			enableIncludeAll = tmplMeta.EnableIncludeAll
		}

		temp := Temp{
			File:             file.Name(),
			Text:             textContent,
			Category:         category,
			RuleSource:       ruleSource,
			UseProxy:         useProxy,
			ProxyLink:        proxyLink,
			EnableIncludeAll: enableIncludeAll,
			CreateDate:       modTime,
		}
		temps = append(temps, temp)
	}

	// 解析分页参数
	page := 0
	pageSize := 0
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	// 如果提供了分页参数，返回分页响应
	if page > 0 && pageSize > 0 {
		total := int64(len(temps))
		offset := (page - 1) * pageSize
		end := offset + pageSize

		var pagedTemps []Temp
		if offset < len(temps) {
			if end > len(temps) {
				end = len(temps)
			}
			pagedTemps = temps[offset:end]
		} else {
			pagedTemps = []Temp{}
		}

		totalPages := 0
		if pageSize > 0 {
			totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
		}
		utils.OkDetailed(c, "ok", gin.H{
			"items":      pagedTemps,
			"total":      total,
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": totalPages,
		})
		return
	}

	// 不带分页参数，返回全部（向后兼容）
	if len(temps) == 0 {
		utils.OkDetailed(c, "ok", []Temp{})
		return
	}
	utils.OkDetailed(c, "ok", temps)
}
func UpdateTemp(c *gin.Context) {
	filename := c.PostForm("filename")
	oldname := c.PostForm("oldname")
	text := c.PostForm("text")
	category := c.PostForm("category")
	ruleSource := c.PostForm("ruleSource")
	useProxy := c.PostForm("useProxy") == "true"
	proxyLink := c.PostForm("proxyLink")
	enableIncludeAll := c.PostForm("enableIncludeAll") == "true"

	if filename == "" || oldname == "" || text == "" {
		utils.FailWithMsg(c, "文件名或内容不能为空")
		return
	}

	// 默认类别为 clash
	if category == "" {
		category = "clash"
	}

	// 验证旧文件名以防止目录遍历
	oldFullPath, err := safeFilePath(oldname)
	if err != nil {
		utils.FailWithMsg(c, "旧文件名非法: "+err.Error())
		return
	}

	// 验证新文件名以防止目录遍历
	newFullPath, err := safeFilePath(filename)
	if err != nil {
		utils.FailWithMsg(c, "新文件名非法: "+err.Error())
		return
	}

	// 检查旧文件是否存在
	if _, err := os.Stat(oldFullPath); os.IsNotExist(err) {
		utils.FailWithMsg(c, "旧文件不存在")
		return
	} else if err != nil {
		utils.Error("检查旧文件存在性失败: %v", err)
		utils.FailWithMsg(c, "服务器错误：检查旧文件失败")
		return
	}

	// 如果新旧文件名不同，则检查新文件是否已存在
	if oldFullPath != newFullPath {
		if _, err := os.Stat(newFullPath); err == nil {
			utils.FailWithMsg(c, "新文件名已存在，请选择其他名称")
			return
		} else if !os.IsNotExist(err) {
			utils.Error("检查新文件存在性失败: %v", err)
			utils.FailWithMsg(c, "服务器错误：检查新文件失败")
			return
		}
	}

	// 如果文件名不同，则进行重命名操作
	if oldFullPath != newFullPath {
		err = os.Rename(oldFullPath, newFullPath)
		if err != nil {
			utils.Error("文件改名失败: %v", err)
			utils.FailWithMsg(c, "改名失败")
			return
		}
	}

	// 写入文件内容到新的安全路径
	err = os.WriteFile(newFullPath, []byte(text), 0666)
	if err != nil {
		utils.Error("修改文件内容失败: %v", err)
		utils.FailWithMsg(c, "修改失败")
		return
	}

	// 同步更新模板内容缓存
	if oldname != filename {
		// 文件名变更，先删除旧缓存
		cache.InvalidateTemplateContent(oldname)
	}
	cache.SetTemplateContent(filename, text)

	// 更新数据库中的模板元数据
	var tmpl models.Template
	if err := tmpl.FindByName(oldname); err != nil {
		// 如果数据库中不存在，创建新记录
		tmpl = models.Template{
			Name:             filename,
			Category:         category,
			RuleSource:       ruleSource,
			UseProxy:         useProxy,
			ProxyLink:        proxyLink,
			EnableIncludeAll: enableIncludeAll,
		}
		if err := tmpl.Add(); err != nil {
			utils.Error("创建模板元数据失败: %v", err)
		}
	} else {
		// 更新现有记录
		tmpl.Name = filename
		tmpl.Category = category
		tmpl.RuleSource = ruleSource
		tmpl.UseProxy = useProxy
		tmpl.ProxyLink = proxyLink
		tmpl.EnableIncludeAll = enableIncludeAll
		if err := tmpl.Update(); err != nil {
			utils.Error("更新模板元数据失败: %v", err)
		}
	}

	utils.OkWithMsg(c, "修改成功")
}

func writeTemplateFileAndMeta(filename, oldname, text, category, ruleSource string, useProxy bool, proxyLink string, enableIncludeAll bool) error {
	if category == "" {
		category = "clash"
	}
	if _, err := os.Stat(baseTemplateDir); os.IsNotExist(err) {
		if err := os.MkdirAll(baseTemplateDir, 0755); err != nil {
			return fmt.Errorf("无法创建模板目录: %w", err)
		}
	}
	newFullPath, err := safeFilePath(filename)
	if err != nil {
		return fmt.Errorf("文件名非法: %w", err)
	}
	oldResolvedName := strings.TrimSpace(oldname)
	if oldResolvedName == "" {
		if _, err := os.Stat(newFullPath); err == nil {
			return fmt.Errorf("文件已存在")
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("检查文件失败: %w", err)
		}
	} else {
		oldFullPath, err := safeFilePath(oldResolvedName)
		if err != nil {
			return fmt.Errorf("旧文件名非法: %w", err)
		}
		if _, err := os.Stat(oldFullPath); os.IsNotExist(err) {
			return fmt.Errorf("旧文件不存在")
		} else if err != nil {
			return fmt.Errorf("检查旧文件失败: %w", err)
		}
		if oldFullPath != newFullPath {
			if _, err := os.Stat(newFullPath); err == nil {
				return fmt.Errorf("新文件名已存在，请选择其他名称")
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("检查新文件失败: %w", err)
			}
			if err := os.Rename(oldFullPath, newFullPath); err != nil {
				return fmt.Errorf("模板改名失败: %w", err)
			}
			cache.InvalidateTemplateContent(oldResolvedName)
		}
	}
	if err := os.WriteFile(newFullPath, []byte(text), 0666); err != nil {
		return fmt.Errorf("写入模板失败: %w", err)
	}
	cache.SetTemplateContent(filename, text)
	var tmpl models.Template
	if lookupName := oldResolvedName; lookupName != "" {
		if err := tmpl.FindByName(lookupName); err == nil {
			tmpl.Name = filename
			tmpl.Category = category
			tmpl.RuleSource = ruleSource
			tmpl.UseProxy = useProxy
			tmpl.ProxyLink = proxyLink
			tmpl.EnableIncludeAll = enableIncludeAll
			return tmpl.Update()
		}
	}
	tmpl = models.Template{
		Name:             filename,
		Category:         category,
		RuleSource:       ruleSource,
		UseProxy:         useProxy,
		ProxyLink:        proxyLink,
		EnableIncludeAll: enableIncludeAll,
	}
	return tmpl.Add()
}
func AddTemp(c *gin.Context) {
	filename := c.PostForm("filename")
	text := c.PostForm("text")
	category := c.PostForm("category")
	ruleSource := c.PostForm("ruleSource")
	useProxy := c.PostForm("useProxy") == "true"
	proxyLink := c.PostForm("proxyLink")
	enableIncludeAll := c.PostForm("enableIncludeAll") == "true"

	if filename == "" || text == "" {
		utils.FailWithMsg(c, "文件名或内容不能为空")
		return
	}

	// 默认类别为 clash
	if category == "" {
		category = "clash"
	}

	if err := writeTemplateFileAndMeta(filename, "", text, category, ruleSource, useProxy, proxyLink, enableIncludeAll); err != nil {
		utils.Error("创建模板失败: %v", err)
		utils.FailWithMsg(c, err.Error())
		return
	}

	utils.OkWithMsg(c, "上传成功")
}

func DelTemp(c *gin.Context) {
	filename := c.PostForm("filename")

	if filename == "" {
		utils.FailWithMsg(c, "文件名不能为空")
		return
	}

	// 获取安全的文件路径
	fullPath, err := safeFilePath(filename)
	if err != nil {
		utils.FailWithMsg(c, "文件名非法: "+err.Error())
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		utils.FailWithMsg(c, "文件不存在")
		return
	} else if err != nil {
		utils.Error("检查文件存在性失败: %v", err)
		utils.FailWithMsg(c, "服务器错误：检查文件失败")
		return
	}

	// 删除文件
	err = os.Remove(fullPath)
	if err != nil {
		utils.Error("删除文件失败: %v", err)
		utils.FailWithMsg(c, "删除失败")
		return
	}

	// 清除模板内容缓存
	cache.InvalidateTemplateContent(filename)

	// 删除数据库记录
	var tmpl models.Template
	if err := tmpl.FindByName(filename); err == nil {
		if err := tmpl.Delete(); err != nil {
			utils.Error("删除模板元数据失败: %v", err)
		}
	}

	utils.OkWithMsg(c, "删除成功")
}

func normalizeTemplateUsageValue(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
}

func buildTemplateMatchValues(filename string) map[string]struct{} {
	normalizedFile := normalizeTemplateUsageValue(filename)
	fileName := filepath.Base(normalizedFile)
	values := []string{normalizedFile, fileName, "./template/" + fileName, "/template/" + fileName}
	matchValues := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = normalizeTemplateUsageValue(value)
		if value == "" {
			continue
		}
		matchValues[value] = struct{}{}
	}
	return matchValues
}

func getTemplateUsageSubscriptions(filename string) ([]string, error) {
	var subs []models.Subcription
	if err := database.DB.Find(&subs).Error; err != nil {
		return nil, err
	}

	matchValues := buildTemplateMatchValues(filename)
	usedBy := make([]string, 0)

	for _, sub := range subs {
		var config struct {
			Clash string `json:"clash"`
			Surge string `json:"surge"`
		}

		if sub.Config != "" {
			if err := json.Unmarshal([]byte(sub.Config), &config); err != nil {
				continue
			}
		}

		clashValue := normalizeTemplateUsageValue(config.Clash)
		surgeValue := normalizeTemplateUsageValue(config.Surge)
		if _, ok := matchValues[clashValue]; ok {
			usedBy = append(usedBy, sub.Name)
			continue
		}
		if _, ok := matchValues[surgeValue]; ok {
			usedBy = append(usedBy, sub.Name)
		}
	}

	return usedBy, nil
}

func GetTemplateUsage(c *gin.Context) {
	filename := c.Query("filename")
	if filename == "" {
		utils.FailWithMsg(c, "文件名不能为空")
		return
	}

	usedBy, err := getTemplateUsageSubscriptions(filename)
	if err != nil {
		utils.FailWithMsg(c, "获取模板使用情况失败")
		return
	}

	utils.OkWithData(c, gin.H{
		"subscriptions": usedBy,
		"count":         len(usedBy),
	})
}

type TemplateAIGenerateRequest struct {
	Filename         string `json:"filename"`
	Category         string `json:"category"`
	CurrentText      string `json:"currentText"`
	UserPrompt       string `json:"userPrompt"`
	RuleSource       string `json:"ruleSource"`
	UseProxy         bool   `json:"useProxy"`
	ProxyLink        string `json:"proxyLink"`
	EnableIncludeAll bool   `json:"enableIncludeAll"`
}

type templateEditSessionPayload struct {
	SessionID          string                       `json:"sessionId"`
	Status             ai.TemplateEditSessionStatus `json:"status"`
	Filename           string                       `json:"filename,omitempty"`
	Category           string                       `json:"category,omitempty"`
	BaseHash           string                       `json:"baseHash"`
	CandidateHash      string                       `json:"candidateHash"`
	CandidateText      string                       `json:"candidateText"`
	Operations         []ai.TemplateEditOperation   `json:"operations"`
	Validation         ai.ValidationResult          `json:"validation"`
	WarningFingerprint string                       `json:"warningFingerprint"`
	CreatedAt          any                          `json:"createdAt,omitempty"`
	ExpiresAt          any                          `json:"expiresAt"`
	LastError          string                       `json:"lastError,omitempty"`
}

var templateEditSessionCleanupStarter = startTemplateEditSessionCleanup

var templateEditSessions = newTemplateEditSessionStoreRuntime()

func newTemplateEditSessionStoreRuntime() *ai.TemplateEditSessionStore {
	store := newTemplateEditSessionStoreFromEnv()
	templateEditSessionCleanupStarter(store)
	return store
}

func startTemplateEditSessionCleanup(store *ai.TemplateEditSessionStore) {
	go store.StartCleanup(context.Background())
}

func newTemplateEditSessionStoreFromEnv() *ai.TemplateEditSessionStore {
	options := []ai.TemplateEditSessionStoreOption{}
	if ai.IsTemplateEditMockProviderAllowed() {
		if ttlSeconds, err := strconv.Atoi(strings.TrimSpace(os.Getenv("SUBLINK_TEMPLATE_EDIT_SESSION_TTL_SECONDS"))); err == nil && ttlSeconds > 0 && ttlSeconds <= 3600 {
			options = append(options, ai.WithTemplateEditSessionTTL(time.Duration(ttlSeconds)*time.Second))
		}
	}
	return ai.NewTemplateEditSessionStore(options...)
}

const (
	templateEditLegacyRemovedCode = "AI_EDIT_LEGACY_REMOVED"
	templateEditStaleBaseCode     = "AI_EDIT_STALE_BASE"
)

func StartTemplateAIEditSessionStream(c *gin.Context) {
	var req TemplateAIGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Category) == "" {
		req.Category = "clash"
	}
	if strings.TrimSpace(req.UserPrompt) == "" {
		utils.FailWithMsg(c, "请输入 AI 指令")
		return
	}
	user, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	ownerKey := ai.TemplateEditSessionOwnerKey(user)
	if ownerKey == "" {
		utils.FailWithData(c, "无法确定当前用户", gin.H{"code": string(ai.TemplateEditInvalidOperation), "message": "template edit session requires an owner"})
		return
	}
	baseHash := ai.BuildRevisionHash(req.CurrentText)
	session, err := templateEditSessions.Create(ai.TemplateEditSessionCreateInput{
		OwnerKey: ownerKey,
		Filename: req.Filename,
		Category: req.Category,
		BaseHash: baseHash,
		BaseText: req.CurrentText,
		Status:   ai.TemplateEditSessionCreated,
	})
	if err != nil {
		respondTemplateEditError(c, err)
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		utils.FailWithMsg(c, "当前环境不支持流式响应")
		return
	}
	writeEvent := func(event string, payload any) error {
		writer := bufio.NewWriter(c.Writer)
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		if _, err := writer.WriteString("event: " + event + "\n"); err != nil {
			return err
		}
		if _, err := writer.WriteString("data: " + string(data) + "\n\n"); err != nil {
			return err
		}
		if err := writer.Flush(); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}
	writeErrorEvent := func(err error) {
		_ = writeEvent("template.edit.error", templateEditErrorPayload(err))
	}

	_ = writeEvent("template.edit.session.created", templateEditSessionResponse(session))
	if _, err := templateEditSessions.UpdateForOwner(ownerKey, session.SessionID, func(current *ai.TemplateEditSession) error {
		current.Status = ai.TemplateEditSessionStreaming
		return nil
	}); err != nil {
		writeErrorEvent(err)
		return
	}

	result, err := ai.GenerateCandidateStream(c.Request.Context(), user, ai.GenerateRequest{
		Filename:         req.Filename,
		Category:         req.Category,
		CurrentText:      req.CurrentText,
		UserPrompt:       req.UserPrompt,
		RuleSource:       req.RuleSource,
		UseProxy:         req.UseProxy,
		ProxyLink:        req.ProxyLink,
		EnableIncludeAll: req.EnableIncludeAll,
	}, func(event ai.ResponsesEvent) error {
		payload, ok := templateEditModelDeltaPayload(event)
		if !ok {
			return nil
		}
		return writeEvent("template.edit.model.delta", payload)
	})
	if err != nil {
		_, _ = templateEditSessions.UpdateForOwner(ownerKey, session.SessionID, func(current *ai.TemplateEditSession) error {
			current.Status = ai.TemplateEditSessionFailed
			current.LastError = err.Error()
			return nil
		})
		writeErrorEvent(err)
		return
	}
	session, err = templateEditSessions.UpdateForOwner(ownerKey, session.SessionID, func(current *ai.TemplateEditSession) error {
		current.Status = ai.TemplateEditSessionOperationsReady
		current.Operations = result.Operations
		return nil
	})
	if err != nil {
		writeErrorEvent(err)
		return
	}
	_ = writeEvent("template.edit.operations.ready", gin.H{
		"sessionId":  session.SessionID,
		"operations": result.Operations,
		"summary":    result.Summary,
		"warnings":   result.Warnings,
	})

	session, err = templateEditSessions.UpdateForOwner(ownerKey, session.SessionID, func(current *ai.TemplateEditSession) error {
		current.Status = ai.TemplateEditSessionValidating
		return nil
	})
	if err != nil {
		writeErrorEvent(err)
		return
	}
	_ = writeEvent("template.edit.preview.validating", gin.H{"sessionId": session.SessionID, "baseHash": baseHash})

	usedBy, _ := getTemplateUsageSubscriptions(req.Filename)
	preview, err := ai.BuildTemplateEditPreviewCandidate(ai.TemplateEditPreviewInput{
		Category:      req.Category,
		BaseText:      req.CurrentText,
		BaseHash:      baseHash,
		Operations:    result.Operations,
		RuleSource:    req.RuleSource,
		Subscriptions: usedBy,
	})
	if err != nil {
		_, _ = templateEditSessions.UpdateForOwner(ownerKey, session.SessionID, func(current *ai.TemplateEditSession) error {
			current.Status = ai.TemplateEditSessionFailed
			current.Validation = preview.Validation
			current.LastError = err.Error()
			return nil
		})
		writeErrorEvent(err)
		return
	}

	session, err = templateEditSessions.UpdateForOwner(ownerKey, session.SessionID, func(current *ai.TemplateEditSession) error {
		current.Status = ai.TemplateEditSessionPreviewReady
		current.CandidateText = preview.CandidateText
		current.Operations = preview.Operations
		current.Validation = preview.Validation
		current.WarningFingerprint = preview.WarningFingerprint
		return nil
	})
	if err != nil {
		writeErrorEvent(err)
		return
	}
	if len(session.Validation.Warnings) > 0 {
		_ = writeEvent("template.edit.warning", gin.H{
			"sessionId":          session.SessionID,
			"warnings":           session.Validation.Warnings,
			"warningFingerprint": session.WarningFingerprint,
		})
	}
	previewPayload := templateEditSessionResponse(session)
	_ = writeEvent("template.edit.preview.ready", previewPayload)
	_ = writeEvent("template.edit.completed", previewPayload)
}

func GetTemplateAIEditSession(c *gin.Context) {
	user, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	session, err := templateEditSessions.GetForOwner(ai.TemplateEditSessionOwnerKey(user), c.Param("sessionId"))
	if err != nil {
		respondTemplateEditError(c, err)
		return
	}
	utils.OkDetailed(c, "ok", templateEditSessionResponse(session))
}

func AcceptTemplateAIEditSession(c *gin.Context) {
	user, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	ownerKey := ai.TemplateEditSessionOwnerKey(user)
	session, err := templateEditSessions.GetForOwner(ownerKey, c.Param("sessionId"))
	if err != nil {
		respondTemplateEditError(c, err)
		return
	}
	if session.Status != ai.TemplateEditSessionPreviewReady {
		respondTemplateEditError(c, ai.NewTemplateEditError(ai.TemplateEditInvalidOperation, "template edit session preview is not ready"))
		return
	}
	var req TemplateEditSessionAcceptRequest
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			respondTemplateEditError(c, ai.NewTemplateEditError(ai.TemplateEditInvalidOperation, "template edit accept request is invalid"))
			return
		}
	}
	fullPath, err := safeFilePath(session.Filename)
	if err != nil {
		respondTemplateEditError(c, ai.NewTemplateEditError(ai.TemplateEditInvalidOperation, "template filename is invalid"))
		return
	}
	clientBaseMatches := strings.TrimSpace(req.CurrentText) != "" && ai.BuildRevisionHash(req.CurrentText) == session.BaseHash
	if !clientBaseMatches {
		currentBytes, err := os.ReadFile(fullPath)
		if err != nil {
			utils.FailWithMsg(c, "读取当前模板失败: "+err.Error())
			return
		}
		if ai.BuildRevisionHash(string(currentBytes)) != session.BaseHash {
			respondTemplateEditCode(c, templateEditStaleBaseCode, "模板已被其他修改更新，请重新生成候选内容")
			return
		}
	}
	if err := ai.ValidateTemplateEditAccept(ai.TemplateEditAcceptValidationInput{
		Validation: session.Validation,
	}); err != nil {
		respondTemplateEditError(c, err)
		return
	}
	session, err = templateEditSessions.UpdateForOwner(ownerKey, session.SessionID, func(current *ai.TemplateEditSession) error {
		current.Status = ai.TemplateEditSessionAccepted
		return nil
	})
	if err != nil {
		respondTemplateEditError(c, err)
		return
	}
	utils.OkDetailed(c, "AI 修改预览已接受", gin.H{
		"sessionId":     session.SessionID,
		"candidateText": session.CandidateText,
		"candidateHash": ai.BuildRevisionHash(session.CandidateText),
		"validation":    session.Validation,
	})
}

func DiscardTemplateAIEditSession(c *gin.Context) {
	user, ok := requireCurrentUser(c)
	if !ok {
		return
	}
	session, err := templateEditSessions.DiscardForOwner(ai.TemplateEditSessionOwnerKey(user), c.Param("sessionId"))
	if err != nil {
		respondTemplateEditError(c, err)
		return
	}
	utils.OkDetailed(c, "AI 修改会话已丢弃", templateEditSessionResponse(session))
}

func TemplateAILegacyRemoved(c *gin.Context) {
	c.JSON(http.StatusGone, gin.H{
		"code":    templateEditLegacyRemovedCode,
		"message": "Legacy full-template AI endpoints have been removed. Use /api/v1/template/ai/edit-sessions/*.",
	})
}

func templateEditSessionResponse(session ai.TemplateEditSession) templateEditSessionPayload {
	return templateEditSessionPayload{
		SessionID:          session.SessionID,
		Status:             session.Status,
		Filename:           session.Filename,
		Category:           session.Category,
		BaseHash:           session.BaseHash,
		CandidateHash:      ai.BuildRevisionHash(session.CandidateText),
		CandidateText:      session.CandidateText,
		Operations:         session.Operations,
		Validation:         session.Validation,
		WarningFingerprint: session.WarningFingerprint,
		CreatedAt:          session.CreatedAt,
		ExpiresAt:          session.ExpiresAt,
		LastError:          session.LastError,
	}
}

func templateEditModelDeltaPayload(event ai.ResponsesEvent) (gin.H, bool) {
	if len(event.Data) == 0 {
		return nil, false
	}
	var payload struct {
		Delta string `json:"delta"`
	}
	if err := json.Unmarshal(event.Data, &payload); err != nil || payload.Delta == "" {
		return nil, false
	}
	return gin.H{"delta": payload.Delta}, true
}

func respondTemplateEditError(c *gin.Context, err error) {
	payload := templateEditErrorPayload(err)
	message, _ := payload["message"].(string)
	utils.FailWithData(c, message, payload)
}

func respondTemplateEditCode(c *gin.Context, code string, message string) {
	utils.FailWithData(c, message, gin.H{"code": code, "message": message})
}

func templateEditErrorPayload(err error) gin.H {
	if err == nil {
		return gin.H{"code": string(ai.TemplateEditInvalidOperation), "message": "template edit failed"}
	}
	var editErr *ai.TemplateEditError
	if errors.As(err, &editErr) {
		return gin.H{"code": string(editErr.Code), "message": editErr.Message}
	}
	var patchErr *ai.PatchError
	if errors.As(err, &patchErr) {
		return gin.H{"code": string(patchErr.Code), "message": patchErr.Message, "operationIndex": patchErr.OperationIndex}
	}
	return gin.H{"code": string(ai.TemplateEditInvalidOperation), "message": err.Error()}
}

// ACL4SSRPreset ACL4SSR 规则预设
type ACL4SSRPreset struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Label string `json:"label"`
}

// GetACL4SSRPresets 获取 ACL4SSR 规则预设列表
func GetACL4SSRPresets(c *gin.Context) {
	presets := []ACL4SSRPreset{
		{
			Name:  "无国家分组",
			URL:   "https://raw.githubusercontent.com/ZeroDeng01/ACL4SSR/master/Clash/config/ACL4SSR_Online_Full_NoCountry.ini",
			Label: "不区分国家",
		},
		{
			Name:  "ACL4SSR",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR.ini",
			Label: "标准版 - 典型分组",
		},
		{
			Name:  "ACL4SSR_AdblockPlus",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_AdblockPlus.ini",
			Label: "标准版 - 典型分组+去广告",
		},
		{
			Name:  "ACL4SSR_BackCN",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_BackCN.ini",
			Label: "回国版 - 回国专用",
		},
		{
			Name:  "ACL4SSR_Mini",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Mini.ini",
			Label: "精简版 - 少量分组",
		},
		{
			Name:  "ACL4SSR_Mini_Fallback",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Mini_Fallback.ini",
			Label: "精简版 - 故障转移",
		},
		{
			Name:  "ACL4SSR_Mini_MultiMode",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Mini_MultiMode.ini",
			Label: "精简版 - 多模式 (自动/手动)",
		},
		{
			Name:  "ACL4SSR_Mini_NoAuto",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Mini_NoAuto.ini",
			Label: "精简版 - 无自动测速",
		},
		{
			Name:  "ACL4SSR_NoApple",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_NoApple.ini",
			Label: "无苹果 - 无苹果分流",
		},
		{
			Name:  "ACL4SSR_NoAuto",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_NoAuto.ini",
			Label: "无测速 - 无自动测速",
		},
		{
			Name:  "ACL4SSR_NoAuto_NoApple",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_NoAuto_NoApple.ini",
			Label: "无测速&苹果 - 无测速&无苹果分流",
		},
		{
			Name:  "ACL4SSR_NoAuto_NoApple_NoMicrosoft",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_NoAuto_NoApple_NoMicrosoft.ini",
			Label: "无测速&苹果&微软 - 无测速&无苹果&无微软分流",
		},
		{
			Name:  "ACL4SSR_NoMicrosoft",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_NoMicrosoft.ini",
			Label: "无微软 - 无微软分流",
		},
		{
			Name:  "ACL4SSR_Online",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online.ini",
			Label: "在线版 - 典型分组",
		},
		{
			Name:  "ACL4SSR_Online_AdblockPlus",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_AdblockPlus.ini",
			Label: "在线版 - 典型分组+去广告",
		},
		{
			Name:  "ACL4SSR_Online_Full",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Full.ini",
			Label: "在线全分组 - 比较全",
		},
		{
			Name:  "ACL4SSR_Online_Full_AdblockPlus",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Full_AdblockPlus.ini",
			Label: "在线全分组 - 带广告拦截",
		},
		{
			Name:  "ACL4SSR_Online_Full_Google",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Full_Google.ini",
			Label: "在线全分组 - 谷歌分流",
		},
		{
			Name:  "ACL4SSR_Online_Full_MultiMode",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Full_MultiMode.ini",
			Label: "在线全分组 - 多模式",
		},
		{
			Name:  "ACL4SSR_Online_Full_Netflix",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Full_Netflix.ini",
			Label: "在线全分组 - 奈飞分流",
		},
		{
			Name:  "ACL4SSR_Online_Full_NoAuto",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Full_NoAuto.ini",
			Label: "在线全分组 - 无自动测速",
		},
		{
			Name:  "ACL4SSR_Online_Mini",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Mini.ini",
			Label: "在线精简版 - 少量分组",
		},
		{
			Name:  "ACL4SSR_Online_Mini_AdblockPlus",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Mini_AdblockPlus.ini",
			Label: "在线精简版 - 带广告拦截",
		},
		{
			Name:  "ACL4SSR_Online_Mini_Ai",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Mini_Ai.ini",
			Label: "在线精简版 - AI",
		},
		{
			Name:  "ACL4SSR_Online_Mini_Fallback",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Mini_Fallback.ini",
			Label: "在线精简版 - 故障转移",
		},
		{
			Name:  "ACL4SSR_Online_Mini_MultiCountry",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Mini_MultiCountry.ini",
			Label: "在线精简版 - 多国家",
		},
		{
			Name:  "ACL4SSR_Online_Mini_MultiMode",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Mini_MultiMode.ini",
			Label: "在线精简版 - 多模式",
		},
		{
			Name:  "ACL4SSR_Online_Mini_NoAuto",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Mini_NoAuto.ini",
			Label: "在线精简版 - 无自动测速",
		},
		{
			Name:  "ACL4SSR_Online_MultiCountry",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_MultiCountry.ini",
			Label: "在线版 - 多国家",
		},
		{
			Name:  "ACL4SSR_Online_NoAuto",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_NoAuto.ini",
			Label: "在线版 - 无自动测速",
		},
		{
			Name:  "ACL4SSR_Online_NoReject",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_NoReject.ini",
			Label: "在线版 - 无拒绝规则",
		},
		{
			Name:  "ACL4SSR_WithChinaIp",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_WithChinaIp.ini",
			Label: "特殊版 - 包含回国IP",
		},
		{
			Name:  "ACL4SSR_WithChinaIp_WithGFW",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_WithChinaIp_WithGFW.ini",
			Label: "特殊版 - 包含回国IP&GFW列表",
		},
		{
			Name:  "ACL4SSR_WithGFW",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_WithGFW.ini",
			Label: "特殊版 - 包含GFW列表",
		},
	}
	utils.OkDetailed(c, "ok", presets)
}
