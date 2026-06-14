# Multi-Factor Authentication (MFA) Security Guide

This guide covers MFA security patterns and common vulnerabilities for SublinkPro.

## What to Check

- **MFA enforcement**: Is MFA required for sensitive operations?
- **MFA bypass prevention**: Can MFA be circumvented through alternate flows?
- **Backup codes security**: Are backup codes securely generated and stored?
- **TOTP secret protection**: Are TOTP secrets encrypted at rest?
- **Rate limiting**: Is there rate limiting on MFA verification attempts?

## Common Vulnerabilities

### MFA Bypass Through Alternate Flows

Ensure all authentication paths enforce MFA when enabled:
- Direct login
- Password reset flows
- Session resumption
- API key authentication (when applicable)

### Weak Backup Code Generation

Backup codes must be:
- Cryptographically random (use `crypto/rand`, not `math/rand`)
- Long enough to resist brute force (at least 8 characters)
- Stored hashed, not in plain text
- One-time use only

### TOTP Secret Exposure

TOTP secrets must be:
- Encrypted at rest in the database
- Never logged or exposed in error messages
- Transmitted only over HTTPS
- Securely deleted when MFA is disabled

## Relevant Files

- `api/auth_mfa.go` - MFA authentication endpoints
- `models/mfa.go` - MFA data models
- `middlewares/mfa.go` - MFA enforcement middleware

## Best Practices

1. **Enforce MFA for admin accounts**: Always require MFA for privileged users
2. **Rate limit verification attempts**: Prevent brute force of TOTP codes
3. **Provide secure backup codes**: Generate cryptographically random backup codes
4. **Encrypt TOTP secrets**: Never store TOTP secrets in plain text
5. **Allow MFA recovery**: Provide secure recovery mechanism for lost devices
6. **Log MFA events**: Track MFA setup, usage, and failures

## References

- OWASP MFA Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Multifactor_Authentication_Cheat_Sheet.html
- RFC 6238 (TOTP): https://datatracker.ietf.org/doc/html/rfc6238
