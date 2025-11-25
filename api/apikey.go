package api

import (
	"log"
	"strconv"
	"sublink/dto"
	"sublink/models"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func GenerateAPIKey(c *gin.Context) {
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
		log.Println(err)
		utils.FailWithMsg(c, "生成API Key失败")
		return
	}
	err = accessKey.Generate()
	if err != nil {
		log.Println(err)
		utils.FailWithMsg(c, "生成API Key失败")
		return
	}
	utils.OkDetailed(c, "API Key生成成功", map[string]string{
		"accessKey": apiKey,
	})
}

func DeleteAPIKey(c *gin.Context) {

	apiKeyIDParam := c.Param("apiKeyId")
	if apiKeyIDParam == "" {
		utils.FailWithMsg(c, "缺少API Key ID")
		return
	}

	var accessKey models.AccessKey
	apiKeyID, err := strconv.Atoi(apiKeyIDParam)
	if err != nil {
		utils.FailWithMsg(c, "删除API Key失败")
		return
	}
	accessKey.ID = apiKeyID
	err = accessKey.Delete()
	if err != nil {
		utils.FailWithMsg(c, "删除API Key失败")
		return
	}

	utils.OkWithMsg(c, "删除API Key成功")

}

func GetAPIKey(c *gin.Context) {
	userIDParam := c.Param("userId")
	if userIDParam == "" {
		utils.FailWithMsg(c, "缺少User ID")
		return
	}

	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		utils.FailWithMsg(c, "删除API Key失败")
		return
	}
	apiKeys, err := models.FindValidAccessKeys(userID)
	if err != nil {
		utils.FailWithMsg(c, "查询API Key失败")
		return
	}
	utils.OkDetailed(c, "查询API Key成功", apiKeys)
}
