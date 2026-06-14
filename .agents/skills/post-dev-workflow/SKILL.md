---
name: post-dev-workflow
description: "Mandatory post-development workflow orchestrating validation, synchronization, and testing phases. Automatically invoked by AI agents after code changes. Not for manual use."
version: "1.0.0"
author: "SublinkPro Team"
user-invocable: false
---

# Post-Development Workflow Skill

**MANDATORY** automated workflow to run after completing ANY code change, before declaring work "complete".

---

## When to Use This Skill

**ALWAYS** use this skill when:
- Any code change is complete (backend, frontend, or both)
- Before declaring work "finished" or "done"
- Before preparing to commit changes
- Before creating or updating a pull request

**This is NOT optional for AI agents.** If an AI agent completes development without running this workflow, the work is incomplete.

---

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

Run validation commands for all changed layers.

### Backend Changes

```bash
gofmt -w <changed-files>    # Format
golangci-lint run           # Lint
go test ./...               # Test
```

**Required exit criteria**:
- [ ] `gofmt` produces no further changes
- [ ] `golangci-lint run` exits with status 0
- [ ] Relevant `go test` passes

### Frontend Changes

```bash
cd webs
yarn run lint               # Always required
yarn run build              # If routing/assets/build affected
```

**Optional auto-fix**:
```bash
yarn run lint:fix
yarn run prettier
```

**Required exit criteria**:
- [ ] `yarn run lint` exits with status 0
- [ ] `yarn run build` succeeds (if applicable)

### When Phase 1 Fails

**DO NOT proceed to Phase 2** until all validation passes. Fix issues first.

**Detailed validation commands and troubleshooting**: See `references/validation-commands.md`

---

## Phase 2: Cross-Layer Synchronization Check

**Trigger condition**: Changes affect multiple layers (backend + frontend, code + docs, code + config)

**Additional checks**: For security-critical or performance-sensitive changes, invoke specialized skills:
- **Security changes**: Also invoke `.agents/skills/security-review/SKILL.md`
- **Performance-critical changes**: Also invoke `.agents/skills/performance-check/SKILL.md`

### Quick Checklist

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
- [ ] Documented which layers were checked

### When to Skip Phase 2

Skip only if the change is truly isolated to one layer. **Document the skip reason** in your change summary.

**Detailed cross-layer sync patterns and verification**: See `references/cross-layer-sync-rules.md`

---

## Phase 3: Documentation Synchronization Check

**Trigger condition**: Changes affect user-visible behavior, APIs, configuration, deployment, or developer workflows

### Quick Checklist

| Change Type | Docs to Update |
|---|---|
| User-facing feature | `README.md` + `.zh-CN.md`, `docs/features/*.md` + `.zh-CN.md` |
| API endpoint | `skill-sublinkpro/reference/api.md` |
| Configuration | `docs/configuration.md` + `.zh-CN.md`, example configs |
| Deployment | `docs/installation.md` + `.zh-CN.md`, `skill-sublinkpro/reference/deploy.md` |
| Developer workflow | `docs/development.md` + `.zh-CN.md`, `CONTRIBUTING.md` + `.zh-CN.md` |
| Architecture | `AGENTS.md` |

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

### When to Skip Phase 3

Skip only if:
- Pure internal refactoring (no user-visible changes)
- Bug fix that restores documented behavior (not new behavior)
- Test-only changes

**Document the skip reason** in your change summary.

**Detailed documentation sync patterns and templates**: See `references/documentation-sync-rules.md`

---

## Phase 4: Test Execution

**Trigger condition**: Changes affect key business logic, APIs, permissions, configuration semantics, migrations, scheduled tasks, mihomo integrations, protocol parsing, or data transformations

### What Needs Tests

**Backend tests required when**:
- [ ] Added or changed business logic in `services/`
- [ ] Added or changed API handler in `api/`
- [ ] Added or changed permission checks in `middlewares/`
- [ ] Added or changed database migration in `models/db_migrate.go`
- [ ] Added or changed scheduled task in `services/scheduler/`
- [ ] Added or changed protocol in `node/protocol/`
- [ ] Fixed a bug (add regression test)

**Frontend tests**:
- [ ] Added or changed utility functions in `webs/src/utils/`
- [ ] Added or changed complex components with business logic
- [ ] Added or changed API client functions in `webs/src/api/`
- [ ] Fixed a bug (add regression test)

### Test Requirements

```bash
# Backend: run relevant tests
go test ./services/scheduler/...  # Example: if scheduler changed
go test ./...                     # Full suite if time permits

# Frontend: run tests with Vitest
cd webs
yarn test                         # Run all tests
yarn test --run                   # Run without watch mode
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

### When to Skip Phase 4

Skip only if:
- Pure documentation changes
- Pure UI styling changes (no logic)
- Refactoring with existing test coverage

**Document the skip reason** in your change summary.

---

## Phase 5: Change Summary

Prepare a comprehensive summary of what was done, why, and how it was validated.

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type**: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `style`, `perf`

**Scope**: Component area (e.g., `airports`, `auth`, `theme`, `i18n`, `scheduler`)

**Subject**: Concise description (≤72 chars, imperative mood, no period)

### Examples

```
feat(airports): add batch subscription update

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

### Change Summary Checklist

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

**Breaking changes** (if any):
- [ ] What breaks
- [ ] Migration path for users
- [ ] Configuration changes required

**Detailed commit message guide and templates**: See `references/commit-message-guide.md` and `assets/commit-template.txt`

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

### Triggering Logic

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

### Reporting Format

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
feat(config): add database connection pooling option

...
```
```

Or if failures occurred:

```
## Post-Development Workflow: Issues Found

### Phase 1: Code Validation ❌
- Frontend lint failed: 3 errors in AirportDialog.jsx

Fixing now...
```

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

## Related Skills

- `.agents/skills/pre-commit-check/SKILL.md` - Detailed pre-commit checklist (subset of this workflow)
- `.agents/skills/cross-layer-sync/SKILL.md` - Detailed cross-layer synchronization guide
- `.agents/skills/doc-sync-check/SKILL.md` - Detailed documentation sync guide
- `.agents/skills/theme-check/SKILL.md` - Theme-specific validation (when UI colors/surfaces changed)
- `.agents/skills/security-review/SKILL.md` - Security review checklist (for auth/sensitive data changes)
- `.agents/skills/performance-check/SKILL.md` - Performance review checklist (for optimization work)

---

## Reference Documentation

Detailed information extracted to separate files for maintainability:

- **`references/validation-commands.md`** - Complete validation command reference, troubleshooting, and quality standards
- **`references/cross-layer-sync-rules.md`** - Detailed cross-layer synchronization patterns, verification commands, and special cases
- **`references/documentation-sync-rules.md`** - Documentation synchronization requirements, templates, and bilingual standards
- **`references/commit-message-guide.md`** - Commit message format, examples, and best practices
- **`assets/commit-template.txt`** - Commit message template for git config

---

## For Human Contributors

This workflow is also useful for manual development:

1. After completing your code change, open this file
2. Go through each phase sequentially
3. Check off items as you complete them
4. Don't skip phases unless justified
5. Document your validation in your commit message

This ensures consistent quality across all contributions, whether from AI or humans.
