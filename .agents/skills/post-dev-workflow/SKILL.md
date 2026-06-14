# Post-Development Workflow Skill

**MANDATORY** automated workflow to run after completing ANY code change, before declaring work "complete".

## Purpose

This skill ensures that every code change goes through a complete validation, synchronization, and documentation pipeline before being considered "done". It prevents incomplete work, missing documentation, and broken cross-layer contracts.

## When to use this skill

**ALWAYS** use this skill when:
- Any code change is complete (backend, frontend, or both)
- Before declaring work "finished" or "done"
- Before preparing to commit changes
- Before creating or updating a pull request

**This is NOT optional for AI agents.** If an AI agent completes development without running this workflow, the work is incomplete.

## Workflow Phases

This skill orchestrates multiple validation phases in sequence:

```
Development Complete
    ↓
1. Code Validation (lint, format, build)
    ↓
2. Cross-Layer Sync Check (if multi-layer change)
    ↓
3. Documentation Sync Check (if behavior/API/config changed)
    ↓
4. Test Execution (if key logic changed)
    ↓
5. Change Summary (prepare commit message and PR description)
    ↓
Ready to Commit
```

---

## Phase 1: Code Validation

### Backend changes

Run from repository root:

```bash
# Format changed Go files
gofmt -w <changed-files>

# Lint
golangci-lint run

# Test (run relevant tests, or full suite if time permits)
go test ./...
```

**Required exit criteria**:
- [ ] `gofmt` produces no further changes
- [ ] `golangci-lint run` exits with status 0
- [ ] Relevant `go test` passes (or full `go test ./...` if possible)

### Frontend changes

Run from `webs/` directory:

```bash
cd webs

# Lint (always required)
yarn run lint

# Build (required if routing, assets, build config, or production integration affected)
yarn run build
```

**Optional auto-fix** (run if lint fails):
```bash
yarn run lint:fix
yarn run prettier
```

**Required exit criteria**:
- [ ] `yarn run lint` exits with status 0
- [ ] `yarn run build` succeeds (if applicable)
- [ ] No console errors during build

### When Phase 1 fails

**DO NOT proceed to Phase 2** until all validation passes. Fix issues first:
- Format errors → run auto-formatters
- Lint errors → fix code or adjust rules if justified
- Build errors → fix imports, paths, or config
- Test failures → fix the actual issue, don't skip tests

---

## Phase 2: Cross-Layer Synchronization Check

**Trigger condition**: Changes affect multiple layers (backend + frontend, or code + docs, or code + config)

### Quick checklist

Use `.agents/skills/cross-layer-sync/SKILL.md` for detailed verification.

**Common cross-layer patterns**:

| Change Type | Must Also Update |
|---|---|
| Backend API endpoint changed | Frontend `webs/src/api/`, `skill-sublinkpro/reference/api.md` |
| Backend response structure changed | Frontend display components, state management |
| Frontend behavior changed | Verify backend supports new flow |
| Configuration option added/changed | Code + `docs/configuration.md` + `.zh-CN.md` + example configs + `skill-sublinkpro/reference/deploy.md` |
| User-facing feature added/changed | Code + `docs/features/*.md` + `.zh-CN.md` + `README.md` |

**Required exit criteria**:
- [ ] All impacted layers identified
- [ ] All impacted layers synchronized
- [ ] Verification commands run for each layer
- [ ] Documented which layers were checked (even if no changes needed)

### When to skip Phase 2

Skip only if the change is truly isolated to one layer:
- Pure documentation changes (no code affected)
- Internal refactoring with no contract changes
- Backend-only logic with no API/config/frontend impact
- Frontend-only UI polish with no API/behavior changes

**Document the skip reason** in your change summary.

---

## Phase 3: Documentation Synchronization Check

**Trigger condition**: Changes affect user-visible behavior, APIs, configuration, deployment, or developer workflows

### Quick checklist

Use `.agents/skills/doc-sync-check/SKILL.md` for detailed verification.

**What needs documentation updates**:

| Change Type | Docs to Update |
|---|---|
| User-facing feature | `README.md` + `.zh-CN.md`, `docs/features/*.md` + `.zh-CN.md` |
| API endpoint | `skill-sublinkpro/reference/api.md` |
| Configuration | `docs/configuration.md` + `.zh-CN.md`, example configs |
| Deployment | `docs/installation.md` + `.zh-CN.md`, `skill-sublinkpro/reference/deploy.md` |
| Developer workflow | `docs/development.md` + `.zh-CN.md`, `CONTRIBUTING.md` + `.zh-CN.md` |
| Architecture | `AGENTS.md` |
| New protocol | `docs/features/*.md`, protocol support matrix |
| New unlock provider | Usually no doc change (dynamically registered) |

**Bilingual requirement**:
- [ ] Both English (`.md`) and Chinese (`.zh-CN.md`) versions updated
- [ ] Language switch links work
- [ ] Content semantically equivalent

**Required exit criteria**:
- [ ] All affected documentation identified
- [ ] Both language versions updated
- [ ] Links verified (no broken references)
- [ ] Code examples accurate and tested
- [ ] Documentation map updated (`skill-sublinkpro/reference/docs.md`) if new docs added

### When to skip Phase 3

Skip only if:
- Pure internal refactoring (no user-visible changes)
- Bug fix that restores documented behavior (not new behavior)
- Test-only changes

**Document the skip reason** in your change summary.

---

## Phase 4: Test Execution

**Trigger condition**: Changes affect key business logic, APIs, permissions, configuration semantics, migrations, scheduled tasks, mihomo integrations, protocol parsing, or data transformations

### What needs tests

**Backend tests required when**:
- [ ] Added or changed business logic in `services/`
- [ ] Added or changed API handler in `api/`
- [ ] Added or changed permission checks in `middlewares/`
- [ ] Added or changed database migration in `models/db_migrate.go`
- [ ] Added or changed scheduled task in `services/scheduler/`
- [ ] Added or changed protocol in `node/protocol/`
- [ ] Added or changed unlock checker in `services/unlock/`
- [ ] Fixed a bug (add regression test)

**Frontend tests**:
- Currently optional (no frontend test framework wired in)
- Only add if explicitly requested or if module already has tests

### Test requirements

```bash
# Backend: run relevant tests
go test ./services/scheduler/...  # Example: if scheduler changed
go test ./...                     # Full suite if time permits
```

**Test quality**:
- [ ] Tests cover happy path, boundaries, and error cases
- [ ] Tests are isolated (no execution order dependency)
- [ ] Test names describe scenario and expected outcome
- [ ] Regression tests added for bug fixes

**Required exit criteria**:
- [ ] Relevant tests exist
- [ ] All tests pass
- [ ] Coverage is reasonable for the changed area

### When to skip Phase 4

Skip only if:
- Pure documentation changes
- Pure UI styling changes (no logic)
- Refactoring with existing test coverage

**Document the skip reason** in your change summary.

---

## Phase 5: Change Summary

Prepare a comprehensive summary of what was done, why, and how it was validated.

### Commit message format

Use semantic commit prefixes:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type**:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation only
- `refactor:` - Code refactoring
- `test:` - Test additions/updates
- `chore:` - Build, deps, tooling

**Scope**: Component area (e.g., `airports`, `auth`, `theme`, `i18n`, `scheduler`, `unlock`)

**Subject**: Concise description (≤72 chars if possible)

**Body** (optional): More details, why the change was needed, what approach was taken

**Footer** (optional): Related issues, breaking changes

### Examples

```
feat(airports): support batch subscription updates

Added batch update dialog with progress tracking.
Users can now select multiple airports and update them in parallel.

Closes #123
```

```
fix(auth): enhance SSE authentication error handling

SSE authentication failures now return proper i18n error messages.
Added retry logic for transient network errors.

Fixes #456
```

```
docs(i18n): update internationalization requirements

Clarified frontend i18n key naming conventions.
Added examples for backend i18nKey + i18nParams usage.
```

### Change summary checklist

Document in your summary:

**What changed**:
- [ ] List files or areas modified
- [ ] Describe the change at a high level

**Why**:
- [ ] What problem this solves
- [ ] What feature this adds
- [ ] What behavior this fixes

**How validated**:
- [ ] Which layers were checked (backend, frontend, docs)
- [ ] Which validation commands were run (lint, test, build)
- [ ] Which cross-layer sync was performed
- [ ] Which documentation was updated
- [ ] Manual testing performed (if applicable)

**Layers not changed** (if cross-layer check performed):
- [ ] Document which layers were checked but didn't need changes
- [ ] Brief reason why no sync was needed

**Breaking changes** (if any):
- [ ] What breaks
- [ ] Migration path for users
- [ ] Configuration changes required

---

## Complete Workflow Example

### Scenario: Adding a new API endpoint for batch airport updates

#### Phase 1: Code Validation

```bash
# Backend
gofmt -w api/clients.go services/subscription_service.go
golangci-lint run
go test ./api/... ./services/...

# Frontend
cd webs
yarn run lint
yarn run build
```

**Result**: All validation passes ✅

#### Phase 2: Cross-Layer Sync

**Layers affected**:
- ✅ Backend: `api/clients.go`, `services/subscription_service.go`
- ✅ Frontend: `webs/src/api/airports.js`, `webs/src/views/airports/AirportBatchUpdateDialog.jsx`
- ✅ i18n: Added `zh-CN` and `en-US` translations for batch update dialog
- ⚠️ Config: Not affected (no new config options)

**Verification**:
- Backend lint, test passed
- Frontend lint, build passed
- Both languages added

**Result**: Cross-layer sync complete ✅

#### Phase 3: Documentation Sync

**Docs updated**:
- ✅ `skill-sublinkpro/reference/api.md` - Added `POST /api/v1/airports/batch-update` endpoint
- ⚠️ `docs/features/` - No user-facing docs needed (feature accessed through UI only)
- ⚠️ `README.md` - No change (not a headline feature)

**Bilingual**: API reference is English-only (technical docs)

**Result**: Documentation sync complete ✅

#### Phase 4: Test Execution

**Tests added**:
- ✅ `api/clients_test.go` - Added `TestBatchUpdateAirports` (happy path, partial failure, all failure cases)
- ✅ `services/subscription_service_test.go` - Added batch update logic tests

**Test run**:
```bash
go test ./api/... ./services/...
# PASS
```

**Result**: Tests added and passing ✅

#### Phase 5: Change Summary

**Commit message**:
```
feat(airports): support batch subscription updates

Added batch update dialog with progress tracking and parallel execution.
Users can select multiple airports and update them concurrently.

Backend:
- Added POST /api/v1/airports/batch-update endpoint
- Added batch update logic in subscription service
- Added concurrent update with error collection

Frontend:
- Added AirportBatchUpdateDialog component
- Added progress tracking UI
- Added i18n for both zh-CN and en-US

Tests:
- Added API handler tests for batch update
- Added service layer tests for parallel execution
- Covered happy path, partial failure, and full failure cases

Docs:
- Updated skill-sublinkpro/reference/api.md with new endpoint

Closes #789
```

**Result**: Change summary prepared ✅

---

## Delivery Checklist

Before declaring work "complete", verify:

- [ ] **Phase 1**: All validation commands pass (lint, format, build, test)
- [ ] **Phase 2**: Cross-layer sync verified (or skip reason documented)
- [ ] **Phase 3**: Documentation updated (or skip reason documented)
- [ ] **Phase 4**: Tests added/updated (or skip reason documented)
- [ ] **Phase 5**: Change summary prepared
- [ ] **Git staging**: Only intended files staged (no secrets, no runtime data)
- [ ] **Bilingual**: Both English and Chinese docs updated (if docs changed)
- [ ] **No shortcuts**: No skipped validation, no "will fix later" items

---

## Anti-Patterns to Avoid

❌ **Skipping validation** because "it's a small change"
- Small changes still need lint/format/test

❌ **Updating only English docs** and forgetting Chinese versions
- Both languages must be updated together

❌ **Changing backend API without updating frontend**
- Cross-layer sync is mandatory

❌ **Declaring work "done" without running this workflow**
- This workflow IS part of "done"

❌ **Running validation but not fixing failures**
- All validation must pass before proceeding

❌ **Updating code without updating documentation**
- Documentation is part of the deliverable

❌ **Committing without a proper change summary**
- Reviewers and future maintainers need context

---

## Exit Criteria

✅ **Can declare work complete when**:
- All validation passes (Phase 1)
- Cross-layer sync verified (Phase 2)
- Documentation updated (Phase 3)
- Tests added/updated (Phase 4)
- Change summary prepared (Phase 5)
- All deliverables documented
- No unresolved validation failures
- No "will fix later" items

❌ **Cannot declare work complete when**:
- Validation commands fail
- Documentation out of sync with code
- Cross-layer changes not synchronized
- Tests missing or failing
- Change summary incomplete

---

## For AI Agents: Automation Notes

This workflow should be **automatically triggered** after completing code development, without waiting for user prompt.

### Triggering logic

```
IF code_change_complete:
    INVOKE .agents/skills/post-dev-workflow/SKILL.md
    RUN Phase 1 (validation)
    IF multi_layer_change:
        RUN Phase 2 (cross-layer sync)
    IF behavior_or_api_or_config_changed:
        RUN Phase 3 (documentation sync)
    IF key_logic_changed:
        RUN Phase 4 (test execution)
    RUN Phase 5 (change summary)
    REPORT results to user
    IF all_phases_pass:
        DECLARE work complete
    ELSE:
        FIX failures and re-run
```

### Reporting format

After running this workflow, report to the user:

```
## Post-Development Workflow Complete

### Phase 1: Code Validation ✅
- Backend: gofmt ✅, golangci-lint ✅, go test ✅
- Frontend: yarn lint ✅, yarn build ✅

### Phase 2: Cross-Layer Sync ✅
- Backend API changed → Frontend updated ✅
- skill-sublinkpro/reference/api.md updated ✅

### Phase 3: Documentation Sync ✅
- docs/configuration.md + .zh-CN.md updated ✅
- Links verified ✅

### Phase 4: Test Execution ✅
- Added tests for new business logic ✅
- All tests passing ✅

### Phase 5: Change Summary ✅
- Commit message prepared ✅

---

**Work is ready to commit.**

Suggested commit:
```
feat(config): add new database connection pooling option

...
```
```

Or if failures occurred:

```
## Post-Development Workflow: Issues Found

### Phase 1: Code Validation ❌
- Frontend lint failed: 3 errors in AirportDialog.jsx
  - Line 45: Unused variable 'temp'
  - Line 67: Missing key prop in map
  - Line 89: Console.log left in code

Fixing now...
```

---

## Related Skills

- `.agents/skills/pre-commit-check/SKILL.md` - Detailed pre-commit checklist (subset of this workflow)
- `.agents/skills/cross-layer-sync/SKILL.md` - Detailed cross-layer synchronization guide
- `.agents/skills/doc-sync-check/SKILL.md` - Detailed documentation sync guide
- `.agents/skills/theme-check/SKILL.md` - Theme-specific validation (when UI colors/surfaces changed)

---

## For Human Contributors

This workflow is also useful for manual development:

1. After completing your code change, open this file
2. Go through each phase sequentially
3. Check off items as you complete them
4. Don't skip phases unless justified
5. Document your validation in your commit message

This ensures consistent quality across all contributions, whether from AI or humans.
