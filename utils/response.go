package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code       int            `json:"code"`
	Msg        string         `json:"msg"`
	Data       any            `json:"data"`
	I18nKey    string         `json:"i18nKey,omitempty"`
	I18nParams map[string]any `json:"i18nParams,omitempty"`
}

const (
	SUCCESS = 200
	ERROR   = 500
)

func Result(c *gin.Context, httpCode int, code int, msg string, data any) {
	ResultI18n(c, httpCode, code, msg, data, "", nil)
}

// ResultI18n 返回带可选国际化 key 的 API 响应。
// msg 保留为旧客户端与诊断场景的兼容回退，前端可优先使用 i18nKey/i18nParams 渲染本地化文案。
func ResultI18n(c *gin.Context, httpCode int, code int, msg string, data any, i18nKey string, i18nParams map[string]any) {
	c.JSON(httpCode, Response{
		Code:       code,
		Msg:        msg,
		Data:       data,
		I18nKey:    i18nKey,
		I18nParams: i18nParams,
	})
}

func Ok(c *gin.Context) {
	Result(c, http.StatusOK, SUCCESS, "操作成功", nil)
}

func OkWithData(c *gin.Context, data any) {
	Result(c, http.StatusOK, SUCCESS, "操作成功", data)
}

func OkWithMsg(c *gin.Context, msg string) {
	Result(c, http.StatusOK, SUCCESS, msg, nil)
}

func OkDetailed(c *gin.Context, msg string, data any) {
	Result(c, http.StatusOK, SUCCESS, msg, data)
}

func OkDetailedI18n(c *gin.Context, msg string, data any, i18nKey string, i18nParams map[string]any) {
	ResultI18n(c, http.StatusOK, SUCCESS, msg, data, i18nKey, i18nParams)
}

func Fail(c *gin.Context) {
	Result(c, http.StatusOK, ERROR, "操作失败", nil)
}

func FailWithMsg(c *gin.Context, msg string) {
	Result(c, http.StatusOK, ERROR, msg, nil)
}

func FailWithI18n(c *gin.Context, msg string, i18nKey string, i18nParams map[string]any) {
	ResultI18n(c, http.StatusOK, ERROR, msg, nil, i18nKey, i18nParams)
}

// FailWithData 返回失败响应并携带额外数据
func FailWithData(c *gin.Context, msg string, data any) {
	Result(c, http.StatusOK, ERROR, msg, data)
}

func FailWithDataI18n(c *gin.Context, msg string, data any, i18nKey string, i18nParams map[string]any) {
	ResultI18n(c, http.StatusOK, ERROR, msg, data, i18nKey, i18nParams)
}

func FailWithCode(c *gin.Context, code int, msg string) {
	Result(c, http.StatusOK, code, msg, nil)
}

func FailWithCodeI18n(c *gin.Context, code int, msg string, i18nKey string, i18nParams map[string]any) {
	ResultI18n(c, http.StatusOK, code, msg, nil, i18nKey, i18nParams)
}

func Forbidden(c *gin.Context, msg string) {
	Result(c, http.StatusForbidden, http.StatusForbidden, msg, nil)
}

func ForbiddenI18n(c *gin.Context, msg string, i18nKey string, i18nParams map[string]any) {
	ResultI18n(c, http.StatusForbidden, http.StatusForbidden, msg, nil, i18nKey, i18nParams)
}
