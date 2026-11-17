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
	"time"

	"sublink/cache"
	"sublink/config"
	"sublink/database"
	"sublink/models"
	backupservice "sublink/services/backup"
	"sublink/services/mihomo"
	"sublink/services/scheduler"
	socks5service "sublink/services/socks5"
	"sublink/services/telegram"
	"sublink/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	maxMigrationZipEntries     = 10000
	maxMigrationFileBytes      = int64(2 << 30)
	maxMigrationExtractedBytes = int64(4 << 30)
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
	ambiguousNodeNames map[string]struct{}
	sourceHasShares    bool
	sourceHasProfiles  bool
	sourceHasWebhooks  bool
	sourceHasTemplates bool
}

// RunDatabaseMigrationTask 执行后台数据库迁移任务。
func RunDatabaseMigrationTask(ctx context.Context, taskID, uploadPath, originalName string, options DatabaseMigrationOptions) {
	tm := GetTaskManager()

	if databaseMigrationRunning.Swap(true) {
		_ = os.Remove(uploadPath)
		_ = tm.FailTask(taskID, "A database migration task is already running; try again later")
		return
	}
	defer databaseMigrationRunning.Store(false)
	defer func() {
		if err := os.Remove(uploadPath); err != nil && !os.IsNotExist(err) {
			utils.Warn("Failed to clean up migration upload file: %v", err)
		}
	}()

	if err := ensureDatabaseMigrationCanRun(taskID); err != nil {
		_ = tm.FailTask(taskID, err.Error())
		return
	}

	result, err := executeDatabaseMigration(ctx, taskID, uploadPath, originalName, options)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.Warn("Database migration task cancelled: %s", taskID)
			return
		}
		_ = tm.FailTask(taskID, err.Error())
		return
	}

	message := "Database migration completed"
	if len(result.Warnings) > 0 {
		message = fmt.Sprintf("Database migration completed with %d warnings", len(result.Warnings))
	}
	_ = tm.CompleteTask(taskID, message, result)
}

func ensureDatabaseMigrationCanRun(currentTaskID string) error {
	for _, task := range GetTaskManager().GetRunningTasksInfo() {
		if task.ID != currentTaskID {
			return fmt.Errorf("task %s is still running; wait for it to finish before migrating", task.Name)
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
				utils.Warn("Failed to clean up migration temp directory: %v", cleanupErr)
			}
		}()
	}

	sourceDB, err := database.OpenMigrationSourceSQLite(sourceBundle.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source SQLite database: %w", err)
	}
	sourceSQLDB, err := sourceDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get source SQLite connection: %w", err)
	}
	defer func() {
		if closeErr := sourceSQLDB.Close(); closeErr != nil {
			utils.Warn("Failed to close source SQLite connection: %v", closeErr)
		}
	}()

	result := &DatabaseMigrationResult{
		SourceName: originalName,
		SourceType: sourceBundle.SourceType,
		Imported:   make(map[string]int),
		Warnings:   make([]string, 0),
		Skipped:    make([]string, 0),
	}

	if sourceBundle.TemplateDir == "" {
		result.Warnings = append(result.Warnings, "No template directory detected; if the old instance had custom templates, import via backup.zip instead")
	}
	if !options.IncludeAccessKeys {
		result.Skipped = append(result.Skipped, "access_keys")
	}
	if !options.IncludeSubLogs {
		result.Skipped = append(result.Skipped, "sub_logs")
	}

	steps := buildDatabaseMigrationSteps(sourceBundle, options)
	if err := GetTaskManager().UpdateTotal(taskID, len(steps)); err != nil {
		utils.Warn("Failed to update total steps of the database migration task: %v", err)
	}

	step := 0
	reportStep := func(label string, payload any) error {
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
		ambiguousNodeNames: make(map[string]struct{}),
		sourceHasShares:    sourceDB.Migrator().HasTable(&models.SubscriptionShare{}),
		sourceHasProfiles:  sourceDB.Migrator().HasTable(&models.NodeCheckProfile{}),
		sourceHasWebhooks:  sourceDB.Migrator().HasTable(&models.Webhook{}),
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
		if err := reportStep("Target business data cleared", nil); err != nil {
			return err
		}

		if err := importUsers(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Users imported: %d", state.result.Imported["users"]), nil); err != nil {
			return err
		}

		if err := importSystemSettings(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("System settings imported: %d", state.result.Imported["system_settings"]), nil); err != nil {
			return err
		}

		if err := importWebhooks(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Webhooks imported: %d", state.result.Imported["webhooks"]), nil); err != nil {
			return err
		}

		if err := importScripts(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Scripts imported: %d", state.result.Imported["scripts"]), nil); err != nil {
			return err
		}

		if err := importTemplates(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Template metadata imported: %d", state.result.Imported["templates"]), nil); err != nil {
			return err
		}

		if err := importNodes(state, ctx); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Nodes imported: %d", state.result.Imported["nodes"]), nil); err != nil {
			return err
		}

		if err := importSubscriptions(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Subscriptions imported: %d", state.result.Imported["subcriptions"]), nil); err != nil {
			return err
		}

		if err := importSubscriptionGroups(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Subscription groups imported: %d", state.result.Imported["subcription_groups"]), nil); err != nil {
			return err
		}

		if err := importSubscriptionScripts(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Subscription-script links imported: %d", state.result.Imported["subcription_scripts"]), nil); err != nil {
			return err
		}

		if err := importSubscriptionNodes(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Subscription-node links imported: %d", state.result.Imported["subcription_nodes"]), nil); err != nil {
			return err
		}

		if err := importTags(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Tags imported: %d", state.result.Imported["tags"]), nil); err != nil {
			return err
		}

		if err := importTagRules(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Tag rules imported: %d", state.result.Imported["tag_rules"]), nil); err != nil {
			return err
		}

		if err := importAccessKeys(state); err != nil {
			return err
		}
		if options.IncludeAccessKeys {
			if err := reportStep(fmt.Sprintf("Access keys imported: %d", state.result.Imported["access_keys"]), nil); err != nil {
				return err
			}
		}

		if err := importSubscriptionShares(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Share links imported: %d", state.result.Imported["subscription_shares"]), nil); err != nil {
			return err
		}

		if err := importSubLogs(state); err != nil {
			return err
		}
		if options.IncludeSubLogs {
			if err := reportStep(fmt.Sprintf("Access logs imported: %d", state.result.Imported["sub_logs"]), nil); err != nil {
				return err
			}
		}

		if err := importSubscriptionChainRules(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Chain proxy rules imported: %d", state.result.Imported["subscription_chain_rules"]), nil); err != nil {
			return err
		}

		if err := importAirports(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Airports imported: %d", state.result.Imported["airports"]), nil); err != nil {
			return err
		}

		if err := importSubscriptionAirports(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Subscription-airport links imported: %d", state.result.Imported["subcription_airports"]), nil); err != nil {
			return err
		}

		if err := importGroupAirportSorts(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Group-airport sort orders imported: %d", state.result.Imported["group_airport_sorts"]), nil); err != nil {
			return err
		}

		if err := importNodeCheckProfiles(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Node check profiles imported: %d", state.result.Imported["node_check_profiles"]), nil); err != nil {
			return err
		}

		if err := importHosts(state); err != nil {
			return err
		}
		if err := reportStep(fmt.Sprintf("Hosts imported: %d", state.result.Imported["hosts"]), nil); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if err := reseedTargetSequences(database.DB); err != nil {
		return nil, err
	}
	if err := reportStep("Auto-increment sequences reset", nil); err != nil {
		return nil, err
	}

	if sourceBundle.TemplateDir != "" {
		if err := restoreTemplateDirectory(sourceBundle.TemplateDir); err != nil {
			result.Warnings = append(result.Warnings, "Failed to restore template files: "+err.Error())
		} else {
			result.TemplatesRestored = true
		}
		label := "Template directory restored"
		if !result.TemplatesRestored {
			label = "Failed to restore template directory; see task warnings for details"
		}
		if err := reportStep(label, nil); err != nil {
			return nil, err
		}
	}

	if err := reloadRuntimeStateAfterMigration(); err != nil {
		return nil, fmt.Errorf("database data imported, but refreshing runtime state failed: %w; restarting the service is recommended", err)
	}
	if err := reportStep("Configuration, caches, and cron jobs refreshed", nil); err != nil {
		return nil, err
	}

	return result, nil
}

func buildDatabaseMigrationSteps(sourceBundle *databaseMigrationSource, options DatabaseMigrationOptions) []string {
	steps := []string{
		"Clear target business data",
		"Import users",
		"Import system settings",
		"Import webhooks",
		"Import scripts",
		"Import template metadata",
		"Import nodes",
		"Import subscriptions",
		"Import subscription groups",
		"Import subscription-script links",
		"Import subscription-node links",
		"Import tags",
		"Import tag rules",
	}

	if options.IncludeAccessKeys {
		steps = append(steps, "Import access keys")
	}

	steps = append(steps, "Import share links")

	if options.IncludeSubLogs {
		steps = append(steps, "Import access logs")
	}

	steps = append(steps,
		"Import chain proxy rules",
		"Import airports",
		"Import group-airport sort orders",
		"Import node check profiles",
		"Import hosts",
		"Reset auto-increment sequences",
	)

	if sourceBundle.TemplateDir != "" {
		steps = append(steps, "Restore template directory")
	}

	steps = append(steps, "Refresh runtime state")
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
		return nil, fmt.Errorf("only a SQLite database file or a backup.zip exported from an old instance is supported")
	}
}

func looksLikeZipFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func() { _ = file.Close() }()

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
	defer func() { _ = file.Close() }()

	header := make([]byte, 16)
	if _, err := io.ReadFull(file, header); err != nil {
		return false
	}
	return string(header) == "SQLite format 3\x00"
}

func extractMigrationZip(zipPath string) (string, error) {
	tempRoot, err := ensureDatabaseMigrationTempRoot()
	if err != nil {
		return "", err
	}
	return extractMigrationZipTo(zipPath, tempRoot)
}

func extractMigrationZipTo(zipPath, tempRoot string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to read migration archive: %w", err)
	}
	defer func() { _ = reader.Close() }()

	tempDir, err := os.MkdirTemp(tempRoot, "bundle-*")
	if err != nil {
		return "", fmt.Errorf("failed to create migration temp directory: %w", err)
	}
	succeeded := false
	defer func() {
		if !succeeded {
			_ = os.RemoveAll(tempDir)
		}
	}()
	if len(reader.File) > maxMigrationZipEntries {
		return "", fmt.Errorf("migration archive exceeds the %d file limit", maxMigrationZipEntries)
	}
	var declaredSize int64
	for _, file := range reader.File {
		if file.FileInfo().Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("migration archive contains a disallowed symbolic link: %s", file.Name)
		}
		if file.UncompressedSize64 > uint64(maxMigrationFileBytes) {
			return "", fmt.Errorf("migration archive file exceeds the %d byte limit: %s", maxMigrationFileBytes, file.Name)
		}
		if file.UncompressedSize64 > uint64(maxMigrationExtractedBytes-declaredSize) {
			return "", fmt.Errorf("migration archive exceeds the %d byte extracted-size limit", maxMigrationExtractedBytes)
		}
		declaredSize += int64(file.UncompressedSize64)
	}

	var extractedSize int64
	for _, file := range reader.File {
		cleanName := filepath.Clean(filepath.FromSlash(file.Name))
		if cleanName == "." || strings.HasPrefix(cleanName, "..") {
			return "", fmt.Errorf("migration archive contains an invalid path: %s", file.Name)
		}
		targetPath := filepath.Join(tempDir, cleanName)
		if !strings.HasPrefix(targetPath, tempDir+string(os.PathSeparator)) && targetPath != tempDir {
			return "", fmt.Errorf("migration archive contains an out-of-bounds path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return "", fmt.Errorf("failed to create migration directory: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return "", fmt.Errorf("failed to create migration directory: %w", err)
		}

		src, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("failed to read migration archive file: %w", err)
		}

		dst, err := os.Create(targetPath)
		if err != nil {
			_ = src.Close()
			return "", fmt.Errorf("failed to create migration temp file: %w", err)
		}

		written, copyErr := io.Copy(dst, io.LimitReader(src, maxMigrationFileBytes+1))
		extractedSize += written
		if copyErr != nil {
			_ = dst.Close()
			_ = src.Close()
			return "", fmt.Errorf("failed to extract migration file: %w", copyErr)
		}
		if written > maxMigrationFileBytes || extractedSize > maxMigrationExtractedBytes {
			_ = dst.Close()
			_ = src.Close()
			return "", fmt.Errorf("migration archive extracted content exceeds the safety limit")
		}

		if err := dst.Close(); err != nil {
			_ = src.Close()
			return "", fmt.Errorf("failed to save migration temp file: %w", err)
		}
		if err := src.Close(); err != nil {
			return "", fmt.Errorf("failed to close migration archive file: %w", err)
		}
	}

	succeeded = true
	return tempDir, nil
}

func ensureDatabaseMigrationTempRoot() (string, error) {
	tempRoot := filepath.Join(config.GetDBPath(), ".tmp", "database-migration")
	if err := os.MkdirAll(tempRoot, 0755); err != nil {
		return "", fmt.Errorf("failed to create migration temp directory: %w", err)
	}
	return tempRoot, nil
}

func findSQLiteDatabaseFile(root string) (string, error) {
	if _, err := os.Stat(root); err != nil {
		return "", fmt.Errorf("no db directory found in the archive")
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
		return "", fmt.Errorf("no SQLite database file found in the archive")
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
	for _, key := range preservedTargetSettingKeys() {
		var setting models.SystemSetting
		err := tx.Where(map[string]any{"key": key}).Take(&setting).Error
		if err == nil {
			result[key] = setting.Value
			continue
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		return nil, fmt.Errorf("failed to load preserved target settings (%s): %w", key, err)
	}
	return result, nil
}

func clearTargetBusinessData(tx *gorm.DB) error {
	modelsInDeleteOrder := []any{
		&models.IPInfo{},
		&models.SubLogs{},
		&models.SubscriptionShare{},
		&models.SubscriptionChainRule{},
		&models.GroupAirportSort{},
		&models.SubcriptionNode{},
		&models.SubcriptionAirport{},
		&models.SubcriptionScript{},
		&models.SubcriptionGroup{},
		&models.TagRule{},
		&models.AccessKey{},
		&models.NodeCheckJob{},
		&models.NodeCheckProfile{},
		&models.Host{},
		&models.Template{},
		&models.Script{},
		&models.Tag{},
		&models.Airport{},
		&models.Subcription{},
		&models.Node{},
		&models.SystemSetting{},
		&models.Webhook{},
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
			return fmt.Errorf("failed to clear table %s: %w", table, err)
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
		return fmt.Errorf("failed to read source users: %w", err)
	}
	if err := insertRecords(state.tx, users); err != nil {
		return fmt.Errorf("failed to import users: %w", err)
	}
	state.result.Imported["users"] = len(users)
	return nil
}

func importSystemSettings(state *databaseMigrationState) error {
	settings := make([]models.SystemSetting, 0)
	if state.source.Migrator().HasTable(&models.SystemSetting{}) {
		if err := state.source.Find(&settings).Error; err != nil {
			return fmt.Errorf("failed to read source system settings: %w", err)
		}
	}

	filtered := make([]models.SystemSetting, 0, len(settings))
	for _, setting := range settings {
		if shouldPreserveTargetSetting(setting.Key) {
			continue
		}
		if len(setting.Key) > 191 {
			return fmt.Errorf("system setting key %q exceeds 191 characters and cannot be migrated", setting.Key)
		}
		filtered = append(filtered, setting)
		state.importedSettings[setting.Key] = setting.Value
	}

	for key, value := range state.preservedSettings {
		filtered = append(filtered, models.SystemSetting{Key: key, Value: value})
		state.importedSettings[key] = value
	}

	if err := insertRecords(state.tx, filtered); err != nil {
		return fmt.Errorf("failed to import system settings: %w", err)
	}
	state.result.Imported["system_settings"] = len(filtered)
	return nil
}

func preservedTargetSettingKeys() []string {
	keys := []string{"jwt_secret", "api_encryption_key", "cloudflared_enabled", "cloudflared_tunnel_token_encrypted"}
	keys = append(keys, backupservice.PreservedSettingKeys()...)
	return append(keys, socks5service.PreservedSettingKeys()...)
}

func shouldPreserveTargetSetting(key string) bool {
	for _, preserved := range preservedTargetSettingKeys() {
		if key == preserved {
			return true
		}
	}
	return false
}

func importWebhooks(state *databaseMigrationState) error {
	webhooks := make([]models.Webhook, 0)
	if state.sourceHasWebhooks {
		if err := state.source.Find(&webhooks).Error; err != nil {
			return fmt.Errorf("failed to read source webhooks: %w", err)
		}
	} else {
		legacyURL := strings.TrimSpace(state.importedSettings["webhook_url"])
		legacyHeaders := state.importedSettings["webhook_headers"]
		legacyBody := state.importedSettings["webhook_body"]
		if legacyURL != "" || strings.TrimSpace(legacyHeaders) != "" || strings.TrimSpace(legacyBody) != "" {
			method := strings.ToUpper(strings.TrimSpace(state.importedSettings["webhook_method"]))
			if method == "" {
				method = "POST"
			}
			contentType := strings.TrimSpace(state.importedSettings["webhook_content_type"])
			if contentType == "" {
				contentType = "application/json"
			}
			webhooks = append(webhooks, models.Webhook{
				Name:        "Default Webhook",
				URL:         legacyURL,
				Method:      method,
				ContentType: contentType,
				Headers:     legacyHeaders,
				Body:        legacyBody,
				Enabled:     parseBool(state.importedSettings["webhook_enabled"], false),
				EventKeys:   state.importedSettings["webhook_event_keys"],
			})
			state.result.Warnings = append(state.result.Warnings, "Source backup has no webhook table; a default webhook was rebuilt from legacy system settings")
		}
	}
	for i := range webhooks {
		normalizeOptionalTimePtr(&webhooks[i].LastTestAt)
	}

	if err := insertRecords(state.tx, webhooks); err != nil {
		return fmt.Errorf("failed to import webhooks: %w", err)
	}
	state.result.Imported["webhooks"] = len(webhooks)
	return nil
}

func importScripts(state *databaseMigrationState) error {
	if !state.source.Migrator().HasTable(&models.Script{}) {
		return nil
	}
	var scripts []models.Script
	if err := state.source.Find(&scripts).Error; err != nil {
		return fmt.Errorf("failed to read source scripts: %w", err)
	}
	if err := insertRecords(state.tx, scripts); err != nil {
		return fmt.Errorf("failed to import scripts: %w", err)
	}
	state.result.Imported["scripts"] = len(scripts)
	return nil
}

func importTemplates(state *databaseMigrationState) error {
	templates := make([]models.Template, 0)
	if state.sourceHasTemplates {
		if err := state.source.Find(&templates).Error; err != nil {
			return fmt.Errorf("failed to read source template metadata: %w", err)
		}
	} else if state.sourceBundle.TemplateDir != "" {
		var err error
		templates, err = buildTemplateMetadataFromDir(state.sourceBundle.TemplateDir)
		if err != nil {
			return err
		}
		if len(templates) > 0 {
			state.result.Warnings = append(state.result.Warnings, "Source backup has no template metadata table; template records were rebuilt from template files")
		}
	}

	for _, template := range templates {
		if len(template.Name) > 191 {
			return fmt.Errorf("template file name %q exceeds 191 characters and cannot be migrated", template.Name)
		}
	}

	if err := insertRecords(state.tx, templates); err != nil {
		return fmt.Errorf("failed to import template metadata: %w", err)
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
		return fmt.Errorf("failed to read source nodes: %w", err)
	}

	for i := range nodes {
		if err := checkDatabaseMigrationContext(ctx); err != nil {
			return err
		}
		models.NormalizeNodeForImport(&nodes[i])
		if nodes[i].Name != "" {
			if existingID, exists := state.nodeNameToID[nodes[i].Name]; !exists {
				state.nodeNameToID[nodes[i].Name] = nodes[i].ID
			} else if existingID != nodes[i].ID {
				state.ambiguousNodeNames[nodes[i].Name] = struct{}{}
			}
		}
	}

	if err := insertRecords(state.tx, nodes); err != nil {
		return fmt.Errorf("failed to import nodes: %w", err)
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
		return fmt.Errorf("failed to read source subscriptions: %w", err)
	}
	if err := insertRecords(state.tx, subs); err != nil {
		return fmt.Errorf("failed to import subscriptions: %w", err)
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
		return fmt.Errorf("failed to read source subscription groups: %w", err)
	}
	for _, group := range groups {
		if len(group.GroupName) > 191 {
			return fmt.Errorf("group name %q exceeds 191 characters and cannot be migrated", group.GroupName)
		}
	}
	if err := insertRecords(state.tx, groups); err != nil {
		return fmt.Errorf("failed to import subscription groups: %w", err)
	}
	state.result.Imported["subcription_groups"] = len(groups)
	return nil
}

func importSubscriptionAirports(state *databaseMigrationState) error {
	if !state.source.Migrator().HasTable(&models.SubcriptionAirport{}) {
		return nil
	}
	var records []models.SubcriptionAirport
	if err := state.source.Find(&records).Error; err != nil {
		return fmt.Errorf("failed to read source subscription-airport links: %w", err)
	}
	if err := insertRecords(state.tx, records); err != nil {
		return fmt.Errorf("failed to import subscription-airport links: %w", err)
	}
	state.result.Imported["subcription_airports"] = len(records)
	return nil
}

func importSubscriptionScripts(state *databaseMigrationState) error {
	if !state.source.Migrator().HasTable(&models.SubcriptionScript{}) {
		return nil
	}
	var records []models.SubcriptionScript
	if err := state.source.Find(&records).Error; err != nil {
		return fmt.Errorf("failed to read source subscription-script links: %w", err)
	}
	if err := insertRecords(state.tx, records); err != nil {
		return fmt.Errorf("failed to import subscription-script links: %w", err)
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
			return fmt.Errorf("failed to read source subscription-node links: %w", err)
		}
	} else if state.source.Migrator().HasColumn("subcription_nodes", "node_name") {
		var legacyRecords []struct {
			SubcriptionID int
			NodeName      string
			Sort          int
		}
		if err := state.source.Table("subcription_nodes").Find(&legacyRecords).Error; err != nil {
			return fmt.Errorf("failed to read legacy subscription-node links: %w", err)
		}
		for _, legacy := range legacyRecords {
			if _, ambiguous := state.ambiguousNodeNames[legacy.NodeName]; ambiguous {
				continue
			}
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
			skippedCount := len(legacyRecords) - len(records)
			warning := fmt.Sprintf("%d legacy subscription-node links skipped because the node was not found", skippedCount)
			if len(state.ambiguousNodeNames) > 0 {
				warning = fmt.Sprintf("%d legacy subscription-node links skipped because the node was missing or its name was duplicated", skippedCount)
			}
			state.result.Warnings = append(state.result.Warnings, warning)
		}
	} else {
		return nil
	}

	if err := insertRecords(state.tx, records); err != nil {
		return fmt.Errorf("failed to import subscription-node links: %w", err)
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
		return fmt.Errorf("failed to read source tags: %w", err)
	}
	if err := insertRecords(state.tx, tags); err != nil {
		return fmt.Errorf("failed to import tags: %w", err)
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
		return fmt.Errorf("failed to read source tag rules: %w", err)
	}
	if err := insertRecords(state.tx, rules); err != nil {
		return fmt.Errorf("failed to import tag rules: %w", err)
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
		return fmt.Errorf("failed to read source access keys: %w", err)
	}
	if err := insertRecords(state.tx, keys); err != nil {
		return fmt.Errorf("failed to import access keys: %w", err)
	}
	if len(keys) > 0 {
		state.result.Warnings = append(state.result.Warnings, "Access keys imported. If old API keys stop working, make sure the new instance uses the same API encryption key as the old one, or regenerate them manually")
	}
	state.result.Imported["access_keys"] = len(keys)
	return nil
}

func importSubscriptionShares(state *databaseMigrationState) error {
	shares := make([]models.SubscriptionShare, 0)
	if state.sourceHasShares {
		var sourceShares []models.SubscriptionShare
		if err := state.source.Find(&sourceShares).Error; err != nil {
			return fmt.Errorf("failed to read source share links: %w", err)
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
					return fmt.Errorf("failed to generate share token: %w", err)
				}
				sourceShares[i].Token = token
			}
			shares = append(shares, sourceShares[i])
		}
		if skippedOrphans > 0 {
			state.result.Warnings = append(state.result.Warnings, fmt.Sprintf("%d share links skipped because they referenced missing subscriptions", skippedOrphans))
		}
	} else {
		for _, sub := range state.importedSubs {
			token, err := models.GenerateToken()
			if err != nil {
				return fmt.Errorf("failed to generate default share token: %w", err)
			}
			share := models.SubscriptionShare{
				SubscriptionID: sub.ID,
				Token:          token,
				Name:           "Default share link",
				ExpireType:     models.ExpireTypeNever,
				IsLegacy:       true,
				Enabled:        true,
			}
			normalizeImportedSubscriptionShare(&share)
			shares = append(shares, share)
		}
		if len(shares) > 0 {
			state.result.Warnings = append(state.result.Warnings, "Source backup has no share table; default share links were generated for every subscription")
		}
	}

	if err := insertRecords(state.tx, shares); err != nil {
		return fmt.Errorf("failed to import share links: %w", err)
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
		return fmt.Errorf("failed to read source access logs: %w", err)
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
		state.result.Warnings = append(state.result.Warnings, fmt.Sprintf("%d access logs skipped because they referenced missing subscriptions", skippedOrphans))
	}
	if normalizedShareRefs > 0 {
		state.result.Warnings = append(state.result.Warnings, fmt.Sprintf("%d access logs reset to subscription-level logs because they referenced missing share links", normalizedShareRefs))
	}
	if err := insertRecords(state.tx, filteredLogs); err != nil {
		return fmt.Errorf("failed to import access logs: %w", err)
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
		return fmt.Errorf("failed to read source chain proxy rules: %w", err)
	}
	if err := insertRecords(state.tx, rules); err != nil {
		return fmt.Errorf("failed to import chain proxy rules: %w", err)
	}
	state.result.Imported["subscription_chain_rules"] = len(rules)
	return nil
}

func importAirports(state *databaseMigrationState) error {
	airports := make([]models.Airport, 0)

	if state.source.Migrator().HasTable(&models.Airport{}) {
		if err := state.source.Find(&airports).Error; err != nil {
			return fmt.Errorf("failed to read source airports: %w", err)
		}
	}

	if len(airports) == 0 && state.source.Migrator().HasTable(&models.SubScheduler{}) {
		var schedulers []models.SubScheduler
		if err := state.source.Find(&schedulers).Error; err != nil {
			return fmt.Errorf("failed to read legacy subscription schedules: %w", err)
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
			state.result.Warnings = append(state.result.Warnings, "Source backup used legacy subscription schedules; converted to airports automatically")
		}
	}
	for i := range airports {
		normalizeOptionalTimePtr(&airports[i].LastRunTime)
		normalizeOptionalTimePtr(&airports[i].NextRunTime)
	}

	if err := insertRecords(state.tx, airports); err != nil {
		return fmt.Errorf("failed to import airports: %w", err)
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
		return fmt.Errorf("failed to read source group-airport sort orders: %w", err)
	}
	if err := insertRecords(state.tx, records); err != nil {
		return fmt.Errorf("failed to import group-airport sort orders: %w", err)
	}
	state.result.Imported["group_airport_sorts"] = len(records)
	return nil
}

func importNodeCheckProfiles(state *databaseMigrationState) error {
	profiles := make([]models.NodeCheckProfile, 0)
	if state.sourceHasProfiles {
		if err := state.source.Find(&profiles).Error; err != nil {
			return fmt.Errorf("failed to read source node check profiles: %w", err)
		}
	}

	for _, profile := range profiles {
		if len(profile.Name) > 191 {
			return fmt.Errorf("node check profile name %q exceeds 191 characters and cannot be migrated", profile.Name)
		}
	}

	if len(profiles) == 0 {
		profiles = append(profiles, buildDefaultNodeCheckProfile(state.importedSettings))
		if !state.sourceHasProfiles {
			state.result.Warnings = append(state.result.Warnings, "Source backup has no node check profiles; a default profile was generated from the existing speed test config")
		}
	}
	for i := range profiles {
		normalizeOptionalTimePtr(&profiles[i].LastRunTime)
		normalizeOptionalTimePtr(&profiles[i].NextRunTime)
	}

	if err := insertRecords(state.tx, profiles); err != nil {
		return fmt.Errorf("failed to import node check profiles: %w", err)
	}
	state.result.Imported["node_check_profiles"] = len(profiles)
	// 独立任务与方案保留原 ID，恢复后继续引用同一方案。
	if state.source.Migrator().HasTable(&models.NodeCheckJob{}) {
		var jobs []models.NodeCheckJob
		if err := state.source.Find(&jobs).Error; err != nil {
			return err
		}
		if err := insertRecords(state.tx, jobs); err != nil {
			return err
		}
		state.result.Imported["node_check_jobs"] = len(jobs)
	}
	return nil
}

func importHosts(state *databaseMigrationState) error {
	if !state.source.Migrator().HasTable(&models.Host{}) {
		return nil
	}
	var hosts []models.Host
	if err := state.source.Find(&hosts).Error; err != nil {
		return fmt.Errorf("failed to read source hosts: %w", err)
	}
	for i := range hosts {
		normalizeOptionalTimePtr(&hosts[i].ExpireAt)
	}
	hosts = models.NormalizeHostsForStorage(hosts)
	if err := insertRecords(state.tx, hosts); err != nil {
		return fmt.Errorf("failed to import hosts: %w", err)
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
		return nil, fmt.Errorf("failed to read template directory: %w", err)
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

func normalizeOptionalTimePtr(value **time.Time) {
	if value == nil || *value == nil {
		return
	}
	if (*value).IsZero() {
		*value = nil
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
		Name:               "Default profile",
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
		return fmt.Errorf("failed to clean up target template directory: %w", err)
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
	defer func() { _ = sourceFile.Close() }()

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer func() { _ = targetFile.Close() }()

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
	// Existing sessions/adapters must not keep routing through pre-restore nodes.
	socks5service.DefaultManager().Stop()
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
		utils.Warn("Failed to sync hosts to mihomo after migration: %v", err)
	}

	if !models.IsDemoMode() {
		if err := socks5service.DefaultManager().StartFromSettings(); err != nil {
			utils.Warn("Failed to reload the SOCKS5 gateway after migration; the gateway remains disabled: %v", err)
		}
	}

	telegramConfig, err := telegram.LoadConfig()
	if err == nil {
		if telegramConfig.Enabled && telegramConfig.BotToken != "" {
			if err := telegram.Reconnect(); err != nil {
				utils.Warn("Failed to reconnect the Telegram bot after migration: %v", err)
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
			return fmt.Errorf("failed to read max ID of table %s: %w", table, err)
		}
		resetSQL := fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = %d", quoteMigrationIdentifier(table), maxID+1)
		if err := tx.Exec(resetSQL).Error; err != nil {
			return fmt.Errorf("failed to reset auto-increment value for table %s: %w", table, err)
		}
	}
	return nil
}

func reseedPostgresSequences(tx *gorm.DB) error {
	for _, table := range sequenceTables(tx) {
		var maxID int64
		query := fmt.Sprintf("SELECT COALESCE(MAX(id), 0) FROM %s", quoteMigrationIdentifier(table))
		if err := tx.Raw(query).Scan(&maxID).Error; err != nil {
			return fmt.Errorf("failed to read max ID of table %s: %w", table, err)
		}

		var sequenceName string
		if err := tx.Raw("SELECT pg_get_serial_sequence(?, 'id')", table).Scan(&sequenceName).Error; err != nil {
			return fmt.Errorf("failed to query sequence for table %s: %w", table, err)
		}
		if sequenceName == "" {
			continue
		}

		if maxID > 0 {
			if err := tx.Exec("SELECT setval(?::regclass, ?, true)", sequenceName, maxID).Error; err != nil {
				return fmt.Errorf("failed to reset sequence for table %s: %w", table, err)
			}
			continue
		}

		if err := tx.Exec("SELECT setval(?::regclass, 1, false)", sequenceName).Error; err != nil {
			return fmt.Errorf("failed to reset sequence for empty table %s: %w", table, err)
		}
	}
	return nil
}

func sequenceTables(tx *gorm.DB) []string {
	modelsWithSequence := []any{
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
		&models.NodeCheckJob{},
		&models.NodeCheckProfile{},
		&models.Webhook{},
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

func tableName(db *gorm.DB, model any) (string, error) {
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
