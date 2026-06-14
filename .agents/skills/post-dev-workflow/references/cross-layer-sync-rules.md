# Cross-Layer Synchronization Rules

Complete reference for cross-layer synchronization requirements in Phase 2 of the post-development workflow.

---

## When Cross-Layer Sync is Required

Cross-layer synchronization is **mandatory** when changes affect multiple layers:
- Backend + Frontend
- Code + Documentation
- Code + Configuration
- Backend + skill-sublinkpro API reference

---

## Synchronization Patterns

### Backend API → Frontend

| Backend Change | Frontend Must Update |
|---|---|
| New endpoint added | `webs/src/api/` - Add API client function |
| Endpoint path changed | `webs/src/api/` - Update request URL |
| Request parameters changed | `webs/src/api/` - Update function signature and request body |
| Response structure changed | Components using the data - Update state management and display logic |
| Error codes changed | Error handling components - Update error messages and i18n keys |
| Authentication requirements changed | `webs/src/api/request.js` - Update auth headers/middleware |

**Verification**:
```bash
# Backend
gofmt -w <changed-files>
golangci-lint run
go test ./api/...

# Frontend
cd webs
yarn run lint
yarn run build
```

### Backend API → skill-sublinkpro

| Backend Change | skill-sublinkpro Must Update |
|---|---|
| New `/api/v1/*` endpoint | `skill-sublinkpro/reference/api.md` - Document endpoint, params, response |
| New `/c/*` subscription endpoint | `skill-sublinkpro/reference/api.md` - Document subscription flow |
| Authentication flow changed | `skill-sublinkpro/reference/api.md` - Update auth section |
| Response structure changed | `skill-sublinkpro/reference/api.md` - Update response examples |

**skill-sublinkpro is a REST API consumer skill. Any changes to public API endpoints must be reflected in the reference documentation.**

### Frontend → Backend

| Frontend Change | Backend Must Verify |
|---|---|
| New data requirement | Backend provides the data in response |
| New validation requirement | Backend enforces validation rules |
| New permission requirement | Backend enforces permission checks |
| New sorting/filtering | Backend supports the query parameters |

**Verification**: Review backend API handlers and ensure they support the new frontend requirements.

### Configuration Changes → Multiple Layers

| Config Change | Must Update |
|---|---|
| New config option | Code + `docs/configuration.md` + `.zh-CN.md` + example configs + `skill-sublinkpro/reference/deploy.md` |
| Config semantics changed | Code + docs (both languages) + migration guide |
| Environment variable added | Code + docs + Dockerfile + docker-compose.yml |
| Default value changed | Code + docs + release notes (breaking change) |

**Verification**:
```bash
# Backend
grep -r "NEW_CONFIG_KEY" .
# Verify all locations updated

# Docs
ls docs/configuration*.md
cat docs/configuration.md | grep -i "new config key"
cat docs/configuration.zh-CN.md | grep -i "新配置项"
```

### User-Facing Feature → Documentation

| Feature Change | Must Update |
|---|---|
| New feature added | `README.md` + `.zh-CN.md` + `docs/features/*.md` + `.zh-CN.md` |
| Feature behavior changed | Relevant feature docs (both languages) |
| Feature removed | Remove from docs, add migration guide |
| Screenshots needed | Update/add screenshots in both language versions |

**Verification**: Review all documentation mentions of the feature area.

---

## Cross-Layer Checklist

### 1. Identify All Affected Layers

- [ ] Backend (`*.go`, `routers/`, `api/`, `services/`, `models/`)
- [ ] Frontend (`webs/src/`)
- [ ] Configuration (`docs/configuration.md`)
- [ ] User documentation (`README.md`, `docs/features/`)
- [ ] Developer documentation (`docs/development.md`, `CONTRIBUTING.md`)
- [ ] skill-sublinkpro API reference (`skill-sublinkpro/reference/api.md`)
- [ ] Deployment documentation (`docs/installation.md`, `skill-sublinkpro/reference/deploy.md`)
- [ ] Build/CI configuration (`.github/workflows/`, `Dockerfile`)

### 2. Synchronize Each Layer

For each affected layer:

- [ ] Make required changes
- [ ] Run validation commands for that layer
- [ ] Verify changes are consistent with other layers
- [ ] Update tests for that layer

### 3. Verify Cross-Layer Consistency

- [ ] Backend and frontend use same data structures
- [ ] Backend and frontend use same i18n keys
- [ ] Code and documentation describe same behavior
- [ ] Configuration and code use same option names
- [ ] Example configs match documented options

### 4. Document What Was Checked

Even if a layer didn't need changes, document that it was checked:

```
Cross-layer sync:
- Backend: Updated API endpoint ✅
- Frontend: Updated API client and components ✅
- skill-sublinkpro: Updated API reference ✅
- Configuration: Not affected (checked, no new config options) ⚠️
- User docs: Not affected (internal API change only) ⚠️
```

---

## Special Cases

### Protocol Changes

When adding or changing protocol support (`node/protocol/`):

- [ ] Backend: Implement protocol parser
- [ ] Backend: Add tests for protocol parsing
- [ ] Frontend: Update protocol support matrix display (if user-visible)
- [ ] Docs: Update `docs/features/` with protocol support (if new protocol)
- [ ] Docs: Update protocol compatibility table

**Verification**:
```bash
go test ./node/protocol/...
```

### Unlock Provider Changes

When adding or changing unlock providers (`services/unlock/`):

- [ ] Backend: Implement unlock checker
- [ ] Backend: Add tests for unlock logic
- [ ] Frontend: Usually no change (dynamically registered)
- [ ] Docs: Usually no change (dynamically registered)

**Note**: Unlock providers are usually dynamically registered. Only update docs if there are configuration requirements.

### Scheduler/Cron Changes

When adding or changing scheduled tasks (`services/scheduler/`):

- [ ] Backend: Implement task logic
- [ ] Backend: Register task in scheduler
- [ ] Backend: Add tests for task execution
- [ ] Configuration: Add cron expression config (if user-configurable)
- [ ] Docs: Document task behavior and configuration (if user-visible)

**Verification**:
```bash
go test ./services/scheduler/...
```

### Migration Changes

When adding database migrations (`models/db_migrate.go`):

- [ ] Backend: Implement up/down migrations
- [ ] Backend: Add tests for migration logic
- [ ] Docs: Add migration notes to release documentation
- [ ] Docs: Add rollback procedure (if breaking change)

**Verification**:
```bash
go test ./models/...
# Manual test: Run migration on test database
```

### mihomo Integration Changes

When changing mihomo integrations (`services/mihomo/`):

- [ ] Backend: Update mihomo adapter/wrapper
- [ ] Backend: Add tests for integration
- [ ] Docs: Update `AGENTS.md` mihomo integration section (if architecture changed)

**Verification**:
```bash
go test ./services/mihomo/...
```

---

## When Cross-Layer Sync Can Be Skipped

Skip Phase 2 only if:

- [ ] Pure documentation changes (no code affected)
- [ ] Internal refactoring (no contracts changed)
- [ ] Backend-only logic (no API/config/frontend impact)
- [ ] Frontend-only UI polish (no API/behavior changes)
- [ ] Test-only changes

**Document the skip reason** in your change summary:

```
Cross-layer sync: Skipped
Reason: Pure internal refactoring, no API contracts changed.
```

---

## Anti-Patterns

❌ **Changing backend API without updating frontend**
- Frontend will break or show stale data

❌ **Updating code without updating skill-sublinkpro API reference**
- AI agent skill will have outdated API documentation

❌ **Adding config option without updating docs**
- Users won't know the option exists

❌ **Changing response structure without updating frontend**
- Frontend will display incorrect/missing data

❌ **Updating only one language version of docs**
- Half the users won't see the update

❌ **Assuming layers are independent**
- Most changes affect multiple layers

---

## Verification Commands

### Full Cross-Layer Validation

```bash
# Backend
gofmt -w .
golangci-lint run
go test ./...

# Frontend
cd webs
yarn run lint
yarn run build
yarn test

# Documentation
# Manual review: Check all affected docs updated
# Manual review: Check both language versions updated

# skill-sublinkpro
# Manual review: Check API reference matches current endpoints
```

### Quick Cross-Layer Check

```bash
# Check for hardcoded strings (should use i18n)
grep -r "error.*:" api/ services/ | grep -v "i18nKey"

# Check for API endpoint references
grep -r "/api/v1/" webs/src/api/
grep -r "/api/v1/" skill-sublinkpro/reference/api.md

# Check for config references
grep -r "viper.Get" . | grep -v "_test.go"
grep -r "SUBLINK_" docs/configuration*.md
```

---

## Related Skills

- `.agents/skills/doc-sync-check/SKILL.md` - Documentation synchronization
- `.agents/skills/security-review/SKILL.md` - Security review for auth changes
- `.agents/skills/performance-check/SKILL.md` - Performance review for optimization
- `.agents/skills/theme-check/SKILL.md` - Theme synchronization for UI changes
