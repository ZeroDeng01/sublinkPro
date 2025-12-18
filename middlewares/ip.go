package middlewares

import (
	"sublink/models"
	"sublink/services/geoip"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func GetIp(c *gin.Context) {
	c.Next()
	func() {
		subname, _ := c.Get("subname")

		ip := c.ClientIP()

		// Get location from local GeoIP database
		addr, err := geoip.GetLocation(ip)
		if err != nil {
			utils.Error("Failed to get location for IP %s: %v", ip, err)
			addr = "Unknown"
		}
		var sub models.Subcription
		if subname, ok := subname.(string); ok {
			sub.Name = subname
		}
		err = sub.Find()
		if err != nil {
			utils.Error("查找订阅失败: %s", err.Error())
			return
		}
		var iplog models.SubLogs
		iplog.IP = ip
		err = iplog.Find(sub.ID)
		// 如果没有找到记录
		if err != nil {
			iplog.Addr = addr
			iplog.SubcriptionID = sub.ID
			iplog.Date = time.Now().Format("2006-01-02 15:04:05")
			iplog.Count = 1
			err = iplog.Add()
			if err != nil {
				utils.Error("Failed to add new IP log: %v", err)
				return
			}
		} else {
			// 更新访问次数
			iplog.Count++
			iplog.Addr = addr
			iplog.Date = time.Now().Format("2006-01-02 15:04:05")
			err = iplog.Update()
			if err != nil {
				utils.Error("更新IP日志失败: %s", err.Error())
				return
			}
		}
	}()

}
