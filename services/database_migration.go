package services

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"

	"sublink/cache"
	"sublink/config"
	"sublink/database"
	"sublink/models"
	"sublink/services/mihomo"
	"sublink/services/scheduler"
	"sublink/services/telegram"
	"sublink/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var databaseMigrationRunning atomic.Bool

// DatabaseMigrationOptions 控制导入时是否包含可选数据。
type DatabaseMigrationOptions struct {
	IncludeSubLogs    bool `json:"includeSubLogs"`
	IncludeAccessKeys bool `json:"includeAccessKeys"`
}

// DatabaseMigrationResult 描述迁移结果，供任务中心和前端展示。
type DatabaseMigrationResult struct {
	SourceName        string         `json:"sourceName"`
	SourceType        string         `json:"sourceType"`
	Imported          map[string]int `json:"imported"`
	TemplatesRestored bool           `json:"templatesRestored"`
	Warnings          []string       `json:"warnings"`
	Skipped           []string       `json:"skipped"`
}

type databaseMigrationSource struct {
	DBPath      string
	TemplateDir string
	SourceType  string
	CleanupDir  string
}

type databaseMigrationState struct {
	tx                 *gorm.DB
	source             *gorm.DB
	sourceBundle       *databaseMigrationSource
	options            DatabaseMigrationOptions
	result             *DatabaseMigrationResult
	preservedSettings  map[string]string
	importedSettings   map[string]string
	importedSubs       []models.Subcription
	importedSubIDs     map[int]struct{}
	importedShareIDs   map[int]struct{}
	nodeNameToID       map[string]int
	sourceHasShares    bool
	sourceHasProfiles  bool
	sourceHasTemplates bool
}

// RunDatabaseMigrationTask 执行后台数据库迁移任务。
func RunDatabaseMigrationTask(ctx context.Context, taskID, uploadPath, originalName string, options DatabaseMigrationOptions) {
	tm := GetTaskManager()

	if databaseMigrationRunning.Swap(true) {
		_ = os.Remove(uploadPath)
		_ = tm.FailTask(taskID, "已有数据库迁移任务在运行，请稍后再试")
		return
	}
	defer databaseMigrationRunning.Store(false)
	defer func() {
		if err := os.Remove(uploadPath); err != nil && !os.IsNotExist(err) {
			utils.Warn("清理迁移上传文件失败: %v", err)
		}
	}()

	if err := ensureDatabaseMigrationCanRun(taskID); err != nil {
		_ = tm.FailTask(taskID, err.Error())
		return
	}

	result, err := executeDatabaseMigration(ctx, taskID, uploadPath, originalName, options)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.Warn("数据库迁移任务已取消: %s", taskID)
			return
		}
		_ = tm.FailTask(taskID, err.Error())
		return
	}

	message := "数据库迁移完成"
	if len(result.Warnings) > 0 {
		message = fmt.Sprintf("数据库迁移完成，但有 %d 条警告", len(result.Warnings))
	}
	_ = tm.CompleteTask(taskID, message, result)
}

func ensureDatabaseMigrationCanRun(currentTaskID string) error {
	for _, task := range GetTaskManager().GetRunningTasksInfo() {
		if task.ID != currentTaskID {
			return fmt.Errorf("存在进行中的任务 %s，请等待其完成后再迁移", task.Name)
		}
	}
	return nil
}

func executeDatabaseMigration(ctx context.Context, taskID, uploadPath, originalName string, options DatabaseMigrationOptions) (*DatabaseMigrationResult, error) {
	sourceBundle, err := prepareDatabaseMigrationSource(uploadPath)
	if err != nil {
		return nil, err
	}
	if sourceBundle.CleanupDir != "" {
		defer func() {
			if cleanupErr := os.RemoveAll(sourceBundle.CleanupDir); cleanupErr != nil {
				utils.Warn("清理迁移临时目录失败: %v", cleanupErr)
			}
		}()
	}

	sourceDB, err := gorm.Open(sqlite.Open(sourceBundle.DBPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("打开源 SQLite 数据库失败: %w", err)
	}

	result := &DatabaseMigrationResult{
		SourceName: originalName,
		SourceType: sourceBundle.SourceType,
		Imported:   make(map[string]int),
		Warnings:   make([]string, 0),
		Skipped:    make([]string, 0),
	}

	if sourceBundle.TemplateDir == "" {
		result.Warnings = append(result.Warnings, "未检测到模板目录；如果旧实例有自定义模板，请改用 backup.zip 导入")
	}
	if !options.IncludeAccessKeys {
		result.Skipped = append(result.Skipped, "access_keys")
	}
	if !options.IncludeSubLogs {
		result.Skipped = append(result.Skipped, "sub_logs")
	}

	steps := buildDatabaseMigrationSteps(sourceBundle, options)
	if err := GetTaskManager().UpdateTotal(taskID, len(steps)); err != nil {
		utils.Warn("更新数据库迁移任务总步骤失败: %v", err)
	}

	step := 0
	reportStep := func(label string, payload interface{}) error {
		if err := checkDatabaseMigrationContext(ctx); err != nil {
			return err
		}
		step++
		return GetTaskManager().UpdateProgress(taskID, step, label, payload)
	}

	state := &databaseMigrationState{
		source:             sourceDB,
		sourceBundle:       sourceBundle,
		options:            options,
		result:             result,
		importedSettings:   make(map[string]string),
		importedSubIDs:     make(map[int]struct{}),
		importedShareIDs:   make(map[int]struct{}),
		nodeNameToID:       make(map[string]int),
		sourceHasShares:    sourceDB.Migrator().HasTable(&models.SubscriptionShare{}),
		sourceHasProfiles:  sourceDB.Migrator().HasTable(&models.NodeCheckProfile{}),
		sourceHasTemplates: sourceDB.Migrator().HasTable(&models.Template{}),
	}

	if err := database.WithTransaction(func(tx *gorm.DB) error {
		state.tx = tx

		preservedSettings, err := loadPreservedTargetSettings(tx)
		if err != nil {
			return err
		}
		state.preservedSettings = preservedSettings

		if err := clearTargetBusinessData(tx); err != nil {
			return err
		}
		if err := reportStep("已清空目标业务数据", nil); err != nil {
			return err
		}

		if err := importUsers(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("用户导入完成，共 %d 条", state.result.Imported["users"]), nil); err != nil {
			return err
		}

		if err := importSystemSettings(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("系统设置导入完成，共 %d 条", state.result.Imported["system_settings"]), nil); err != nil {
			return err
		}

		if err := importScripts(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("脚本导入完成，共 %d 条", state.result.Imported["scripts"]), nil); err != nil {
			return err
		}

		if err := importTemplates(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("模板元数据导入完成，共 %d 条", state.result.Imported["templates"]), nil); err != nil {
			return err
		}

		if err := importNodes(state, ctx); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("节点导入完成，共 %d 条", state.result.Imported["nodes"]), nil); err != nil {
			return err
		}

		if err := importSubscriptions(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("订阅导入完成，共 %d 条", state.result.Imported["subcriptions"]), nil); err != nil {
			return err
		}

		if err := importSubscriptionGroups(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("订阅分组导入完成，共 %d 条", state.result.Imported["subcription_groups"]), nil); err != nil {
			return err
		}

		if err := importSubscriptionScripts(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("订阅脚本关联导入完成，共 %d 条", state.result.Imported["subcription_scripts"]), nil); err != nil {
			return err
		}

		if err := importSubscriptionNodes(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("订阅节点关联导入完成，共 %d 条", state.result.Imported["subcription_nodes"]), nil); err != nil {
			return err
		}

		if err := importTags(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("标签导入完成，共 %d 条", state.result.Imported["tags"]), nil); err != nil {
			return err
		}

		if err := importTagRules(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("标签规则导入完成，共 %d 条", state.result.Imported["tag_rules"]), nil); err != nil {
			return err
		}

		if err := importAccessKeys(state); err != nil {
			return err
		}
		if options.IncludeAccessKeys {
			if err := reportStep(fmt.Sprintf("AccessKey 导入完成，共 %d 条", state.result.Imported["access_keys"]), nil); err != nil {
				return err
			}
		}

		if err := importSubscriptionShares(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("分享链接导入完成，共 %d 条", state.result.Imported["subscription_shares"]), nil); err != nil {
			return err
		}

		if err := importSubLogs(state); err != nil {
			return err
		}
		if options.IncludeSubLogs {
			if err := reportStep(fmt.Sprintf("访问日志导入完成，共 %d 条", state.result.Imported["sub_logs"]), nil); err != nil {
				return err
			}
		}

		if err := importSubscriptionChainRules(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("链式代理规则导入完成，共 %d 条", state.result.Imported["subscription_chain_rules"]), nil); err != nil {
			return err
		}

		if err := importAirports(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("机场导入完成，共 %d 条", state.result.Imported["airports"]), nil); err != nil {
			return err
		}

		if err := importGroupAirportSorts(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("分组机场排序导入完成，共 %d 条", state.result.Imported["group_airport_sorts"]), nil); err != nil {
			return err
		}

		if err := importNodeCheckProfiles(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("节点检测策略导入完成，共 %d 条", state.result.Imported["node_check_profiles"]), nil); err != nil {
			return err
		}

		if err := importHosts(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Host 导入完成，共 %d 条", state.result.Imported["hosts"]), nil); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if err := reseedTargetSequences(database.DB); err != nil {
		return nil, err
	}
	if err := reportStep("自增序列已重置", nil); err != nil {
		return nil, err
	}

	if sourceBundle.TemplateDir != "" {
		if err := restoreTemplateDirectory(sourceBundle.TemplateDir); err != nil {
			result.Warnings = append(result.Warnings, "模板文件恢复失败: "+err.Error())
		} else {
			result.TemplatesRestored = true
		}
		label := "模板目录已恢复"
		if !result.TemplatesRestored {
			label = "模板目录恢复失败，详情见任务警告"
		}
		if err := reportStep(label, nil); err != nil {
			return nil, err
		}
	}

	if err := reloadRuntimeStateAfterMigration(); err != nil {
		return nil, fmt.Errorf("数据库数据已导入，但刷新运行时状态失败: %w；建议重启服务", err)
	}
	if err := reportStep("配置、缓存和定时任务已刷新", nil); err != nil {
		return nil, err
	}

	return result, nil
}

func buildDatabaseMigrationSteps(sourceBundle *databaseMigrationSource, options DatabaseMigrationOptions) []string {
	steps := []string{
		"清空目标业务数据",
		"导入用户",
		"导入系统设置",
		"导入脚本",
		"导入模板元数据",
		"导入节点",
		"导入订阅",
		"导入订阅分组",
		"导入订阅脚本关联",
		"导入订阅节点关联",
		"导入标签",
		"导入标签规则",
	}

	if options.IncludeAccessKeys {
		steps = append(steps, "导入AccessKey")
	}

	steps = append(steps, "导入分享链接")

	if options.IncludeSubLogs {
		steps = append(steps, "导入访问日志")
	}

	steps = append(steps,
		"导入链式代理规则",
		"导入机场",
		"导入分组机场排序",
		"导入节点检测策略",
		"导入Host",
		"重置自增序列",
	)

	if sourceBundle.TemplateDir != "" {
		steps = append(steps, "恢复模板目录")
	}

	steps = append(steps, "刷新运行时状态")
	return steps
}

func prepareDatabaseMigrationSource(uploadPath string) (*databaseMigrationSource, error) {
	switch {
	case looksLikeZipFile(uploadPath):
		extractDir, err := extractMigrationZip(uploadPath)
		if err != nil {
			return nil, err
		}
		dbPath, err := findSQLiteDatabaseFile(filepath.Join(extractDir, "db"))
		if err != nil {
			_ = os.RemoveAll(extractDir)
			return nil, err
		}

		templateDir := filepath.Join(extractDir, "template")
		if _, err := os.Stat(templateDir); err != nil {
			templateDir = ""
		}

		return &databaseMigrationSource{
			DBPath:      dbPath,
			TemplateDir: templateDir,
			SourceType:  "backup_zip",
			CleanupDir:  extractDir,
		}, nil
	case looksLikeSQLiteFile(uploadPath):
		return &databaseMigrationSource{
			DBPath:     uploadPath,
			SourceType: "sqlite_db",
		}, nil
	default:
		return nil, fmt.Errorf("仅支持 SQLite 数据库文件或旧实例导出的 backup.zip")
	}
}

func looksLikeZipFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	header := make([]byte, 4)
	if _, err := io.ReadFull(file, header); err != nil {
		return false
	}
	return string(header) == "PK\x03\x04"
}

func looksLikeSQLiteFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	header := make([]byte, 16)
	if _, err := io.ReadFull(file, header); err != nil {
		return false
	}
	return string(header) == "SQLite format 3\x00"
}

func extractMigrationZip(zipPath string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("读取迁移压缩包失败: %w", err)
	}
	defer reader.Close()

	tempRoot, err := ensureDatabaseMigrationTempRoot()
	if err != nil {
		return "", err
	}

	tempDir, err := os.MkdirTemp(tempRoot, "bundle-*")
	if err != nil {
		return "", fmt.Errorf("创建迁移临时目录失败: %w", err)
	}

	for _, file := range reader.File {
		cleanName := filepath.Clean(filepath.FromSlash(file.Name))
		if cleanName == "." || strings.HasPrefix(cleanName, "..") {
			return "", fmt.Errorf("迁移压缩包包含非法路径: %s", file.Name)
		}
		targetPath := filepath.Join(tempDir, cleanName)
		if !strings.HasPrefix(targetPath, tempDir+string(os.PathSeparator)) && targetPath != tempDir {
			return "", fmt.Errorf("迁移压缩包包含越界路径: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return "", fmt.Errorf("创建迁移目录失败: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return "", fmt.Errorf("创建迁移目录失败: %w", err)
		}

		src, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("读取迁移压缩包文件失败: %w", err)
		}

		dst, err := os.Create(targetPath)
		if err != nil {
			src.Close()
			return "", fmt.Errorf("创建迁移临时文件失败: %w", err)
		}

		if _, err := io.Copy(dst, src); err != nil {
			dst.Close()
			src.Close()
			return "", fmt.Errorf("解压迁移文件失败: %w", err)
		}

		dst.Close()
		src.Close()
	}

	return tempDir, nil
}

func ensureDatabaseMigrationTempRoot() (string, error) {
	tempRoot := filepath.Join(config.GetDBPath(), ".tmp", "database-migration")
	if err := os.MkdirAll(tempRoot, 0755); err != nil {
		return "", fmt.Errorf("创建迁移临时目录失败: %w", err)
	}
	return tempRoot, nil
}

func findSQLiteDatabaseFile(root string) (string, error) {
	if _, err := os.Stat(root); err != nil {
		return "", fmt.Errorf("压缩包中未找到 db 目录")
	}

	candidates := make([]string, 0)
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if ext == ".db" || ext == ".sqlite" || ext == ".sqlite3" {
			candidates = append(candidates, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("压缩包中未找到 SQLite 数据库文件")
	}
	return candidates[0], nil
}

func checkDatabaseMigrationContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return context.Canceled
	default:
		return nil
	}
}

func loadPreservedTargetSettings(tx *gorm.DB) (map[string]string, error) {
	result := make(map[string]string)
	for _, key := range []string{"jwt_secret"} {
		var setting models.SystemSetting
		err := tx.Where(map[string]interface{}{"key": key}).Take(&setting).Error
		if err == nil {
			result[key] = setting.Value
			continue
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		return nil, fmt.Errorf("读取目标保留设置失败(%s): %w", key, err)
	}
	return result, nil
}

func clearTargetBusinessData(tx *gorm.DB) error {
	modelsInDeleteOrder := []interface{}{
		&models.IPInfo{},
		&models.SubLogs{},
		&models.SubscriptionShare{},
		&models.SubscriptionChainRule{},
		&models.GroupAirportSort{},
		&models.SubcriptionNode{},
		&models.SubcriptionScript{},
		&models.SubcriptionGroup{},
		&models.TagRule{},
		&models.AccessKey{},
		&models.NodeCheckProfile{},
		&models.Host{},
		&models.Template{},
		&models.Script{},
		&models.Tag{},
		&models.Airport{},
		&models.Subcription{},
		&models.Node{},
		&models.SystemSetting{},
		&models.User{},
	}

	for _, model := range modelsInDeleteOrder {
		table, err := tableName(tx, model)
		if err != nil {
			return err
		}
		if !tx.Migrator().HasTable(table) {
			continue
		}
		if err := tx.Exec("DELETE FROM " + quoteMigrationIdentifier(table)).Error; err != nil {
			return fmt.Errorf("清空数据表 %s 失败: %w", table, err)
		}
	}

	return nil
}

func importUsers(state *databaseMigrationState) error {
	if !state.source.Migrator().HasTable(&models.User{}) {
		return nil
	}
	var users []models.User
	if err := state.source.Find(&users).Error; err != nil {
		return fmt.Errorf("读取源用户失败: %w", err)
	}
	if err := insertRecords(state.tx, users); err != nil {
		return fmt.Errorf("导入用户失败: %w", err)
	}
	state.result.Imported["users"] = len(users)
	return nil
}

func importSystemSettings(state *databaseMigrationState) error {
	settings := make([]models.SystemSetting, 0)
	if state.source.Migrator().HasTable(&models.SystemSetting{}) {
		if err := state.source.Find(&settings).Error; err != nil {
			return fmt.Errorf("读取源系统设置失败: %w", err)
		}
	}

	filtered := make([]models.SystemSetting, 0, len(settings))
	for _, setting := range settings {
		if setting.Key == "jwt_secret" {
			continue
		}
		if len(setting.Key) > 191 {
			return fmt.Errorf("系统设置键 %q 长度超过 191，无法迁移", setting.Key)
		}
		filtered = append(filtered, setting)
		state.importedSettings[setting.Key] = setting.Value
	}

	for key, value := range state.preservedSettings {
		filtered = append(filtered, models.SystemSetting{Key: key, Value: value})
		state.importedSettings[key] = value
	}

	if err := insertRecords(state.tx, filtered); err != nil {
		return fmt.Errorf("导入系统设置失败: %w", err)
	}
	state.result.Imported["system_settings"] = len(filtered)
	return nil
}

func importScripts(state *databaseMigrationState) error {
	if !state.source.Migrator().HasTable(&models.Script{}) {
		return nil
	}
	var scripts []models.Script
	if err := state.source.Find(&scripts).Error; err != nil {
		return fmt.Errorf("读取源脚本失败: %w", err)
	}
	if err := insertRecords(state.tx, scripts); err != nil {
		return fmt.Errorf("导入脚本失败: %w", err)
	}
	state.result.Imported["scripts"] = len(scripts)
	return nil
}

func importTemplates(state *databaseMigrationState) error {
	templates := make([]models.Template, 0)
	if state.sourceHasTemplates {
		if err := state.source.Find(&templates).Error; err != nil {
			return fmt.Errorf("读取源模板元数据失败: %w", err)
		}
	} else if state.sourceBundle.TemplateDir != "" {
		var err error
		templates, err = buildTemplateMetadataFromDir(state.sourceBundle.TemplateDir)
		if err != nil {
			return err
		}
		if len(templates) > 0 {
			state.result.Warnings = append(state.result.Warnings, "源备份未包含模板元数据表，已根据模板文件重建模板记录")
		}
	}

	for _, template := range templates {
		if len(template.Name) > 191 {
			return fmt.Errorf("模板文件名 %q 长度超过 191，无法迁移", template.Name)
		}
	}

	if err := insertRecords(state.tx, templates); err != nil {
		return fmt.Errorf("导入模板元数据失败: %w", err)
	}
	state.result.Imported["templates"] = len(templates)
	return nil
}

func importNodes(state *databaseMigrationState, ctx context.Context) error {
	if !state.source.Migrator().HasTable(&models.Node{}) {
		return nil
	}
	var nodes []models.Node
	if err := state.source.Find(&nodes).Error; err != nil {
		return fmt.Errorf("读取源节点失败: %w", err)
	}

	for i := range nodes {
		if err := checkDatabaseMigrationContext(ctx); err != nil {
			return err
		}
		models.NormalizeNodeForImport(&nodes[i])
		if nodes[i].Name != "" {
			if _, exists := state.nodeNameToID[nodes[i].Name]; !exists {
				state.nodeNameToID[nodes[i].Name] = nodes[i].ID
			}
		}
	}

	if err := insertRecords(state.tx, nodes); err != nil {
		return fmt.Errorf("导入节点失败: %w", err)
	}
	state.result.Imported["nodes"] = len(nodes)
	return nil
}

func importSubscriptions(state *databaseMigrationState) error {
	if !state.source.Migrator().HasTable(&models.Subcription{}) {
		return nil
	}
	var subs []models.Subcription
	if err := state.source.Unscoped().Find(&subs).Error; err != nil {
		return fmt.Errorf("读取源订阅失败: %w", err)
	}
	if err := insertRecords(state.tx, subs); err != nil {
		return fmt.Errorf("导入订阅失败: %w", err)
	}
	state.importedSubs = subs
	for _, sub := range subs {
		state.importedSubIDs[sub.ID] = struct{}{}
	}
	state.result.Imported["subcriptions"] = len(subs)
	return nil
}

func importSubscriptionGroups(state *databaseMigrationState) error {
	if !state.source.Migrator().HasTable(&models.SubcriptionGroup{}) {
		return nil
	}
	var groups []models.SubcriptionGroup
	if err := state.source.Find(&groups).Error; err != nil {
		return fmt.Errorf("读取源订阅分组失败: %w", err)
	}
	for _, group := range groups {
		if len(group.GroupName) > 191 {
			return fmt.Errorf("分组名称 %q 长度超过 191，无法迁移", group.GroupName)
		}
	}
	if err := insertRecords(state.tx, groups); err != nil {
		return fmt.Errorf("导入订阅分组失败: %w", err)
	}
	state.result.Imported["subcription_groups"] = len(groups)
	return nil
}

func importSubscriptionScripts(state *databaseMigrationState) error {
	if !state.source.Migrator().HasTable(&models.SubcriptionScript{}) {
		return nil
	}
	var records []models.SubcriptionScript
	if err := state.source.Find(&records).Error; err != nil {
		return fmt.Errorf("读取源订阅脚本关联失败: %w", err)
	}
	if err := insertRecords(state.tx, records); err != nil {
		return fmt.Errorf("导入订阅脚本关联失败: %w", err)
	}
	state.result.Imported["subcription_scripts"] = len(records)
	return nil
}

func importSubscriptionNodes(state *databaseMigrationState) error {
	if !state.source.Migrator().HasTable(&models.SubcriptionNode{}) && !state.source.Migrator().HasTable("subcription_nodes") {
		return nil
	}

	var records []models.SubcriptionNode
	if state.source.Migrator().HasColumn("subcription_nodes", "node_id") {
		if err := state.source.Table("subcription_nodes").Find(&records).Error; err != nil {
			return fmt.Errorf("读取源订阅节点关联失败: %w", err)
		}
	} else if state.source.Migrator().HasColumn("subcription_nodes", "node_name") {
		var legacyRecords []struct {
			SubcriptionID int
			NodeName      string
			Sort          int
		}
		if err := state.source.Table("subcription_nodes").Find(&legacyRecords).Error; err != nil {
			return fmt.Errorf("读取源旧版订阅节点关联失败: %w", err)
		}
		for _, legacy := range legacyRecords {
			nodeID := state.nodeNameToID[legacy.NodeName]
			if nodeID == 0 {
				continue
			}
			records = append(records, models.SubcriptionNode{
				SubcriptionID: legacy.SubcriptionID,
				NodeID:        nodeID,
				Sort:          legacy.Sort,
			})
		}
		if len(legacyRecords) > len(records) {
			state.result.Warnings = append(state.result.Warnings, fmt.Sprintf("有 %d 条旧版订阅节点关联因找不到对应节点而被跳过", len(legacyRecords)-len(records)))
		}
	} else {
		return nil
	}

	if err := insertRecords(state.tx, records); err != nil {
		return fmt.Errorf("导入订阅节点关联失败: %w", err)
	}
	state.result.Imported["subcription_nodes"] = len(records)
	return nil
}

func importTags(state *databaseMigrationState) error {
	if !state.source.Migrator().HasTable(&models.Tag{}) {
		return nil
	}
	var tags []models.Tag
	if err := state.source.Find(&tags).Error; err != nil {
		return fmt.Errorf("读取源标签失败: %w", err)
	}
	if err := insertRecords(state.tx, tags); err != nil {
		return fmt.Errorf("导入标签失败: %w", err)
	}
	state.result.Imported["tags"] = len(tags)
	return nil
}

func importTagRules(state *databaseMigrationState) error {
	if !state.source.Migrator().HasTable(&models.TagRule{}) {
		return nil
	}
	var rules []models.TagRule
	if err := state.source.Find(&rules).Error; err != nil {
		return fmt.Errorf("读取源标签规则失败: %w", err)
	}
	if err := insertRecords(state.tx, rules); err != nil {
		return fmt.Errorf("导入标签规则失败: %w", err)
	}
	state.result.Imported["tag_rules"] = len(rules)
	return nil
}

func importAccessKeys(state *databaseMigrationState) error {
	if !state.options.IncludeAccessKeys {
		return nil
	}
	if !state.source.Migrator().HasTable(&models.AccessKey{}) {
		return nil
	}
	var keys []models.AccessKey
	if err := state.source.Find(&keys).Error; err != nil {
		return fmt.Errorf("读取源 AccessKey 失败: %w", err)
	}
	if err := insertRecords(state.tx, keys); err != nil {
		return fmt.Errorf("导入 AccessKey 失败: %w", err)
	}
	if len(keys) > 0 {
		state.result.Warnings = append(state.result.Warnings, "已导入 AccessKey。如旧 API Key 无法使用，请确认新实例的 API 加密密钥与旧实例一致，或手动重新生成")
	}
	state.result.Imported["access_keys"] = len(keys)
	return nil
}

func importSubscriptionShares(state *databaseMigrationState) error {
	shares := make([]models.SubscriptionShare, 0)
	if state.sourceHasShares {
		var sourceShares []models.SubscriptionShare
		if err := state.source.Find(&sourceShares).Error; err != nil {
			return fmt.Errorf("读取源分享链接失败: %w", err)
		}
		skippedOrphans := 0
		for i := range sourceShares {
			if _, ok := state.importedSubIDs[sourceShares[i].SubscriptionID]; !ok {
				skippedOrphans++
				continue
			}
			normalizeImportedSubscriptionShare(&sourceShares[i])
			if sourceShares[i].Token == "" {
				token, err := models.GenerateToken()
				if err != nil {
					return fmt.Errorf("生成分享 Token 失败: %w", err)
				}
				sourceShares[i].Token = token
			}
			shares = append(shares, sourceShares[i])
		}
		if skippedOrphans > 0 {
			state.result.Warnings = append(state.result.Warnings, fmt.Sprintf("有 %d 条分享链接因引用不存在的订阅而被跳过", skippedOrphans))
		}
	} else {
		for _, sub := range state.importedSubs {
			token, err := models.GenerateToken()
			if err != nil {
				return fmt.Errorf("生成默认分享 Token 失败: %w", err)
			}
			share := models.SubscriptionShare{
				SubscriptionID: sub.ID,
				Token:          token,
				Name:           "默认分享链接",
				ExpireType:     models.ExpireTypeNever,
				IsLegacy:       true,
				Enabled:        true,
			}
			normalizeImportedSubscriptionShare(&share)
			shares = append(shares, share)
		}
		if len(shares) > 0 {
			state.result.Warnings = append(state.result.Warnings, "源备份未包含分享表，已为每个订阅生成默认分享链接")
		}
	}

	if err := insertRecords(state.tx, shares); err != nil {
		return fmt.Errorf("导入分享链接失败: %w", err)
	}
	for _, share := range shares {
		state.importedShareIDs[share.ID] = struct{}{}
	}
	state.result.Imported["subscription_shares"] = len(shares)
	return nil
}

func importSubLogs(state *databaseMigrationState) error {
	if !state.options.IncludeSubLogs {
		return nil
	}
	if !state.source.Migrator().HasTable(&models.SubLogs{}) {
		return nil
	}
	var logs []models.SubLogs
	if err := state.source.Find(&logs).Error; err != nil {
		return fmt.Errorf("读取源访问日志失败: %w", err)
	}
	filteredLogs := make([]models.SubLogs, 0, len(logs))
	skippedOrphans := 0
	normalizedShareRefs := 0
	for i := range logs {
		if _, ok := state.importedSubIDs[logs[i].SubcriptionID]; !ok {
			skippedOrphans++
			continue
		}
		if logs[i].ShareID > 0 {
			if _, ok := state.importedShareIDs[logs[i].ShareID]; !ok {
				logs[i].ShareID = 0
				normalizedShareRefs++
			}
		}
		filteredLogs = append(filteredLogs, logs[i])
	}
	if skippedOrphans > 0 {
		state.result.Warnings = append(state.result.Warnings, fmt.Sprintf("有 %d 条访问日志因引用不存在的订阅而被跳过", skippedOrphans))
	}
	if normalizedShareRefs > 0 {
		state.result.Warnings = append(state.result.Warnings, fmt.Sprintf("有 %d 条访问日志因引用不存在的分享链接而被重置为订阅级日志", normalizedShareRefs))
	}
	if err := insertRecords(state.tx, filteredLogs); err != nil {
		return fmt.Errorf("导入访问日志失败: %w", err)
	}
	state.result.Imported["sub_logs"] = len(filteredLogs)
	return nil
}

func importSubscriptionChainRules(state *databaseMigrationState) error {
	if !state.source.Migrator().HasTable(&models.SubscriptionChainRule{}) {
		return nil
	}
	var rules []models.SubscriptionChainRule
	if err := state.source.Find(&rules).Error; err != nil {
		return fmt.Errorf("读取源链式代理规则失败: %w", err)
	}
	if err := insertRecords(state.tx, rules); err != nil {
		return fmt.Errorf("导入链式代理规则失败: %w", err)
	}
	state.result.Imported["subscription_chain_rules"] = len(rules)
	return nil
}

func importAirports(state *databaseMigrationState) error {
	airports := make([]models.Airport, 0)

	if state.source.Migrator().HasTable(&models.Airport{}) {
		if err := state.source.Find(&airports).Error; err != nil {
			return fmt.Errorf("读取源机场失败: %w", err)
		}
	}

	if len(airports) == 0 && state.source.Migrator().HasTable(&models.SubScheduler{}) {
		var schedulers []models.SubScheduler
		if err := state.source.Find(&schedulers).Error; err != nil {
			return fmt.Errorf("读取源旧版订阅调度失败: %w", err)
		}
		for _, schedulerItem := range schedulers {
			airports = append(airports, models.Airport{
				ID:                schedulerItem.ID,
				Name:              schedulerItem.Name,
				URL:               schedulerItem.URL,
				CronExpr:          schedulerItem.CronExpr,
				Enabled:           schedulerItem.Enabled,
				SuccessCount:      schedulerItem.SuccessCount,
				LastRunTime:       schedulerItem.LastRunTime,
				NextRunTime:       schedulerItem.NextRunTime,
				CreatedAt:         schedulerItem.CreatedAt,
				UpdatedAt:         schedulerItem.UpdatedAt,
				Group:             schedulerItem.Group,
				DownloadWithProxy: schedulerItem.DownloadWithProxy,
				ProxyLink:         schedulerItem.ProxyLink,
				UserAgent:         schedulerItem.UserAgent,
			})
		}
		if len(airports) > 0 {
			state.result.Warnings = append(state.result.Warnings, "源备份使用旧版订阅调度表，已自动转换为机场数据")
		}
	}

	if err := insertRecords(state.tx, airports); err != nil {
		return fmt.Errorf("导入机场失败: %w", err)
	}
	state.result.Imported["airports"] = len(airports)
	return nil
}

func importGroupAirportSorts(state *databaseMigrationState) error {
	if !state.source.Migrator().HasTable(&models.GroupAirportSort{}) {
		return nil
	}
	var records []models.GroupAirportSort
	if err := state.source.Find(&records).Error; err != nil {
		return fmt.Errorf("读取源分组机场排序失败: %w", err)
	}
	if err := insertRecords(state.tx, records); err != nil {
		return fmt.Errorf("导入分组机场排序失败: %w", err)
	}
	state.result.Imported["group_airport_sorts"] = len(records)
	return nil
}

func importNodeCheckProfiles(state *databaseMigrationState) error {
	profiles := make([]models.NodeCheckProfile, 0)
	if state.sourceHasProfiles {
		if err := state.source.Find(&profiles).Error; err != nil {
			return fmt.Errorf("读取源节点检测策略失败: %w", err)
		}
	}

	for _, profile := range profiles {
		if len(profile.Name) > 191 {
			return fmt.Errorf("节点检测策略名称 %q 长度超过 191，无法迁移", profile.Name)
		}
	}

	if len(profiles) == 0 {
		profiles = append(profiles, buildDefaultNodeCheckProfile(state.importedSettings))
		if !state.sourceHasProfiles {
			state.result.Warnings = append(state.result.Warnings, "源备份未包含节点检测策略，已根据现有测速配置生成默认策略")
		}
	}

	if err := insertRecords(state.tx, profiles); err != nil {
		return fmt.Errorf("导入节点检测策略失败: %w", err)
	}
	state.result.Imported["node_check_profiles"] = len(profiles)
	return nil
}

func importHosts(state *databaseMigrationState) error {
	if !state.source.Migrator().HasTable(&models.Host{}) {
		return nil
	}
	var hosts []models.Host
	if err := state.source.Find(&hosts).Error; err != nil {
		return fmt.Errorf("读取源 Host 失败: %w", err)
	}
	if err := insertRecords(state.tx, hosts); err != nil {
		return fmt.Errorf("导入 Host 失败: %w", err)
	}
	state.result.Imported["hosts"] = len(hosts)
	return nil
}

func insertRecords[T any](tx *gorm.DB, records []T) error {
	if len(records) == 0 {
		return nil
	}
	return tx.Session(&gorm.Session{SkipHooks: true}).Omit(clause.Associations).CreateInBatches(&records, database.BatchSize).Error
}

func buildTemplateMetadataFromDir(templateDir string) ([]models.Template, error) {
	entries, err := os.ReadDir(templateDir)
	if err != nil {
		return nil, fmt.Errorf("读取模板目录失败: %w", err)
	}

	templates := make([]models.Template, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		templates = append(templates, models.Template{
			Name:       entry.Name(),
			Category:   models.InferTemplateCategory(entry.Name()),
			RuleSource: "",
		})
	}
	return templates, nil
}

func normalizeImportedSubscriptionShare(share *models.SubscriptionShare) {
	if share == nil {
		return
	}

	if share.ExpireType != models.ExpireTypeDateTime || share.ExpireAt == nil || share.ExpireAt.IsZero() {
		share.ExpireAt = nil
	}
	if share.AccessCount <= 0 || share.LastAccessAt == nil || share.LastAccessAt.IsZero() {
		share.LastAccessAt = nil
	}
}

func buildDefaultNodeCheckProfile(settings map[string]string) models.NodeCheckProfile {
	timeout := parsePositiveInt(settings["speed_test_timeout"], 5)
	latencyConcurrency := parseInt(settings["speed_test_latency_concurrency"], 0)
	if latencyConcurrency == 0 && settings["speed_test_latency_concurrency"] == "" {
		latencyConcurrency = parseInt(settings["speed_test_concurrency"], 0)
	}
	speedConcurrency := parsePositiveInt(settings["speed_test_speed_concurrency"], 1)
	peakSampleInterval := parseRangedInt(settings["speed_test_peak_sample_interval"], 100, 50, 200)

	mode := strings.TrimSpace(settings["speed_test_mode"])
	if mode == "" {
		mode = "tcp"
	}

	speedRecordMode := strings.TrimSpace(settings["speed_test_speed_record_mode"])
	if speedRecordMode == "" {
		speedRecordMode = "average"
	}

	return models.NodeCheckProfile{
		Name:               "默认策略",
		Enabled:            parseBool(settings["speed_test_enabled"], false),
		CronExpr:           strings.TrimSpace(settings["speed_test_cron"]),
		Mode:               mode,
		TestURL:            strings.TrimSpace(settings["speed_test_url"]),
		LatencyURL:         strings.TrimSpace(settings["speed_test_latency_url"]),
		Timeout:            timeout,
		Groups:             strings.TrimSpace(settings["speed_test_groups"]),
		Tags:               strings.TrimSpace(settings["speed_test_tags"]),
		LatencyConcurrency: latencyConcurrency,
		SpeedConcurrency:   speedConcurrency,
		DetectCountry:      parseBool(settings["speed_test_detect_country"], false),
		LandingIPURL:       strings.TrimSpace(settings["speed_test_landing_ip_url"]),
		IncludeHandshake:   parseBool(settings["speed_test_include_handshake"], true),
		SpeedRecordMode:    speedRecordMode,
		PeakSampleInterval: peakSampleInterval,
		TrafficByGroup:     parseBool(settings["speed_test_traffic_by_group"], true),
		TrafficBySource:    parseBool(settings["speed_test_traffic_by_source"], true),
		TrafficByNode:      parseBool(settings["speed_test_traffic_by_node"], false),
	}
}

func parseBool(raw string, defaultValue bool) bool {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return defaultValue
	}
	return raw == "1" || raw == "true" || raw == "yes" || raw == "on"
}

func parseInt(raw string, defaultValue int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}
	return value
}

func parsePositiveInt(raw string, defaultValue int) int {
	value := parseInt(raw, defaultValue)
	if value <= 0 {
		return defaultValue
	}
	return value
}

func parseRangedInt(raw string, defaultValue, minValue, maxValue int) int {
	value := parseInt(raw, defaultValue)
	if value < minValue || value > maxValue {
		return defaultValue
	}
	return value
}

func restoreTemplateDirectory(sourceDir string) error {
	targetDir := templateDirPath()
	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("清理目标模板目录失败: %w", err)
	}

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(targetDir, relativePath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		return copyFile(path, targetPath)
	})
}

func copyFile(sourcePath, targetPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		return err
	}
	return targetFile.Sync()
}

func templateDirPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "template"
	}
	return filepath.Join(cwd, "template")
}

func reloadRuntimeStateAfterMigration() error {
	config.Load()

	cache.InvalidateAllTemplateContent()

	initializers := []func() error{
		models.InitNodeCache,
		models.InitSettingCache,
		models.InitUserCache,
		models.InitScriptCache,
		models.InitAirportCache,
		models.InitGroupAirportSortCache,
		models.InitAccessKeyCache,
		models.InitNodeCheckProfileCache,
		models.InitSubLogsCache,
		models.InitSubcriptionCache,
		models.InitTemplateCache,
		models.InitTagCache,
		models.InitTagRuleCache,
		models.InitTaskCache,
		models.InitIPInfoCache,
		models.InitHostCache,
		models.InitSubscriptionShareCache,
		models.InitChainRuleCache,
	}

	for _, initializer := range initializers {
		if err := initializer(); err != nil {
			return err
		}
	}

	cache.InitTemplateContentCache()
	utils.SetTagGroupTagsFunc(models.GetTagNamesByGroupName)

	if !models.IsDemoMode() {
		if err := scheduler.GetSchedulerManager().ReloadFromDatabase(); err != nil {
			return err
		}
	}

	if err := mihomo.SyncHostsFromDB(); err != nil {
		utils.Warn("迁移后同步 Host 到 mihomo 失败: %v", err)
	}

	telegramConfig, err := telegram.LoadConfig()
	if err == nil {
		if telegramConfig.Enabled && telegramConfig.BotToken != "" {
			if err := telegram.Reconnect(); err != nil {
				utils.Warn("迁移后重连 Telegram 机器人失败: %v", err)
			}
		} else {
			telegram.StopBot()
		}
	}

	return nil
}

func reseedTargetSequences(tx *gorm.DB) error {
	switch {
	case database.IsMySQL():
		return reseedMySQLAutoIncrement(tx)
	case database.IsPostgres():
		return reseedPostgresSequences(tx)
	default:
		return nil
	}
}

func reseedMySQLAutoIncrement(tx *gorm.DB) error {
	for _, table := range sequenceTables(tx) {
		var maxID int64
		query := fmt.Sprintf("SELECT COALESCE(MAX(id), 0) FROM %s", quoteMigrationIdentifier(table))
		if err := tx.Raw(query).Scan(&maxID).Error; err != nil {
			return fmt.Errorf("读取表 %s 最大ID失败: %w", table, err)
		}
		resetSQL := fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = %d", quoteMigrationIdentifier(table), maxID+1)
		if err := tx.Exec(resetSQL).Error; err != nil {
			return fmt.Errorf("重置表 %s 自增值失败: %w", table, err)
		}
	}
	return nil
}

func reseedPostgresSequences(tx *gorm.DB) error {
	for _, table := range sequenceTables(tx) {
		var maxID int64
		query := fmt.Sprintf("SELECT COALESCE(MAX(id), 0) FROM %s", quoteMigrationIdentifier(table))
		if err := tx.Raw(query).Scan(&maxID).Error; err != nil {
			return fmt.Errorf("读取表 %s 最大ID失败: %w", table, err)
		}

		var sequenceName string
		if err := tx.Raw("SELECT pg_get_serial_sequence(?, 'id')", table).Scan(&sequenceName).Error; err != nil {
			return fmt.Errorf("查询表 %s 序列失败: %w", table, err)
		}
		if sequenceName == "" {
			continue
		}

		if maxID > 0 {
			if err := tx.Exec("SELECT setval(?::regclass, ?, true)", sequenceName, maxID).Error; err != nil {
				return fmt.Errorf("重置表 %s 序列失败: %w", table, err)
			}
			continue
		}

		if err := tx.Exec("SELECT setval(?::regclass, 1, false)", sequenceName).Error; err != nil {
			return fmt.Errorf("重置空表 %s 序列失败: %w", table, err)
		}
	}
	return nil
}

func sequenceTables(tx *gorm.DB) []string {
	modelsWithSequence := []interface{}{
		&models.User{},
		&models.Subcription{},
		&models.Node{},
		&models.SubLogs{},
		&models.AccessKey{},
		&models.Script{},
		&models.Template{},
		&models.TagRule{},
		&models.IPInfo{},
		&models.Host{},
		&models.SubscriptionShare{},
		&models.SubscriptionChainRule{},
		&models.Airport{},
		&models.GroupAirportSort{},
		&models.NodeCheckProfile{},
	}

	tables := make([]string, 0, len(modelsWithSequence))
	for _, model := range modelsWithSequence {
		table, err := tableName(tx, model)
		if err != nil || !tx.Migrator().HasTable(table) {
			continue
		}
		tables = append(tables, table)
	}
	return tables
}

func tableName(db *gorm.DB, model interface{}) (string, error) {
	statement := &gorm.Statement{DB: db}
	if err := statement.Parse(model); err != nil {
		return "", err
	}
	return statement.Schema.Table, nil
}

func quoteMigrationIdentifier(identifier string) string {
	switch {
	case database.IsMySQL():
		return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
	case database.IsPostgres():
		return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
	default:
		return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
	}
}
