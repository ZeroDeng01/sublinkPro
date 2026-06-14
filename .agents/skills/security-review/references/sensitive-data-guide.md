# Sensitive Data Handling Guide

This guide covers secure handling of sensitive data (passwords, tokens, API keys, secrets) in SublinkPro.

## What to Check

- **Password storage**: Are passwords hashed with bcrypt/argon2 (never plain text)?
- **Secrets in logs**: Are secrets/tokens masked in logs and error messages?
- **Secrets in responses**: Are sensitive fields (passwords, tokens) excluded from API responses?
- **Secrets in transit**: Are sensitive data transmitted over HTTPS only?
- **Secrets in database**: Are secrets encrypted at rest (API keys, tokens)?

## Common Vulnerabilities

### Plain Text Password Storage

**Bad Pattern**: Password stored in plain text
```go
// ❌ BAD: Password in plain text
func CreateUserBad(req CreateUserRequest) error {
    user := &models.User{
        Username: req.Username,
        Password: req.Password, // Plain text!
    }
    return db.Create(user).Error
}
```

**Good Pattern**: Password properly hashed
```go
// ✅ GOOD: Password properly hashed
func CreateUserGood(req CreateUserRequest) error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    user := &models.User{
        Username: req.Username,
        Password: string(hashedPassword),
    }
    return db.Create(user).Error
}
```

### Exposing Sensitive Fields in API Responses

**Bad Pattern**: Exposing sensitive fields
```go
// ❌ BAD: Exposing sensitive fields
type UserResponseBad struct {
    ID       uint   `json:"id"`
    Username string `json:"username"`
    Password string `json:"password"` // Never expose this!
    APIKey   string `json:"api_key"`  // Never expose this!
}
```

**Good Pattern**: Exclude sensitive fields
```go
// ✅ GOOD: Exclude sensitive fields
type UserResponseGood struct {
    ID       uint   `json:"id"`
    Username string `json:"username"`
    // Password and APIKey intentionally omitted
}
```

### Logging Sensitive Data

**Bad Pattern**: Logging passwords or tokens
```go
// ❌ BAD: Logging sensitive data
func LoginBad(username, password string) {
    log.Info("User login attempt: ", username, password) // ❌ Password exposed!
}
```

**Good Pattern**: Never log sensitive data
```go
// ✅ GOOD: Never log sensitive data
func LoginGood(username string) {
    log.Info("User login attempt: ", username) // ✅ No sensitive data
}
```

## Best Practices

1. **Hash all passwords**: Use bcrypt or argon2, never plain text or weak hashes (MD5, SHA1)
2. **Encrypt secrets at rest**: Encrypt API keys, tokens, and other secrets in the database
3. **Exclude from responses**: Never include passwords or secrets in API responses
4. **Mask in logs**: Mask or exclude secrets from all logging output
5. **Use HTTPS**: Always transmit sensitive data over encrypted connections
6. **Separate storage**: Consider using a secrets manager for highly sensitive data

## Relevant Files

- `models/user.go` - User password handling
- `models/mfa.go` - TOTP secret storage
- `utils/crypto.go` - Encryption utilities (if applicable)

## References

- OWASP Password Storage Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
- OWASP Cryptographic Storage Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html
