package api

import (
	"strconv"
	"sublink/dto"
	"sublink/models"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func GenerateAccessKey(c *gin.Context) {
	var userAccessKey dto.UserAccessKey
	if err := c.BindJSON(&userAccessKey); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	user := &models.User{Username: userAccessKey.UserName}
	err := user.Find()
	if err != nil {
		utils.FailWithMsg(c, "用户不存在")
		return
	}

	var accessKey models.AccessKey
	accessKey.ExpiredAt = userAccessKey.ExpiredAt
	accessKey.Description = userAccessKey.Description
	accessKey.UserID = user.ID
	accessKey.CreatedAt = time.Now()
	accessKey.Username = user.Username

	apiKey, err := accessKey.GenerateAPIKey()
	if err != nil {
		utils.Error("生成 API Key 失败: %v", err)
		utils.FailWithMsg(c, "生成API Key失败")
		return
	}
	err = accessKey.Generate()
	if err != nil {
		utils.Error("生成 API Key 失败: %v", err)
		utils.FailWithMsg(c, "生成API Key失败")
		return
	}
	utils.OkDetailed(c, "API Key生成成功", map[string]string{
		"accessKey": apiKey,
	})
}

func DeleteAccessKey(c *gin.Context) {

	accessKeyIDParam := c.Param("accessKeyId")
	if accessKeyIDParam == "" {
		utils.FailWithMsg(c, "缺少Access Key ID")
		return
	}

	var accessKey models.AccessKey
	accessKeyID, err := strconv.Atoi(accessKeyIDParam)
	if err != nil {
		utils.FailWithMsg(c, "删除Access Key失败")
		return
	}
	accessKey.ID = accessKeyID
	err = accessKey.Delete()
	if err != nil {
		utils.FailWithMsg(c, "删除Access Key失败")
		return
	}

	utils.OkWithMsg(c, "删除Access Key成功")

}

func GetAccessKey(c *gin.Context) {
	userIDParam := c.Param("userId")
	if userIDParam == "" {
		utils.FailWithMsg(c, "缺少User ID")
		return
	}

	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		utils.FailWithMsg(c, "查询Access Key失败")
		return
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
		accessKeys, total, err := models.FindValidAccessKeysPaginated(userID, page, pageSize)
		if err != nil {
			utils.FailWithMsg(c, "查询Access Key失败")
			return
		}
		totalPages := 0
		if pageSize > 0 {
			totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
		}
		utils.OkDetailed(c, "查询Access Key成功", gin.H{
			"items":      accessKeys,
			"total":      total,
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": totalPages,
		})
		return
	}

	// 不带分页参数，返回全部（向后兼容）
	accessKeys, err := models.FindValidAccessKeys(userID)
	if err != nil {
		utils.FailWithMsg(c, "查询Access Key失败")
		return
	}
	utils.OkDetailed(c, "查询Access Key成功", accessKeys)
}
