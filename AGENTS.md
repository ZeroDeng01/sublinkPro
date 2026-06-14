# AGENTS.md

This file provides AI and human contributors with a high-level architectural guide and navigation map for SublinkPro.

---

## 0. MANDATORY: Post-Development Workflow for AI Agents

**CRITICAL INSTRUCTION FOR AI AGENTS**: After completing ANY code change, you MUST automatically run the post-development workflow before declaring work "complete". This is NOT optional.

### Workflow Overview

```
Code Change Complete
    ↓
Automatically invoke: .agents/skills/post-dev-workflow/SKILL.md
    ↓
Phase 1: Code Validation (lint, format, build, test)
    ↓
Phase 2: Cross-Layer Sync Check (if multi-layer change)
    ↓
Phase 3: Documentation Sync Check (if behavior/API/config changed)
    ↓
Phase 4: Test Execution (if key logic changed)
    ↓
Phase 5: Change Summary (prepare commit message)
    ↓
Report Results to User
    ↓
Ready to Commit (only if all phases pass)
```

### When to Trigger

**Always** trigger after:
- Backend code changes (any `.go` file)
- Frontend code changes (any file in `webs/`)
- Configuration changes
- Documentation changes that affect code behavior
- Any combination of the above

### What "Complete" Means

Work is NOT complete until:
- ✅ All validation passes (lint, format, build, test)
- ✅ Cross-layer synchronization verified (if applicable)
- ✅ Documentation updated (if applicable)
- ✅ Tests added/updated (if applicable)
- ✅ Change summary prepared

**Do not declare work "done" or "finished" until the post-development workflow completes successfully.**

### Detailed Workflow

See `.agents/skills/post-dev-workflow/SKILL.md` for:
- Complete phase-by-phase workflow
- Validation commands for each layer
- Cross-layer sync requirements
- Documentation sync requirements
- Test execution guidelines
- Change summary format
- Exit criteria

---

## 1. Project Shape

**Architecture**: Single full-stack application (not a monorepo)
- **Backend**: Go at repository root
- **Frontend**: React + Vite in `webs/`
- **Production**: Frontend built first, then embedded into Go binary

**Entry points**:
- Backend: `main.go`
- Frontend: `webs/src/index.jsx`

**Key boundaries**:
- `routers/` - Route registration
- `api/` - HTTP handlers
- `services/` - Business logic and background workflows
- `models/` - Persistence and migrations
- `middlewares/` - Auth and request pipeline
- `node/` - Subscription and protocol parsing
- `webs/src/api/` - Frontend request boundary
- `webs/src/views/` - Page-level features
- `skill-sublinkpro/` - User-facing AI agent skill (portable SKILL.md format, consumes REST API via X-API-Key)

## 2. Source of Truth

When instructions conflict, trust these files in order:

1. `webs/package.json` - Frontend commands and dependencies
2. `docs/development.md` - Development workflow and structure
3. `docs/configuration.md` - Config precedence and runtime behavior
4. `.github/workflows/pr-checks.yml` - PR automated checks
5. `.github/workflows/build-release.yml` - CI and release build
6. `Dockerfile` - Production build sequence

**For frontend commands, output paths, and toolchain**: Trust repository files over generic framework assumptions.

## 3. Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go + Gin + GORM |
| Core network/proxy | mihomo (MetaCubeX) |
| Database | SQLite (default), MySQL/PostgreSQL (optional) |
| Frontend | React 19 + Vite |
| UI | Material UI |
| Package manager | Yarn 4 |
| Scheduler | robfig/cron |

**About mihomo**: Core integration point for proxy adapters, speed testing, DNS resolution, Host injection, and proxied outbound requests. Not a minor dependency.

## 4. Getting Started

### Development
See `docs/development.md` for:
- Local setup (backend and frontend)
- Validation commands
- Testing requirements
- Protocol extension guide
- Scheduled task development
- Cron format

### Configuration
See `docs/configuration.md` for:
- Environment variables
- Config file precedence
- Default values
- Runtime directories

### Deployment
See `docs/installation.md` and `skill-sublinkpro/reference/deploy.md` for:
- Docker installation
- docker-compose setup
- One-line script
- Update procedures

## 5. Contribution Workflow

### Quick reference
See `CONTRIBUTING.md` for:
- Branch and commit conventions
- PR submission process
- Cross-layer synchronization summary
- Testing and validation overview

### Detailed guidelines

**Theme changes**: See `docs/frontend-theme-guidelines.md`
- Light/dark mode requirements
- Surface layering principles
- Component coverage expectations

**Internationalization**: See `docs/internationalization.md`
- Frontend i18n infrastructure (i18next)
- Backend i18n patterns (i18nKey + i18nParams)
- Bilingual documentation requirements

**Code quality**: See `docs/development.md` sections:
- Commenting standards
- Testing standards
- Validation commands

## 6. Operational Checklists (Skills)

For task-specific validation procedures, see `.agents/skills/`:

- **post-dev-workflow** - **MANDATORY** automated workflow after completing ANY code change (AI agents must run this)
- **pre-commit-check** - Pre-commit validation steps (subset of post-dev-workflow)
- **cross-layer-sync** - Multi-layer synchronization guide
- **doc-sync-check** - Documentation sync requirements
- **theme-check** - UI theme adaptation checklist
- **security-review** - Security review for authentication, authorization, sensitive data changes
- **performance-check** - Performance review for optimization, scalability, resource usage

Skills are reusable checklists for operational tasks. See `.agents/README.md` for details.

**For AI agents**: The `post-dev-workflow` skill is the master workflow that orchestrates all other skills. It must be invoked automatically after every code change.

## 7. Core Architectural Principles

### Minimal, boundary-respecting changes

- Fix issues in the correct layer
- When a change affects multiple layers, synchronize all impacted layers in the same PR
- Don't move business logic into `routers/`
- Don't put persistence details into UI code
- Keep scheduler logic in `services/scheduler/`

### Cross-layer synchronization is mandatory

**Rule**: When a change affects multiple layers (backend, frontend, docs), all impacted layers must be updated together.

**For detailed requirements**: See `.agents/skills/cross-layer-sync/SKILL.md`

**Core principle**: Backend API changes → update frontend + docs. Frontend contract changes → verify backend + docs. Config/deployment changes → update all relevant docs.

### Follow existing patterns

- Backend: Standard Go style and package structure
- Frontend: Current Vite + MUI + ESLint/Prettier setup
- Match surrounding naming and organization before introducing new patterns

### Documentation is part of the deliverable

**Rule**: Code changes that alter behavior, APIs, or configuration must update documentation in the same PR.

**For detailed requirements**: See `.agents/skills/doc-sync-check/SKILL.md`

**Bilingual**: All documentation maintained in English and Chinese (`*.zh-CN.md`)

## 8. mihomo Integration

Several core features depend directly on mihomo capabilities:

**Key capabilities**:
- Node delay and speed testing
- Proxy adapter construction and dialing
- DNS resolution and DoH flows
- Host mapping synchronization
- Proxied downloads and outbound requests

**Key files**:
- `services/mihomo/mihomo.go` - Core adapter and testing wrapper
- `services/mihomo/dns_resolver.go` - DNS resolution entry point
- `services/mihomo/host_resolver.go` - Host mapping sync
- `services/scheduler/speedtest_task.go` - Speed test integration
- `utils/proxy_client.go` - Shared proxy HTTP client
- `node/sub.go`, `node/usage.go` - Proxied subscription flows

**When to check**: Changes involving proxying, speed tests, DNS, Host mappings, chain proxy, or proxied downloads.

## 9. High-Value Entry Points

Start here when changing behavior:

| Area | Files |
|---|---|
| Auth/MFA | `api/auth.go`, `api/auth_mfa.go`, `middlewares/` |
| Subscriptions | `api/clients.go` |
| mihomo core | `services/mihomo/` |
| Scheduled tasks | `services/scheduler/` |
| Tags | `services/tag_service.go` |
| Protocols | `node/protocol/` |
| Proxy client | `utils/proxy_client.go` |
| Host management | `models/host.go` |
| DB migration | `models/db_migrate.go` |
| Frontend pages | `webs/src/views/` |
| Frontend routing | `webs/src/routes/` |
| Frontend API | `webs/src/api/` |
| Theme infrastructure | `webs/src/themes/`, `webs/src/utils/colorUtils.js` |

**Theme reference patterns**:
- Node preview: `NodePreviewDialog.jsx`, `NodePreviewCard.jsx`, `NodePreviewDetailsPanel.jsx`
- Chain proxy: `ChainProxyDialog.jsx`, `ChainPreviewDialog.jsx`, `MobileChainBuilder.jsx`, `ConditionBuilder.jsx`
- Airport management: `views/airports/index.jsx`, `AirportFormDialog.jsx`, `AirportBatchEditDialog.jsx`, `AirportMobileList.jsx`

## 10. Documentation Map

### For users
- `README.md` - Project overview and quick start
- `docs/installation.md` - Installation methods
- `docs/configuration.md` - Configuration reference
- `docs/security-guidelines.md` - Security best practices
- `docs/features/` - Feature-specific guides
- `skill-sublinkpro/` - AI agent skill for REST API interaction

### For developers
- `CONTRIBUTING.md` - Contribution workflow
- `docs/development.md` - Development guide
- `docs/build-and-deployment.md` - Build process and deployment
- `docs/practical-recipes.md` - Common development patterns
- `docs/frontend-theme-guidelines.md` - Theme adaptation rules
- `docs/internationalization.md` - i18n guidelines
- `.agents/skills/` - Operational checklists

### For AI agents
- `AGENTS.md` (this file) - Architectural overview and navigation
- `.agents/skills/` - Task-specific validation procedures
- `.agents/README.md` - Skills system documentation

## 11. Configuration Precedence

**Order** (highest to lowest):
1. Command-line flags
2. Environment variables
3. Config file `db/config.yaml`
4. Database-stored settings
5. Defaults

**Common defaults**:
- Port: `8000`
- DB path: `./db`
- Log path: `./logs`
- Default SQLite DSN: `sqlite://./db/sublink.db`

**For complete reference**: See `docs/configuration.md`

## 12. Frontend-Backend Contract

- Frontend requests go through `webs/src/api/request.js`
- API boundary: `/api/*`
- Subscription access: `/c/*`
- `SUBLINK_WEB_BASE_PATH` affects Web UI routing, not `/api/*` or `/c/*`

**skill-sublinkpro is also a REST API consumer**: 
- Any changes to `/api/v1/*` or `/c/*` must update `skill-sublinkpro/reference/api.md`
- Deployment/config changes must update `skill-sublinkpro/reference/deploy.md`
- New/moved docs must update `skill-sublinkpro/reference/docs.md`

## 13. Safety Notes

- Default admin credentials: `admin / 123456` (remind users to change)
- SQLite → MySQL/PostgreSQL migration requires manual restart
- Not compatible with upstream project databases
- Docker runtime directories: `/app/db`, `/app/template`, `/app/logs`

## 14. Branch Conventions

- `main` - Stable branch
- `dev` - Development branch (target for PRs)
- Semantic commit prefixes: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`

**For complete workflow**: See `CONTRIBUTING.md`

## 15. Validation Quick Reference

**Frontend changes**:
```bash
cd webs
yarn run lint                 # Always
yarn run build               # When routing/assets/build affected
```

**Backend changes**:
```bash
gofmt -w <changed-files>     # Format
golangci-lint run            # Lint
go test ./...                # Test
```

**For complete checklist**: See `.agents/skills/pre-commit-check/SKILL.md`

## 16. Need More Detail?

This file provides architectural overview and navigation. For specific procedures:

- **Setup and development**: `docs/development.md`
- **Build and deployment**: `docs/build-and-deployment.md`
- **Configuration**: `docs/configuration.md`
- **Security**: `docs/security-guidelines.md`
- **Common development patterns**: `docs/practical-recipes.md`
- **Contributing**: `CONTRIBUTING.md`
- **Theme work**: `docs/frontend-theme-guidelines.md`
- **i18n work**: `docs/internationalization.md`
- **Validation checklists**: `.agents/skills/`
- **Features**: `docs/features/`
- **Skill API**: `skill-sublinkpro/`

Don't reinvent or duplicate content that exists in these references. Read the relevant file when you need details.
