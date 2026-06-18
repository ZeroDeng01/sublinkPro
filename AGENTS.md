# AGENTS.md

This file provides AI and human contributors with a high-level architectural guide and navigation map for SublinkPro.

---

## 0. MANDATORY PRE-COMMIT VALIDATION - NON-NEGOTIABLE REQUIREMENT

**⚠️ ABSOLUTE REQUIREMENT FOR ALL AI AGENTS ⚠️**

This is a **HARD STOP** requirement. No exceptions. No shortcuts. No "will fix later". No "just this once".

### Red Line Rule: NO COMMITS WITHOUT VALIDATION

```
┌─────────────────────────────────────────────────────────────┐
│  🛑 BLOCKING REQUIREMENT - READ THIS FIRST                  │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  BEFORE any git commit, git add, or PR creation:            │
│                                                              │
│  1. You MUST invoke: pre-commit-check skill                 │
│  2. You MUST execute ALL applicable validation commands     │
│  3. You MUST fix ALL failures before proceeding             │
│  4. You MUST document what was validated                    │
│                                                              │
│  Violations of this rule constitute incomplete work.        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Mandatory Trigger Points

You **MUST** invoke `.agents/skills/pre-commit-check/SKILL.md` **BEFORE**:

1. **Running `git add` or `git commit`** - Whether initiated by you or requested by user
2. **Creating or updating a PR** - Via `gh pr create` or any PR command
3. **Declaring work "complete", "done", or "finished"** - In any form
4. **User explicitly asks to commit** - Even if they say "skip checks" (refuse and explain why)
5. **Any code change is ready for commit** - Backend, frontend, docs, config, or mixed

### The Validation-Before-Commit Workflow

```
Code Change Complete
    ↓
    ├─> Did you invoke pre-commit-check skill? 
    │   ├─> NO  → STOP. Invoke it now. DO NOT PROCEED.
    │   └─> YES → Continue
    ↓
    ├─> Did ALL validation commands pass?
    │   ├─> NO  → STOP. Fix failures. Re-run validation. DO NOT PROCEED.
    │   └─> YES → Continue
    ↓
    ├─> Did you stage ONLY intended changes?
    │   ├─> NO  → STOP. Review staged files. Remove unintended. DO NOT PROCEED.
    │   └─> YES → Continue
    ↓
    ├─> Did you prepare a proper commit message?
    │   ├─> NO  → STOP. Prepare semantic commit message. DO NOT PROCEED.
    │   └─> YES → Continue
    ↓
Ready to Commit (present to user for final review)
```

### What "Validation Complete" Means

Validation is NOT complete until **ALL** of the following are verified:

#### ✅ Backend Changes (if any .go files changed):
- [ ] `gofmt -w` executed on all changed Go files
- [ ] `gofmt -l` returns no files (all formatted)
- [ ] `golangci-lint run` exits with status 0 (no errors)
- [ ] `go test ./...` passes (or at minimum, relevant package tests pass)
- [ ] No compilation errors

#### ✅ Frontend Changes (if any webs/ files changed):
- [ ] `cd webs && yarn run lint` exits with status 0 (no errors)
- [ ] `cd webs && yarn run build` succeeds (if routing/assets/build affected)
- [ ] No build warnings about missing imports or broken dependencies

#### ✅ Cross-Layer Sync (if multi-layer change):
- [ ] Backend API changes → Frontend updated
- [ ] Frontend contract changes → Backend verified
- [ ] Config changes → Code + docs updated
- [ ] Skill verified: `.agents/skills/cross-layer-sync/SKILL.md` (if complex)

#### ✅ Documentation Sync (if behavior/API/config changed):
- [ ] User-facing docs updated (both `.md` and `.zh-CN.md`)
- [ ] API docs updated (`skill-sublinkpro/reference/api.md`)
- [ ] Config docs updated (`docs/configuration.md` + `.zh-CN.md`)
- [ ] Links verified (no broken references)
- [ ] Skill verified: `.agents/skills/doc-sync-check/SKILL.md` (if complex)

#### ✅ Test Coverage (if key logic changed):
- [ ] Tests added for new business logic
- [ ] Tests updated for changed behavior
- [ ] Regression tests added for bug fixes
- [ ] All tests passing

#### ✅ Git Staging Verification:
- [ ] Only intended files staged (no accidental includes)
- [ ] No sensitive files staged (`.env`, credentials, keys)
- [ ] No runtime data staged (`db/`, `logs/`, `cache/`, `out/`)
- [ ] No large binary files unintentionally staged
- [ ] No AI agent temporary files staged:
  - No `*_SUMMARY.md`, `*_REPORT.md` in skill directories
  - No `QUICK_REFERENCE.md` files (execution artifacts)
  - No `.claude/projects/`, `.claude/sessions/`, `.claude/plans/` runtime files
  - No agent execution logs or temporary outputs

#### ✅ Commit Message Prepared:
- [ ] Semantic prefix used (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`)
- [ ] Subject line ≤72 characters
- [ ] Body explains what, why, and how validated
- [ ] References issues if applicable (`Closes #123`, `Fixes #456`)

### Forbidden Actions (Will Be Rejected)

❌ **"Let me commit this first, I'll fix validation later"**
- **Rejection**: No. Fix validation now. Commits must be clean.

❌ **"This is a small change, we can skip validation"**
- **Rejection**: No. Small changes still need validation. No exceptions.

❌ **"The user said skip checks"**
- **Rejection**: No. Explain to user why checks are mandatory. Refuse to skip.

❌ **"Validation failed but I'll commit anyway"**
- **Rejection**: No. Fix failures first. Do not proceed with broken validation.

❌ **"I'll update docs in a separate commit"**
- **Rejection**: No. Docs are part of the same change. Update together.

❌ **"Tests will be added later"**
- **Rejection**: No. Tests are part of the deliverable. Add them now.

❌ **Auto-committing without user verification**
- **Rejection**: No. Stage changes and let user verify before committing.

### How to Handle User Requests to Skip Validation

When a user says "just commit it" or "skip the checks":

**Your response should be**:

```
I cannot commit without running validation checks. This is a mandatory 
requirement for code quality and consistency.

The checks take only a few minutes and ensure:
- No linting errors that will fail CI
- No broken builds
- No test failures
- Cross-layer synchronization
- Documentation consistency

I'll run the validation now. If everything passes, I'll stage the changes 
and you can commit immediately.
```

**Then proceed to invoke the pre-commit-check skill and run all validations.**

### Detailed Validation Procedures

For complete validation procedures, checklists, and commands, see:

**Primary Skill**: `.agents/skills/pre-commit-check/SKILL.md`
- Complete validation command reference
- Layer-by-layer validation steps
- Troubleshooting guide
- Exit criteria

**Supporting Skills**:
- `.agents/skills/post-dev-workflow/SKILL.md` - Full post-development orchestration
- `.agents/skills/cross-layer-sync/SKILL.md` - Cross-layer synchronization guide
- `.agents/skills/doc-sync-check/SKILL.md` - Documentation synchronization
- `.agents/skills/theme-check/SKILL.md` - Theme/UI validation
- `.agents/skills/security-review/SKILL.md` - Security validation
- `.agents/skills/performance-check/SKILL.md` - Performance validation

### For Human Contributors

This rule applies to AI agents. Human contributors should also follow these practices, but are trusted to use their judgment.

If you're a human reading this:
- We recommend using the same validation workflow
- Consider setting up a git pre-commit hook to automate checks
- See `CONTRIBUTING.md` for detailed contribution guidelines

### Summary: The Three Laws of Pre-Commit Validation

1. **Validation First**: No commit without validation. No exceptions.
2. **All Checks Pass**: All validation commands must exit successfully. Fix failures before proceeding.
3. **User Verification**: Stage changes and present to user. Do not auto-commit.

---

## 0.1. CRITICAL: USE PROJECT-LOCAL `.agents/.tmp/` FOR TEMPORARY FILES

**⚠️ ABSOLUTE PROHIBITION ⚠️**

AI agents are **FORBIDDEN** from creating temporary, debug, scratch, or generated explanation files anywhere in the repository except the project-local `.agents/.tmp/` directory.

Do **not** write agent work files to the operating system temp directory. For this project, `/tmp`, `/var/tmp`, and other system temp locations are not allowed for agent debugging, scratch work, temporary tests, logs, screenshots, reports, or intermediate outputs.

### Prohibited Actions

❌ **DO NOT create ANY of the following outside `.agents/.tmp/`:**
- Documentation files: `*.md` (except when explicitly requested by user to update existing docs)
- Summary files: `SUMMARY.md`, `IMPLEMENTATION_*.md`, `FEATURE_*.md`, `TEST_*.md`, etc.
- Temporary test files: `test_*.go`, `test_*.js`, `debug_*.py`, etc.
- Debug output files: `output.txt`, `result.json`, `debug.log`, etc.
- Notes or scratch files: `notes.txt`, `TODO.md`, `scratch.py`, etc.
- ANY file with names like: `QUICK_REFERENCE.md`, `REPORT.md`, etc.

### Allowed Locations for Temporary Files

✅ **ONLY use this location for agent temporary/test/debug files:**
- `.agents/.tmp/` at the repository root

Create `.agents/.tmp/` if it does not exist, keep any contents disposable, and remove files when they are no longer needed. `.agents/.tmp/` is ignored by git and must never be used for source code, documentation deliverables, configuration, or assets that need to ship.

### Examples

**❌ WRONG - DO NOT DO THIS:**
```bash
# Creating summary in project root - FORBIDDEN
echo "# Summary" > IMPLEMENTATION_SUMMARY.md

# Creating test file in project - FORBIDDEN  
cat > test_feature.go << 'EOF'

# Using system temp directories - FORBIDDEN for this project
cat > /tmp/test_feature.go << 'EOF'
```

**✅ CORRECT - Do this instead:**
```bash
# Use project-local .tmp/ for temporary files
mkdir -p .tmp
echo "# Summary" > .agents/.tmp/implementation_summary.md

# Use project-local .tmp/ for test files
cat > .agents/.tmp/test_feature.go << 'EOF'
```

### Clean Up After Yourself

- If you accidentally create temporary files outside `.agents/.tmp/`, DELETE them immediately
- Use `rm -f <file>` to remove any temporary files you created
- Check `git status --short` before finishing to ensure no temporary files are tracked or visible outside `.agents/.tmp/`

### Violation Consequences

Creating temporary files outside `.agents/.tmp/`, or using system temp directories for this project, is **UNACCEPTABLE** and shows lack of respect for the codebase.

**DO NOT make this mistake.**

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
- `webs/src/components/` - Shared frontend components; feature-local components must stay under their owning `webs/src/views/<feature>/component(s)/` directory unless intentionally shared
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

- **post-dev-workflow** - Mandatory post-development workflow orchestrating validation, synchronization, and testing phases (automatically invoked by AI agents)
- **pre-commit-check** - Pre-commit validation checklist for linting, formatting, building, and testing
- **cross-layer-sync** - Cross-layer synchronization guide for backend, frontend, and documentation changes
- **doc-sync-check** - Documentation synchronization checklist for API, configuration, and feature changes
- **theme-check** - Theme adaptation checklist for light/dark mode and UI component changes
- **security-review** - Security review checklist for authentication, authorization, and sensitive data handling
- **performance-check** - Performance review checklist for optimization, scalability, and resource usage

Skills follow the official Claude Code skill specification - freeform markdown with optional frontmatter. Each skill includes metadata indicating whether it's user-invocable or workflow-invoked.

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
