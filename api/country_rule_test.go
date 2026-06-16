package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"sublink/cache"
	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTestRouter 设置测试路由
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// 注册路由（不使用认证中间件）
	v1 := router.Group("/api/v1")
	{
		countryRules := v1.Group("/country-rules")
		{
			countryRules.GET("", ListCountryRules)
			countryRules.POST("", CreateCountryRule)
			countryRules.PUT("/:id", UpdateCountryRule)
			countryRules.DELETE("/:id", DeleteCountryRule)
			countryRules.POST("/test", TestCountryRule)
			countryRules.POST("/batch-test", BatchTestCountryRules)
			countryRules.POST("/batch", BatchCountryRules)
			countryRules.GET("/export", ExportCountryRules)
			countryRules.POST("/sync", SyncCountryRules)
		}
	}

	return router
}

// resetCountryRuleCacheForAPITest 重置国家规则缓存（用于 API 测试隔离）
func resetCountryRuleCacheForAPITest() {
	countryRuleCache := cache.NewMapCache(func(r models.CountryRule) int { return r.ID })
	countryRuleCache.AddIndex("enabled", func(r models.CountryRule) string {
		if r.Enabled {
			return "true"
		}
		return "false"
	})
	countryRuleCache.AddIndex("countryCode", func(r models.CountryRule) string { return r.CountryCode })
}

// setupTestDBForAPI 设置测试数据库
func setupTestDBForAPI(t *testing.T) {
	t.Helper()

	// 保存原有状态
	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	// 创建独立的测试数据库
	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "country_rule_api_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	// 运行迁移
	if err := db.AutoMigrate(&models.CountryRule{}); err != nil {
		t.Fatalf("auto migrate country_rules: %v", err)
	}

	// 设置测试环境
	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	resetCountryRuleCacheForAPITest()

	// 初始化国家规则缓存
	if err := models.InitCountryRuleCache(); err != nil {
		t.Fatalf("init country rule cache: %v", err)
	}

	// 注册清理函数
	t.Cleanup(func() {
		testutil.CloseDB(t, db)
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
	})
}

// cleanupTestDBForAPI 清理测试数据库（已废弃，使用 t.Cleanup）
func cleanupTestDBForAPI(t *testing.T) {
	// 保留空函数以兼容现有测试，实际清理在 setupTestDBForAPI 的 t.Cleanup 中
}

// createTestRuleForAPI 创建测试规则
func createTestRuleForAPI(t *testing.T, code string, name string, pattern string, priority int) *models.CountryRule {
	t.Helper()
	rule := &models.CountryRule{
		CountryCode: code,
		CountryName: name,
		Pattern:     pattern,
		Priority:    priority,
		Enabled:     true,
	}
	if err := rule.Add(); err != nil {
		t.Fatalf("create test rule: %v", err)
	}
	return rule
}

// ========== API 层测试 ==========

func TestListCountryRules(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 创建测试数据
	createTestRuleForAPI(t, "CN", "中国", "中国|China|CN", 100)
	createTestRuleForAPI(t, "US", "美国", "美国|USA?|US", 90)

	// 发送请求
	req, _ := http.NewRequest("GET", "/api/v1/country-rules", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["code"].(float64) != 200 {
		t.Errorf("Expected code 200, got %v", response["code"])
	}

	data := response["data"].([]interface{})
	if len(data) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(data))
	}
}

func TestCreateCountryRule(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 准备请求数据
	ruleData := map[string]interface{}{
		"countryCode": "CN",
		"countryName": "中国",
		"pattern":     "中国|China|CN",
		"priority":    100,
		"enabled":     true,
	}
	body, _ := json.Marshal(ruleData)

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/v1/country-rules", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d, body: %s", resp.Code, resp.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["code"].(float64) != 200 {
		t.Errorf("Expected code 200, got %v", response["code"])
	}

	// 验证数据已保存
	var count int64
	database.DB.Model(&models.CountryRule{}).Where("country_code = ?", "CN").Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 rule in database, got %d", count)
	}
}

func TestCreateCountryRule_DuplicateCode(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 创建第一条规则
	createTestRuleForAPI(t, "CN", "中国", "中国|China|CN", 100)

	// 尝试创建重复的国家代码
	ruleData := map[string]interface{}{
		"countryCode": "CN",
		"countryName": "中国2",
		"pattern":     "中国2|China2|CN2",
		"priority":    90,
		"enabled":     true,
	}
	body, _ := json.Marshal(ruleData)

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/v1/country-rules", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应
	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["code"].(float64) != 400 {
		t.Errorf("Expected code 400, got %v", response["code"])
	}
}

func TestCreateCountryRule_InvalidData(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 准备无效数据（空国家代码）
	ruleData := map[string]interface{}{
		"countryCode": "",
		"countryName": "中国",
		"pattern":     "中国|China|CN",
		"priority":    100,
		"enabled":     true,
	}
	body, _ := json.Marshal(ruleData)

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/v1/country-rules", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应
	if resp.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.Code)
	}
}

func TestUpdateCountryRule(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 创建初始规则
	rule := createTestRuleForAPI(t, "CN", "中国", "中国|China|CN", 100)

	// 准备更新数据
	updateData := map[string]interface{}{
		"countryCode": "CN",
		"countryName": "中国更新",
		"pattern":     "中国|China|CN|🇨🇳",
		"priority":    110,
		"enabled":     true,
	}
	body, _ := json.Marshal(updateData)

	// 发送请求
	req, _ := http.NewRequest("PUT", "/api/v1/country-rules/"+strconv.Itoa(rule.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d, body: %s", resp.Code, resp.Body.String())
	}

	// 验证数据已更新
	var updated models.CountryRule
	database.DB.First(&updated, rule.ID)
	if updated.CountryName != "中国更新" {
		t.Errorf("Expected CountryName to be updated to '中国更新', got %q", updated.CountryName)
	}
	if updated.Priority != 110 {
		t.Errorf("Expected Priority to be updated to 110, got %d", updated.Priority)
	}
}

func TestUpdateCountryRule_InvalidID(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 准备更新数据
	updateData := map[string]interface{}{
		"countryCode": "CN",
		"countryName": "中国",
		"pattern":     "中国|China|CN",
		"priority":    100,
		"enabled":     true,
	}
	body, _ := json.Marshal(updateData)

	// 发送请求（使用不存在的 ID）
	req, _ := http.NewRequest("PUT", "/api/v1/country-rules/99999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应
	if resp.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.Code)
	}
}

func TestDeleteCountryRule(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 创建测试规则
	rule := createTestRuleForAPI(t, "CN", "中国", "中国|China|CN", 100)

	// 发送请求
	req, _ := http.NewRequest("DELETE", "/api/v1/country-rules/"+strconv.Itoa(rule.ID), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}

	// 验证数据已删除
	var count int64
	database.DB.Model(&models.CountryRule{}).Where("id = ?", rule.ID).Count(&count)
	if count != 0 {
		t.Error("Expected rule to be deleted")
	}
}

func TestDeleteCountryRule_InvalidID(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 发送请求（使用不存在的 ID）
	req, _ := http.NewRequest("DELETE", "/api/v1/country-rules/99999", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应 - GORM 的 Delete 操作是幂等的，删除不存在的记录也返回成功
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200 (idempotent delete), got %d", resp.Code)
	}
}

func TestTestCountryRule(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 准备测试数据
	testData := map[string]interface{}{
		"pattern":  "(?i)(香港|HK|Hong\\s*Kong)",
		"testName": "Hong Kong Server 01",
	}
	body, _ := json.Marshal(testData)

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/v1/country-rules/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d, body: %s", resp.Code, resp.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	data := response["data"].(map[string]interface{})
	if !data["matched"].(bool) {
		t.Error("Expected pattern to match")
	}
}

func TestTestCountryRule_InvalidPattern(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 准备测试数据（无效正则）
	testData := map[string]interface{}{
		"pattern":  "[invalid(regex",
		"testName": "Test Name",
	}
	body, _ := json.Marshal(testData)

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/v1/country-rules/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应
	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.Code)
	}
}

func TestBatchTestCountryRules(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 创建测试规则
	createTestRuleForAPI(t, "HK", "香港", "(?i)(香港|HK|Hong\\s*Kong)", 100)
	createTestRuleForAPI(t, "US", "美国", "(?i)(美国|USA?|US)", 90)

	// 准备测试数据
	testData := map[string]interface{}{
		"nodeNames": []string{
			"Hong Kong Server 01",
			"US Node 02",
			"Unknown Node",
		},
	}
	body, _ := json.Marshal(testData)

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/v1/country-rules/batch-test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d, body: %s", resp.Code, resp.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	data := response["data"].([]interface{})
	if len(data) != 3 {
		t.Errorf("Expected 3 results, got %d", len(data))
	}

	// 验证第一个结果（香港）
	result1 := data[0].(map[string]interface{})
	if result1["country"].(string) != "HK" {
		t.Errorf("Expected country 'HK', got %v", result1["country"])
	}
	if !result1["matched"].(bool) {
		t.Error("Expected first node to match")
	}

	// 验证第二个结果（美国）
	result2 := data[1].(map[string]interface{})
	if result2["country"].(string) != "US" {
		t.Errorf("Expected country 'US', got %v", result2["country"])
	}
	if !result2["matched"].(bool) {
		t.Error("Expected second node to match")
	}

	// 验证第三个结果（未匹配）
	result3 := data[2].(map[string]interface{})
	if result3["country"].(string) != "" {
		t.Errorf("Expected country to be empty, got %v", result3["country"])
	}
	if result3["matched"].(bool) {
		t.Error("Expected third node not to match")
	}
}

func TestBatchCountryRules_ImportMode(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 准备批量导入数据
	batchData := map[string]interface{}{
		"mode": "import",
		"rules": []map[string]interface{}{
			{
				"countryCode": "CN",
				"countryName": "中国",
				"pattern":     "中国|China|CN",
				"priority":    100,
				"enabled":     true,
			},
			{
				"countryCode": "US",
				"countryName": "美国",
				"pattern":     "美国|USA?|US",
				"priority":    90,
				"enabled":     true,
			},
		},
	}
	body, _ := json.Marshal(batchData)

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/v1/country-rules/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d, body: %s", resp.Code, resp.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	data := response["data"].(map[string]interface{})
	if data["imported"].(float64) != 2 {
		t.Errorf("Expected 2 imported, got %v", data["imported"])
	}

	// 验证数据已保存
	var count int64
	database.DB.Model(&models.CountryRule{}).Count(&count)
	if count != 2 {
		t.Errorf("Expected 2 rules in database, got %d", count)
	}
}

func TestBatchCountryRules_ReplaceMode(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 创建初始规则
	createTestRuleForAPI(t, "JP", "日本", "日本|Japan|JP", 80)

	// 准备批量覆盖数据
	batchData := map[string]interface{}{
		"mode": "replace",
		"rules": []map[string]interface{}{
			{
				"countryCode": "CN",
				"countryName": "中国",
				"pattern":     "中国|China|CN",
				"priority":    100,
				"enabled":     true,
			},
		},
	}
	body, _ := json.Marshal(batchData)

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/v1/country-rules/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d, body: %s", resp.Code, resp.Body.String())
	}

	// 验证旧数据已删除，新数据已添加
	var count int64
	database.DB.Model(&models.CountryRule{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 rule in database after replace, got %d", count)
	}

	var rule models.CountryRule
	database.DB.First(&rule)
	if rule.CountryCode != "CN" {
		t.Errorf("Expected CountryCode 'CN', got %q", rule.CountryCode)
	}
}

func TestExportCountryRules(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 创建测试规则
	createTestRuleForAPI(t, "CN", "中国", "中国|China|CN", 100)
	createTestRuleForAPI(t, "US", "美国", "美国|USA?|US", 90)

	// 发送请求
	req, _ := http.NewRequest("GET", "/api/v1/country-rules/export", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	data := response["data"].(map[string]interface{})
	text := data["text"].(string)

	// 验证导出的文本包含规则
	if text == "" {
		t.Error("Expected non-empty text")
	}
	if len(text) < 100 {
		t.Errorf("Expected text to contain rules, got: %s", text)
	}
}

func TestSyncCountryRules(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 准备同步数据
	syncData := map[string]interface{}{
		"text": `CN 中国 100 true 中国|China|CN
US 美国 90 true 美国|USA?|US`,
	}
	body, _ := json.Marshal(syncData)

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/v1/country-rules/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d, body: %s", resp.Code, resp.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	data := response["data"].(map[string]interface{})
	if data["added"].(float64) != 2 {
		t.Errorf("Expected 2 added, got %v", data["added"])
	}

	// 验证数据已同步
	var count int64
	database.DB.Model(&models.CountryRule{}).Count(&count)
	if count != 2 {
		t.Errorf("Expected 2 rules in database, got %d", count)
	}
}

func TestSyncCountryRules_InvalidText(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 准备无效的同步数据
	syncData := map[string]interface{}{
		"text": `CN 中国 abc true 中国|China|CN`,
	}
	body, _ := json.Marshal(syncData)

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/v1/country-rules/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应
	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.Code)
	}

	// 验证没有数据被添加
	var count int64
	database.DB.Model(&models.CountryRule{}).Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 rules in database after failed sync, got %d", count)
	}
}

func TestSyncCountryRules_ComplexRegex(t *testing.T) {
	setupTestDBForAPI(t)
	defer cleanupTestDBForAPI(t)

	router := setupTestRouter()

	// 准备包含复杂正则的同步数据
	syncData := map[string]interface{}{
		"text": `HK 香港 100 true (?i)(香港|HK|Hong\s*Kong)
US 美国 90 true (?i)(美国|USA?|United\s*States)`,
	}
	body, _ := json.Marshal(syncData)

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/v1/country-rules/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 验证响应
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d, body: %s", resp.Code, resp.Body.String())
	}

	// 验证数据已同步
	var rules []models.CountryRule
	database.DB.Find(&rules)
	if len(rules) != 2 {
		t.Errorf("Expected 2 rules in database, got %d", len(rules))
	}

	// 验证复杂正则表达式被正确保存
	var hkRule models.CountryRule
	database.DB.Where("country_code = ?", "HK").First(&hkRule)
	expectedPattern := `(?i)(香港|HK|Hong\s*Kong)`
	if hkRule.Pattern != expectedPattern {
		t.Errorf("Expected pattern %q, got %q", expectedPattern, hkRule.Pattern)
	}
}
