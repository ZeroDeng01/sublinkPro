package api

import (
	"errors"
	"strings"

	"sublink/node/happ"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

type happDecryptRequest struct {
	Link string `json:"link"`
}

// HappDecrypt 将 happ:// 加密链接还原为明文地址
func HappDecrypt(c *gin.Context) {
	var req happDecryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithI18n(c, "Invalid parameters", "backend.common.invalidParams", nil)
		return
	}

	link := strings.TrimSpace(req.Link)
	if link == "" {
		utils.FailWithI18n(c, "Please enter a happ link", "backend.happ.linkRequired", nil)
		return
	}

	resolved, err := happ.Decrypt(link)
	if err != nil {
		utils.FailWithI18n(c, happErrorMessage(err), happErrorKey(err), happErrorParams(link))
		return
	}

	utils.OkDetailedI18n(c, "Decrypted successfully", gin.H{
		"url":        resolved,
		"generation": happ.Generation(link),
	}, "backend.happ.decryptSuccess", nil)
}

// resolveHappURLValue 在输入阶段把 happ:// 链接换成明文地址。
// 返回值 resolved 表示是否可以继续处理，失败时已经写入了错误响应。
func resolveHappURLValue(c *gin.Context, rawURL string) (resolved string, ok bool) {
	if !happ.IsLink(rawURL) {
		return rawURL, true
	}

	decrypted, err := happ.Decrypt(rawURL)
	if err != nil {
		utils.FailWithI18n(c, happErrorMessage(err), happErrorKey(err), happErrorParams(rawURL))
		return "", false
	}
	return decrypted, true
}

func happErrorKey(err error) string {
	switch {
	case errors.Is(err, happ.ErrNotALink):
		return "backend.happ.invalidLink"
	case errors.Is(err, happ.ErrUnsupported):
		return "backend.happ.unsupportedLink"
	default:
		return "backend.happ.decryptFailed"
	}
}

func happErrorMessage(err error) string {
	switch {
	case errors.Is(err, happ.ErrNotALink):
		return "Not a valid happ link; the link must start with happ://"
	case errors.Is(err, happ.ErrUnsupported):
		return "Unsupported happ link type"
	default:
		return "Failed to decrypt happ link; make sure the link is complete and not expired"
	}
}

func happErrorParams(link string) map[string]any {
	generation := happ.Generation(link)
	if generation == "" {
		return nil
	}
	return map[string]any{"type": generation}
}
