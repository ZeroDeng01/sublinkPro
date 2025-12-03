package settings

import (
	"log"
	"sublink/models"
)

// 重置默认用户
func ResetUser(username string, password string) {
	// 如果账号或者密码为空
	if username == "" || password == "" {
		log.Println("账号或者密码不能为空")
		return
	}
	if len(password) < 6 {
		log.Println("密码不能小于6位数")
		return
	}
	User := &models.User{}
	// 获取所有用户
	users, err := User.All()
	if err != nil {
		log.Printf("用户存在")
	}
	// 遍历所有用户
	for _, user := range users {
		// 删除所有用户
		user.Del()
	}

	User = &models.User{Username: username, Password: password, Role: "admin", Nickname: "管理员"}
	User.Create()
}
