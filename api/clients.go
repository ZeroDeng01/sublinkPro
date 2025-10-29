package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sublink/models"
	"sublink/node"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

var SunName string

// md5加密
func Md5(src string) string {
	m := md5.New()
	m.Write([]byte(src))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}
func GetClient(c *gin.Context) {
	// 获取协议头
	token := c.Query("token")
	ClientIndex := c.Query("client") // 客户端标识
	if token == "" {
		log.Println("token为空")
		c.Writer.WriteString("token为空")
		return
	}
	// fmt.Println(c.Query("token"))
	Sub := new(models.Subcription)
	// 获取所有订阅
	list, _ := Sub.List()
	// 查找订阅是否包含此名字
	for _, sub := range list {
		// 数据库订阅名字赋值变量
		SunName = sub.Name
		//查找token的md5是否匹配并且转换成小写
		if Md5(SunName) == strings.ToLower(token) {

			if sub.IPBlacklist != "" && utils.IsIpInCidr(c.ClientIP(), sub.IPBlacklist) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"msg": "IP受限(IP已被加入黑名单)",
				})
				return
			}
			if sub.IPWhitelist != "" && !utils.IsIpInCidr(c.ClientIP(), sub.IPWhitelist) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"msg": "IP受限(您的IP不在允许访问列表)",
				})
				return
			}

			// 判断是否带客户端参数
			switch ClientIndex {
			case "clash":
				GetClash(c)
				return
			case "surge":
				GetSurge(c)
				return
			case "v2ray":
				GetV2ray(c)
				return
			}
			// 自动识别客户端
			ClientList := []string{"clash", "surge"}
			for k, v := range c.Request.Header {
				if k == "User-Agent" {
					for _, UserAgent := range v {
						if UserAgent == "" {
							fmt.Println("User-Agent为空")
						}
						// fmt.Println("协议头:", UserAgent)
						// 遍历客户端列表
						// SunName = sub.Name
						for _, client := range ClientList {
							// fmt.Println(strings.ToLower(UserAgent), strings.ToLower(client))
							// fmt.Println(strings.Contains(strings.ToLower(UserAgent), strings.ToLower(client)))
							if strings.Contains(strings.ToLower(UserAgent), strings.ToLower(client)) {
								// fmt.Println("客户端", client)
								switch client {
								case "clash":
									GetClash(c)
									return
								case "surge":
									GetSurge(c)
									return
								default:
									fmt.Println("未知客户端") // 这个应该是不能达到的，因为已经在上面列出所有情况
								}
								// 找到匹配的客户端后退出循环

							}
						}
						GetV2ray(c)
					}

				}
			}
		}
	}

}
func GetV2ray(c *gin.Context) {
	var sub models.Subcription
	if SunName == "" {
		c.Writer.WriteString("订阅名为空")
		return
	}
	// subname := c.Param("subname")
	// subname := SunName
	// subname = node.Base64Decode(subname)
	sub.Name = SunName
	err := sub.Find()
	if err != nil {
		c.Writer.WriteString("找不到这个订阅:" + SunName)
		return
	}
	err = sub.GetSub()
	if err != nil {
		c.Writer.WriteString("读取错误")
		return
	}
	baselist := ""
	for _, v := range sub.Nodes {
		switch {
		// 如果包含多条节点
		case strings.Contains(v.Link, ","):
			links := strings.Split(v.Link, ",")
			baselist += strings.Join(links, "\n") + "\n"
			continue
		//如果是订阅转换
		case strings.Contains(v.Link, "http://") || strings.Contains(v.Link, "https://"):
			resp, err := http.Get(v.Link)
			if err != nil {
				log.Println(err)
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			nodes := utils.Base64Decode(string(body))
			baselist += nodes + "\n"
		// 默认
		default:
			baselist += v.Link + "\n"
		}
	}
	c.Set("subname", SunName)
	filename := fmt.Sprintf("%s.txt", SunName)
	encodedFilename := url.QueryEscape(filename)
	c.Writer.Header().Set("Content-Disposition", "inline; filename*=utf-8''"+encodedFilename)
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteString(utils.Base64Encode(baselist))
}
func GetClash(c *gin.Context) {
	var sub models.Subcription
	// subname := c.Param("subname")
	// subname := node.Base64Decode(SunName)
	sub.Name = SunName
	err := sub.Find()
	if err != nil {
		c.Writer.WriteString("找不到这个订阅:" + SunName)
		return
	}
	err = sub.GetSub()
	if err != nil {
		c.Writer.WriteString("读取错误")
		return
	}
	var urls []node.Urls
	for _, v := range sub.Nodes {
		switch {
		// 如果包含多条节点
		case strings.Contains(v.Link, ","):
			links := strings.Split(v.Link, ",")
			for _, link := range links {
				urls = append(urls, node.Urls{
					Url:             link,
					DialerProxyName: strings.TrimSpace(v.DialerProxyName),
				})
			}
			continue
		//如果是订阅转换
		case strings.Contains(v.Link, "http://") || strings.Contains(v.Link, "https://"):
			resp, err := http.Get(v.Link)
			if err != nil {
				log.Println(err)
				continue
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			nodes := utils.Base64Decode(string(body))
			links := strings.Split(nodes, "\n")
			for _, link := range links {
				urls = append(urls, node.Urls{
					Url:             link,
					DialerProxyName: strings.TrimSpace(v.DialerProxyName),
				})
			}
		// 默认
		default:
			urls = append(urls, node.Urls{
				Url:             v.Link,
				DialerProxyName: strings.TrimSpace(v.DialerProxyName),
			})
		}
	}

	var configs utils.SqlConfig
	err = json.Unmarshal([]byte(sub.Config), &configs)
	if err != nil {
		c.Writer.WriteString("配置读取错误")
		return
	}
	DecodeClash, err := node.EncodeClash(urls, configs)
	if err != nil {
		c.Writer.WriteString(err.Error())
		return
	}
	c.Set("subname", SunName)
	filename := fmt.Sprintf("%s.yaml", SunName)
	encodedFilename := url.QueryEscape(filename)
	c.Writer.Header().Set("Content-Disposition", "inline; filename*=utf-8''"+encodedFilename)
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Writer.WriteString(string(DecodeClash))
}
func GetSurge(c *gin.Context) {
	var sub models.Subcription
	// subname := c.Param("subname")
	// subname := node.Base64Decode(SunName)
	sub.Name = SunName
	err := sub.Find()
	if err != nil {
		c.Writer.WriteString("找不到这个订阅:" + SunName)
		return
	}
	err = sub.GetSub()
	if err != nil {
		c.Writer.WriteString("读取错误")
		return
	}
	urls := []string{}
	for _, v := range sub.Nodes {
		switch {
		// 如果包含多条节点
		case strings.Contains(v.Link, ","):
			links := strings.Split(v.Link, ",")
			urls = append(urls, links...)
			continue
		//如果是订阅转换
		case strings.Contains(v.Link, "http://") || strings.Contains(v.Link, "https://"):
			resp, err := http.Get(v.Link)
			if err != nil {
				log.Println(err)
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			nodes := utils.Base64Decode(string(body))
			links := strings.Split(nodes, "\n")
			urls = append(urls, links...)
		// 默认
		default:
			urls = append(urls, v.Link)
		}
	}

	var configs utils.SqlConfig
	err = json.Unmarshal([]byte(sub.Config), &configs)
	if err != nil {
		c.Writer.WriteString("配置读取错误")
		return
	}
	// log.Println("surge路径:", configs)
	DecodeClash, err := node.EncodeSurge(urls, configs)
	if err != nil {
		c.Writer.WriteString(err.Error())
		return
	}
	c.Set("subname", SunName)
	filename := fmt.Sprintf("%s.conf", SunName)
	encodedFilename := url.QueryEscape(filename)
	c.Writer.Header().Set("Content-Disposition", "inline; filename*=utf-8''"+encodedFilename)
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	host := c.Request.Host
	url := c.Request.URL.String()
	// 如果包含头部更新信息
	if strings.Contains(DecodeClash, "#!MANAGED-CONFIG") {
		c.Writer.WriteString(DecodeClash)
		return
	}
	// 否则就插入头部更新信息
	interval := fmt.Sprintf("#!MANAGED-CONFIG %s interval=86400 strict=false", host+url)
	c.Writer.WriteString(string(interval + "\n" + DecodeClash))
}
