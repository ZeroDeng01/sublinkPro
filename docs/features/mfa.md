English | [简体中文](mfa.zh-CN.md)

# TOTP Multi Factor Authentication, MFA

SublinkPro supports TOTP based multi factor authentication to improve admin account security.

## Usage Flow

### 1. Enable TOTP

Open:

`Settings -> Personal Settings -> Security Settings`

The Personal Settings page currently has two sections:

- **Basic Profile**: change username and nickname
- **Security Settings**: manage TOTP, recovery codes, and other security related operations

Password changes are no longer mixed into the profile form. Use the **Change Password** button at the top of the page to open a separate dialog.

Then complete these steps:

1. Open `Security Settings`.
2. Enter the current password.
3. Scan the QR code with an authenticator app.
4. Enter the current 6 digit dynamic code.
5. Save the generated recovery codes.

Supported authenticators include:

- Google Authenticator
- Microsoft Authenticator
- 1Password
- Aegis
- Other apps compatible with standard TOTP

### 2. Verify during login

After enabling TOTP, login has two steps:

1. Username + password + CAPTCHA
2. TOTP dynamic code or recovery code

If you cannot access the authenticator, use a recovery code to sign in.

## Recovery Codes

Recovery codes are for emergency login when the authenticator is lost, the phone is broken, or temporary access is unavailable.

Notes:

- Each recovery code can be used once.
- Recovery codes can be used for login only **after TOTP is fully enabled**.
- Old recovery codes become invalid immediately after regeneration.
- Printing or storing recovery codes offline is recommended.

## Second Confirmation for Sensitive Operations

When TOTP is enabled on an account, these operations require the current dynamic code again:

- Change password
- Change username or nickname
- Disable TOTP
- Regenerate recovery codes
- Rebind TOTP

Where:

- **Username / nickname changes** are done in `Basic Profile`.
- **Password changes** use the top button and separate dialog.
- **TOTP operations** are done in `Security Settings`.

This prevents someone who already has a login session from directly changing account security settings.

## What if the Authenticator Is Lost

Recommended recovery order:

1. **Use a recovery code first**.
2. After signing in, rebind TOTP and save new recovery codes.
3. If recovery codes are also lost, contact operations staff to use the restricted emergency reset flow.

## Operations Emergency Reset

SublinkPro provides a restricted break glass mechanism, not a global MFA bypass.

### Prerequisite

Operations staff must set this environment variable:

```bash
SUBLINK_MFA_RESET_SECRET=your-break-glass-secret
```

### Characteristics

- Environment variable only
- Tokens should have expiration times
- Reset still requires the target user's username and password
- Only clears TOTP for that account, with no direct login capability

### Suitable scenarios

- User lost the authenticator
- User has no usable recovery code
- Operations staff need to help restore access

## Security Advice

- Enable TOTP for every admin account.
- Use Cloudflare Turnstile as the preferred login CAPTCHA mode.
- Protect `SUBLINK_JWT_SECRET`, `SUBLINK_API_ENCRYPTION_KEY`, and `SUBLINK_MFA_RESET_SECRET`.
- Don't store recovery codes in plaintext browser notes, chats, or the same place as the account password.
- After emergency reset work is done, rotate `SUBLINK_MFA_RESET_SECRET` as soon as possible.
