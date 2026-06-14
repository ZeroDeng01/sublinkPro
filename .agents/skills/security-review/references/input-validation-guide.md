# Input Validation & Sanitization Guide

This guide covers input validation and sanitization patterns to prevent injection attacks in SublinkPro.

## What to Check

- **SQL injection**: Are SQL queries parameterized (no string concatenation)?
- **XSS (Cross-Site Scripting)**: Is user input sanitized before rendering?
- **Command injection**: Are shell commands parameterized (no user input in commands)?
- **Path traversal**: Are file paths validated (no `../` attacks)?
- **Type validation**: Are input types validated (strings, numbers, emails, URLs)?
- **Length validation**: Are input lengths limited to prevent DoS?
- **Whitelist validation**: Are inputs validated against allowed values?

## Common Vulnerabilities

### SQL Injection

**Bad Pattern**: String concatenation in queries
```go
// ❌ BAD: SQL injection vulnerability
func FindUserByUsernameBad(username string) (*models.User, error) {
    var user models.User
    query := "SELECT * FROM users WHERE username = '" + username + "'"
    err := db.Raw(query).Scan(&user).Error
    return &user, err
}
```

**Good Pattern**: Parameterized queries
```go
// ✅ GOOD: Parameterized query
func FindUserByUsernameGood(username string) (*models.User, error) {
    var user models.User
    err := db.Where("username = ?", username).First(&user).Error
    return &user, err
}

// ✅ ALSO GOOD: Named parameters
func FindUserByUsernameAlsoGood(username string) (*models.User, error) {
    var user models.User
    err := db.Where("username = @username", sql.Named("username", username)).First(&user).Error
    return &user, err
}
```

### Path Traversal

**Bad Pattern**: Unsanitized file paths
```go
// ❌ BAD: Path traversal vulnerability
func DownloadFileBad(c *gin.Context) error {
    filename := c.Param("filename")
    filePath := filepath.Join("/uploads", filename)

    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    c.File(filePath)
    return nil
}
```

**Good Pattern**: Validate and sanitize paths
```go
// ✅ GOOD: Validate and sanitize paths
func DownloadFileGood(c *gin.Context) error {
    filename := c.Param("filename")

    // Remove directory components
    safeFilename := filepath.Base(filename)

    // Build path
    filePath := filepath.Join("/uploads", safeFilename)

    // Ensure path is within allowed directory
    if !strings.HasPrefix(filePath, "/uploads/") {
        return errors.New("invalid path")
    }

    // Check file exists
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        return errors.New("file not found")
    }

    c.File(filePath)
    return nil
}
```

### Cross-Site Scripting (XSS)

**Bad Pattern (Frontend)**: Unsafe HTML injection
```jsx
// ❌ BAD: Unsafe HTML injection
function DisplayUserInputBad({ userInput }) {
  return <div dangerouslySetInnerHTML={{__html: userInput}} />;
}
```

**Good Pattern (Frontend)**: React auto-escapes by default
```jsx
// ✅ GOOD: React auto-escapes
function DisplayUserInputGood({ userInput }) {
  return <div>{userInput}</div>;
}

// ✅ GOOD: Sanitize if HTML is needed
import DOMPurify from 'dompurify';

function DisplayUserHTMLGood({ userInput }) {
  const sanitized = DOMPurify.sanitize(userInput);
  return <div dangerouslySetInnerHTML={{__html: sanitized}} />;
}
```

### Command Injection

**Bad Pattern**: User input in shell commands
```go
cmd := exec.Command("sh", "-c", "ls "+userInput)
```

**Good Pattern**: Use command arrays, not shell interpolation
```go
// Don't use shell at all
cmd := exec.Command("ls", userInput)

// Or use proper escaping libraries
```

## Best Practices

1. **Validate all user input**: Never trust client-side validation alone
2. **Use parameterized queries**: Always use `?` placeholders for SQL queries
3. **Whitelist validation**: Validate against allowed values when possible
4. **Length limits**: Enforce maximum lengths to prevent DoS
5. **Type validation**: Ensure inputs match expected types
6. **Sanitize file paths**: Use `filepath.Base()` and validate prefixes
7. **Escape output**: Let frameworks auto-escape (React, templates)
8. **Avoid shell execution**: Use command arrays instead of shell strings

## Relevant Files

- `api/*.go` - API handlers with user input
- `models/*.go` - Database query patterns
- `webs/src/components/*.jsx` - Frontend input handling

## References

- OWASP Input Validation Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html
- OWASP SQL Injection Prevention: https://cheatsheetseries.owasp.org/cheatsheets/SQL_Injection_Prevention_Cheat_Sheet.html
- OWASP XSS Prevention: https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html
