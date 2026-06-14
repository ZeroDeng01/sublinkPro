---
name: security-review
description: "Security review checklist for authentication, authorization, and sensitive data handling. Use when changing auth, MFA, secrets, input validation, or security-critical features. Not for general code review."
version: "2.0.0"
author: "SublinkPro Team"
user-invocable: true
---

# Security Review Skill

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

**Detailed guide**: `references/authentication-guide.md` (includes code examples)

**Relevant files**:
- `api/auth.go`
- `api/auth_mfa.go`
- `middlewares/auth.go`
- `middlewares/mfa.go`

---

### 🔑 2. MFA (Multi-Factor Authentication)

**Check for**:
- [ ] **MFA enforcement**: Is MFA required for sensitive operations?
- [ ] **MFA bypass prevention**: Can MFA be circumvented through alternate flows?
- [ ] **Backup codes security**: Are backup codes securely generated and stored?
- [ ] **TOTP secret protection**: Are TOTP secrets encrypted at rest?
- [ ] **Rate limiting**: Is there rate limiting on MFA verification attempts?

**Detailed guide**: `references/mfa-guide.md`

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

**Detailed guide**: `references/sensitive-data-guide.md` (includes code examples)

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

**Detailed guide**: `references/input-validation-guide.md` (includes code examples)

---

### 🗄️ 5. Database Security

**Check for**:
- [ ] **Parameterized queries**: Are all queries parameterized (GORM `Where` with `?`)?
- [ ] **Mass assignment protection**: Are only allowed fields updated?
- [ ] **Soft delete leaks**: Are soft-deleted records excluded from queries?
- [ ] **Permission checks before queries**: Is authorization checked before database access?
- [ ] **Transactions for critical ops**: Are multi-step operations wrapped in transactions?

**Detailed guide**: `references/database-security-guide.md` (includes code examples)

---

### 🌐 6. API Security

**Check for**:
- [ ] **CORS configuration**: Are CORS origins properly restricted (not `*` in production)?
- [ ] **Rate limiting**: Are endpoints rate-limited to prevent abuse?
- [ ] **API key validation**: Are API keys validated before processing requests?
- [ ] **HTTPS enforcement**: Is HTTPS required for sensitive endpoints?
- [ ] **Content-Type validation**: Are Content-Type headers validated?
- [ ] **Error information leakage**: Do error messages avoid exposing internal details?

**Detailed guide**: `references/api-security-guide.md` (includes code examples)

---

### 🔐 7. Cryptography

**Check for**:
- [ ] **Strong algorithms**: Are modern algorithms used (AES-256, bcrypt, argon2)?
- [ ] **Avoid weak algorithms**: No MD5/SHA1 for passwords, no DES/RC4 for encryption
- [ ] **Proper key management**: Are encryption keys stored securely (not in code)?
- [ ] **Secure random generation**: Is `crypto/rand` used (not `math/rand`)?
- [ ] **IV/Salt usage**: Are IVs/salts unique per encryption/hash?

**Detailed guide**: `references/cryptography-guide.md` (includes code examples)

---

### 📤 8. File Upload/Download Security

**Check for**:
- [ ] **File type validation**: Are file types validated by content (not just extension)?
- [ ] **File size limits**: Are file sizes limited to prevent DoS?
- [ ] **Filename sanitization**: Are filenames sanitized to prevent path traversal?
- [ ] **Virus scanning**: Are uploaded files scanned for malware (if applicable)?
- [ ] **Storage location**: Are files stored outside the web root?
- [ ] **Download authorization**: Is authorization checked before file download?

**Detailed guide**: `references/file-security-guide.md` (includes code examples)

---

### 🔒 9. Session Management

**Check for**:
- [ ] **Session timeout**: Are sessions expired after inactivity?
- [ ] **Secure flags**: Are session cookies marked as Secure and HttpOnly?
- [ ] **SameSite attribute**: Is SameSite attribute set to prevent CSRF?
- [ ] **Session regeneration**: Are session IDs regenerated after login?
- [ ] **Logout functionality**: Does logout properly invalidate the session?

**Detailed guide**: `references/session-management-guide.md` (includes code examples)

---

### 🛡️ 10. Dependency Security

**Check for**:
- [ ] **Known vulnerabilities**: Are dependencies scanned for CVEs?
- [ ] **Dependency versions**: Are dependencies pinned to specific versions?
- [ ] **Minimal dependencies**: Are only necessary dependencies included?
- [ ] **License compliance**: Are dependency licenses compatible with project license?

**Detailed guide**: `references/dependency-security-guide.md`

**Scan commands**:
```bash
# Go: Check for known vulnerabilities
govulncheck ./...

# Frontend: Check npm dependencies
cd webs && yarn audit
```

---

## Security Review Process

### Step 1: Identify Security-Sensitive Changes
Review the diff and identify if the change involves authentication, sensitive data, user input, database queries, API endpoints, file operations, or cryptographic operations.

### Step 2: Apply Relevant Checklists
Go through the relevant sections above and verify each item. Consult detailed guides in `references/` for in-depth coverage.

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

- **Trusting user input**: Never trust user input directly; always validate and sanitize
- **Security by obscurity**: Obscure endpoints still need proper authentication
- **Client-side validation only**: Always validate on backend; frontend validation can be bypassed
- **Logging sensitive data**: Never log passwords, tokens, or API keys

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

- **Project security guidelines**: `docs/security-guidelines.md`
- **Detailed guides with code examples**: `references/*.md`
- **OWASP Top 10**: https://owasp.org/www-project-top-ten/
- **Go Security Best Practices**: https://github.com/securego/gosec
- **React Security**: https://cheatsheetseries.owasp.org/cheatsheets/React_Security_Cheat_Sheet.html
