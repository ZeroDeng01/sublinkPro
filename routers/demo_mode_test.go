package routers

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDemoModeRouteRestrictions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pathParameter := regexp.MustCompile(`:[^/]+`)

	groups := []struct {
		name            string
		register        func(*gin.Engine)
		restrictAll     bool
		restrictedReads []string
		allowedWrites   []string
	}{
		{name: "airports", register: Airport, restrictAll: true},
		{name: "hosts", register: Host, restrictAll: true},
		{name: "shares", register: Share, restrictAll: true},
		{name: "backup", register: Backup, restrictAll: true},
		{name: "nodes", register: Nodes, restrictedReads: []string{"/api/v1/nodes/ip-info"}},
		{name: "node checks", register: NodeCheck},
		{name: "subscriptions", register: Subcription, allowedWrites: []string{"/api/v1/subcription/preview"}},
		{
			name: "country rules", register: CountryRule,
			allowedWrites: []string{"/api/v1/country-rules/test", "/api/v1/country-rules/batch-test"},
		},
		{name: "tags", register: Tag},
		{name: "group sort", register: GroupSort},
		{name: "tasks", register: Tasks},
		{name: "GeoIP", register: GeoIP, restrictedReads: []string{"/api/v1/geoip/config"}},
		{name: "scripts", register: Script, restrictedReads: []string{"/api/v1/script/list"}},
		{name: "statistics", register: Total, restrictedReads: []string{"/api/v1/total/system-stats"}},
		{
			name: "templates", register: Templates,
			restrictedReads: []string{"/api/v1/template/get", "/api/v1/template/ai/edit-sessions/:sessionId"},
			allowedWrites: []string{
				"/api/v1/template/ai/generate", "/api/v1/template/ai/generate-stream",
				"/api/v1/template/ai/validate", "/api/v1/template/ai/apply",
			},
		},
		{
			name: "settings", register: Settings,
			restrictedReads: []string{
				"/api/v1/settings/webhooks", "/api/v1/settings/base-templates",
				"/api/v1/settings/telegram", "/api/v1/settings/telegram/status",
				"/api/v1/settings/ai-assistant", "/api/v1/settings/cloudflared", "/api/v1/settings/substore",
			},
		},
	}
	for _, group := range groups {
		t.Run(group.name, func(t *testing.T) {
			router := gin.New()
			var handlers []string
			router.Use(func(c *gin.Context) {
				handlers = c.HandlerNames()
			})
			group.register(router)

			for _, route := range router.Routes() {
				t.Run(route.Method+" "+route.Path, func(t *testing.T) {
					path := pathParameter.ReplaceAllString(route.Path, "1")
					response := httptest.NewRecorder()
					router.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), route.Method, path, nil))
					if response.Code != http.StatusForbidden {
						t.Fatalf("anonymous request returned HTTP %d, want %d", response.Code, http.StatusForbidden)
					}

					wantRestricted := group.restrictAll || slices.Contains(group.restrictedReads, route.Path)
					if route.Method != http.MethodGet && !slices.Contains(group.allowedWrites, route.Path) {
						wantRestricted = true
					}
					auth := slices.IndexFunc(handlers, func(name string) bool {
						return strings.HasSuffix(name, ".AuthToken")
					})
					restriction := slices.IndexFunc(handlers, func(name string) bool {
						return strings.HasSuffix(name, ".DemoModeRestrict")
					})
					if wantRestricted && (auth < 0 || restriction <= auth || restriction >= len(handlers)-1) {
						t.Errorf("demo restriction must run after authentication and before the endpoint handler, got %v", handlers)
					}
					if !wantRestricted && restriction >= 0 {
						t.Error("route unexpectedly restricts demo access")
					}
				})
			}
		})
	}
}
