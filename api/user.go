package api

import (
	"log"
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
		log.Println("创建用户失败")
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
		"avatar":   "static/avatar.gif",
		"nickname": user.Nickname,
		"userId":   user.ID,
		"username": user.Username,
		"roles":    []string{"ADMIN"},
		// "perms": []string{
		// 	"sys:menu:delete", "sys:dept:edit", "sys:dict_type:add",
		// 	"sys:dict:edit", "sys:dict:delete", "sys:dict_type:edit",
		// 	"sys:menu:add", "sys:user:add", "sys:role:edit",
		// 	"sys:dept:delete", "sys:user:password_reset", "sys:user:edit",
		// 	"sys:user:delete", "sys:dept:add", "sys:role:delete",
		// 	"sys:dict_type:delete", "sys:menu:edit", "sys:dict:add",
		// 	"sys:role:add",
		// },
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
		log.Println("获取用户信息失败")
	}
	list := []*User{}
	for i := range users {
		list = append(list, &User{
			ID:       users[i].ID,
			Username: users[i].Username,
			Nickname: users[i].Nickname,
			Avatar:   "static/avatar.gif",
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
	log.Println(NewUsername, NewPassword)
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
		log.Println(err)
		utils.FailWithMsg(c, err.Error())
		return
	}
	// 修改成功
	utils.OkWithMsg(c, "修改成功")

}
