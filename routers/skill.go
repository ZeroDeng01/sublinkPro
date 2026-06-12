package routers

import (
	"io/fs"
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

// Skill 注册 AI 技能包下载路由。需登录鉴权，避免匿名滥用。
func Skill(r *gin.Engine, skillFS fs.FS) {
	skillGroup := r.Group("/api/v1/skill")
	skillGroup.Use(middlewares.AuthToken)
	{
		skillGroup.GET("/download", api.DownloadSkill(skillFS))
	}
}
