# Dependency Security Guide

This guide covers secure dependency management for SublinkPro.

## What to Check

- **Known vulnerabilities**: Are dependencies scanned for CVEs?
- **Dependency versions**: Are dependencies pinned to specific versions?
- **Minimal dependencies**: Are only necessary dependencies included?
- **License compliance**: Are dependency licenses compatible with project license?

## Scanning for Vulnerabilities

### Backend (Go)

**Using govulncheck** (official Go tool):
```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

**Using nancy**:
```bash
go list -json -m all | nancy sleuth
```

**Using go mod** (check for updates):
```bash
go list -m -u all
```

### Frontend (JavaScript/Yarn)

**Using yarn audit**:
```bash
cd webs
yarn audit

# Fix automatically (when possible)
yarn audit --fix
```

**Check for outdated packages**:
```bash
cd webs
yarn outdated
```

## Best Practices

1. **Regular scanning**: Run vulnerability scans in CI/CD pipeline
2. **Pin versions**: Use exact versions in `go.mod` and `package.json`
3. **Minimal dependencies**: Avoid unnecessary dependencies
4. **Review updates**: Don't blindly update; review changelogs
5. **License compliance**: Ensure licenses are compatible (MIT, Apache, BSD)
6. **Monitor advisories**: Subscribe to security advisories for critical dependencies
7. **Update regularly**: Keep dependencies reasonably up-to-date

## Version Pinning

### Go (go.mod)
```
module github.com/example/project

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1 // Exact version
    gorm.io/gorm v1.25.5           // Exact version
)
```

### JavaScript (package.json)
```json
{
  "dependencies": {
    "react": "19.0.0",      // Exact version (no ^, ~)
    "react-dom": "19.0.0"
  }
}
```

Use exact versions for production stability.

## Handling Vulnerabilities

### When a vulnerability is found:

1. **Assess severity**: Critical, High, Medium, Low
2. **Check exploitability**: Does it affect your usage?
3. **Find patches**: Is a patched version available?
4. **Update or mitigate**: Update dependency or apply workaround
5. **Test thoroughly**: Ensure updates don't break functionality
6. **Document**: Add notes about security updates in changelog

### If no patch exists:

1. **Find alternatives**: Can you replace the dependency?
2. **Mitigate**: Can you limit exposure (input validation, sandboxing)?
3. **Fork and patch**: As a last resort, maintain your own patched version
4. **Report upstream**: Notify maintainers of the issue

## License Compliance

Common license types:
- ✅ **MIT, Apache 2.0, BSD**: Very permissive, usually safe
- ⚠️ **LGPL**: Requires dynamic linking (check with legal)
- ❌ **GPL**: May require open-sourcing your code (check with legal)
- ❌ **Proprietary/Unknown**: Avoid unless cleared by legal

Check licenses:
```bash
# Go
go-licenses report ./... --template=csv

# JavaScript
cd webs
yarn licenses list
```

## CI/CD Integration

Add to `.github/workflows/security-scan.yml`:
```yaml
name: Security Scan
on: [push, pull_request]
jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...
      - name: Run yarn audit
        run: |
          cd webs
          yarn audit
```

## Relevant Files

- `go.mod`, `go.sum` - Go dependencies
- `webs/package.json`, `webs/yarn.lock` - Frontend dependencies
- `.github/workflows/` - CI/CD security scans

## References

- Go Vulnerability Database: https://pkg.go.dev/vuln/
- GitHub Advisory Database: https://github.com/advisories
- Snyk Vulnerability Database: https://snyk.io/vuln/
- OWASP Dependency Check: https://owasp.org/www-project-dependency-check/
