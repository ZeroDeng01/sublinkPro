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
		shareIDVal, _ := c.Get("shareID")

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

		// 获取shareID
		var shareID int
		if sid, ok := shareIDVal.(int); ok {
			shareID = sid
		}

		var iplog models.SubLogs
		iplog.IP = ip

		// 使用 FindByShare 精确查找
		err = iplog.FindByShare(sub.ID, shareID)
		// 如果没有找到记录
		if err != nil {
			iplog.Addr = addr
			iplog.SubcriptionID = sub.ID
			iplog.ShareID = shareID
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
