package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sublink/cache"
	"sublink/database"
	"sublink/models"

	"github.com/gin-gonic/gin"
)

func performTemplateFormRequest(t *testing.T, handler gin.HandlerFunc, values url.Values) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(values.Encode()))
	ginContext.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ginContext.Set("username", "admin")
	handler(ginContext)
	return recorder
}

func performTemplateQueryRequest(t *testing.T, handler gin.HandlerFunc, query url.Values, params gin.Params) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?"+query.Encode(), nil)
	ginContext.Set("username", "admin")
	ginContext.Params = params
	handler(ginContext)
	return recorder
}

func loadTemplateVersionsForTest(t *testing.T) []models.TemplateVersion {
	t.Helper()
	var versions []models.TemplateVersion
	if err := database.DB.Order("version ASC").Find(&versions).Error; err != nil {
		t.Fatalf("load template versions: %v", err)
	}
	return versions
}

func TestTemplateSaveCreatesVersionsRestoresAndDeletesHistory(t *testing.T) {
	setupTemplateAPITestDB(t)
	templateDir := t.TempDir()
	baseTemplateDir = templateDir

	addResponse := decodeAPIResponse(t, performTemplateFormRequest(t, AddTemp, url.Values{
		"filename": {"clash.yaml"},
		"text":     {"version-one\n"},
		"category": {"clash"},
	}))
	if addResponse.Code != 200 {
		t.Fatalf("add template failed: %s", string(addResponse.Data))
	}
	versions := loadTemplateVersionsForTest(t)
	if len(versions) != 1 || versions[0].Version != 1 || versions[0].Source != models.TemplateVersionSourceAdd {
		t.Fatalf("unexpected versions after add: %#v", versions)
	}

	updateResponse := decodeAPIResponse(t, performTemplateFormRequest(t, UpdateTemp, url.Values{
		"oldname":  {"clash.yaml"},
		"filename": {"clash.yaml"},
		"text":     {"version-two\n"},
		"category": {"clash"},
	}))
	if updateResponse.Code != 200 {
		t.Fatalf("update template failed: %s", updateResponse.Msg)
	}
	versions = loadTemplateVersionsForTest(t)
	if len(versions) != 2 || versions[1].Version != 2 || versions[1].Source != models.TemplateVersionSourceSave {
		t.Fatalf("unexpected versions after update: %#v", versions)
	}

	renameResponse := decodeAPIResponse(t, performTemplateFormRequest(t, UpdateTemp, url.Values{
		"oldname":  {"clash.yaml"},
		"filename": {"clash-renamed.yaml"},
		"text":     {"version-three\n"},
		"category": {"clash"},
	}))
	if renameResponse.Code != 200 {
		t.Fatalf("rename template failed: %s", renameResponse.Msg)
	}
	versions = loadTemplateVersionsForTest(t)
	if len(versions) != 3 || versions[2].Source != models.TemplateVersionSourceRename {
		t.Fatalf("unexpected versions after rename: %#v", versions)
	}

	listResponse := decodeAPIResponse(t, performTemplateQueryRequest(t, GetTemplateVersions, url.Values{
		"filename": {"clash-renamed.yaml"},
	}, nil))
	if listResponse.Code != 200 {
		t.Fatalf("list versions failed: %s", listResponse.Msg)
	}
	var listPayload struct {
		Items      []templateVersionItem `json:"items"`
		Total      int64                 `json:"total"`
		TemplateID int                   `json:"templateId"`
	}
	if err := json.Unmarshal(listResponse.Data, &listPayload); err != nil {
		t.Fatalf("unmarshal version list: %v", err)
	}
	if listPayload.Total != 3 || len(listPayload.Items) != 3 || listPayload.TemplateID == 0 {
		t.Fatalf("unexpected version list payload: %#v", listPayload)
	}
	if listPayload.Items[0].Version != 3 || !listPayload.Items[0].IsCurrent {
		t.Fatalf("latest version must be current: %#v", listPayload.Items[0])
	}
	if listPayload.Items[2].Version != 1 || listPayload.Items[2].IsCurrent {
		t.Fatalf("oldest version must not be current: %#v", listPayload.Items[2])
	}

	detailResponse := decodeAPIResponse(t, performTemplateQueryRequest(t, GetTemplateVersion, url.Values{
		"filename": {"clash-renamed.yaml"},
	}, gin.Params{{Key: "version", Value: "1"}}))
	if detailResponse.Code != 200 {
		t.Fatalf("get version failed: %s", detailResponse.Msg)
	}
	var detailPayload struct {
		Version models.TemplateVersion `json:"version"`
		Content string                 `json:"content"`
	}
	if err := json.Unmarshal(detailResponse.Data, &detailPayload); err != nil {
		t.Fatalf("unmarshal version detail: %v", err)
	}
	if detailPayload.Content != "version-one\n" {
		t.Fatalf("version content = %q", detailPayload.Content)
	}

	downloadRecorder := performTemplateQueryRequest(t, DownloadTemplateVersion, url.Values{
		"filename": {"clash-renamed.yaml"},
	}, gin.Params{{Key: "version", Value: "1"}})
	if downloadRecorder.Code != http.StatusOK {
		t.Fatalf("download status = %d", downloadRecorder.Code)
	}
	if body := downloadRecorder.Body.String(); body != "version-one\n" {
		t.Fatalf("download body = %q", body)
	}
	if disposition := downloadRecorder.Header().Get("Content-Disposition"); !strings.Contains(disposition, `clash-renamed.v1.yaml`) {
		t.Fatalf("unexpected content disposition: %q", disposition)
	}

	diffResponse := decodeAPIResponse(t, performTemplateQueryRequest(t, GetTemplateVersionDiff, url.Values{
		"filename": {"clash-renamed.yaml"},
		"from":     {"1"},
		"to":       {"3"},
	}, nil))
	if diffResponse.Code != 200 {
		t.Fatalf("diff versions failed: %s", diffResponse.Msg)
	}
	var diffPayload struct {
		From        int    `json:"from"`
		To          int    `json:"to"`
		FromContent string `json:"fromContent"`
		ToContent   string `json:"toContent"`
		Diff        string `json:"diff"`
		Additions   int    `json:"additions"`
		Deletions   int    `json:"deletions"`
		Identical   bool   `json:"identical"`
	}
	if err := json.Unmarshal(diffResponse.Data, &diffPayload); err != nil {
		t.Fatalf("unmarshal diff: %v", err)
	}
	if diffPayload.FromContent != "version-one\n" || diffPayload.ToContent != "version-three\n" {
		t.Fatalf("unexpected diff contents: %#v", diffPayload)
	}
	if diffPayload.Identical || diffPayload.Additions == 0 || diffPayload.Deletions == 0 {
		t.Fatalf("unexpected diff stats: %#v", diffPayload)
	}

	restoreResponse := decodeAPIResponse(t, performTemplateFormRequest(t, RestoreTemplateVersion, url.Values{
		"filename": {"clash-renamed.yaml"},
		"version":  {"1"},
	}))
	if restoreResponse.Code != 200 {
		t.Fatalf("restore version failed: %s", restoreResponse.Msg)
	}
	restoredFile, err := os.ReadFile(filepath.Join(templateDir, "clash-renamed.yaml"))
	if err != nil {
		t.Fatalf("read restored template: %v", err)
	}
	if string(restoredFile) != "version-one\n" {
		t.Fatalf("restored content = %q", string(restoredFile))
	}
	if cached, ok := cache.GetTemplateContent("clash-renamed.yaml"); !ok || cached != "version-one\n" {
		t.Fatalf("template cache not updated after restore: %q ok=%v", cached, ok)
	}
	versions = loadTemplateVersionsForTest(t)
	if len(versions) != 4 || versions[3].Source != models.TemplateVersionSourceRestore {
		t.Fatalf("unexpected versions after restore: %#v", versions)
	}

	var tmpl models.Template
	if err := tmpl.FindByName("clash-renamed.yaml"); err != nil {
		t.Fatalf("find template before delete: %v", err)
	}
	templateID := tmpl.ID

	deleteResponse := decodeAPIResponse(t, performTemplateFormRequest(t, DelTemp, url.Values{
		"filename": {"clash-renamed.yaml"},
	}))
	if deleteResponse.Code != 200 {
		t.Fatalf("delete template failed: %s", deleteResponse.Msg)
	}
	if remaining := loadTemplateVersionsForTest(t); len(remaining) != 0 {
		t.Fatalf("versions still present after delete: %#v", remaining)
	}
	if _, err := os.Stat(models.TemplateVersionDir(templateDir, templateID)); !os.IsNotExist(err) {
		t.Fatalf("version dir still exists after delete: %v", err)
	}
}

func TestTemplateUpdateCreatesBaselineForExistingTemplate(t *testing.T) {
	setupTemplateAPITestDB(t)
	templateDir := t.TempDir()
	baseTemplateDir = templateDir
	writeTemplateFileForTest(t, templateDir, "clash.yaml", "original-content\n")

	tmpl := models.Template{Name: "clash.yaml", Category: "clash"}
	if err := tmpl.Add(); err != nil {
		t.Fatalf("add template metadata: %v", err)
	}

	response := decodeAPIResponse(t, performTemplateFormRequest(t, UpdateTemp, url.Values{
		"oldname":  {"clash.yaml"},
		"filename": {"clash.yaml"},
		"text":     {"updated-content\n"},
		"category": {"clash"},
	}))
	if response.Code != 200 {
		t.Fatalf("update template failed: %s", response.Msg)
	}

	versions := loadTemplateVersionsForTest(t)
	if len(versions) != 2 {
		t.Fatalf("expected baseline + save versions, got %#v", versions)
	}
	if versions[0].Source != models.TemplateVersionSourceBaseline {
		t.Fatalf("first version source = %q, want baseline", versions[0].Source)
	}
	baselineContent, err := models.ReadTemplateVersionContent(templateDir, &versions[0])
	if err != nil {
		t.Fatalf("read baseline content: %v", err)
	}
	if string(baselineContent) != "original-content\n" {
		t.Fatalf("baseline content = %q", string(baselineContent))
	}
	if versions[1].Source != models.TemplateVersionSourceSave {
		t.Fatalf("second version source = %q, want save", versions[1].Source)
	}
}

func TestTemplateUpdateWithIdenticalContentKeepsSingleVersion(t *testing.T) {
	setupTemplateAPITestDB(t)
	templateDir := t.TempDir()
	baseTemplateDir = templateDir

	if response := decodeAPIResponse(t, performTemplateFormRequest(t, AddTemp, url.Values{
		"filename": {"clash.yaml"},
		"text":     {"same-content\n"},
		"category": {"clash"},
	})); response.Code != 200 {
		t.Fatalf("add template failed: %s", response.Msg)
	}

	if response := decodeAPIResponse(t, performTemplateFormRequest(t, UpdateTemp, url.Values{
		"oldname":  {"clash.yaml"},
		"filename": {"clash.yaml"},
		"text":     {"same-content\n"},
		"category": {"clash"},
	})); response.Code != 200 {
		t.Fatalf("update template failed: %s", response.Msg)
	}

	if versions := loadTemplateVersionsForTest(t); len(versions) != 1 {
		t.Fatalf("expected single version for identical content, got %#v", versions)
	}
}

func TestGetTemplateVersionsWithoutMetadataReturnsEmptyHistory(t *testing.T) {
	setupTemplateAPITestDB(t)
	templateDir := t.TempDir()
	baseTemplateDir = templateDir
	writeTemplateFileForTest(t, templateDir, "orphan.yaml", "rules: []\n")

	response := decodeAPIResponse(t, performTemplateQueryRequest(t, GetTemplateVersions, url.Values{
		"filename": {"orphan.yaml"},
	}, nil))
	if response.Code != 200 {
		t.Fatalf("list versions failed: %s", response.Msg)
	}
	var payload struct {
		Items      []templateVersionItem `json:"items"`
		Total      int64                 `json:"total"`
		TemplateID int                   `json:"templateId"`
	}
	if err := json.Unmarshal(response.Data, &payload); err != nil {
		t.Fatalf("unmarshal empty history: %v", err)
	}
	if payload.Total != 0 || len(payload.Items) != 0 || payload.TemplateID != 0 {
		t.Fatalf("unexpected empty history payload: %#v", payload)
	}
}

func TestTemplateVersionEndpointsRejectPathTraversal(t *testing.T) {
	setupTemplateAPITestDB(t)
	baseTemplateDir = t.TempDir()

	response := decodeAPIResponse(t, performTemplateQueryRequest(t, GetTemplateVersions, url.Values{
		"filename": {"../evil.yaml"},
	}, nil))
	if response.Code == 200 {
		t.Fatalf("expected path traversal to be rejected, got %#v", response)
	}

	if versions := loadTemplateVersionsForTest(t); len(versions) != 0 {
		t.Fatalf("unexpected versions created: %#v", versions)
	}
}

func TestTemplateVersionDiffAgainstCurrentContent(t *testing.T) {
	setupTemplateAPITestDB(t)
	templateDir := t.TempDir()
	baseTemplateDir = templateDir

	if response := decodeAPIResponse(t, performTemplateFormRequest(t, AddTemp, url.Values{
		"filename": {"clash.yaml"},
		"text":     {"line-one\n"},
		"category": {"clash"},
	})); response.Code != 200 {
		t.Fatalf("add template failed: %s", response.Msg)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "clash.yaml"), []byte("line-one\nline-two\n"), 0600); err != nil {
		t.Fatalf("write current template: %v", err)
	}

	response := decodeAPIResponse(t, performTemplateQueryRequest(t, GetTemplateVersionDiff, url.Values{
		"filename": {"clash.yaml"},
		"from":     {"1"},
		"to":       {"0"},
	}, nil))
	if response.Code != 200 {
		t.Fatalf("diff against current failed: %s", response.Msg)
	}
	var payload struct {
		To        int  `json:"to"`
		Additions int  `json:"additions"`
		Identical bool `json:"identical"`
	}
	if err := json.Unmarshal(response.Data, &payload); err != nil {
		t.Fatalf("unmarshal diff payload: %v", err)
	}
	if payload.To != 0 || payload.Identical || payload.Additions == 0 {
		t.Fatalf("unexpected diff against current: %#v", payload)
	}
}

func TestDelTempRemovesMetadataWhenFileAlreadyDeleted(t *testing.T) {
	setupTemplateAPITestDB(t)
	templateDir := t.TempDir()
	baseTemplateDir = templateDir

	if response := decodeAPIResponse(t, performTemplateFormRequest(t, AddTemp, url.Values{
		"filename": {"clash.yaml"},
		"text":     {"rules: []\n"},
		"category": {"clash"},
	})); response.Code != 200 {
		t.Fatalf("add template failed: %s", response.Msg)
	}

	// 模拟文件被外部删除（例如用户直接在挂载目录中删除了默认模板），
	// 此时界面上的删除操作仍应清理元数据与历史版本，而不是报错。
	if err := os.Remove(filepath.Join(templateDir, "clash.yaml")); err != nil {
		t.Fatalf("remove template file: %v", err)
	}

	response := decodeAPIResponse(t, performTemplateFormRequest(t, DelTemp, url.Values{
		"filename": {"clash.yaml"},
	}))
	if response.Code != 200 {
		t.Fatalf("delete template with missing file failed: %s", response.Msg)
	}
	if versions := loadTemplateVersionsForTest(t); len(versions) != 0 {
		t.Fatalf("versions still present after delete: %#v", versions)
	}

	var tmpl models.Template
	if err := tmpl.FindByName("clash.yaml"); err == nil {
		t.Fatal("template metadata still present after delete")
	}
}
