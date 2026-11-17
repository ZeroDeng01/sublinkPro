package models

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"sublink/database"
	"sublink/utils"

	"gorm.io/gorm"
)

// 模板版本来源
const (
	TemplateVersionSourceBaseline = "baseline" // 启用版本功能前的原始内容
	TemplateVersionSourceAdd      = "add"      // 新建模板
	TemplateVersionSourceSave     = "save"     // 保存内容
	TemplateVersionSourceRename   = "rename"   // 保存并重命名
	TemplateVersionSourceRestore  = "restore"  // 恢复历史版本
)

// templateVersionDirName 版本快照目录名，位于模板根目录下并作为子目录被
// 模板列表与文件迁移逻辑跳过。
const templateVersionDirName = ".versions"

// TemplateVersion 模板内容历史版本，内容快照存放在模板目录的 .versions 下
type TemplateVersion struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`
	TemplateID  int       `gorm:"not null;uniqueIndex:idx_template_version,priority:1" json:"templateId"`
	Version     int       `gorm:"not null;uniqueIndex:idx_template_version,priority:2" json:"version"`
	ContentHash string    `gorm:"size:64;not null" json:"contentHash"`
	Size        int       `gorm:"not null;default:0" json:"size"`
	Source      string    `gorm:"size:16;not null;default:'save'" json:"source"`
	Comment     string    `gorm:"size:255" json:"comment"`
	Author      string    `gorm:"size:191" json:"author"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// BuildTemplateContentHash 计算模板内容摘要，用于版本去重
func BuildTemplateContentHash(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// TemplateVersionsRoot 返回版本快照根目录
func TemplateVersionsRoot(templateDir string) string {
	return filepath.Join(templateDir, templateVersionDirName)
}

// TemplateVersionDir 返回单个模板的版本快照目录
func TemplateVersionDir(templateDir string, templateID int) string {
	return filepath.Join(TemplateVersionsRoot(templateDir), strconv.Itoa(templateID))
}

// templateVersionFilePath 返回版本快照文件路径，文件名只由版本号生成
func templateVersionFilePath(templateDir string, templateID, version int) string {
	return filepath.Join(TemplateVersionDir(templateDir, templateID), fmt.Sprintf("%06d.snapshot", version))
}

// CountTemplateVersions 统计模板历史版本数量
func CountTemplateVersions(templateID int) (int64, error) {
	var count int64
	if err := database.DB.Model(&TemplateVersion{}).Where("template_id = ?", templateID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// LatestTemplateVersion 查询模板最新版本，没有历史时返回 gorm.ErrRecordNotFound
func LatestTemplateVersion(templateID int) (*TemplateVersion, error) {
	var version TemplateVersion
	if err := database.DB.Where("template_id = ?", templateID).Order("version DESC").First(&version).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

// nextTemplateVersionNumber 计算下一个版本号
func nextTemplateVersionNumber(templateID int) (int, error) {
	var maxVersion int
	if err := database.DB.Model(&TemplateVersion{}).
		Where("template_id = ?", templateID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion).Error; err != nil {
		return 0, err
	}
	return maxVersion + 1, nil
}

// AppendTemplateVersion 追加一个历史版本
// 内容与最新版本一致时不重复创建，返回已存在的版本与 created=false。
// 版本号通过唯一索引保证不重复，冲突时重试。
func AppendTemplateVersion(templateDir string, templateID int, content []byte, source, comment, author string) (*TemplateVersion, bool, error) {
	if templateID <= 0 {
		return nil, false, errors.New("invalid template ID")
	}
	if source == "" {
		source = TemplateVersionSourceSave
	}

	hash := BuildTemplateContentHash(content)
	if latest, err := LatestTemplateVersion(templateID); err == nil {
		if latest.ContentHash == hash {
			return latest, false, nil
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		nextVersion, err := nextTemplateVersionNumber(templateID)
		if err != nil {
			return nil, false, err
		}

		snapshotPath := templateVersionFilePath(templateDir, templateID, nextVersion)
		if err := writeTemplateVersionSnapshot(snapshotPath, content); err != nil {
			return nil, false, err
		}

		version := TemplateVersion{
			TemplateID:  templateID,
			Version:     nextVersion,
			ContentHash: hash,
			Size:        len(content),
			Source:      source,
			Comment:     truncateTemplateVersionText(comment, 255),
			Author:      truncateTemplateVersionText(author, 191),
		}
		if err := database.DB.Create(&version).Error; err != nil {
			_ = os.Remove(snapshotPath)
			if isTemplateVersionConflict(err) {
				lastErr = err
				continue
			}
			return nil, false, err
		}
		return &version, true, nil
	}

	if lastErr == nil {
		lastErr = errors.New("failed to create template version")
	}
	return nil, false, lastErr
}

// EnsureTemplateBaseline 在首次写入版本历史时保存改动前的原始内容
func EnsureTemplateBaseline(templateDir string, templateID int, content []byte, author string) error {
	if templateID <= 0 || len(content) == 0 {
		return nil
	}
	count, err := CountTemplateVersions(templateID)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, _, err = AppendTemplateVersion(templateDir, templateID, content, TemplateVersionSourceBaseline, "baseline", author)
	return err
}

// ListTemplateVersions 分页查询模板版本，pageSize 小于等于 0 时返回全部
func ListTemplateVersions(templateID int, page, pageSize int) ([]TemplateVersion, int64, error) {
	query := database.DB.Model(&TemplateVersion{}).Where("template_id = ?", templateID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	versions := make([]TemplateVersion, 0)
	listQuery := database.DB.Where("template_id = ?", templateID).Order("version DESC")
	if page > 0 && pageSize > 0 {
		listQuery = listQuery.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	if err := listQuery.Find(&versions).Error; err != nil {
		return nil, 0, err
	}
	return versions, total, nil
}

// FindTemplateVersion 查询指定版本
func FindTemplateVersion(templateID, version int) (*TemplateVersion, error) {
	var record TemplateVersion
	if err := database.DB.Where("template_id = ? AND version = ?", templateID, version).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// ReadTemplateVersionContent 读取版本快照内容
func ReadTemplateVersionContent(templateDir string, version *TemplateVersion) ([]byte, error) {
	if version == nil {
		return nil, errors.New("version not found")
	}
	path := templateVersionFilePath(templateDir, version.TemplateID, version.Version)
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("version snapshot file not found: %w", err)
		}
		return nil, err
	}
	return content, nil
}

// DeleteTemplateVersions 删除模板的全部版本记录与快照文件
func DeleteTemplateVersions(templateDir string, templateID int) error {
	if templateID <= 0 {
		return nil
	}

	var firstErr error
	if err := database.DB.Where("template_id = ?", templateID).Delete(&TemplateVersion{}).Error; err != nil {
		firstErr = err
	}
	if err := os.RemoveAll(TemplateVersionDir(templateDir, templateID)); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

// writeTemplateVersionSnapshot 原子写入快照文件
func writeTemplateVersionSnapshot(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	temp, err := os.CreateTemp(dir, ".snapshot-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	cleanup := func() {
		_ = temp.Close()
		_ = os.Remove(tempPath)
	}

	if _, err := temp.Write(content); err != nil {
		cleanup()
		return err
	}
	if err := temp.Close(); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	return nil
}

// isTemplateVersionConflict 判断是否为版本号唯一索引冲突
func isTemplateVersionConflict(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint failed") ||
		strings.Contains(message, "duplicate entry") ||
		strings.Contains(message, "duplicate key")
}

// truncateTemplateVersionText 按字符截断字符串，避免超出字段长度
func truncateTemplateVersionText(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 || value == "" {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

// LogTemplateVersionError 记录版本操作错误但不影响主流程
func LogTemplateVersionError(action string, err error) {
	if err != nil {
		utils.Error("Template version %s failed: %v", action, err)
	}
}
