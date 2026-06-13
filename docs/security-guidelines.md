English | [简体中文](security-guidelines.zh-CN.md)

# Security Guidelines

This guide outlines security considerations, best practices, and safety checks for SublinkPro development and deployment.

---

## 🔐 Default Credentials

### ⚠️ Critical Security Warning

**Default admin credentials**: `admin / 123456`

**Action Required**:
- Change default password immediately after first login
- Use strong, unique passwords (minimum 12 characters)
- Consider enabling MFA (Multi-Factor Authentication)

**For production deployments**:
```bash
# Login with default credentials
# Navigate to: Settings → Security → Change Password
# Or use API to change password programmatically
```

**Never**:
- Use default credentials in production
- Share admin credentials
- Store credentials in plain text
- Commit credentials to version control

---

## 🔑 Sensitive Configuration

### Environment Variables

These variables contain sensitive data and must be protected:

#### JWT Secret
```bash
SUBLINK_JWT_SECRET=<your-secret-key>
```

**Purpose**: Signs JWT tokens for authentication
**Security**:
- Must be at least 32 characters
- Use cryptographically random values
- Never reuse across instances
- Rotate periodically

**Generate strong secret**:
```bash
openssl rand -base64 32
```

#### API Encryption Key
```bash
SUBLINK_API_ENCRYPTION_KEY=<your-encryption-key>
```

**Purpose**: Encrypts sensitive API data
**Security**:
- Must be exactly 32 bytes (AES-256)
- Keep separate from JWT secret
- Store securely (secrets manager recommended)

**Generate**:
```bash
openssl rand -hex 32
```

#### MFA Reset Secret
```bash
SUBLINK_MFA_RESET_SECRET=<your-mfa-secret>
```

**Purpose**: Emergency MFA reset capability
**Security**:
- Only for emergency recovery
- Restrict access to system administrators
- Audit all uses of this secret
- Rotate after each use (recommended)

### Multi-Instance Deployment

When running multiple instances:

**Critical**: All instances must use the **same** secrets:
- `SUBLINK_JWT_SECRET`
- `SUBLINK_API_ENCRYPTION_KEY`
- `SUBLINK_MFA_RESET_SECRET`

**Why**: JWT tokens and encrypted data must be portable across instances.

**Deployment pattern**:
```bash
# Load from secure secrets manager
export SUBLINK_JWT_SECRET=$(vault kv get -field=jwt_secret secret/sublinkpro)
export SUBLINK_API_ENCRYPTION_KEY=$(vault kv get -field=api_key secret/sublinkpro)
export SUBLINK_MFA_RESET_SECRET=$(vault kv get -field=mfa_secret secret/sublinkpro)
```

---

## 🗄️ Database Security

### SQLite

**Default location**: `./db/sublink.db`

**Security considerations**:
- File permissions: `600` (owner read/write only)
- Backup regularly
- Encrypt at rest (filesystem level)
- Not recommended for multi-instance deployments

**Secure permissions**:
```bash
chmod 600 db/sublink.db
chown sublinkpro:sublinkpro db/sublink.db
```

### MySQL / PostgreSQL Migration

**Warning**: Migration from SQLite to MySQL/PostgreSQL requires manual restart.

**Migration checklist**:
1. ✅ Backup SQLite database
2. ✅ Configure MySQL/PostgreSQL connection
3. ✅ Update `DSN` configuration
4. ✅ Restart application
5. ✅ Verify data integrity
6. ✅ Test authentication

**Security during migration**:
- Use TLS for database connections
- Restrict database user permissions
- Enable audit logging
- Verify no data loss

### Database Connection Security

```bash
# MySQL with TLS
DSN="username:password@tcp(host:3306)/dbname?tls=true"

# PostgreSQL with SSL
DSN="postgres://username:password@host:5432/dbname?sslmode=require"
```

**Never**:
- Expose database ports to public internet
- Use root database accounts
- Store plaintext passwords in config files
- Disable SSL/TLS in production

---

## 🐳 Docker Security

### Runtime Directories

Docker mounts these directories with sensitive data:

```yaml
volumes:
  - ./db:/app/db           # Database and config files
  - ./template:/app/template   # Template files
  - ./logs:/app/logs       # Application logs (may contain sensitive data)
```

**Security requirements**:
- Host directories: `700` or `750` permissions
- Files: `600` or `640` permissions
- Run container as non-root user
- Use read-only mounts where possible

**Secure docker-compose example**:
```yaml
version: '3.8'
services:
  sublinkpro:
    image: zerodeng/sublink-pro:latest
    user: "1000:1000"  # Non-root user
    volumes:
      - ./db:/app/db:rw
      - ./template:/app/template:ro  # Read-only
      - ./logs:/app/logs:rw
    environment:
      - SUBLINK_JWT_SECRET=${SUBLINK_JWT_SECRET}
    secrets:
      - jwt_secret
      - api_key

secrets:
  jwt_secret:
    external: true
  api_key:
    external: true
```

### Container Security

**Best practices**:
- Use specific version tags, not `latest`
- Scan images for vulnerabilities
- Enable Docker content trust
- Limit container resources
- Enable security options:

```yaml
security_opt:
  - no-new-privileges:true
  - seccomp:unconfined
read_only: true
tmpfs:
  - /tmp
```

---

## 🔒 MFA (Multi-Factor Authentication)

### Setup Security

**When enabling MFA**:
- Generate backup codes immediately
- Store backup codes securely (password manager)
- Test TOTP code generation before finalizing
- Document recovery procedure

**For users**:
```
Settings → Security → Multi-Factor Authentication
→ Scan QR code → Save backup codes → Verify TOTP
```

### Emergency Recovery

**MFA Reset Secret** (`SUBLINK_MFA_RESET_SECRET`):
- Only for emergency MFA disable
- Requires system administrator access
- Audit all uses
- Rotate after each emergency use

**Recovery procedure**:
1. Verify user identity through alternative channel
2. Use MFA reset secret to disable MFA
3. Force password reset
4. User re-enables MFA with new codes
5. Rotate MFA reset secret

---

## 🌐 Network Security

### Port Exposure

**Default port**: `8000`

**Recommendations**:
- Use reverse proxy (nginx/Caddy) in production
- Enable HTTPS/TLS at proxy level
- Restrict direct access to backend port
- Use firewall rules

**Nginx example**:
```nginx
server {
    listen 443 ssl http2;
    server_name sublink.example.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### API Security

**Authentication**:
- All `/api/v1/*` endpoints require authentication
- Use `X-API-Key` header or JWT token
- Subscription endpoints `/c/*` use token-based auth

**Rate limiting** (recommended):
```nginx
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;

location /api/ {
    limit_req zone=api burst=20 nodelay;
    proxy_pass http://localhost:8000;
}
```

---

## 📝 Logging and Audit

### Log Security

**Log location**: `./logs/`

**Security considerations**:
- Logs may contain sensitive data (IPs, user agents, errors)
- Rotate logs regularly
- Restrict log file permissions: `640`
- Redact sensitive data in logs

**Log rotation**:
```bash
# Setup logrotate
cat > /etc/logrotate.d/sublinkpro <<EOF
/app/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 640 sublinkpro sublinkpro
}
EOF
```

### Audit Events

**Monitor for**:
- Failed login attempts
- MFA changes
- Password changes
- Admin privilege escalations
- Bulk data exports
- Configuration changes

---

## 🛡️ Security Checklist for Changes

### Before Deploying Changes

- [ ] No hardcoded credentials
- [ ] No sensitive data in logs
- [ ] No SQL injection vulnerabilities
- [ ] No XSS vulnerabilities
- [ ] Authentication/authorization unchanged or improved
- [ ] Input validation present
- [ ] Error messages don't leak sensitive info
- [ ] Secrets use environment variables
- [ ] Dependencies updated and scanned
- [ ] No known CVEs in dependencies

### Code Review Security Focus

- [ ] Review authentication logic changes
- [ ] Verify permission checks
- [ ] Check for information disclosure
- [ ] Validate input sanitization
- [ ] Ensure secure defaults
- [ ] Check for timing attacks
- [ ] Verify cryptographic operations
- [ ] Review database queries for injection

---

## 🚨 Incident Response

### Security Incident Checklist

1. **Identify**: What happened?
2. **Contain**: Stop the breach
3. **Eradicate**: Remove threat
4. **Recover**: Restore service
5. **Lessons**: Document and improve

### Immediate Actions

**If credentials compromised**:
1. Rotate all secrets immediately
2. Force logout all sessions
3. Audit access logs
4. Notify affected users
5. Review recent changes

**If data breach suspected**:
1. Isolate affected systems
2. Preserve logs and evidence
3. Contact security team
4. Document timeline
5. Prepare disclosure (if required)

---

## 🔧 Security Maintenance

### Regular Tasks

**Weekly**:
- Review access logs
- Check for failed login attempts
- Monitor error rates

**Monthly**:
- Update dependencies
- Scan for vulnerabilities
- Review permissions
- Test backup restoration

**Quarterly**:
- Rotate secrets
- Security audit
- Penetration test (if applicable)
- Update security documentation

### Dependency Security

**Scan dependencies**:
```bash
# Backend
go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth

# Frontend
cd webs
yarn audit
```

**Update dependencies**:
```bash
# Backend
go get -u ./...
go mod tidy

# Frontend
cd webs
yarn upgrade-interactive
```

---

## 📚 Related Documentation

- **Configuration**: See `configuration.md`
- **MFA Setup**: See `docs/features/mfa.md`
- **Deployment**: See `installation.md`
- **Docker Security**: See `skill-sublinkpro/reference/deploy.md`

---

## ⚠️ Known Limitations

### Database Compatibility

**Not compatible** with upstream project databases.

Do not attempt to:
- Import data from upstream projects
- Share databases between versions
- Use unsupported database migrations

### Production Recommendations

- Use PostgreSQL or MySQL in production (not SQLite)
- Enable HTTPS/TLS
- Use secrets manager for sensitive config
- Implement rate limiting
- Enable audit logging
- Regular backups
- Monitoring and alerting

---

## 📞 Security Contacts

For security vulnerabilities:
- GitHub Security Advisories (preferred)
- GitHub Issues (for non-sensitive security questions)
- Project maintainers (for sensitive disclosures)

**Do not**:
- Disclose vulnerabilities publicly before patch
- Test vulnerabilities on production systems
- Share exploit code publicly
