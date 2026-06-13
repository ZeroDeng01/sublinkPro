# Documentation Index

Welcome to SublinkPro documentation. This index helps you find the right document for your needs.

---

## 📚 For Users

### Getting Started
- **[README.md](../README.md)** - Project overview, features, quick start
- **[Installation Guide](installation.md)** - Docker, docker-compose, one-line script installation

### Configuration & Operation
- **[Configuration Reference](configuration.md)** - Environment variables, config file, precedence
- **[Security Guidelines](security-guidelines.md)** - Best practices, default credentials, MFA, incident response

### Features
- **[Tags System](features/tags.md)** - Smart auto-tagging, rule-based groups
- **[Speed Test](features/speedtest.md)** - Two-stage testing, latency/speed, IP quality
- **[Unlock Checks](features/unlock-check.md)** - Streaming & AI availability testing
- **[Chain Proxy](features/chain-proxy.md)** - Condition-based node selection
- **[Template AI Editing](features/template-ai.md)** - AI-assisted template generation
- **[Airport Management](features/airport.md)** - Import, scheduled updates, traffic monitoring
- **[Subscription Sharing](features/subscription-share.md)** - Multiple links, expiration, stats
- **[Host Management](features/host.md)** - Domain mappings, DNS, CDN preferred IPs
- **[Cloudflare Tunnel](features/cloudflare-tunnel.md)** - Secure public access
- **[Telegram Bot](features/telegram-bot.md)** - Command list, setup guide
- **[Multi-Factor Auth (MFA)](features/mfa.md)** - TOTP, recovery codes, emergency reset
- **[Script Support](script_support.md)** - Node filtering, post-processing functions

---

## 🛠️ For Developers

### Development Guides
- **[Development Guide](development.md)** - Project structure, local setup, validation commands
- **[Build & Deployment](build-and-deployment.md)** - Build process, Docker, CI/CD, troubleshooting
- **[Practical Recipes](practical-recipes.md)** - Common development patterns and step-by-step guides

### Guidelines & Standards
- **[Contributing Guide](../CONTRIBUTING.md)** - Branch conventions, PR process, cross-layer sync
- **[Frontend Theme Guidelines](frontend-theme-guidelines.md)** - Light/dark mode, surface layering
- **[Internationalization (i18n)](internationalization.md)** - Translation workflow, bilingual requirements
- **[Security Guidelines](security-guidelines.md)** - Security checklist, incident response
- **[Code of Conduct](../CODE_OF_CONDUCT.md)** - Community standards

---

## 🤖 For AI Agents

### Agent Guides
- **[AGENTS.md](../AGENTS.md)** - Architectural overview, tech stack, navigation map
- **[.agents/README.md](../.agents/README.md)** - Skills system documentation

### Operational Skills
Located in `.agents/skills/`:
- **theme-check** - UI theme adaptation checklist
- **cross-layer-sync** - Multi-layer synchronization guide
- **pre-commit-check** - Pre-commit validation steps
- **doc-sync-check** - Documentation sync requirements

### AI Agent Skill (User-Facing)
- **[skill-sublinkpro/](../../skill-sublinkpro/)** - Portable AI skill for REST API interaction
  - `SKILL.md` - Skill definition and workflows
  - `reference/api.md` - API endpoint reference
  - `reference/deploy.md` - Deployment guide
  - `reference/docs.md` - Documentation map

---

## 📖 Quick Navigation

### By Task

| I want to... | Read this |
|---|---|
| Install SublinkPro | [Installation Guide](installation.md) |
| Configure environment variables | [Configuration Reference](configuration.md) |
| Set up local development | [Development Guide](development.md) |
| Build for production | [Build & Deployment](build-and-deployment.md) |
| Add a backend feature | [Practical Recipes](practical-recipes.md) - Recipe 1 |
| Add a scheduled task | [Practical Recipes](practical-recipes.md) - Recipe 2 |
| Change mihomo behavior | [Practical Recipes](practical-recipes.md) - Recipe 3 |
| Adapt UI theme | [Frontend Theme Guidelines](frontend-theme-guidelines.md) |
| Add translations | [Internationalization](internationalization.md) |
| Secure my deployment | [Security Guidelines](security-guidelines.md) |
| Contribute code | [Contributing Guide](../CONTRIBUTING.md) |
| Use speed testing | [Speed Test Feature](features/speedtest.md) |
| Set up MFA | [MFA Feature](features/mfa.md) |

### By Role

**New User**:
1. [README.md](../README.md) - What is SublinkPro?
2. [Installation Guide](installation.md) - How to install
3. [Configuration Reference](configuration.md) - How to configure
4. [Features](features/) - What can it do?

**Developer**:
1. [Development Guide](development.md) - Local setup
2. [AGENTS.md](../AGENTS.md) - Project architecture
3. [Practical Recipes](practical-recipes.md) - Common patterns
4. [Contributing Guide](../CONTRIBUTING.md) - How to contribute

**DevOps/SRE**:
1. [Installation Guide](installation.md) - Deployment methods
2. [Build & Deployment](build-and-deployment.md) - Build process
3. [Configuration Reference](configuration.md) - Config options
4. [Security Guidelines](security-guidelines.md) - Security best practices

---

## 🌍 Language Versions

All documentation is maintained in both English and Simplified Chinese:
- English: `<filename>.md`
- Chinese: `<filename>.zh-CN.md`

Example:
- `installation.md` (English)
- `installation.zh-CN.md` (简体中文)

---

## 🔄 Documentation Updates

When code changes affect documentation:
1. Update relevant docs in the same PR
2. Update both English and Chinese versions
3. Use the doc-sync-check skill: `.agents/skills/doc-sync-check/SKILL.md`

See [Contributing Guide](../CONTRIBUTING.md) for details.

---

## 📞 Need Help?

- **GitHub Issues**: https://github.com/ZeroDeng01/sublinkPro/issues
- **Telegram Group**: See README for link
- **Skill API**: Use `skill-sublinkpro/` for AI-assisted help

---

## 📝 Note

This is the documentation index. For the actual GitHub repository structure, see the project's README.md.
