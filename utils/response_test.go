package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func decodeResponseRecorder(t *testing.T, recorder *httptest.ResponseRecorder) Response {
	t.Helper()

	var response Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return response
}

func TestResultKeepsLegacyShapeWithoutI18nFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	Result(ctx, http.StatusOK, SUCCESS, "操作成功", gin.H{"ok": true})

	if strings.Contains(recorder.Body.String(), "i18nKey") || strings.Contains(recorder.Body.String(), "i18nParams") {
		t.Fatalf("legacy response should omit i18n fields: %s", recorder.Body.String())
	}

	response := decodeResponseRecorder(t, recorder)
	if response.Code != SUCCESS || response.Msg != "操作成功" {
		t.Fatalf("legacy response = %#v", response)
	}
}

func TestResultI18nIncludesKeyAndParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	ResultI18n(ctx, http.StatusBadRequest, ERROR, "参数错误", nil, "errors.invalidRequest", map[string]any{"field": "name"})

	response := decodeResponseRecorder(t, recorder)
	if response.I18nKey != "errors.invalidRequest" {
		t.Fatalf("i18n key = %q", response.I18nKey)
	}
	if response.I18nParams["field"] != "name" {
		t.Fatalf("i18n params = %#v", response.I18nParams)
	}
}
