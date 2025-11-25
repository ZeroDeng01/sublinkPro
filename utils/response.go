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

func Result(c *gin.Context, code int, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}

func Ok(c *gin.Context) {
	Result(c, SUCCESS, "操作成功", nil)
}

func OkWithData(c *gin.Context, data interface{}) {
	Result(c, SUCCESS, "操作成功", data)
}

func OkWithMsg(c *gin.Context, msg string) {
	Result(c, SUCCESS, msg, nil)
}

func OkDetailed(c *gin.Context, msg string, data interface{}) {
	Result(c, SUCCESS, msg, data)
}

func Fail(c *gin.Context) {
	Result(c, ERROR, "操作失败", nil)
}

func FailWithMsg(c *gin.Context, msg string) {
	Result(c, ERROR, msg, nil)
}

func FailWithCode(c *gin.Context, code int, msg string) {
	Result(c, code, msg, nil)
}
