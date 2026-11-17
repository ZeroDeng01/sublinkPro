package api

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sublink/cache"
	"sublink/models"
	"sublink/utils"

	"github.com/gin-gonic/gin"
	"github.com/pmezard/go-difflib/difflib"
	"gorm.io/gorm"
)

// templateVersionItem 列表项，额外标记该版本是否与当前模板内容一致
type templateVersionItem struct {
	models.TemplateVersion
	IsCurrent bool `json:"isCurrent"`
}

const (
	defaultTemplateVersionPageSize = 20
	maxTemplateVersionPageSize     = 200
)

// templateAuthorFromContext 读取当前登录用户名，未认证场景下返回空字符串
func templateAuthorFromContext(c *gin.Context) string {
	if value, exists := c.Get("username"); exists {
		if username, ok := value.(string); ok {
			return strings.TrimSpace(username)
		}
	}
	return ""
}

func parseTemplateVersionNumber(value string) (int, error) {
	version, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || version <= 0 {
		return 0, errors.New("invalid version number")
	}
	return version, nil
}

// findTemplateByFilename 校验文件名并查询模板元数据
func findTemplateByFilename(filename string) (*models.Template, string, error) {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return nil, "", errors.New("file name is required")
	}
	fullPath, err := safeFilePath(filename)
	if err != nil {
		return nil, "", fmt.Errorf("invalid file name: %w", err)
	}

	var tmpl models.Template
	if err := tmpl.FindByName(filename); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fullPath, errors.New("template not found")
		}
		return nil, fullPath, err
	}
	return &tmpl, fullPath, nil
}

func templateVersionTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	return int((total + int64(pageSize) - 1) / int64(pageSize))
}

// GetTemplateVersions 获取模板版本列表
func GetTemplateVersions(c *gin.Context) {
	filename := strings.TrimSpace(c.Query("filename"))
	if filename == "" {
		utils.FailWithMsg(c, "File name is required")
		return
	}
	fullPath, err := safeFilePath(filename)
	if err != nil {
		utils.FailWithMsg(c, "Invalid file name: "+err.Error())
		return
	}

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
			page = parsed
		}
	}
	pageSize := defaultTemplateVersionPageSize
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if parsed, err := strconv.Atoi(pageSizeStr); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}
	if pageSize > maxTemplateVersionPageSize {
		pageSize = maxTemplateVersionPageSize
	}

	var tmpl models.Template
	if err := tmpl.FindByName(filename); err != nil {
		// 模板文件存在但尚未生成元数据时没有历史版本
		utils.OkDetailed(c, "ok", gin.H{
			"items":      []templateVersionItem{},
			"total":      int64(0),
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": 0,
			"templateId": 0,
		})
		return
	}

	versions, total, err := models.ListTemplateVersions(tmpl.ID, page, pageSize)
	if err != nil {
		utils.Error("Failed to load template version list: %v", err)
		utils.FailWithMsg(c, "Failed to read version history")
		return
	}

	currentHash := ""
	if content, err := os.ReadFile(fullPath); err == nil {
		currentHash = models.BuildTemplateContentHash(content)
	}

	items := make([]templateVersionItem, 0, len(versions))
	currentMarked := false
	for _, version := range versions {
		isCurrent := !currentMarked && currentHash != "" && version.ContentHash == currentHash
		if isCurrent {
			currentMarked = true
		}
		items = append(items, templateVersionItem{TemplateVersion: version, IsCurrent: isCurrent})
	}

	utils.OkDetailed(c, "ok", gin.H{
		"items":      items,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": templateVersionTotalPages(total, pageSize),
		"templateId": tmpl.ID,
	})
}

// GetTemplateVersion 获取单个版本的元数据与内容
func GetTemplateVersion(c *gin.Context) {
	filename := c.Query("filename")
	versionNo, err := parseTemplateVersionNumber(c.Param("version"))
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	tmpl, _, err := findTemplateByFilename(filename)
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	version, err := models.FindTemplateVersion(tmpl.ID, versionNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.FailWithMsg(c, "Version not found")
			return
		}
		utils.Error("Failed to query template version: %v", err)
		utils.FailWithMsg(c, "Failed to read version")
		return
	}

	content, err := models.ReadTemplateVersionContent(baseTemplateDir, version)
	if err != nil {
		utils.Error("Failed to read template version content: %v", err)
		utils.FailWithMsg(c, "Failed to read version content")
		return
	}

	utils.OkDetailed(c, "ok", gin.H{
		"version": version,
		"content": string(content),
	})
}

// DownloadTemplateVersion 下载指定版本的快照内容
func DownloadTemplateVersion(c *gin.Context) {
	filename := c.Query("filename")
	versionNo, err := parseTemplateVersionNumber(c.Param("version"))
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	tmpl, _, err := findTemplateByFilename(filename)
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	version, err := models.FindTemplateVersion(tmpl.ID, versionNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.FailWithMsg(c, "Version not found")
			return
		}
		utils.Error("Failed to query template version: %v", err)
		utils.FailWithMsg(c, "Failed to read version")
		return
	}

	content, err := models.ReadTemplateVersionContent(baseTemplateDir, version)
	if err != nil {
		utils.Error("Failed to read template version content: %v", err)
		utils.FailWithMsg(c, "Failed to read version content")
		return
	}

	downloadName := buildTemplateVersionDownloadName(filename, versionNo)
	c.Header("Content-Disposition", `attachment; filename="`+downloadName+`"`)
	c.Data(http.StatusOK, "application/octet-stream", content)
}

// GetTemplateVersionDiff 对比两个版本，to=0 表示当前模板内容
func GetTemplateVersionDiff(c *gin.Context) {
	filename := strings.TrimSpace(c.Query("filename"))
	fromNo, err := parseTemplateVersionNumber(c.Query("from"))
	if err != nil {
		utils.FailWithMsg(c, "Invalid source version number")
		return
	}

	toNo := 0
	if toStr := strings.TrimSpace(c.Query("to")); toStr != "" {
		parsed, err := strconv.Atoi(toStr)
		if err != nil || parsed < 0 {
			utils.FailWithMsg(c, "Invalid target version number")
			return
		}
		toNo = parsed
	}

	tmpl, fullPath, err := findTemplateByFilename(filename)
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	fromVersion, err := models.FindTemplateVersion(tmpl.ID, fromNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.FailWithMsg(c, "Source version does not exist")
			return
		}
		utils.Error("Failed to query template version: %v", err)
		utils.FailWithMsg(c, "Failed to read version")
		return
	}
	fromContent, err := models.ReadTemplateVersionContent(baseTemplateDir, fromVersion)
	if err != nil {
		utils.Error("Failed to read template version content: %v", err)
		utils.FailWithMsg(c, "Failed to read version content")
		return
	}

	fromLabel := fmt.Sprintf("%s v%d", filename, fromNo)
	var toContent []byte
	var toLabel string
	if toNo == 0 {
		toContent, err = os.ReadFile(fullPath)
		if err != nil {
			utils.Error("Failed to read current template content: %v", err)
			utils.FailWithMsg(c, "Failed to read current template")
			return
		}
		toLabel = filename + " (current)"
	} else {
		toVersion, err := models.FindTemplateVersion(tmpl.ID, toNo)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				utils.FailWithMsg(c, "Target version does not exist")
				return
			}
			utils.Error("Failed to query template version: %v", err)
			utils.FailWithMsg(c, "Failed to read version")
			return
		}
		toContent, err = models.ReadTemplateVersionContent(baseTemplateDir, toVersion)
		if err != nil {
			utils.Error("Failed to read template version content: %v", err)
			utils.FailWithMsg(c, "Failed to read version content")
			return
		}
		toLabel = fmt.Sprintf("%s v%d", filename, toNo)
	}

	identical := bytes.Equal(fromContent, toContent)
	diffText := ""
	if !identical {
		diffText, err = difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(fromContent)),
			B:        difflib.SplitLines(string(toContent)),
			FromFile: fromLabel,
			ToFile:   toLabel,
			Context:  3,
		})
		if err != nil {
			utils.Error("Failed to generate version diff: %v", err)
			utils.FailWithMsg(c, "Failed to generate version diff")
			return
		}
	}
	additions, deletions := countTemplateDiffStat(diffText)

	utils.OkDetailed(c, "ok", gin.H{
		"from":        fromNo,
		"to":          toNo,
		"fromLabel":   fromLabel,
		"toLabel":     toLabel,
		"fromContent": string(fromContent),
		"toContent":   string(toContent),
		"diff":        diffText,
		"additions":   additions,
		"deletions":   deletions,
		"identical":   identical,
	})
}

// RestoreTemplateVersion 将指定版本内容恢复为当前模板内容
// 只恢复内容，不修改模板的分类、规则源等元数据。
func RestoreTemplateVersion(c *gin.Context) {
	filename := strings.TrimSpace(c.PostForm("filename"))
	versionNo, err := parseTemplateVersionNumber(c.PostForm("version"))
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	comment := strings.TrimSpace(c.PostForm("comment"))

	tmpl, fullPath, err := findTemplateByFilename(filename)
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	version, err := models.FindTemplateVersion(tmpl.ID, versionNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.FailWithMsg(c, "Version not found")
			return
		}
		utils.Error("Failed to query template version: %v", err)
		utils.FailWithMsg(c, "Failed to read version")
		return
	}

	content, err := models.ReadTemplateVersionContent(baseTemplateDir, version)
	if err != nil {
		utils.Error("Failed to read template version content: %v", err)
		utils.FailWithMsg(c, "Failed to read version content")
		return
	}

	mode := os.FileMode(0666)
	if info, err := os.Stat(fullPath); err == nil {
		mode = info.Mode().Perm()
	}
	if err := os.WriteFile(fullPath, content, mode); err != nil {
		utils.Error("Failed to write restored template version: %v", err)
		utils.FailWithMsg(c, "Failed to restore version")
		return
	}
	cache.SetTemplateContent(filename, string(content))

	if comment == "" {
		comment = fmt.Sprintf("restore v%d", versionNo)
	}
	newVersion, created, err := models.AppendTemplateVersion(
		baseTemplateDir,
		tmpl.ID,
		content,
		models.TemplateVersionSourceRestore,
		comment,
		templateAuthorFromContext(c),
	)
	if err != nil {
		models.LogTemplateVersionError("record restored version", err)
		newVersion = nil
		created = false
	}

	utils.OkDetailed(c, "Restored successfully", gin.H{
		"restoredVersion": versionNo,
		"newVersion":      newVersion,
		"created":         created,
	})
}

func countTemplateDiffStat(diffText string) (int, int) {
	additions, deletions := 0, 0
	for _, line := range strings.Split(diffText, "\n") {
		switch {
		case strings.HasPrefix(line, "+++"), strings.HasPrefix(line, "---"):
			continue
		case strings.HasPrefix(line, "+"):
			additions++
		case strings.HasPrefix(line, "-"):
			deletions++
		}
	}
	return additions, deletions
}

func buildTemplateVersionDownloadName(filename string, version int) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	name := fmt.Sprintf("%s.v%d%s", base, version, ext)
	name = strings.ReplaceAll(name, `"`, "")
	name = strings.ReplaceAll(name, "\r", "")
	name = strings.ReplaceAll(name, "\n", "")
	return name
}
