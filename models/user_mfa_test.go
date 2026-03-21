package models

import (
	"strings"
	"testing"
	"time"

	"sublink/config"
	"sublink/database"
	"sublink/internal/testutil"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupUserMFATestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized
	oldCfg := *config.Get()

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "user_mfa_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&User{}, &MFALoginChallenge{}); err != nil {
		t.Fatalf("auto migrate users: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	config.UpdateConfig(func(cfg *config.AppConfig) {
		cfg.APIEncryptionKey = "test-api-encryption-key"
		cfg.JwtSecret = "test-jwt-secret"
	})
	if err := InitUserCache(); err != nil {
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
			_ = InitUserCache()
		}
		testutil.CloseDB(t, db)
	})
}

func TestTOTPEnrollmentConfirmationAndRecoveryCodeUse(t *testing.T) {
	setupUserMFATestDB(t)

	user := &User{Username: "admin", Password: "123456", Role: "admin", Nickname: "管理员"}
	if err := user.Create(); err != nil {
		t.Fatalf("create user: %v", err)
	}

	secret, provisioningURI, recoveryCodes, err := user.BeginTOTPEnrollment()
	if err != nil {
		t.Fatalf("begin enrollment: %v", err)
	}
	if secret == "" {
		t.Fatal("expected secret to be returned")
	}
	if !strings.Contains(provisioningURI, "otpauth://totp/") {
		t.Fatalf("unexpected provisioning URI: %s", provisioningURI)
	}
	if len(recoveryCodes) != TOTPRecoveryCodeCount {
		t.Fatalf("unexpected recovery code count: %d", len(recoveryCodes))
	}
	if user.CountRecoveryCodes() != 0 {
		t.Fatalf("expected active recovery codes to remain unavailable before confirmation, got %d", user.CountRecoveryCodes())
	}

	now := time.Unix(1711111111, 0)
	code, err := totpCode(secret, now)
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}
	if err := user.ConfirmTOTPEnrollment(code, now); err != nil {
		t.Fatalf("confirm enrollment: %v", err)
	}
	if !user.TOTPEnabled {
		t.Fatal("expected TOTP to be enabled after confirmation")
	}
	if user.CountRecoveryCodes() != TOTPRecoveryCodeCount {
		t.Fatalf("unexpected recovery count after confirmation: %d", user.CountRecoveryCodes())
	}

	if err := user.VerifyTOTPChallenge(code, now); err != nil {
		t.Fatalf("verify totp challenge: %v", err)
	}
	if err := user.UseRecoveryCode(recoveryCodes[0]); err != nil {
		t.Fatalf("use recovery code: %v", err)
	}
	if user.CountRecoveryCodes() != TOTPRecoveryCodeCount-1 {
		t.Fatalf("unexpected remaining recovery count: %d", user.CountRecoveryCodes())
	}
	if err := user.UseRecoveryCode(recoveryCodes[0]); err == nil {
		t.Fatal("expected used recovery code to be rejected")
	}
}

func TestEncryptTOTPSecretUsesVersionedAEADFormat(t *testing.T) {
	setupUserMFATestDB(t)

	encrypted, err := EncryptTOTPSecret("JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("encrypt secret: %v", err)
	}
	if !strings.HasPrefix(encrypted, "v1:") {
		t.Fatalf("expected versioned encrypted payload, got %s", encrypted)
	}
	decrypted, err := DecryptTOTPSecret(encrypted)
	if err != nil {
		t.Fatalf("decrypt secret: %v", err)
	}
	if decrypted != "JBSWY3DPEHPK3PXP" {
		t.Fatalf("unexpected decrypted value: %s", decrypted)
	}
}

func TestDisableTOTPResetsSecretsAndRecoveryCodes(t *testing.T) {
	setupUserMFATestDB(t)

	user := &User{Username: "mfa-user", Password: "abcdef", Role: "admin", Nickname: "MFA"}
	if err := user.Create(); err != nil {
		t.Fatalf("create user: %v", err)
	}

	secret, _, _, err := user.BeginTOTPEnrollment()
	if err != nil {
		t.Fatalf("begin enrollment: %v", err)
	}
	now := time.Unix(1712222222, 0)
	code, err := totpCode(secret, now)
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}
	if err := user.ConfirmTOTPEnrollment(code, now); err != nil {
		t.Fatalf("confirm enrollment: %v", err)
	}
	if err := user.DisableTOTP(); err != nil {
		t.Fatalf("disable totp: %v", err)
	}
	if user.TOTPEnabled {
		t.Fatal("expected TOTP to be disabled")
	}
	if user.TOTPSecret != "" || user.TOTPPendingSecret != "" {
		t.Fatal("expected secrets to be cleared")
	}
	if user.CountRecoveryCodes() != 0 {
		t.Fatalf("expected recovery codes to be cleared, got %d", user.CountRecoveryCodes())
	}
	if err := user.VerifyTOTPChallenge(code, now); err == nil {
		t.Fatal("expected TOTP challenge to fail after disable")
	}
}
