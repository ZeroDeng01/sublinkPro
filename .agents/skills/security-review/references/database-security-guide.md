# Database Security Guide

This guide covers database security patterns for SublinkPro, focusing on GORM usage.

## What to Check

- **Parameterized queries**: Are all queries parameterized (GORM `Where` with `?`)?
- **Mass assignment protection**: Are only allowed fields updated?
- **Soft delete leaks**: Are soft-deleted records excluded from queries?
- **Permission checks before queries**: Is authorization checked before database access?
- **Transactions for critical ops**: Are multi-step operations wrapped in transactions?

## Common Vulnerabilities

### Mass Assignment Vulnerability

**Bad Pattern**: Binding directly to model
```go
// ❌ BAD: Mass assignment vulnerability
func UpdateUserBad(c *gin.Context) error {
    var user models.User
    c.BindJSON(&user) // User can set any field, including IsAdmin!
    return db.Save(&user).Error
}
```

**Good Pattern**: Explicit field assignment
```go
// ✅ GOOD: Explicit field assignment
func UpdateUserGood(c *gin.Context) error {
    var req struct {
        Username string `json:"username"`
        Email    string `json:"email"`
    }
    if err := c.BindJSON(&req); err != nil {
        return err
    }

    var user models.User
    if err := db.First(&user, c.Param("id")).Error; err != nil {
        return err
    }

    // Only update allowed fields
    user.Username = req.Username
    user.Email = req.Email
    // IsAdmin cannot be modified by user

    return db.Save(&user).Error
}

// ✅ ALTERNATIVE: Using GORM's Select for partial updates
func UpdateUserAlsoGood(c *gin.Context) error {
    var req struct {
        Username string `json:"username"`
        Email    string `json:"email"`
    }
    if err := c.BindJSON(&req); err != nil {
        return err
    }

    return db.Model(&models.User{}).
        Where("id = ?", c.Param("id")).
        Select("Username", "Email").
        Updates(req).Error
}
```

### Missing Parameterization

Even with GORM, be careful:

**Bad Pattern**: String concatenation
```go
db.Where("username = '" + username + "'").First(&user)
```

**Good Pattern**: Use placeholders
```go
db.Where("username = ?", username).First(&user)
```

### Soft Delete Data Leaks

GORM's soft delete adds `deleted_at` field. Ensure:
- Queries automatically exclude soft-deleted records (GORM does this by default)
- Hard delete is used only when necessary: `db.Unscoped().Delete(&user)`
- Check if sensitive data should be hard deleted for compliance

## Best Practices

1. **Always use parameterized queries**: Use `?` placeholders with GORM
2. **Explicit field assignment**: Don't bind JSON directly to models for updates
3. **Check authorization first**: Verify permissions before database operations
4. **Use transactions**: Wrap multi-step operations in `db.Transaction()`
5. **Validate foreign keys**: Ensure referenced records exist and user has access
6. **Audit sensitive operations**: Log changes to critical data (users, permissions, etc.)

## Relevant Files

- `models/*.go` - Database models and queries
- `api/*.go` - API handlers calling database operations

## References

- GORM Security: https://gorm.io/docs/security.html
- OWASP Query Parameterization: https://cheatsheetseries.owasp.org/cheatsheets/Query_Parameterization_Cheat_Sheet.html
