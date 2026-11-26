package routers

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestProxyGatewayRouteRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	var handlers []string
	router.Use(func(c *gin.Context) {
		handlers = c.HandlerNames()
	})
	ProxyGateway(router)

	for _, route := range router.Routes() {
		if !strings.HasPrefix(route.Path, "/api/v1/proxy-gateway/") {
			t.Errorf("unexpected gateway route %s %s", route.Method, route.Path)
		}
		path := strings.ReplaceAll(route.Path, ":id", "example")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), route.Method, path, nil))
		if response.Code != http.StatusForbidden {
			t.Errorf("anonymous %s %s returned %d", route.Method, path, response.Code)
		}
		restricted := slices.ContainsFunc(handlers, func(name string) bool {
			return strings.HasSuffix(name, ".DemoModeRestrict")
		})
		wantRestricted := route.Method != http.MethodGet || route.Path == "/api/v1/proxy-gateway/connections"
		if restricted != wantRestricted {
			t.Errorf("unexpected demo-mode policy for %s %s", route.Method, path)
		}
	}
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), method, "/api/v1/settings/socks5", nil))
		if response.Code != http.StatusNotFound {
			t.Errorf("retired route %s returned %d", method, response.Code)
		}
	}
}
