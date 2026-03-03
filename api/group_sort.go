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
		utils.FailWithMsg(c, "分组名称不能为空")
		return
	}

	detail, err := models.GetGroupDetail(groupName)
	if err != nil {
		utils.FailWithMsg(c, "获取分组详情失败: "+err.Error())
		return
	}

	utils.OkWithData(c, detail)
}

// GroupSortSave 保存分组内机场排序
func GroupSortSave(c *gin.Context) {
	var req models.SaveGroupAirportSortsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	if req.GroupName == "" {
		utils.FailWithMsg(c, "分组名称不能为空")
		return
	}

	if err := models.SaveGroupAirportSorts(req.GroupName, req.AirportSorts); err != nil {
		utils.FailWithMsg(c, "保存排序失败: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "保存成功")
}
