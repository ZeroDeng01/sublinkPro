# API Security Guide

This guide covers API security patterns for SublinkPro's REST APIs.

## What to Check

- **CORS configuration**: Are CORS origins properly restricted (not `*` in production)?
- **Rate limiting**: Are endpoints rate-limited to prevent abuse?
- **API key validation**: Are API keys validated before processing requests?
- **HTTPS enforcement**: Is HTTPS required for sensitive endpoints?
- **Content-Type validation**: Are Content-Type headers validated?
- **Error information leakage**: Do error messages avoid exposing internal details?

## Common Vulnerabilities

### Overly Permissive CORS

**Bad Pattern**: Allow all origins
```go
// ❌ BAD: Allow all origins
func SetupRouterBad() *gin.Engine {
    router := gin.Default()
    router.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowCredentials: true,
    }))
    return router
}
```

**Good Pattern**: Restrict to specific origins
```go
// ✅ GOOD: Restrict to specific origins
func SetupRouterGood() *gin.Engine {
    router := gin.Default()
    router.Use(cors.New(cors.Config{
        AllowOrigins: []string{
            "https://example.com",
            "https://app.example.com",
        },
        AllowCredentials: true,
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    }))
    return router
}

// ✅ BEST: Environment-based CORS configuration
func SetupRouterBest() *gin.Engine {
    router := gin.Default()

    allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
    if allowedOrigins == "" {
        allowedOrigins = "https://example.com"
    }

    router.Use(cors.New(cors.Config{
        AllowOrigins:     strings.Split(allowedOrigins, ","),
        AllowCredentials: true,
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    }))
    return router
}
```

### Information Leakage in Errors

**Bad Pattern**: Exposing internal error details
```go
// ❌ BAD: Exposing internal error details
func HandleRequestBad(c *gin.Context) {
    result, err := someOperation()
    if err != nil {
        // May leak stack trace, DB schema, file paths, etc.
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, result)
}
```

**Good Pattern**: Generic error messages
```go
// ✅ GOOD: Generic error messages
func HandleRequestGood(c *gin.Context) {
    result, err := someOperation()
    if err != nil {
        // Log detailed error for debugging
        log.Error("Operation failed: ", err)
        // Return generic message to client
        c.JSON(500, gin.H{"error": "Internal server error"})
        return
    }
    c.JSON(200, result)
}

// ✅ BEST: Categorized error handling
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func HandleRequestBest(c *gin.Context) {
    result, err := someOperation()
    if err != nil {
        log.Error("Operation failed: ", err)

        // Categorize errors
        var apiErr APIError
        switch {
        case errors.Is(err, ErrNotFound):
            apiErr = APIError{Code: "NOT_FOUND", Message: "Resource not found"}
            c.JSON(404, apiErr)
        case errors.Is(err, ErrUnauthorized):
            apiErr = APIError{Code: "UNAUTHORIZED", Message: "Authentication required"}
            c.JSON(401, apiErr)
        default:
            apiErr = APIError{Code: "INTERNAL_ERROR", Message: "An error occurred"}
            c.JSON(500, apiErr)
        }
        return
    }
    c.JSON(200, result)
}
```

### Missing Rate Limiting

Endpoints susceptible to abuse should be rate-limited:
- Login endpoints
- Password reset
- API key generation
- Resource-intensive operations

## Best Practices

1. **Restrict CORS origins**: Never use `*` in production with credentials
2. **Rate limit sensitive endpoints**: Prevent brute force and DoS attacks
3. **Validate API keys**: Check validity and permissions before processing
4. **Use HTTPS**: Enforce HTTPS for all sensitive endpoints
5. **Generic error messages**: Don't expose internal details to clients
6. **Validate Content-Type**: Ensure requests have expected content types
7. **Log security events**: Track suspicious activity (failed auth, rate limits hit)

## Relevant Files

- `routers/*.go` - Route registration and middleware application
- `middlewares/auth.go` - API key validation
- `api/*.go` - API endpoint handlers

## References

- OWASP API Security Top 10: https://owasp.org/www-project-api-security/
- OWASP REST Security Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html
- CORS Best Practices: https://cheatsheetseries.owasp.org/cheatsheets/HTML5_Security_Cheat_Sheet.html#cross-origin-resource-sharing
