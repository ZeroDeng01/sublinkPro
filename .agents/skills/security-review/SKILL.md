# Security Review Skill

## Purpose

This skill provides a comprehensive security review checklist for code changes that involve authentication, authorization, sensitive data handling, or security-critical features.

**When to use**: When making changes to:
- Authentication/authorization logic
- MFA (Multi-Factor Authentication) functionality
- Sensitive data handling (passwords, tokens, API keys, secrets)
- API endpoints with permission requirements
- Input validation and sanitization
- Database queries
- Cryptographic operations
- CORS/CSP configurations
- Session management
- File upload/download functionality

---

## Security Review Checklist

### 🔐 1. Authentication & Authorization

**Check for**:
- [ ] **Authentication bypass**: Can the endpoint/feature be accessed without proper authentication?
- [ ] **Authorization checks**: Are user permissions verified before allowing access?
- [ ] **Role-based access control**: Are roles (admin, user, guest) properly enforced?
- [ ] **Token validation**: Are JWT/API tokens properly validated (signature, expiration, issuer)?
- [ ] **Session security**: Are sessions properly managed (timeout, secure flags, HttpOnly)?

**Example violations**:
```go
// ❌ BAD: No authentication check
func GetUserData(c *gin.Context) {
    userID := c.Param("id")
    user := models.GetUser(userID)
    c.JSON(200, user)
}

// ✅ GOOD: Authentication and authorization
func GetUserData(c *gin.Context) {
    currentUser := middlewares.GetCurrentUser(c)
    if currentUser == nil {
        c.JSON(401, gin.H{"error": "Unauthorized"})
        return
    }
    
    userID := c.Param("id")
    // Users can only access their own data, admins can access any
    if userID != currentUser.ID && !currentUser.IsAdmin {
        c.JSON(403, gin.H{"error": "Forbidden"})
        return
    }
    
    user := models.GetUser(userID)
    c.JSON(200, user)
}
```

---

### 🔑 2. MFA (Multi-Factor Authentication)

**Check for**:
- [ ] **MFA enforcement**: Is MFA required for sensitive operations?
- [ ] **MFA bypass prevention**: Can MFA be circumvented through alternate flows?
- [ ] **Backup codes security**: Are backup codes securely generated and stored?
- [ ] **TOTP secret protection**: Are TOTP secrets encrypted at rest?
- [ ] **Rate limiting**: Is there rate limiting on MFA verification attempts?

**Relevant files**:
- `api/auth_mfa.go`
- `models/mfa.go`
- `middlewares/mfa.go`

---

### 🛡️ 3. Sensitive Data Handling

**Check for**:
- [ ] **Password storage**: Are passwords hashed with bcrypt/argon2 (never plain text)?
- [ ] **Secrets in logs**: Are secrets/tokens masked in logs and error messages?
- [ ] **Secrets in responses**: Are sensitive fields (passwords, tokens) excluded from API responses?
- [ ] **Secrets in transit**: Are sensitive data transmitted over HTTPS only?
- [ ] **Secrets in database**: Are secrets encrypted at rest (API keys, tokens)?

**Example violations**:
```go
// ❌ BAD: Password in plain text
user := &models.User{
    Username: req.Username,
    Password: req.Password, // Plain text!
}

// ✅ GOOD: Password hashed
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
if err != nil {
    return err
}
user := &models.User{
    Username: req.Username,
    Password: string(hashedPassword),
}
```

```go
// ❌ BAD: Exposing sensitive fields
type UserResponse struct {
    ID       uint   `json:"id"`
    Username string `json:"username"`
    Password string `json:"password"` // Never expose this!
    APIKey   string `json:"api_key"`  // Never expose this!
}

// ✅ GOOD: Exclude sensitive fields
type UserResponse struct {
    ID       uint   `json:"id"`
    Username string `json:"username"`
    // Password and APIKey intentionally omitted
}
```

---

### 🔍 4. Input Validation & Sanitization

**Check for**:
- [ ] **SQL injection**: Are SQL queries parameterized (no string concatenation)?
- [ ] **XSS (Cross-Site Scripting)**: Is user input sanitized before rendering?
- [ ] **Command injection**: Are shell commands parameterized (no user input in commands)?
- [ ] **Path traversal**: Are file paths validated (no `../` attacks)?
- [ ] **Type validation**: Are input types validated (strings, numbers, emails, URLs)?
- [ ] **Length validation**: Are input lengths limited to prevent DoS?
- [ ] **Whitelist validation**: Are inputs validated against allowed values?

**Example violations**:
```go
// ❌ BAD: SQL injection vulnerability
query := "SELECT * FROM users WHERE username = '" + username + "'"
db.Raw(query).Scan(&user)

// ✅ GOOD: Parameterized query
db.Where("username = ?", username).First(&user)
```

```go
// ❌ BAD: Path traversal vulnerability
filePath := filepath.Join("/uploads", c.Param("filename"))
file, _ := os.Open(filePath)

// ✅ GOOD: Validate path
filename := filepath.Base(c.Param("filename")) // Remove directory components
filePath := filepath.Join("/uploads", filename)
if !strings.HasPrefix(filePath, "/uploads/") {
    return errors.New("invalid path")
}
file, _ := os.Open(filePath)
```

**Frontend XSS prevention**:
```jsx
{/* ❌ BAD: Unsafe HTML injection */}
<div dangerouslySetInnerHTML={{__html: userInput}} />

{/* ✅ GOOD: React auto-escapes */}
<div>{userInput}</div>

{/* ✅ GOOD: Sanitize if HTML is needed */}
<div dangerouslySetInnerHTML={{__html: DOMPurify.sanitize(userInput)}} />
```

---

### 🗄️ 5. Database Security

**Check for**:
- [ ] **Parameterized queries**: Are all queries parameterized (GORM `Where` with `?`)?
- [ ] **Mass assignment protection**: Are only allowed fields updated?
- [ ] **Soft delete leaks**: Are soft-deleted records excluded from queries?
- [ ] **Permission checks before queries**: Is authorization checked before database access?
- [ ] **Transactions for critical ops**: Are multi-step operations wrapped in transactions?

**Example violations**:
```go
// ❌ BAD: Mass assignment vulnerability
var user models.User
c.BindJSON(&user) // User can set any field, including IsAdmin!
db.Save(&user)

// ✅ GOOD: Explicit field assignment
var req struct {
    Username string `json:"username"`
    Email    string `json:"email"`
}
c.BindJSON(&req)
user.Username = req.Username
user.Email = req.Email
// IsAdmin cannot be modified by user
db.Save(&user)
```

---

### 🌐 6. API Security

**Check for**:
- [ ] **CORS configuration**: Are CORS origins properly restricted (not `*` in production)?
- [ ] **Rate limiting**: Are endpoints rate-limited to prevent abuse?
- [ ] **API key validation**: Are API keys validated before processing requests?
- [ ] **HTTPS enforcement**: Is HTTPS required for sensitive endpoints?
- [ ] **Content-Type validation**: Are Content-Type headers validated?
- [ ] **Error information leakage**: Do error messages avoid exposing internal details?

**Example violations**:
```go
// ❌ BAD: CORS allows all origins
router.Use(cors.New(cors.Config{
    AllowOrigins: []string{"*"},
    AllowCredentials: true,
}))

// ✅ GOOD: CORS restricted to specific origins
router.Use(cors.New(cors.Config{
    AllowOrigins: []string{
        "https://example.com",
        "https://app.example.com",
    },
    AllowCredentials: true,
}))
```

```go
// ❌ BAD: Exposing internal error details
if err != nil {
    c.JSON(500, gin.H{"error": err.Error()}) // May leak stack trace, DB schema, etc.
}

// ✅ GOOD: Generic error message
if err != nil {
    log.Error("Internal error: ", err)
    c.JSON(500, gin.H{"error": "Internal server error"})
}
```

---

### 🔐 7. Cryptography

**Check for**:
- [ ] **Strong algorithms**: Are modern algorithms used (AES-256, bcrypt, argon2)?
- [ ] **Avoid weak algorithms**: No MD5/SHA1 for passwords, no DES/RC4 for encryption
- [ ] **Proper key management**: Are encryption keys stored securely (not in code)?
- [ ] **Secure random generation**: Is `crypto/rand` used (not `math/rand`)?
- [ ] **IV/Salt usage**: Are IVs/salts unique per encryption/hash?

**Example violations**:
```go
// ❌ BAD: Weak hash for passwords
import "crypto/md5"
hash := md5.Sum([]byte(password))

// ✅ GOOD: Strong hash for passwords
import "golang.org/x/crypto/bcrypt"
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

```go
// ❌ BAD: Insecure random
import "math/rand"
token := rand.Intn(999999)

// ✅ GOOD: Cryptographically secure random
import "crypto/rand"
b := make([]byte, 32)
rand.Read(b)
token := hex.EncodeToString(b)
```

---

### 📤 8. File Upload/Download Security

**Check for**:
- [ ] **File type validation**: Are file types validated by content (not just extension)?
- [ ] **File size limits**: Are file sizes limited to prevent DoS?
- [ ] **Filename sanitization**: Are filenames sanitized to prevent path traversal?
- [ ] **Virus scanning**: Are uploaded files scanned for malware (if applicable)?
- [ ] **Storage location**: Are files stored outside the web root?
- [ ] **Download authorization**: Is authorization checked before file download?

**Example violations**:
```go
// ❌ BAD: No file type validation
file, _ := c.FormFile("upload")
c.SaveUploadedFile(file, "/uploads/"+file.Filename)

// ✅ GOOD: Validate file type and sanitize filename
file, _ := c.FormFile("upload")

// Check file size
if file.Size > 10*1024*1024 { // 10MB limit
    return errors.New("file too large")
}

// Check file type by content
fileHeader, _ := file.Open()
defer fileHeader.Close()
buffer := make([]byte, 512)
fileHeader.Read(buffer)
contentType := http.DetectContentType(buffer)
if contentType != "image/png" && contentType != "image/jpeg" {
    return errors.New("invalid file type")
}

// Sanitize filename
safeFilename := filepath.Base(file.Filename)
safeFilename = strings.ReplaceAll(safeFilename, "..", "")
newPath := filepath.Join("/uploads", uuid.New().String()+filepath.Ext(safeFilename))
c.SaveUploadedFile(file, newPath)
```

---

### 🔒 9. Session Management

**Check for**:
- [ ] **Session timeout**: Are sessions expired after inactivity?
- [ ] **Secure flags**: Are session cookies marked as Secure and HttpOnly?
- [ ] **SameSite attribute**: Is SameSite attribute set to prevent CSRF?
- [ ] **Session regeneration**: Are session IDs regenerated after login?
- [ ] **Logout functionality**: Does logout properly invalidate the session?

**Example configuration**:
```go
// ✅ GOOD: Secure session configuration
store := cookie.NewStore([]byte(secretKey))
store.Options(sessions.Options{
    Path:     "/",
    MaxAge:   3600, // 1 hour
    HttpOnly: true,
    Secure:   true, // HTTPS only
    SameSite: http.SameSiteStrictMode,
})
```

---

### 🛡️ 10. Dependency Security

**Check for**:
- [ ] **Known vulnerabilities**: Are dependencies scanned for CVEs?
- [ ] **Dependency versions**: Are dependencies pinned to specific versions?
- [ ] **Minimal dependencies**: Are only necessary dependencies included?
- [ ] **License compliance**: Are dependency licenses compatible with project license?

**Tools**:
```bash
# Go: Check for known vulnerabilities
go list -json -m all | nancy sleuth

# Or use govulncheck (official Go tool)
govulncheck ./...

# Frontend: Check npm dependencies
cd webs && yarn audit
```

---

## Security Review Process

### Step 1: Identify Security-Sensitive Changes
Review the diff and identify if the change involves:
- Authentication/authorization logic
- Sensitive data handling
- User input processing
- Database queries
- API endpoints
- File operations
- Cryptographic operations

### Step 2: Apply Relevant Checklists
Go through the relevant sections above and verify each item.

### Step 3: Test Security Controls
- [ ] **Manual testing**: Try to bypass security controls
- [ ] **Automated testing**: Run security linters (`gosec`, `eslint-plugin-security`)
- [ ] **Dependency scanning**: Check for known vulnerabilities

**Run security linters**:
```bash
# Backend: gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...

# Frontend: eslint with security plugin (if configured)
cd webs
yarn lint
```

### Step 4: Document Security Implications
If the change has security implications:
- [ ] Update `docs/security-guidelines.md` if introducing new security patterns
- [ ] Add security notes to PR description
- [ ] Request security review from another team member for critical changes

---

## Common Security Anti-Patterns

### ❌ Trusting User Input
```go
// Never trust user input directly
userID := c.Query("user_id") // Can be manipulated!
db.Where("id = ?", userID).First(&user) // But parameterized query is safe
```

### ❌ Security by Obscurity
```go
// Don't hide security behind obscure endpoints
router.GET("/admin_secret_panel_xyz", adminHandler) // Still needs proper auth!
```

### ❌ Client-Side Validation Only
```jsx
// Never rely on frontend validation alone
<input type="number" min="0" max="100" /> // Can be bypassed!
// Always validate on backend too
```

### ❌ Logging Sensitive Data
```go
// Never log sensitive data
log.Info("User login: ", username, password) // ❌ Password exposed!
log.Info("User login: ", username) // ✅ No sensitive data
```

---

## Exit Criteria

Before completing the security review:
- [ ] All relevant checklist items have been verified
- [ ] Security linters (`gosec`, ESLint) pass with no critical issues
- [ ] No sensitive data is exposed in logs, errors, or responses
- [ ] Input validation is implemented for all user-controlled data
- [ ] Authorization checks are in place for all sensitive operations
- [ ] Security implications are documented (if any)
- [ ] Tests cover security scenarios (auth failures, permission denials, invalid input)

---

## References

- `docs/security-guidelines.md` - Project security guidelines
- `api/auth.go`, `api/auth_mfa.go` - Authentication implementation
- `middlewares/auth.go`, `middlewares/mfa.go` - Auth middleware
- OWASP Top 10: https://owasp.org/www-project-top-ten/
- Go Security Best Practices: https://github.com/securego/gosec
- React Security Best Practices: https://cheatsheetseries.owasp.org/cheatsheets/React_Security_Cheat_Sheet.html
