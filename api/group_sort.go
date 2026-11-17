package api

import (
	"sublink/models"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// GroupSortGroups 获取所有分组概要信息
func GroupSortGroups(c *gin.Context) {
	infos := models.GetAllGroupInfos()
	utils.OkWithData(c, infos)
}

// GroupSortDetail 获取分组详情（query: group=分组名）
func GroupSortDetail(c *gin.Context) {
	groupName := c.Query("group")
	if groupName == "" {
		utils.FailWithMsg(c, "Group name is required")
		return
	}

	detail, err := models.GetGroupDetail(groupName)
	if err != nil {
		utils.FailWithMsg(c, "Failed to fetch group details: "+err.Error())
		return
	}

	utils.OkWithData(c, detail)
}

// GroupSortSave 保存分组内机场排序
func GroupSortSave(c *gin.Context) {
	var req models.SaveGroupAirportSortsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}

	if req.GroupName == "" {
		utils.FailWithMsg(c, "Group name is required")
		return
	}

	if err := models.SaveGroupAirportSorts(req.GroupName, req.AirportSorts); err != nil {
		utils.FailWithMsg(c, "Failed to save sort order: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "Saved successfully")
}
