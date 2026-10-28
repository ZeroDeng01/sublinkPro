package backup

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"
)

func TestWriteArchiveIncludesExpectedFilesAndSkipsGeoIP(t *testing.T) {
	root := t.TempDir()
	dbDir := filepath.Join(root, "db-source")
	templateDir := filepath.Join(root, "template-source")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "sublink.db"), []byte("database"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "GeoLite2-City.mmdb"), []byte("large"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "custom.yaml"), []byte("template"), 0o600); err != nil {
		t.Fatal(err)
	}

	var buffer bytes.Buffer
	err := writeArchive(context.Background(), &buffer, []archiveFolder{{dbDir, "db"}, {templateDir, "template"}})
	if err != nil {
		t.Fatalf("writeArchive failed: %v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	if err != nil {
		t.Fatal(err)
	}
	names := make(map[string]bool)
	for _, file := range reader.File {
		names[file.Name] = true
	}
	if !names["db/sublink.db"] || !names["template/custom.yaml"] {
		t.Fatalf("missing expected backup files: %#v", names)
	}
	if names["db/GeoLite2-City.mmdb"] {
		t.Fatal("GeoLite2-City.mmdb should be excluded")
	}
}

func TestWebDAVClientUploadListDownload(t *testing.T) {
	var mu sync.Mutex
	stored := map[string][]byte{}
	modified := time.Now().UTC().Format(http.TimeFormat)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "user" || password != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.Method {
		case "MKCOL":
			w.WriteHeader(http.StatusMethodNotAllowed)
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			mu.Lock()
			stored[r.URL.Path] = body
			mu.Unlock()
			w.WriteHeader(http.StatusCreated)
		case "PROPFIND":
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusMultiStatus)
			_, _ = io.WriteString(w, `<?xml version="1.0"?><d:multistatus xmlns:d="DAV:"><d:response><d:href>/dav/backups/sublink-pro-backup-20260910-120000.zip</d:href><d:propstat><d:prop><d:getcontentlength>7</d:getcontentlength><d:getlastmodified>`+modified+`</d:getlastmodified><d:resourcetype/></d:prop><d:status>HTTP/1.1 200 OK</d:status></d:propstat></d:response></d:multistatus>`)
		case http.MethodGet:
			mu.Lock()
			body, exists := stored[r.URL.Path]
			mu.Unlock()
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = w.Write(body)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL + "/dav", Username: "user", Password: "secret", RemotePath: "backups", TimeoutSeconds: 5, AllowInsecureHTTP: true, AllowPrivateNetwork: true})
	if err != nil {
		t.Fatal(err)
	}
	archivePath := filepath.Join(t.TempDir(), "backup.zip")
	if err := os.WriteFile(archivePath, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	archive := Archive{Path: archivePath, Name: "sublink-pro-backup-20260910-120000.zip", Size: 7, Modified: time.Now()}
	if _, err := client.Upload(context.Background(), archive); err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	files, err := client.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(files) != 1 || files[0].Name != archive.Name || files[0].Size != 7 {
		t.Fatalf("unexpected list: %#v", files)
	}
	var downloaded bytes.Buffer
	if _, err := client.Download(context.Background(), archive.Name, &downloaded); err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	if downloaded.String() != "content" {
		t.Fatalf("downloaded %q", downloaded.String())
	}
}

func TestWebDAVValidationRejectsUnsafeInput(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		config Config
	}{
		{"http without opt-in", Config{BaseURL: "http://example.com/dav", RemotePath: "backup"}},
		{"userinfo", Config{BaseURL: "https://user:pass@example.com/dav", RemotePath: "backup"}},
		{"query", Config{BaseURL: "https://example.com/dav?token=x", RemotePath: "backup"}},
		{"parent path", Config{BaseURL: "https://example.com/dav", RemotePath: "../backup"}},
		{"oversized username", Config{BaseURL: "https://example.com/dav", Username: strings.Repeat("u", 513), RemotePath: "backup"}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := NewClient(testCase.config); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
	for _, filename := range []string{"../backup.zip", `..\\backup.zip`, "backup.db", ""} {
		if err := validateBackupFilename(filename); err == nil {
			t.Fatalf("expected filename %q to be rejected", filename)
		}
	}
}

func TestWebDAVClientDoesNotFollowRedirects(t *testing.T) {
	secondCalled := false
	second := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		secondCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer second.Close()
	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Redirect(w, &http.Request{}, second.URL, http.StatusFound)
	}))
	defer first.Close()

	client, err := NewClient(Config{BaseURL: first.URL, RemotePath: "backup", TimeoutSeconds: 5, AllowInsecureHTTP: true, AllowPrivateNetwork: true})
	if err != nil {
		t.Fatal(err)
	}
	err = client.Test(context.Background())
	if err == nil || !strings.Contains(err.Error(), "重定向") {
		t.Fatalf("expected redirect error, got %v", err)
	}
	if secondCalled {
		t.Fatal("redirect target should not be called")
	}
}

func TestWebDAVEnsureDirectoryAcceptsCollectionRedirect(t *testing.T) {
	var mkcolTargets []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "MKCOL" {
			mkcolTargets = append(mkcolTargets, r.URL.Path)
			w.Header().Set("Location", r.URL.Path)
			w.WriteHeader(http.StatusMovedPermanently)
			return
		}
		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL + "/dav", Username: "user", Password: "secret", RemotePath: "SublinkPro", TimeoutSeconds: 5, AllowInsecureHTTP: true, AllowPrivateNetwork: true})
	if err != nil {
		t.Fatal(err)
	}
	archivePath := filepath.Join(t.TempDir(), "backup.zip")
	if err := os.WriteFile(archivePath, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	archive := Archive{Path: archivePath, Name: "sublink-pro-backup-20260910-120000.zip", Size: 7, Modified: time.Now()}
	if _, err := client.Upload(context.Background(), archive); err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	if len(mkcolTargets) != 1 || mkcolTargets[0] != "/dav/SublinkPro/" {
		t.Fatalf("unexpected MKCOL targets: %#v", mkcolTargets)
	}
}

func TestWebDAVMKCOLRejectsCrossOriginRedirect(t *testing.T) {
	redirectCalled := false
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		redirectCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer redirect.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "MKCOL" {
			t.Fatalf("unexpected request: %s", r.Method)
		}
		w.Header().Set("Location", redirect.URL+"/stolen")
		w.WriteHeader(http.StatusMovedPermanently)
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, Username: "user", Password: "secret", RemotePath: "backup", TimeoutSeconds: 5, AllowInsecureHTTP: true, AllowPrivateNetwork: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.EnsureDirectory(context.Background()); err == nil {
		t.Fatal("expected cross-origin MKCOL redirect to be rejected")
	}
	if redirectCalled {
		t.Fatal("redirect target should not receive credentials or requests")
	}
}

func TestWriteArchiveRejectsSymlinks(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "secret.txt")
	if err := os.WriteFile(target, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(source, "link.txt")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	var buffer bytes.Buffer
	if err := writeArchive(context.Background(), &buffer, []archiveFolder{{source, "db"}}); err == nil {
		t.Fatal("expected symlink to be rejected")
	}
}

func TestWebDAVBlocksPrivateTargetsByDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusMultiStatus)
	}))
	defer server.Close()
	client, err := NewClient(Config{BaseURL: server.URL, RemotePath: "backup", TimeoutSeconds: 5, AllowInsecureHTTP: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Test(context.Background()); err == nil || !strings.Contains(err.Error(), "私有或保留地址") {
		t.Fatalf("expected private target rejection, got %v", err)
	}
}

func TestResolveConfigCanReplaceCorruptedStoredPassword(t *testing.T) {
	oldDB := database.DB
	oldDialect := database.Dialect
	db := testutil.OpenMemoryDB(t, "webdav_settings_test")
	if err := db.AutoMigrate(&models.SystemSetting{}); err != nil {
		t.Fatal(err)
	}
	database.DB = db
	database.Dialect = database.DialectSQLite
	if err := models.InitSettingCache(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		if oldDB != nil {
			_ = models.InitSettingCache()
		}
		testutil.CloseDB(t, db)
	})
	if err := models.SetSetting(settingPasswordEncrypted, "corrupted-ciphertext"); err != nil {
		t.Fatal(err)
	}
	cfg, err := ResolveConfig(ConfigUpdate{
		BaseURL:        "https://dav.example.com",
		Password:       "replacement-password",
		RemotePath:     "backup",
		TimeoutSeconds: 60,
	})
	if err != nil {
		t.Fatalf("replace corrupted password: %v", err)
	}
	if cfg.Password != "replacement-password" {
		t.Fatalf("password = %q", cfg.Password)
	}
	cfg, err = ResolveConfig(ConfigUpdate{
		BaseURL:        "https://dav.example.com",
		ClearPassword:  true,
		RemotePath:     "backup",
		TimeoutSeconds: 60,
	})
	if err != nil {
		t.Fatalf("clear corrupted password: %v", err)
	}
	if cfg.Password != "" {
		t.Fatal("password should be cleared")
	}
}

func TestIsPublicWebDAVIPRejectsSpecialUseRanges(t *testing.T) {
	for _, value := range []string{"127.0.0.1", "10.0.0.1", "100.64.0.1", "192.0.2.1", "198.18.0.1", "203.0.113.1", "2001:db8::1", "fc00::1"} {
		if isPublicWebDAVIP(net.ParseIP(value)) {
			t.Fatalf("special-use address %s should be rejected", value)
		}
	}
	for _, value := range []string{"8.8.8.8", "1.1.1.1", "2606:4700:4700::1111"} {
		if !isPublicWebDAVIP(net.ParseIP(value)) {
			t.Fatalf("public address %s should be allowed", value)
		}
	}
}

func TestWebDAVScheduleValidation(t *testing.T) {
	_, err := normalizeAndValidateConfig(Config{BaseURL: "https://dav.example.com", RemotePath: "backup", ScheduleEnabled: true}, true)
	if err == nil {
		t.Fatal("expected error when schedule enabled without cron")
	}
	cfg, err := normalizeAndValidateConfig(Config{BaseURL: "https://dav.example.com", RemotePath: "backup", ScheduleEnabled: true, CronExpr: "0  3 * * *"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CronExpr != "0 3 * * *" {
		t.Fatalf("cron not cleaned: %q", cfg.CronExpr)
	}
	_, err = normalizeAndValidateConfig(Config{BaseURL: "https://dav.example.com", RemotePath: "backup", ScheduleEnabled: true, CronExpr: "not-a-cron"}, true)
	if err == nil {
		t.Fatal("expected invalid cron error")
	}
	cfg, err = normalizeAndValidateConfig(Config{RemotePath: "backup", ScheduleEnabled: false, CronExpr: ""}, false)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ScheduleEnabled {
		t.Fatal("schedule should remain disabled")
	}
}
