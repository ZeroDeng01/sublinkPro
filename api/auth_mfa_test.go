package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"sublink/config"
	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupAuthMFATestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized
	oldCfg := *config.Get()

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "auth_mfa_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.MFALoginChallenge{}); err != nil {
		t.Fatalf("auto migrate users: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	config.UpdateConfig(func(cfg *config.AppConfig) {
		cfg.APIEncryptionKey = "test-api-encryption-key"
		cfg.JwtSecret = "test-jwt-secret"
		cfg.CaptchaMode = config.CaptchaModeDisabled
		cfg.MFAResetSecret = "reset-secret"
	})
	if err := models.InitUserCache(); err != nil {
		t.Fatalf("init user cache: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Exec("DELETE FROM users").Error
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		config.UpdateConfig(func(cfg *config.AppConfig) {
			*cfg = oldCfg
		})
		if oldDB != nil {
			_ = models.InitUserCache()
		}
		testutil.CloseDB(t, db)
	})
}

func performFormRequest(t *testing.T, handler gin.HandlerFunc, values map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	body := bytes.NewBufferString(encodeForm(values))
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/", body)
	context.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler(context)
	return recorder
}

func encodeForm(values map[string]string) string {
	encoded := url.Values{}
	for key, value := range values {
		encoded.Set(key, value)
	}
	return encoded.Encode()
}

func performJSONRequestWithContext(t *testing.T, handler gin.HandlerFunc, body interface{}, username string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	requestBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(requestBody))
	context.Request.Header.Set("Content-Type", "application/json")
	if username != "" {
		context.Set("username", username)
	}
	handler(context)
	return recorder
}

func createMFAEnabledUser(t *testing.T, username, password string) (*models.User, string) {
	t.Helper()
	user := &models.User{Username: username, Password: password, Role: "admin", Nickname: username}
	if err := user.Create(); err != nil {
		t.Fatalf("create user: %v", err)
	}
	secret, _, _, err := user.BeginTOTPEnrollment()
	if err != nil {
		t.Fatalf("begin enrollment: %v", err)
	}
	now := time.Unix(1711111111, 0)
	currentCode, err := totpCodeForTest(secret, now)
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}
	if err := user.ConfirmTOTPEnrollment(currentCode, now); err != nil {
		t.Fatalf("confirm enrollment: %v", err)
	}
	return user, secret
}

func totpCodeForTest(secret string, now time.Time) (string, error) {
	decoded, err := base32.StdEncoding.DecodeString(normalizeBase32SecretForTest(secret))
	if err != nil {
		return "", err
	}
	counter := uint64(now.Unix() / models.TOTPPeriod)
	var counterBuf [8]byte
	binary.BigEndian.PutUint64(counterBuf[:], counter)
	h := hmac.New(sha1.New, decoded)
	if _, err := h.Write(counterBuf[:]); err != nil {
		return "", err
	}
	sum := h.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	binaryCode := (int(sum[offset])&0x7f)<<24 |
		(int(sum[offset+1])&0xff)<<16 |
		(int(sum[offset+2])&0xff)<<8 |
		(int(sum[offset+3]) & 0xff)
	return fmt.Sprintf("%06d", binaryCode%1000000), nil
}

func normalizeBase32SecretForTest(secret string) string {
	for len(secret)%8 != 0 {
		secret += "="
	}
	return secret
}

func TestUserLoginRequiresMFAChallengeForTOTPEnabledUser(t *testing.T) {
	setupAuthMFATestDB(t)
	_, _ = createMFAEnabledUser(t, "admin", "123456")

	recorder := performFormRequest(t, UserLogin, map[string]string{
		"username": "admin",
		"password": "123456",
	})
	response := decodeAPIResponse(t, recorder)
	if response.Code != 200 {
		t.Fatalf("expected success code 200, got %d", response.Code)
	}
	var data struct {
		RequiresMFA    bool     `json:"requiresMFA"`
		ChallengeToken string   `json:"challengeToken"`
		Methods        []string `json:"methods"`
	}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatalf("unmarshal response data: %v", err)
	}
	if !data.RequiresMFA {
		t.Fatal("expected login to require MFA")
	}
	if data.ChallengeToken == "" {
		t.Fatal("expected challenge token to be returned")
	}
	if len(data.Methods) != 2 {
		t.Fatalf("unexpected mfa methods: %v", data.Methods)
	}
}

func TestVerifyRecoveryCodeLoginConsumesRecoveryCodeAndReturnsAccessToken(t *testing.T) {
	setupAuthMFATestDB(t)
	user := &models.User{Username: "admin", Password: "123456", Role: "admin", Nickname: "管理员"}
	if err := user.Create(); err != nil {
		t.Fatalf("create user: %v", err)
	}
	enrollmentSecret, _, recoveryCodes, err := user.BeginTOTPEnrollment()
	if err != nil {
		t.Fatalf("begin enrollment: %v", err)
	}
	now := time.Unix(1711111111, 0)
	code, err := totpCodeForTest(enrollmentSecret, now)
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}
	if err := user.ConfirmTOTPEnrollment(code, now); err != nil {
		t.Fatalf("confirm enrollment: %v", err)
	}
	challengeToken, err := issuePendingMFAChallenge(user)
	if err != nil {
		t.Fatalf("issue challenge: %v", err)
	}

	recorder := performJSONRequestWithContext(t, VerifyRecoveryCodeLogin, map[string]string{
		"challengeToken": challengeToken,
		"recoveryCode":   recoveryCodes[0],
	}, "")
	response := decodeAPIResponse(t, recorder)
	if response.Code != 200 {
		t.Fatalf("expected success code 200, got %d", response.Code)
	}
	var data struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatalf("unmarshal response data: %v", err)
	}
	if data.AccessToken == "" {
		t.Fatal("expected access token after recovery code login")
	}
	refreshed, err := models.FindUserByUsername("admin")
	if err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if refreshed.CountRecoveryCodes() != models.TOTPRecoveryCodeCount-1 {
		t.Fatalf("expected recovery code to be consumed, got %d remaining", refreshed.CountRecoveryCodes())
	}
}

func TestRecoveryCodeCannotBeUsedBeforeEnrollmentConfirmation(t *testing.T) {
	setupAuthMFATestDB(t)
	user := &models.User{Username: "pending-user", Password: "123456", Role: "admin", Nickname: "待确认"}
	if err := user.Create(); err != nil {
		t.Fatalf("create user: %v", err)
	}
	_, _, recoveryCodes, err := user.BeginTOTPEnrollment()
	if err != nil {
		t.Fatalf("begin enrollment: %v", err)
	}
	challengeToken, err := issuePendingMFAChallenge(user)
	if err != nil {
		t.Fatalf("issue challenge: %v", err)
	}
	recorder := performJSONRequestWithContext(t, VerifyRecoveryCodeLogin, map[string]string{
		"challengeToken": challengeToken,
		"recoveryCode":   recoveryCodes[0],
	}, "")
	response := decodeAPIResponse(t, recorder)
	if response.Code == 200 {
		t.Fatal("expected pending enrollment recovery code login to be rejected")
	}
}

func TestMFAChallengeCannotBeReusedAfterSuccess(t *testing.T) {
	setupAuthMFATestDB(t)
	user, secret := createMFAEnabledUser(t, "reuse-user", "123456")
	now := time.Unix(1711111111, 0)
	code, err := totpCodeForTest(secret, now)
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}
	challengeToken, err := issuePendingMFAChallenge(user)
	if err != nil {
		t.Fatalf("issue challenge: %v", err)
	}
	first := performJSONRequestWithContext(t, VerifyTOTPLogin, map[string]string{
		"challengeToken": challengeToken,
		"code":           code,
	}, "")
	if decodeAPIResponse(t, first).Code != 200 {
		t.Fatal("expected first MFA verification to succeed")
	}
	second := performJSONRequestWithContext(t, VerifyTOTPLogin, map[string]string{
		"challengeToken": challengeToken,
		"code":           code,
	}, "")
	if decodeAPIResponse(t, second).Code == 200 {
		t.Fatal("expected reused MFA challenge to be rejected")
	}
}
