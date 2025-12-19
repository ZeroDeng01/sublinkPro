package api

import (
	"net/http"
	"strconv"
	"sublink/models"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// ShareListReq 获取分享列表请求
type ShareListReq struct {
	SubID int `form:"subId" binding:"required"`
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

// ShareGet 获取订阅的所有分享列表
func ShareGet(c *gin.Context) {
	var req ShareListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	shares := models.GetSharesBySubscriptionID(req.SubID)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": shares})
}

// ShareAdd 创建新分享
func ShareAdd(c *gin.Context) {
	var req ShareCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 解析过期时间
	var expireAt time.Time
	if req.ExpireAt != "" && req.ExpireType == models.ExpireTypeDateTime {
		parsed, err := time.Parse(time.RFC3339, req.ExpireAt)
		if err != nil {
			// 尝试其他常见格式
			parsed, err = time.ParseInLocation("2006-01-02 15:04:05", req.ExpireAt, time.Local)
			if err != nil {
				parsed, err = time.ParseInLocation("2006-01-02T15:04", req.ExpireAt, time.Local)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "过期时间格式错误"})
					return
				}
			}
		}
		expireAt = parsed
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

	// 解析过期时间
	var expireAt time.Time
	if req.ExpireAt != "" && req.ExpireType == models.ExpireTypeDateTime {
		parsed, err := time.Parse(time.RFC3339, req.ExpireAt)
		if err != nil {
			parsed, err = time.ParseInLocation("2006-01-02 15:04:05", req.ExpireAt, time.Local)
			if err != nil {
				parsed, err = time.ParseInLocation("2006-01-02T15:04", req.ExpireAt, time.Local)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "过期时间格式错误"})
					return
				}
			}
		}
		expireAt = parsed
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
