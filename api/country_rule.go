package api

import (
	"net/http"
	"strconv"
	"strings"
	"sublink/database"
	"sublink/middlewares"
	"sublink/models"

	"github.com/gin-gonic/gin"
)

// ListCountryRules 获取国家规则列表
// @Summary 获取国家规则列表
// @Description 获取所有国家规则，按优先级降序排列
// @Tags 国家规则
// @Accept json
// @Produce json
// @Success 200 {object} gin.H{"code": 200, "data": []models.CountryRule}
// @Router /api/v1/country-rules [get]
func ListCountryRules(c *gin.Context) {
	var rule models.CountryRule
	rules, err := rule.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取国家规则列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    rules,
	})
}

// CreateCountryRule 创建国家规则
// @Summary 创建国家规则
// @Description 创建新的国家规则
// @Tags 国家规则
// @Accept json
// @Produce json
// @Param rule body models.CountryRule true "国家规则"
// @Success 200 {object} gin.H{"code": 200, "data": models.CountryRule}
// @Router /api/v1/country-rules [post]
func CreateCountryRule(c *gin.Context) {
	var rule models.CountryRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 添加规则（包含验证和重复检查）
	if err := rule.Add(); err != nil {
		// Check if it's a duplicate key error
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "国家代码已存在",
				"error":   "country code already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建国家规则失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    rule,
	})
}

// UpdateCountryRule 更新国家规则
// @Summary 更新国家规则
// @Description 更新指定ID的国家规则
// @Tags 国家规则
// @Accept json
// @Produce json
// @Param id path int true "规则ID"
// @Param rule body models.CountryRule true "国家规则"
// @Success 200 {object} gin.H{"code": 200, "data": models.CountryRule}
// @Router /api/v1/country-rules/{id} [put]
func UpdateCountryRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的规则ID",
			"error":   err.Error(),
		})
		return
	}

	var rule models.CountryRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 设置ID
	rule.ID = id

	// 检查国家代码是否已被其他规则使用
	var existingRule models.CountryRule
	if err := database.DB.Where("country_code = ? AND id != ?", rule.CountryCode, id).First(&existingRule).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "国家代码已被其他规则使用",
			"error":   "country code already exists",
		})
		return
	}

	// 更新规则
	if err := rule.Update(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新国家规则失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    rule,
	})
}

// DeleteCountryRule 删除国家规则
// @Summary 删除国家规则
// @Description 删除指定ID的国家规则
// @Tags 国家规则
// @Accept json
// @Produce json
// @Param id path int true "规则ID"
// @Success 200 {object} gin.H{"code": 200, "message": "删除成功"}
// @Router /api/v1/country-rules/{id} [delete]
func DeleteCountryRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的规则ID",
			"error":   err.Error(),
		})
		return
	}

	var rule models.CountryRule
	rule.ID = id

	// 删除规则
	if err := rule.Delete(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除国家规则失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// TestCountryRuleRequest 测试规则请求
type TestCountryRuleRequest struct {
	Pattern  string `json:"pattern" binding:"required"`
	TestName string `json:"testName" binding:"required"`
}

// TestCountryRuleResponse 测试规则响应
type TestCountryRuleResponse struct {
	Matched     bool   `json:"matched"`
	CountryCode string `json:"countryCode,omitempty"`
	CountryName string `json:"countryName,omitempty"`
}

// TestCountryRule 测试国家规则
// @Summary 测试国家规则
// @Description 测试指定规则是否匹配节点名称
// @Tags 国家规则
// @Accept json
// @Produce json
// @Param request body TestCountryRuleRequest true "测试请求"
// @Success 200 {object} gin.H{"code": 200, "data": TestCountryRuleResponse}
// @Router /api/v1/country-rules/test [post]
func TestCountryRule(c *gin.Context) {
	var req TestCountryRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 测试规则
	matched, err := models.TestPattern(req.Pattern, req.TestName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "测试失败",
			"error":   err.Error(),
		})
		return
	}

	response := TestCountryRuleResponse{
		Matched: matched,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    response,
	})
}

// BatchTestCountryRulesRequest 批量测试请求
type BatchTestCountryRulesRequest struct {
	NodeNames []string `json:"nodeNames" binding:"required"`
}

// BatchTestResult 批量测试结果项
type BatchTestResult struct {
	Name        string `json:"name"`
	Country     string `json:"country"`
	CountryName string `json:"countryName"`
	Matched     bool   `json:"matched"`
}

// BatchTestCountryRules 批量测试节点名称
// @Summary 批量测试节点名称
// @Description 批量测试多个节点名称的国家匹配结果
// @Tags 国家规则
// @Accept json
// @Produce json
// @Param request body BatchTestCountryRulesRequest true "批量测试请求"
// @Success 200 {object} gin.H{"code": 200, "data": []BatchTestResult}
// @Router /api/v1/country-rules/batch-test [post]
func BatchTestCountryRules(c *gin.Context) {
	var req BatchTestCountryRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	results := make([]BatchTestResult, 0, len(req.NodeNames))

	// 获取所有启用的规则（用于查找国家名称）
	allRules := models.GetEnabledCountryRules()
	countryNameMap := make(map[string]string)
	for _, rule := range allRules {
		if _, exists := countryNameMap[rule.CountryCode]; !exists {
			countryNameMap[rule.CountryCode] = rule.CountryName
		}
	}

	for _, nodeName := range req.NodeNames {
		country := models.ParseCountryFromNodeName(nodeName)
		result := BatchTestResult{
			Name:    nodeName,
			Country: country,
			Matched: country != "",
		}
		if country != "" {
			if name, ok := countryNameMap[country]; ok {
				result.CountryName = name
			}
		}
		results = append(results, result)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    results,
	})
}

// BatchCountryRuleItem 批量操作规则项
type BatchCountryRuleItem struct {
	CountryCode string `json:"countryCode" binding:"required,min=2,max=10"`
	CountryName string `json:"countryName" binding:"required"`
	Pattern     string `json:"pattern" binding:"required"`
	Priority    int    `json:"priority"`
	Enabled     bool   `json:"enabled"`
}

// BatchCountryRulesRequest 批量操作规则请求
type BatchCountryRulesRequest struct {
	Mode  string                 `json:"mode" binding:"required,oneof=import replace"`
	Rules []BatchCountryRuleItem `json:"rules" binding:"required,min=1"`
}

// BatchCountryRules 批量操作规则
// @Summary 批量操作规则
// @Description 批量导入或覆盖国家规则
// @Tags 国家规则
// @Accept json
// @Produce json
// @Param request body BatchCountryRulesRequest true "批量操作请求"
// @Success 200 {object} gin.H{"code": 200, "data": map[string]int}
// @Router /api/v1/country-rules/batch [post]
func BatchCountryRules(c *gin.Context) {
	var req BatchCountryRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 验证
	if len(req.Rules) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "规则列表不能为空",
		})
		return
	}

	// 开启事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 覆盖模式：先删除所有规则
	if req.Mode == "replace" {
		if err := tx.Where("1 = 1").Delete(&models.CountryRule{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "清空规则失败",
				"error":   err.Error(),
			})
			return
		}
	}

	// 批量插入
	importCount := 0
	skipCount := 0

	for _, ruleData := range req.Rules {
		rule := models.CountryRule{
			CountryCode: ruleData.CountryCode,
			CountryName: ruleData.CountryName,
			Pattern:     ruleData.Pattern,
			Priority:    ruleData.Priority,
			Enabled:     ruleData.Enabled,
		}

		// 导入模式：检查重复
		if req.Mode == "import" {
			var exists models.CountryRule
			if err := tx.Where("country_code = ?", rule.CountryCode).First(&exists).Error; err == nil {
				skipCount++
				continue // 跳过已存在的
			}
		}

		if err := tx.Create(&rule).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "插入规则失败",
				"error":   err.Error(),
			})
			return
		}
		importCount++
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "批量操作成功",
		"data": gin.H{
			"imported": importCount,
			"skipped":  skipCount,
			"mode":     req.Mode,
		},
	})
}

// ExportCountryRules 导出所有国家规则为文本格式
// @Summary 导出国家规则
// @Description 导出所有国家规则为文本格式（用于文本编辑模式）
// @Tags 国家规则
// @Accept json
// @Produce json
// @Success 200 {object} gin.H{"code": 200, "data": map[string]string}
// @Router /api/v1/country-rules/export [get]
func ExportCountryRules(c *gin.Context) {
	text := models.ExportCountryRulesToText()
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "导出成功",
		"data": gin.H{
			"text": text,
		},
	})
}

// SyncCountryRulesRequest 同步国家规则请求
type SyncCountryRulesRequest struct {
	Text string `json:"text" binding:"required"`
}

// SyncCountryRules 从文本全量同步国家规则
// @Summary 同步国家规则
// @Description 从文本全量同步国家规则（用于文本编辑模式）
// @Tags 国家规则
// @Accept json
// @Produce json
// @Param request body SyncCountryRulesRequest true "同步请求"
// @Success 200 {object} gin.H{"code": 200, "data": map[string]int}
// @Router /api/v1/country-rules/sync [post]
func SyncCountryRules(c *gin.Context) {
	var req SyncCountryRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	added, updated, deleted, err := models.SyncCountryRulesFromText(req.Text)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "同步失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "同步成功",
		"data": gin.H{
			"added":   added,
			"updated": updated,
			"deleted": deleted,
		},
	})
}

// RegisterCountryRuleRoutes 注册国家规则路由
func RegisterCountryRuleRoutes(router *gin.RouterGroup) {
	// 国家规则路由需要认证
	authorized := router.Group("/country-rules")
	authorized.Use(middlewares.AuthToken)
	{
		authorized.GET("", ListCountryRules)
		authorized.POST("", middlewares.DemoModeRestrict, CreateCountryRule)
		authorized.PUT("/:id", middlewares.DemoModeRestrict, UpdateCountryRule)
		authorized.DELETE("/:id", middlewares.DemoModeRestrict, DeleteCountryRule)
		authorized.POST("/test", TestCountryRule)
		authorized.POST("/batch-test", BatchTestCountryRules)
		authorized.POST("/batch", middlewares.DemoModeRestrict, BatchCountryRules)
		authorized.GET("/export", ExportCountryRules)
		authorized.POST("/sync", middlewares.DemoModeRestrict, SyncCountryRules)
	}
}
