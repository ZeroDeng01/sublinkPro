# Cryptography Guide

This guide covers cryptographic best practices for SublinkPro.

## What to Check

- **Strong algorithms**: Are modern algorithms used (AES-256, bcrypt, argon2)?
- **Avoid weak algorithms**: No MD5/SHA1 for passwords, no DES/RC4 for encryption
- **Proper key management**: Are encryption keys stored securely (not in code)?
- **Secure random generation**: Is `crypto/rand` used (not `math/rand`)?
- **IV/Salt usage**: Are IVs/salts unique per encryption/hash?

## Common Vulnerabilities

### Weak Hash Algorithms for Passwords

**Bad Pattern**: Using MD5 or SHA1
```go
// ❌ BAD: Using MD5 or SHA1 for passwords
import "crypto/md5"

func HashPasswordBad(password string) string {
    hash := md5.Sum([]byte(password))
    return hex.EncodeToString(hash[:])
}
```

**Good Pattern**: Using bcrypt
```go
// ✅ GOOD: Using bcrypt
import "golang.org/x/crypto/bcrypt"

func HashPasswordGood(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(hash), nil
}

func VerifyPasswordGood(hashedPassword, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}

// ✅ ALSO GOOD: Using argon2 (stronger but more resource-intensive)
import "golang.org/x/crypto/argon2"

func HashPasswordArgon2(password string) string {
    salt := make([]byte, 16)
    rand.Read(salt)

    hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
    return base64.StdEncoding.EncodeToString(append(salt, hash...))
}
```

### Insecure Random Generation

**Bad Pattern**: Using `math/rand`
```go
// ❌ BAD: Using math/rand for security
import "math/rand"

func GenerateTokenBad() string {
    token := rand.Intn(999999)
    return fmt.Sprintf("%06d", token)
}

func GenerateAPIKeyBad() string {
    const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    b := make([]byte, 32)
    for i := range b {
        b[i] = chars[rand.Intn(len(chars))]
    }
    return string(b)
}
```

**Good Pattern**: Using `crypto/rand`
```go
// ✅ GOOD: Using crypto/rand
import "crypto/rand"

func GenerateTokenGood() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}

func GenerateAPIKeyGood() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}

// ✅ GOOD: Using UUIDs for identifiers (not secrets)
import "github.com/google/uuid"

func GenerateUUID() string {
    return uuid.New().String()
}
```

### Hardcoded Encryption Keys

**Bad Pattern**: Keys in source code
```go
var encryptionKey = []byte("my-secret-key-12345")
```

**Good Pattern**: Keys from environment or secrets manager
```go
encryptionKey := []byte(os.Getenv("ENCRYPTION_KEY"))
if len(encryptionKey) == 0 {
    log.Fatal("ENCRYPTION_KEY not set")
}
```

## Algorithm Recommendations

### Password Hashing
- ✅ **bcrypt** (default choice, well-tested)
- ✅ **argon2** (strongest, higher resource usage)
- ❌ MD5, SHA1, SHA256 (too fast, not designed for passwords)

### Symmetric Encryption
- ✅ **AES-256-GCM** (authenticated encryption)
- ✅ **ChaCha20-Poly1305** (modern, fast)
- ❌ DES, 3DES, RC4 (broken or weak)

### Random Generation
- ✅ **crypto/rand** (cryptographically secure)
- ❌ **math/rand** (predictable, not secure)

### Token Generation
- ✅ **32+ bytes from crypto/rand** (for session tokens, API keys)
- ✅ **UUID v4** (for identifiers, not secrets)

## Best Practices

1. **Use standard libraries**: Prefer Go's `crypto/*` packages over custom implementations
2. **Generate unique IVs/salts**: Never reuse IVs or salts
3. **Use authenticated encryption**: Prefer GCM or Poly1305 modes
4. **Store keys securely**: Use environment variables or secrets managers
5. **Rotate keys regularly**: Have a key rotation strategy for long-term secrets
6. **Use sufficient key lengths**: AES-256, not AES-128
7. **Hash passwords with work factor**: Use bcrypt's cost parameter appropriately

## Relevant Files

- `models/user.go` - Password hashing
- `models/mfa.go` - TOTP secret encryption
- `utils/crypto.go` - Encryption utilities (if applicable)

## References

- Go Cryptography: https://pkg.go.dev/crypto
- OWASP Cryptographic Storage: https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html
- OWASP Password Storage: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
