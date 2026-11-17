package settings

import (
	"sublink/models"
	"sublink/utils"
)

// 重置默认用户
func ResetUser(username string, password string) {
	// 如果账号或者密码为空
	if username == "" || password == "" {
		utils.Error("Username or password cannot be empty")
		return
	}
	if len(password) < 6 {
		utils.Error("Password must be at least 6 characters")
		return
	}
	User := &models.User{}
	// 获取所有用户
	users, err := User.All()
	if err != nil {
		utils.Info("User exists")
	}
	// 遍历所有用户
	for _, user := range users {
		// 删除所有用户
		_ = user.Del()
	}

	User = &models.User{Username: username, Password: password, Role: "admin", Nickname: "Admin"}
	_ = User.Create()
}
