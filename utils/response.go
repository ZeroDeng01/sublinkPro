package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

const (
	SUCCESS = 200
	ERROR   = 500
)

func Result(c *gin.Context, httpCode int, code int, msg string, data interface{}) {
	c.JSON(httpCode, Response{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}

func Ok(c *gin.Context) {
	Result(c, http.StatusOK, SUCCESS, "操作成功", nil)
}

func OkWithData(c *gin.Context, data interface{}) {
	Result(c, http.StatusOK, SUCCESS, "操作成功", data)
}

func OkWithMsg(c *gin.Context, msg string) {
	Result(c, http.StatusOK, SUCCESS, msg, nil)
}

func OkDetailed(c *gin.Context, msg string, data interface{}) {
	Result(c, http.StatusOK, SUCCESS, msg, data)
}

func Fail(c *gin.Context) {
	Result(c, http.StatusOK, ERROR, "操作失败", nil)
}

func FailWithMsg(c *gin.Context, msg string) {
	Result(c, http.StatusOK, ERROR, msg, nil)
}

// FailWithData 返回失败响应并携带额外数据
func FailWithData(c *gin.Context, msg string, data interface{}) {
	Result(c, http.StatusOK, ERROR, msg, data)
}

func FailWithCode(c *gin.Context, code int, msg string) {
	Result(c, http.StatusOK, code, msg, nil)
}

func Forbidden(c *gin.Context, msg string) {
	Result(c, http.StatusForbidden, http.StatusForbidden, msg, nil)
}
