package api

import (
	"bytes"
	"errors"
	"net/http"
	"strconv"

	"sublink/models"

	"github.com/gin-gonic/gin"
)

// subscriptionBuffer keeps headers and content private until device admission
// succeeds. Existing renderers are buffered only on the public subscription route.
type subscriptionBuffer struct {
	gin.ResponseWriter
	header http.Header
	body   bytes.Buffer
	status int
}

func (w *subscriptionBuffer) Header() http.Header               { return w.header }
func (w *subscriptionBuffer) WriteHeader(status int)            { w.status = status }
func (w *subscriptionBuffer) WriteHeaderNow()                   {}
func (w *subscriptionBuffer) Write(b []byte) (int, error)       { return w.body.Write(b) }
func (w *subscriptionBuffer) WriteString(s string) (int, error) { return w.body.WriteString(s) }
func (w *subscriptionBuffer) Status() int                       { return w.status }
func (w *subscriptionBuffer) Size() int                         { return w.body.Len() }
func (w *subscriptionBuffer) Written() bool                     { return w.body.Len() > 0 }
func (w *subscriptionBuffer) Flush()                            {}

func subscriptionDeviceIdentity(c *gin.Context) models.ShareDeviceIdentity {
	identity := models.ShareDeviceIdentity{UserAgent: c.GetHeader("User-Agent"), HWID: c.GetHeader("X-HWID"), OS: c.GetHeader("X-Device-OS"), Model: c.GetHeader("X-Device-Model")}
	if len(c.Request.Header.Values("X-HWID")) != 1 {
		identity.HWID = ""
	}
	return identity
}

func writeShareAccessError(c *gin.Context, err error) {
	c.Header("Cache-Control", "private, no-store")
	var denied *models.ShareAccessError
	if errors.As(err, &denied) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": denied.Code, "msg": denied.Message})
		return
	}
	c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"code": "device_check_unavailable", "msg": "设备校验暂不可用，请稍后重试"})
}

func writeSubscriptionContent(c *gin.Context, content string) {
	c.Set("subscriptionContentReady", true)
	_, _ = c.Writer.WriteString(content)
}

func dispatchDeviceCheckedResponse(c *gin.Context, prepared preparedClientResponse, token string) {
	original := c.Writer
	buffer := &subscriptionBuffer{ResponseWriter: original, header: make(http.Header), status: http.StatusOK}
	c.Writer = buffer
	defer func() { c.Writer = original }()
	dispatchPreparedClientResponse(c, prepared)
	c.Writer = original
	if buffer.status >= 400 {
		c.Data(buffer.status, "text/plain; charset=utf-8", buffer.body.Bytes())
		return
	}
	if c.Request.Method != http.MethodHead && !c.GetBool("subscriptionContentReady") {
		c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"code": "subscription_generation_failed", "msg": "订阅生成失败"})
		return
	}
	if err := models.CheckShareDevice(prepared.ShareID, token, subscriptionDeviceIdentity(c), c.Request.Method != http.MethodHead); err != nil {
		writeShareAccessError(c, err)
		return
	}
	for key, values := range buffer.header {
		original.Header()[key] = values
	}
	c.Header("Cache-Control", "private, no-store")
	c.Header("Vary", "User-Agent, X-HWID")
	if c.Request.Method == http.MethodHead {
		c.Status(buffer.status)
		return
	}
	c.Data(buffer.status, buffer.header.Get("Content-Type"), buffer.body.Bytes())
}

func shareDeviceID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Query("shareId"))
	if err != nil || id <= 0 {
		c.AbortWithStatusJSON(400, gin.H{"code": 400, "msg": "无效的分享ID"})
		return 0, false
	}
	share := models.SubscriptionShare{ID: id}
	if err := share.Find(); err != nil {
		c.AbortWithStatusJSON(404, gin.H{"code": 404, "msg": "分享不存在"})
		return 0, false
	}
	return id, true
}

func ShareDevices(c *gin.Context) {
	id, ok := shareDeviceID(c)
	if !ok {
		return
	}
	devices, err := models.ListShareDevices(id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "读取设备失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": devices})
}

func ShareDeviceUpdate(c *gin.Context) {
	id, ok := shareDeviceID(c)
	if !ok {
		return
	}
	var req struct {
		ID      int     `json:"id" binding:"required,min=1"`
		Revoked *bool   `json:"revoked"`
		Name    *string `json:"name" binding:"omitempty,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "设备参数无效"})
		return
	}
	if err := models.UpdateShareDevice(id, req.ID, req.Revoked, req.Name); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "msg": "设备已更新"})
}

func ShareDevicesReset(c *gin.Context) {
	id, ok := shareDeviceID(c)
	if !ok {
		return
	}
	token, err := models.ResetShareDevices(id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "重置设备失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": gin.H{"token": token}})
}
