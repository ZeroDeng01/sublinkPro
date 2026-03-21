package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sublink/config"
	"sublink/database"
	"time"

	"golang.org/x/crypto/pbkdf2"
	"gorm.io/gorm"
)

const (
	TOTPPeriod               = 30
	TOTPCodeDigits           = 6
	TOTPValidateWindow       = 1
	TOTPRecoveryCodeCount    = 8
	TOTPChallengeTTL         = 5 * time.Minute
	TOTPChallengeMaxAttempts = 5
	totpEncryptionVersion    = "v1"
	totpRecoverySaltLength   = 16
	totpRecoveryIterations   = 120000
)

type TOTPRecoveryCode struct {
	Salt string `json:"salt"`
	Hash string `json:"hash"`
}

func generateRandomBytes(size int) ([]byte, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func GenerateTOTPSecret() (string, error) {
	buf, err := generateRandomBytes(20)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(buf), "="), nil
}

func normalizeBase32Secret(secret string) string {
	secret = strings.ToUpper(strings.TrimSpace(secret))
	if secret == "" {
		return ""
	}
	if mod := len(secret) % 8; mod != 0 {
		secret += strings.Repeat("=", 8-mod)
	}
	return secret
}

func totpCode(secret string, at time.Time) (string, error) {
	decoded, err := base32.StdEncoding.DecodeString(normalizeBase32Secret(secret))
	if err != nil {
		return "", err
	}

	counter := uint64(at.Unix() / TOTPPeriod)
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
	code := binaryCode % 1000000
	return fmt.Sprintf("%06d", code), nil
}

func VerifyTOTPCode(secret, code string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != TOTPCodeDigits {
		return false
	}
	for i := -TOTPValidateWindow; i <= TOTPValidateWindow; i++ {
		candidate, err := totpCode(secret, now.Add(time.Duration(i*TOTPPeriod)*time.Second))
		if err != nil {
			return false
		}
		if hmac.Equal([]byte(candidate), []byte(code)) {
			return true
		}
	}
	return false
}

func encryptionKey() ([]byte, error) {
	keyMaterial := strings.TrimSpace(config.GetAPIEncryptionKey())
	if len(keyMaterial) < 32 {
		return nil, fmt.Errorf("API_ENCRYPTION_KEY 未设置或长度不足，无法安全存储 TOTP 密钥")
	}
	sum := sha256.Sum256([]byte(keyMaterial))
	return sum[:], nil
}

func EncryptTOTPSecret(secret string) (string, error) {
	key, err := encryptionKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce, err := generateRandomBytes(gcm.NonceSize())
	if err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(secret), nil)
	payload := append(nonce, ciphertext...)
	return totpEncryptionVersion + ":" + base64.StdEncoding.EncodeToString(payload), nil
}

func DecryptTOTPSecret(secret string) (string, error) {
	parts := strings.SplitN(strings.TrimSpace(secret), ":", 2)
	if len(parts) != 2 || parts[0] != totpEncryptionVersion {
		return "", fmt.Errorf("不支持的 TOTP 密钥格式")
	}
	key, err := encryptionKey()
	if err != nil {
		return "", err
	}
	payload, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(payload) < gcm.NonceSize() {
		return "", fmt.Errorf("TOTP 密钥数据损坏")
	}
	nonce := payload[:gcm.NonceSize()]
	ciphertext := payload[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func normalizeRecoveryCode(code string) string {
	return strings.TrimSpace(strings.ToUpper(code))
}

func hashRecoveryCode(code, salt string) string {
	derived := pbkdf2.Key([]byte(normalizeRecoveryCode(code)), []byte(salt), totpRecoveryIterations, 32, sha256.New)
	return hex.EncodeToString(derived)
}

func GenerateRecoveryCodes() ([]string, []TOTPRecoveryCode, error) {
	plainCodes := make([]string, 0, TOTPRecoveryCodeCount)
	hashedCodes := make([]TOTPRecoveryCode, 0, TOTPRecoveryCodeCount)
	for i := 0; i < TOTPRecoveryCodeCount; i++ {
		buf, err := generateRandomBytes(5)
		if err != nil {
			return nil, nil, err
		}
		raw := strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf))
		code := raw[:4] + "-" + raw[4:8]
		saltBytes, err := generateRandomBytes(totpRecoverySaltLength)
		if err != nil {
			return nil, nil, err
		}
		salt := base64.StdEncoding.EncodeToString(saltBytes)
		plainCodes = append(plainCodes, code)
		hashedCodes = append(hashedCodes, TOTPRecoveryCode{Salt: salt, Hash: hashRecoveryCode(code, salt)})
	}
	return plainCodes, hashedCodes, nil
}

func marshalRecoveryCodes(codes []TOTPRecoveryCode) (string, error) {
	if len(codes) == 0 {
		return "[]", nil
	}
	buf, err := json.Marshal(codes)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func parseRecoveryCodes(raw string) ([]TOTPRecoveryCode, error) {
	if strings.TrimSpace(raw) == "" {
		return []TOTPRecoveryCode{}, nil
	}
	var codes []TOTPRecoveryCode
	if err := json.Unmarshal([]byte(raw), &codes); err != nil {
		return nil, err
	}
	return codes, nil
}

func (user *User) recoveryCodes() ([]TOTPRecoveryCode, error) {
	return parseRecoveryCodes(user.TOTPRecoveryCodes)
}

func (user *User) pendingRecoveryCodes() ([]TOTPRecoveryCode, error) {
	return parseRecoveryCodes(user.TOTPPendingRecoveryCodes)
}

func (user *User) currentTOTPSecret() (string, error) {
	if strings.TrimSpace(user.TOTPSecret) == "" {
		return "", nil
	}
	return DecryptTOTPSecret(user.TOTPSecret)
}

func (user *User) pendingTOTPSecret() (string, error) {
	if strings.TrimSpace(user.TOTPPendingSecret) == "" {
		return "", nil
	}
	return DecryptTOTPSecret(user.TOTPPendingSecret)
}

func TOTPProvisioningURI(issuer, accountName, secret string) string {
	label := url.PathEscape(fmt.Sprintf("%s:%s", issuer, accountName))
	values := url.Values{}
	values.Set("secret", secret)
	values.Set("issuer", issuer)
	values.Set("algorithm", "SHA1")
	values.Set("digits", strconv.Itoa(TOTPCodeDigits))
	values.Set("period", strconv.Itoa(TOTPPeriod))
	return fmt.Sprintf("otpauth://totp/%s?%s", label, values.Encode())
}

func (user *User) BeginTOTPEnrollment() (string, string, []string, error) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		return "", "", nil, err
	}
	encryptedSecret, err := EncryptTOTPSecret(secret)
	if err != nil {
		return "", "", nil, err
	}
	plainCodes, hashedCodes, err := GenerateRecoveryCodes()
	if err != nil {
		return "", "", nil, err
	}
	encodedCodes, err := marshalRecoveryCodes(hashedCodes)
	if err != nil {
		return "", "", nil, err
	}
	updates := map[string]interface{}{
		"totp_pending_secret":         encryptedSecret,
		"totp_pending_recovery_codes": encodedCodes,
	}
	if err := database.DB.Model(&User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return "", "", nil, err
	}
	user.TOTPPendingSecret = encryptedSecret
	user.TOTPPendingRecoveryCodes = encodedCodes
	userCache.Set(user.ID, *user)
	issuer := "SublinkPro"
	return secret, TOTPProvisioningURI(issuer, user.Username, secret), plainCodes, nil
}

func (user *User) ConfirmTOTPEnrollment(code string, now time.Time) error {
	pendingSecret, err := user.pendingTOTPSecret()
	if err != nil {
		return err
	}
	if pendingSecret == "" {
		return fmt.Errorf("请先开始 TOTP 绑定")
	}
	if !VerifyTOTPCode(pendingSecret, code, now) {
		return fmt.Errorf("验证码错误或已过期")
	}
	updates := map[string]interface{}{
		"totp_enabled":                true,
		"totp_secret":                 user.TOTPPendingSecret,
		"totp_pending_secret":         "",
		"totp_recovery_codes":         user.TOTPPendingRecoveryCodes,
		"totp_pending_recovery_codes": "[]",
	}
	if err := database.DB.Model(&User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return err
	}
	user.TOTPEnabled = true
	user.TOTPSecret = user.TOTPPendingSecret
	user.TOTPPendingSecret = ""
	user.TOTPRecoveryCodes = user.TOTPPendingRecoveryCodes
	user.TOTPPendingRecoveryCodes = "[]"
	userCache.Set(user.ID, *user)
	return nil
}

func (user *User) VerifyTOTPChallenge(code string, now time.Time) error {
	secret, err := user.currentTOTPSecret()
	if err != nil {
		return err
	}
	if secret == "" || !user.TOTPEnabled {
		return fmt.Errorf("用户未启用 TOTP")
	}
	if !VerifyTOTPCode(secret, code, now) {
		return fmt.Errorf("验证码错误或已过期")
	}
	return nil
}

func (user *User) UseRecoveryCode(code string) error {
	if !user.TOTPEnabled {
		return fmt.Errorf("恢复码不可用")
	}
	codes, err := user.recoveryCodes()
	if err != nil {
		return err
	}
	if len(codes) == 0 {
		return fmt.Errorf("恢复码不可用")
	}
	normalized := normalizeRecoveryCode(code)
	remaining := make([]TOTPRecoveryCode, 0, len(codes))
	matched := false
	for _, recoveryCode := range codes {
		candidateHash := hashRecoveryCode(normalized, recoveryCode.Salt)
		if !matched && hmac.Equal([]byte(recoveryCode.Hash), []byte(candidateHash)) {
			matched = true
			continue
		}
		remaining = append(remaining, recoveryCode)
	}
	if !matched {
		return fmt.Errorf("恢复码无效")
	}
	encoded, err := marshalRecoveryCodes(remaining)
	if err != nil {
		return err
	}
	if err := database.DB.Model(&User{}).Where("id = ?", user.ID).Update("totp_recovery_codes", encoded).Error; err != nil {
		return err
	}
	user.TOTPRecoveryCodes = encoded
	userCache.Set(user.ID, *user)
	return nil
}

func (user *User) RegenerateRecoveryCodes() ([]string, error) {
	plainCodes, hashedCodes, err := GenerateRecoveryCodes()
	if err != nil {
		return nil, err
	}
	encoded, err := marshalRecoveryCodes(hashedCodes)
	if err != nil {
		return nil, err
	}
	if err := database.DB.Model(&User{}).Where("id = ?", user.ID).Update("totp_recovery_codes", encoded).Error; err != nil {
		return nil, err
	}
	user.TOTPRecoveryCodes = encoded
	userCache.Set(user.ID, *user)
	return plainCodes, nil
}

func (user *User) DisableTOTP() error {
	updates := map[string]interface{}{
		"totp_enabled":                false,
		"totp_secret":                 "",
		"totp_pending_secret":         "",
		"totp_recovery_codes":         "[]",
		"totp_pending_recovery_codes": "[]",
	}
	if err := database.DB.Model(&User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return err
	}
	user.TOTPEnabled = false
	user.TOTPSecret = ""
	user.TOTPPendingSecret = ""
	user.TOTPRecoveryCodes = "[]"
	user.TOTPPendingRecoveryCodes = "[]"
	userCache.Set(user.ID, *user)
	return nil
}

func (user *User) CountRecoveryCodes() int {
	codes, err := user.recoveryCodes()
	if err != nil {
		return 0
	}
	return len(codes)
}

func FindUserByUsername(username string) (*User, error) {
	user := &User{Username: username}
	if err := user.Find(); err != nil {
		return nil, err
	}
	return user, nil
}

func LoadUserForMFAChallenge(username string) (*User, error) {
	var user User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	userCache.Set(user.ID, user)
	return &user, nil
}

func IsRecordNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}
