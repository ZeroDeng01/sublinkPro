package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sublink/models"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// ShareListReq 获取分享列表请求
type ShareListReq struct {
	SubID    int    `form:"subId" binding:"required"`
	Page     int    `form:"page"`     // 可选，分页页码
	PageSize int    `form:"pageSize"` // 可选，每页数量
	Keyword  string `form:"keyword"`  // 可选，搜索关键词
}

// ShareBatchCreateReq 批量创建分享请求
type ShareBatchCreateReq struct {
	SubscriptionID int    `json:"subscription_id" binding:"required"`
	BaseName       string `json:"base_name" binding:"required"`
	Count          int    `json:"count" binding:"required,min=1,max=100"`
	ExpireType     int    `json:"expire_type"`
	ExpireDays     int    `json:"expire_days"`
	ExpireAt       string `json:"expire_at"`
	Enabled        bool   `json:"enabled"`
}

// ShareBatchDeleteReq 批量删除分享请求
type ShareBatchDeleteReq struct {
	IDs []int `json:"ids" binding:"required,min=1"`
}

// ShareBatchUpdateReq 批量更新分享请求
type ShareBatchUpdateReq struct {
	IDs        []int  `json:"ids" binding:"required,min=1"`
	Enabled    *bool  `json:"enabled"` // 指针允许 null
	ExpireType *int   `json:"expire_type"`
	ExpireDays *int   `json:"expire_days"`
	ExpireAt   string `json:"expire_at"`
}

// ShareCreateReq 创建分享请求
type ShareCreateReq struct {
	SubscriptionID int    `json:"subscription_id" binding:"required"`
	Name           string `json:"name"`
	Token          string `json:"token"` // 可选，为空则自动生成
	ExpireType     int    `json:"expire_type"`
	ExpireDays     int    `json:"expire_days"`
	ExpireAt       string `json:"expire_at"` // ISO格式日期时间字符串
}

// ShareUpdateReq 更新分享请求
type ShareUpdateReq struct {
	ID         int    `json:"id" binding:"required"`
	Name       string `json:"name"`
	Token      string `json:"token"`
	ExpireType int    `json:"expire_type"`
	ExpireDays int    `json:"expire_days"`
	ExpireAt   string `json:"expire_at"`
	Enabled    bool   `json:"enabled"`
}

func parseShareExpireAt(expireType int, raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if expireType != models.ExpireTypeDateTime {
		return nil, nil
	}
	if raw == "" {
		return nil, fmt.Errorf("指定时间过期时必须提供过期时间")
	}

	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		parsed, err = time.ParseInLocation("2006-01-02 15:04:05", raw, time.Local)
		if err != nil {
			parsed, err = time.ParseInLocation("2006-01-02T15:04", raw, time.Local)
			if err != nil {
				return nil, fmt.Errorf("过期时间格式错误")
			}
		}
	}

	return &parsed, nil
}

// ShareGet 获取订阅的所有分享列表（支持分页和搜索）
func ShareGet(c *gin.Context) {
	var req ShareListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 如果提供了分页参数，返回分页数据
	if req.Page > 0 && req.PageSize > 0 {
		shares, total, err := models.GetSharesBySubscriptionIDPaginated(req.SubID, req.Page, req.PageSize, req.Keyword)
		if err != nil {
			utils.Error("获取分享列表失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取分享列表失败"})
			return
		}

		totalPages := (total + req.PageSize - 1) / req.PageSize
		hasMore := req.Page < totalPages

		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": gin.H{
				"items":      shares,
				"total":      total,
				"page":       req.Page,
				"pageSize":   req.PageSize,
				"totalPages": totalPages,
				"hasMore":    hasMore,
			},
		})
		return
	}

	// 不分页，返回所有分享（向后兼容，也支持搜索）
	shares := models.GetSharesBySubscriptionID(req.SubID, req.Keyword)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": shares})
}

// ShareAdd 创建新分享
func ShareAdd(c *gin.Context) {
	var req ShareCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	expireAt, err := parseShareExpireAt(req.ExpireType, req.ExpireAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	share := &models.SubscriptionShare{
		SubscriptionID: req.SubscriptionID,
		Name:           req.Name,
		Token:          req.Token,
		ExpireType:     req.ExpireType,
		ExpireDays:     req.ExpireDays,
		ExpireAt:       expireAt,
		Enabled:        true,
	}

	if err := share.Add(); err != nil {
		utils.Error("创建分享失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "创建成功", "data": share})
}

// ShareUpdate 更新分享设置
func ShareUpdate(c *gin.Context) {
	var req ShareUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 获取现有分享
	share := &models.SubscriptionShare{ID: req.ID}
	if err := share.Find(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "分享不存在"})
		return
	}

	expireAt, err := parseShareExpireAt(req.ExpireType, req.ExpireAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// 更新字段
	share.Name = req.Name
	share.Token = req.Token
	share.ExpireType = req.ExpireType
	share.ExpireDays = req.ExpireDays
	share.ExpireAt = expireAt
	share.Enabled = req.Enabled

	if err := share.Update(); err != nil {
		utils.Error("更新分享失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

// ShareDelete 删除分享
func ShareDelete(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的分享ID"})
		return
	}

	share := &models.SubscriptionShare{ID: id}
	if err := share.Find(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "分享不存在"})
		return
	}

	// 禁止删除默认分享链接
	if share.IsLegacy {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "默认分享链接不可删除，如链接泄漏请使用刷新Token功能"})
		return
	}

	if err := share.Delete(); err != nil {
		utils.Error("删除分享失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// ShareRefreshToken 刷新分享Token
func ShareRefreshToken(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的分享ID"})
		return
	}

	share := &models.SubscriptionShare{ID: id}
	if err := share.Find(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "分享不存在"})
		return
	}

	// 生成新Token
	newToken, err := models.GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "生成Token失败"})
		return
	}

	share.Token = newToken
	if err := share.Update(); err != nil {
		utils.Error("刷新Token失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "Token已刷新", "data": gin.H{"token": newToken}})
}

// ShareLogs 获取分享的访问日志
func ShareLogs(c *gin.Context) {
	shareIdStr := c.Query("shareId")
	shareId, err := strconv.Atoi(shareIdStr)
	if err != nil || shareId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的分享ID"})
		return
	}

	logs := models.GetSubLogsByShareID(shareId)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": logs})
}

// ShareBatchAdd 批量创建分享
func ShareBatchAdd(c *gin.Context) {
	var req ShareBatchCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 验证订阅是否存在
	var sub models.Subcription
	sub.ID = req.SubscriptionID
	if err := sub.Find(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "订阅不存在"})
		return
	}

	// 解析过期时间
	expireAt, err := parseShareExpireAt(req.ExpireType, req.ExpireAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// 检查名称冲突
	existingShares := models.GetSharesBySubscriptionID(req.SubscriptionID)
	existingNames := make(map[string]bool)
	for _, s := range existingShares {
		existingNames[s.Name] = true
	}

	var conflictNames []string
	for i := 1; i <= req.Count; i++ {
		name := fmt.Sprintf("%s-%d", req.BaseName, i)
		if existingNames[name] {
			conflictNames = append(conflictNames, name)
		}
	}

	if len(conflictNames) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("以下名称已存在: %s", strings.Join(conflictNames, ", ")),
		})
		return
	}

	// 开始批量创建
	createdShares := make([]models.SubscriptionShare, 0, req.Count)
	for i := 1; i <= req.Count; i++ {
		name := fmt.Sprintf("%s-%d", req.BaseName, i)

		// 生成唯一token
		token, err := models.GenerateToken()
		if err != nil {
			utils.Error("生成Token失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "生成Token失败"})
			return
		}

		// 确保token唯一（极少情况下会冲突）
		retries := 0
		for models.IsTokenExists(token, 0) && retries < 3 {
			token, err = models.GenerateToken()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "生成Token失败"})
				return
			}
			retries++
		}

		share := &models.SubscriptionShare{
			SubscriptionID: req.SubscriptionID,
			Name:           name,
			Token:          token,
			ExpireType:     req.ExpireType,
			ExpireDays:     req.ExpireDays,
			ExpireAt:       expireAt,
			Enabled:        req.Enabled,
			IsLegacy:       false,
		}

		if err := share.Add(); err != nil {
			utils.Error("创建分享失败 [%s]: %v", name, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  fmt.Sprintf("创建分享失败: %s (%s)", name, err.Error()),
			})
			return
		}

		createdShares = append(createdShares, *share)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  fmt.Sprintf("成功创建 %d 个分享", len(createdShares)),
		"data": createdShares,
	})
}

// ShareBatchDelete 批量删除分享
func ShareBatchDelete(c *gin.Context) {
	var req ShareBatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 验证所有ID并检查是否为legacy
	var legacyNames []string
	for _, id := range req.IDs {
		share := &models.SubscriptionShare{ID: id}
		if err := share.Find(); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": fmt.Sprintf("分享ID %d 不存在", id)})
			return
		}
		if share.IsLegacy {
			legacyNames = append(legacyNames, share.Name)
		}
	}

	if len(legacyNames) > 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"code": 403,
			"msg":  fmt.Sprintf("以下默认分享链接不可删除: %s", strings.Join(legacyNames, ", ")),
		})
		return
	}

	// 批量删除
	successCount := 0
	var errors []string
	for _, id := range req.IDs {
		share := &models.SubscriptionShare{ID: id}
		if err := share.Find(); err == nil {
			if err := share.Delete(); err != nil {
				errors = append(errors, fmt.Sprintf("ID %d: %s", id, err.Error()))
			} else {
				successCount++
			}
		}
	}

	if len(errors) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  fmt.Sprintf("删除了 %d/%d 个分享，部分失败: %s", successCount, len(req.IDs), strings.Join(errors, "; ")),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  fmt.Sprintf("成功删除 %d 个分享", successCount),
	})
}

// ShareBatchUpdate 批量更新分享
func ShareBatchUpdate(c *gin.Context) {
	var req ShareBatchUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 验证所有ID是否存在
	shares := make([]*models.SubscriptionShare, 0, len(req.IDs))
	for _, id := range req.IDs {
		share := &models.SubscriptionShare{ID: id}
		if err := share.Find(); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": fmt.Sprintf("分享ID %d 不存在", id)})
			return
		}
		shares = append(shares, share)
	}

	// 解析过期时间（如果提供了ExpireType）
	var expireAt *time.Time
	if req.ExpireType != nil {
		var err error
		expireAt, err = parseShareExpireAt(*req.ExpireType, req.ExpireAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
	}

	// 批量更新
	successCount := 0
	var errors []string
	for _, share := range shares {
		// 只更新提供的字段
		if req.Enabled != nil {
			share.Enabled = *req.Enabled
		}
		if req.ExpireType != nil {
			share.ExpireType = *req.ExpireType
		}
		if req.ExpireDays != nil {
			share.ExpireDays = *req.ExpireDays
		}
		if expireAt != nil {
			share.ExpireAt = expireAt
		}

		if err := share.Update(); err != nil {
			errors = append(errors, fmt.Sprintf("ID %d: %s", share.ID, err.Error()))
		} else {
			successCount++
		}
	}

	if len(errors) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  fmt.Sprintf("更新了 %d/%d 个分享，部分失败: %s", successCount, len(req.IDs), strings.Join(errors, "; ")),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  fmt.Sprintf("成功更新 %d 个分享", successCount),
	})
}
