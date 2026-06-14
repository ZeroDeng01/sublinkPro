# Authentication & Authorization Guide

This guide covers authentication and authorization security patterns for SublinkPro.

## What to Check

- **Authentication bypass**: Can the endpoint/feature be accessed without proper authentication?
- **Authorization checks**: Are user permissions verified before allowing access?
- **Role-based access control**: Are roles (admin, user, guest) properly enforced?
- **Token validation**: Are JWT/API tokens properly validated (signature, expiration, issuer)?
- **Session security**: Are sessions properly managed (timeout, secure flags, HttpOnly)?

## Common Vulnerabilities

### Missing Authentication Check

**Bad Pattern**: No authentication check
```go
// ❌ BAD: No authentication check
func GetUserData(c *gin.Context) {
    userID := c.Param("id")
    user := models.GetUser(userID)
    c.JSON(200, user)
}
```

**Good Pattern**: Proper authentication and authorization
```go
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

## Relevant Files

- `api/auth.go` - Main authentication handlers
- `api/auth_mfa.go` - MFA authentication
- `middlewares/auth.go` - Authentication middleware
- `middlewares/mfa.go` - MFA middleware

## Best Practices

1. **Always authenticate first**: Check authentication before any business logic
2. **Separate authentication from authorization**: Authentication verifies identity, authorization verifies permissions
3. **Use middleware for common checks**: Apply authentication middleware to route groups
4. **Fail securely**: Default to denying access if auth state is unclear
5. **Log auth failures**: Track failed authentication attempts for security monitoring

## References

- OWASP Authentication Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
- OWASP Authorization Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html
