package api

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDownloadSkillArchiveHasInstallableRoot(t *testing.T) {
	engine := gin.New()
	engine.GET("/download", DownloadSkill(os.DirFS("..")))
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/download", nil))
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/zip" {
		t.Fatalf("unexpected response: status %d, type %q", response.Code, response.Header().Get("Content-Type"))
	}
	if got := response.Header().Get("Content-Disposition"); got != "attachment; filename=skill-sublinkpro.zip" {
		t.Fatalf("unexpected Content-Disposition: %q", got)
	}
	archive, err := zip.NewReader(bytes.NewReader(response.Body.Bytes()), int64(response.Body.Len()))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range archive.File {
		if !strings.HasPrefix(entry.Name, "sublinkpro/") {
			t.Errorf("unexpected ZIP entry %q", entry.Name)
		}
	}
	skill, err := fs.ReadFile(archive, "sublinkpro/SKILL.md")
	if err != nil || !bytes.Contains(skill, []byte("\nname: sublinkpro\n")) {
		t.Fatalf("SKILL.md missing or inconsistent with directory name: %v", err)
	}
	for _, name := range []string{"reference/api.md", "reference/deploy.md", "reference/docs.md", "scripts/sublink.py"} {
		if _, err := fs.Stat(archive, "sublinkpro/"+name); err != nil {
			t.Error(err)
		}
	}
}
