# Session Management Security Guide

This guide covers secure session management patterns for SublinkPro.

## What to Check

- **Session timeout**: Are sessions expired after inactivity?
- **Secure flags**: Are session cookies marked as Secure and HttpOnly?
- **SameSite attribute**: Is SameSite attribute set to prevent CSRF?
- **Session regeneration**: Are session IDs regenerated after login?
- **Logout functionality**: Does logout properly invalidate the session?

## Secure Session Configuration

**Good Pattern**: Secure session setup with full lifecycle management
```go
// ✅ GOOD: Secure session configuration
import (
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"
    "github.com/gin-gonic/gin"
    "net/http"
)

func SetupSessionsGood(router *gin.Engine, secretKey string) {
    store := cookie.NewStore([]byte(secretKey))

    // Configure secure session options
    store.Options(sessions.Options{
        Path:     "/",
        MaxAge:   3600, // 1 hour
        HttpOnly: true, // Prevent JavaScript access
        Secure:   true, // HTTPS only
        SameSite: http.SameSiteStrictMode, // Prevent CSRF
    })

    router.Use(sessions.Sessions("session", store))
}

// Session lifecycle management
func LoginHandler(c *gin.Context) {
    // After successful authentication
    session := sessions.Default(c)

    // Clear old session (prevent fixation)
    session.Clear()

    // Set new session data
    session.Set("user_id", user.ID)
    session.Set("username", user.Username)
    session.Set("login_time", time.Now().Unix())

    if err := session.Save(); err != nil {
        log.Error("Failed to save session: ", err)
        c.JSON(500, gin.H{"error": "Session error"})
        return
    }

    c.JSON(200, gin.H{"message": "Login successful"})
}

func LogoutHandler(c *gin.Context) {
    session := sessions.Default(c)

    // Clear all session data
    session.Clear()

    if err := session.Save(); err != nil {
        log.Error("Failed to clear session: ", err)
    }

    c.JSON(200, gin.H{"message": "Logout successful"})
}

// Session validation middleware
func SessionAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        session := sessions.Default(c)

        userID := session.Get("user_id")
        if userID == nil {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }

        // Optional: Check session age
        loginTime := session.Get("login_time")
        if loginTime != nil {
            if time.Now().Unix()-loginTime.(int64) > 3600 {
                session.Clear()
                session.Save()
                c.JSON(401, gin.H{"error": "Session expired"})
                c.Abort()
                return
            }
        }

        c.Next()
    }
}
```

## Cookie Attributes Explained

### HttpOnly
- Prevents JavaScript access to cookies
- Protects against XSS cookie theft
- **Always set to true** for session cookies

### Secure
- Ensures cookies are only sent over HTTPS
- Prevents man-in-the-middle attacks
- **Always set to true** in production

### SameSite
- Prevents CSRF attacks
- Options:
  - `Strict`: Never sent on cross-site requests (most secure)
  - `Lax`: Sent on top-level navigation (balanced)
  - `None`: Sent on all requests (requires Secure flag)
- **Use Strict or Lax** for session cookies

### MaxAge / Expires
- Limits session lifetime
- Forces re-authentication after timeout
- **Set appropriate timeout** based on sensitivity

## Best Practices

1. **Short session lifetimes**: 30-60 minutes for sensitive apps
2. **Regenerate on privilege change**: New session ID after login/logout
3. **Secure cookie flags**: HttpOnly, Secure, SameSite=Strict/Lax
4. **Server-side storage**: Store session data server-side, not in cookies
5. **Proper logout**: Clear session from server and client
6. **Concurrent session limits**: Optionally limit active sessions per user
7. **Session fixation prevention**: Generate new session ID after authentication

## Session Lifecycle

### Login
1. Authenticate user
2. **Generate new session ID** (prevent fixation)
3. Store session server-side
4. Set secure cookie with session ID

### Request
1. Validate session cookie
2. Check session expiration
3. Refresh activity timestamp
4. Process request

### Logout
1. Delete session from server-side store
2. Clear session cookie
3. Optional: Redirect to login page

## Relevant Files

- `middlewares/auth.go` - Session validation middleware
- `api/auth.go` - Login/logout handlers
- Session configuration in router setup

## References

- OWASP Session Management Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
- Go session libraries: gorilla/sessions, gin-contrib/sessions
