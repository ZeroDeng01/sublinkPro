package api

import (
	"sublink/database"
	"sublink/models"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID       int
	Username string
	Nickname string
	Avatar   string
	Mobile   string
	Email    string
}

// 新增用户
func UserAdd(c *gin.Context) {
	user := &models.User{
		Username: "test",
		Password: "test",
	}
	err := user.Create()
	if err != nil {
		utils.Error("创建用户失败: %v", err)
	}
	utils.OkWithMsg(c, "创建用户成功")
}

// 获取用户信息
func UserMe(c *gin.Context) {
	// 获取jwt中的username
	// 返回用户信息
	username, _ := c.Get("username")
	user := &models.User{Username: username.(string)}
	err := user.Find()
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkDetailed(c, "获取用户信息成功", gin.H{
		"avatar":   "",
		"nickname": user.Nickname,
		"userId":   user.ID,
		"username": user.Username,
		"roles":    []string{"ADMIN"},
	})
}

// 获取所有用户
func UserPages(c *gin.Context) {
	// 获取jwt中的username
	// 返回用户信息
	username, _ := c.Get("username")
	user := &models.User{Username: username.(string)}
	users, err := user.All()
	if err != nil {
		utils.Error("获取用户信息失败: %v", err)
	}
	list := []*User{}
	for i := range users {
		list = append(list, &User{
			ID:       users[i].ID,
			Username: users[i].Username,
			Nickname: users[i].Nickname,
			Avatar:   "",
		})
	}
	utils.OkDetailed(c, "获取用户信息成功", gin.H{
		"list": list,
	})
}

// 更新用户信息

func UserSet(c *gin.Context) {
	NewUsername := c.PostForm("username")
	NewPassword := c.PostForm("password")
	if NewUsername == "" || NewPassword == "" {
		utils.FailWithMsg(c, "用户名或密码不能为空")
		return
	}
	username, _ := c.Get("username")
	user := &models.User{Username: username.(string)}
	err := user.Set(&models.User{
		Username: NewUsername,
		Password: NewPassword,
	})
	if err != nil {
		utils.Error("修改密码失败: %v", err)
		utils.FailWithMsg(c, err.Error())
		return
	}
	// 修改成功
	utils.OkWithMsg(c, "修改成功")

}

// 修改密码
func UserChangePassword(c *gin.Context) {
	type ChangePasswordRequest struct {
		OldPassword     string `json:"oldPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required"`
		ConfirmPassword string `json:"confirmPassword" binding:"required"`
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "请求参数错误: "+err.Error())
		return
	}

	// 验证两次密码是否一致
	if req.NewPassword != req.ConfirmPassword {
		utils.FailWithMsg(c, "两次密码输入不一致")
		return
	}

	// 验证密码长度
	if len(req.NewPassword) < 6 {
		utils.FailWithMsg(c, "密码长度不能小于6位")
		return
	}

	// 获取当前用户
	username, _ := c.Get("username")
	user := &models.User{
		Username: username.(string),
		Password: req.OldPassword,
	}

	// 验证旧密码是否正确
	if err := user.Verify(); err != nil {
		utils.FailWithMsg(c, "当前密码错误")
		return
	}

	// 更新密码
	updateUser := &models.User{Password: req.NewPassword}
	if err := user.Set(updateUser); err != nil {
		utils.Error("密码修改失败: %v", err)
		utils.FailWithMsg(c, "密码修改失败")
		return
	}

	// 删除该用户的所有记住密码令牌，强制重新登录
	if err := models.DeleteUserRememberTokens(user.ID); err != nil {
		utils.Error("清除记住密码令牌失败: %v", err)
		// 不影响密码修改成功的返回
	}

	utils.OkWithMsg(c, "密码修改成功")
}

// 更新个人资料（用户名、昵称）
func UserUpdateProfile(c *gin.Context) {
	type UpdateProfileRequest struct {
		Username string `json:"username"`
		Nickname string `json:"nickname"`
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "请求参数错误: "+err.Error())
		return
	}

	// 验证用户名不能为空
	if req.Username == "" {
		utils.FailWithMsg(c, "用户名不能为空")
		return
	}

	// 获取当前用户
	username, _ := c.Get("username")
	user := &models.User{Username: username.(string)}

	// 使用 map 更新字段，避免 GORM 忽略零值
	// 这样可以更新 nickname 为空字符串
	updates := map[string]interface{}{
		"username": req.Username,
		"nickname": req.Nickname,
	}

	if err := database.DB.Where("username = ?", user.Username).Model(&models.User{}).Updates(updates).Error; err != nil {
		utils.Error("个人资料更新失败: %v", err)
		utils.FailWithMsg(c, "个人资料更新失败: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "个人资料更新成功")
}
