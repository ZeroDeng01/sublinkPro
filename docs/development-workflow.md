English | [简体中文](development-workflow.zh-CN.md)

# Development Workflow

This document describes the complete development workflow for SublinkPro, from starting a task to committing changes.

---

## Overview

The development workflow ensures that every code change goes through consistent validation, synchronization, and documentation processes before being considered "complete".

```
┌─────────────────────────────────────────────────────────────┐
│                    Development Workflow                      │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Phase 1: Task Understanding & Planning                      │
│  - Read relevant documentation                               │
│  - Understand architectural boundaries                       │
│  - Identify impacted layers                                  │
│  - Plan approach                                             │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Phase 2: Implementation                                     │
│  - Write code following project patterns                     │
│  - Respect layer boundaries                                  │
│  - Add comments for complex logic                            │
│  - Follow naming conventions                                 │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Phase 3: Post-Development Workflow (MANDATORY)              │
│  → Automatically triggered after code completion             │
│  → See .agents/skills/post-dev-workflow/SKILL.md             │
│                                                               │
│  Sub-Phase 3.1: Code Validation                              │
│    - Backend: gofmt, golangci-lint, go test                  │
│    - Frontend: yarn lint, yarn build (if applicable)         │
│                                                               │
│  Sub-Phase 3.2: Cross-Layer Synchronization                  │
│    - Verify all impacted layers are updated                  │
│    - Backend ↔ Frontend sync                                 │
│    - Code ↔ Documentation sync                               │
│    - API ↔ skill-sublinkpro sync                             │
│                                                               │
│  Sub-Phase 3.3: Documentation Synchronization                │
│    - Update relevant documentation                           │
│    - Maintain bilingual consistency (EN + ZH)                │
│    - Verify links and examples                               │
│                                                               │
│  Sub-Phase 3.4: Test Execution                               │
│    - Add tests for key logic changes                         │
│    - Run relevant test suites                                │
│    - Verify all tests pass                                   │
│                                                               │
│  Sub-Phase 3.5: Change Summary                               │
│    - Prepare commit message                                  │
│    - Document what/why/how                                   │
│    - List validation results                                 │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Phase 4: Commit & Pull Request                              │
│  - Review git staging (only intended files)                  │
│  - Commit with semantic message                              │
│  - Push to dev branch                                        │
│  - Create pull request with complete description             │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Phase 5: Review & Merge                                     │
│  - Automated PR checks run                                   │
│  - Address review feedback                                   │
│  - Merge to dev after approval                               │
└─────────────────────────────────────────────────────────────┘
```

---

## Phase 1: Task Understanding & Planning

### Before writing code

1. **Read architectural guidelines**:
   - `AGENTS.md` - Architectural overview
   - `docs/development.md` - Development setup and patterns
   - `CONTRIBUTING.md` - Contribution workflow

2. **Understand the scope**:
   - Which layers are affected? (backend, frontend, docs, config)
   - What are the entry points?
   - What existing patterns should be followed?

3. **Identify cross-layer impacts**:
   - Does this change affect APIs? → Frontend and skill-sublinkpro need updates
   - Does this change affect configuration? → Docs and examples need updates
   - Does this change affect user behavior? → Feature docs need updates

4. **Plan the approach**:
   - What files need to be modified?
   - What new files need to be created?
   - What tests need to be added?
   - What documentation needs to be updated?

### Key questions to answer

- [ ] What problem am I solving?
- [ ] Which layer(s) does this change affect?
- [ ] What existing code/patterns can I reuse?
- [ ] What are the architectural constraints?
- [ ] What documentation will need updates?

---

## Phase 2: Implementation

### Code quality standards

**Follow existing patterns**:
- Match surrounding code style
- Respect layer boundaries (don't put business logic in `routers/`)
- Use existing utilities and helpers
- Follow naming conventions

**Add meaningful comments**:
- Explain **why**, not **what**
- Document complex business logic
- Document cross-layer contracts
- Document non-obvious algorithms
- Keep comments updated with code changes

**Write tests for**:
- Key business logic
- API contracts
- Permission checks
- Configuration semantics
- Database migrations
- Protocol parsing
- Data transformations

### Backend development

```bash
# Run backend locally
go mod download
go run main.go
```

**Key principles**:
- Standard Go style
- Handle errors properly
- Use GORM for database operations
- Follow existing service patterns
- Add comments in Chinese for business logic

### Frontend development

```bash
# Run frontend locally
cd webs
yarn install
yarn run start
```

**Key principles**:
- React 19 + Vite
- Material UI components
- Use i18next for translations
- Follow existing theme patterns
- Responsive design (desktop + mobile)
- Light + dark mode support

### i18n requirements

**Frontend**:
- Add translations for both `zh-CN` and `en-US`
- Use `useTranslation()` hook or `<Trans>` component
- Translation keys should be semantic, not literal text

**Backend**:
- API responses: use `i18nKey` + `i18nParams` for Web UI display
- Keep `msg` field for backward compatibility
- Error messages: provide both machine-readable codes and i18n keys

**Documentation**:
- Maintain both `.md` (English) and `.zh-CN.md` (Chinese)
- Keep language switch links working
- Content should be semantically equivalent, not machine-translated

---

## Phase 3: Post-Development Workflow (MANDATORY)

**This phase is MANDATORY and must run automatically after completing ANY code change.**

For complete details, see: `.agents/skills/post-dev-workflow/SKILL.md`

### Sub-Phase 3.1: Code Validation

**Backend validation**:
```bash
# Format
gofmt -w <changed-files>

# Lint
golangci-lint run

# Test
go test ./...
```

**Frontend validation**:
```bash
cd webs

# Lint (always required)
yarn run lint

# Build (required when routing/assets/build affected)
yarn run build
```

**Exit criteria**: All validation commands pass with no errors.

### Sub-Phase 3.2: Cross-Layer Synchronization

**Verify synchronization when**:
- Backend API changed → Update frontend + skill-sublinkpro/reference/api.md
- Frontend behavior changed → Verify backend supports it
- Configuration changed → Update docs + example configs
- Feature added/changed → Update feature docs

**See**: `.agents/skills/cross-layer-sync/SKILL.md` for detailed checklist

**Exit criteria**: All impacted layers identified and synchronized.

### Sub-Phase 3.3: Documentation Synchronization

**Update documentation when**:
- User-facing behavior changed → Update feature docs
- API endpoints changed → Update skill-sublinkpro/reference/api.md
- Configuration options changed → Update docs/configuration.md
- Deployment process changed → Update docs/installation.md

**Bilingual requirement**:
- Update both `.md` (English) and `.zh-CN.md` (Chinese)
- Keep language switch links working
- Verify all links and code examples

**See**: `.agents/skills/doc-sync-check/SKILL.md` for detailed checklist

**Exit criteria**: All relevant docs updated in both languages.

### Sub-Phase 3.4: Test Execution

**Add tests when**:
- Changed key business logic
- Changed API contracts
- Changed permission checks
- Changed configuration semantics
- Changed database migrations
- Changed protocol parsing
- Fixed a bug (add regression test)

**Test requirements**:
```bash
# Run relevant tests
go test ./services/...
go test ./api/...

# Run full suite (if time permits)
go test ./...
```

**Exit criteria**: Tests added/updated for key changes, all tests pass.

### Sub-Phase 3.5: Change Summary

**Prepare commit message**:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type**: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

**Document in summary**:
- What changed (high-level)
- Why (problem being solved)
- How validated (which commands were run)
- Which layers were synchronized
- Which documentation was updated
- Any breaking changes

**Example**:
```
feat(airports): support batch subscription updates

Added batch update dialog with progress tracking.
Users can select multiple airports and update them in parallel.

Validation:
- Backend: gofmt ✅, golangci-lint ✅, go test ./... ✅
- Frontend: yarn lint ✅, yarn build ✅

Cross-layer sync:
- Backend: api/clients.go, services/subscription_service.go
- Frontend: webs/src/api/airports.js, AirportBatchUpdateDialog.jsx
- i18n: Added zh-CN and en-US translations
- Docs: Updated skill-sublinkpro/reference/api.md

Tests:
- Added TestBatchUpdateAirports
- Added service layer batch update tests
- All tests passing ✅

Closes #789
```

---

## Phase 4: Commit & Pull Request

### Before committing

**Review staging**:
```bash
git status
git diff --staged
```

**Verify**:
- [ ] Only intended files are staged
- [ ] No secrets (.env, credentials, tokens)
- [ ] No runtime data (db/, logs/, out/)
- [ ] No node_modules or vendor directories
- [ ] No large binary files

**Commit**:
```bash
git add <specific-files>
git commit -m "feat(scope): description"
```

### Creating pull request

**Target branch**: Always target `dev` (not `main`)

**PR description should include**:

```markdown
## Summary
Brief description of what this PR does

## Changes
- Backend: List backend changes
- Frontend: List frontend changes
- Documentation: List doc changes

## Validation
- [ ] Backend: gofmt ✅, golangci-lint ✅, go test ✅
- [ ] Frontend: yarn lint ✅, yarn build ✅
- [ ] Cross-layer sync verified
- [ ] Documentation updated (EN + ZH)
- [ ] Tests added/updated

## Testing
How to test this change:
1. Step-by-step instructions
2. Expected behavior

## Related Issues
Closes #123
Fixes #456

## Breaking Changes
List any breaking changes (or "None")

## Screenshots (if UI change)
Before/After screenshots
```

---

## Phase 5: Review & Merge

### Automated PR checks

`.github/workflows/pr-checks.yml` runs automatically when PR is opened:

- Backend: `golangci-lint run`, `go test ./...`
- Frontend: `yarn run lint`, `yarn run build`

**If checks fail**:
1. Fix the issues
2. Push the fix commits
3. Comment `/recheck` on the PR to re-trigger checks

### Code review

**Reviewers will check**:
- Code quality and correctness
- Test coverage
- Documentation completeness
- Cross-layer synchronization
- Bilingual consistency
- Architectural compliance

**Address feedback promptly**:
- Make requested changes
- Respond to comments
- Re-request review after addressing feedback

### Merge

After approval and passing checks:
- Merge to `dev` branch
- Delete feature branch (if desired)
- Monitor for any issues

---

## Quick Reference: Common Workflows

### Adding a new API endpoint

1. **Implementation**:
   - Backend: Add route in `routers/`, handler in `api/`, logic in `services/`
   - Frontend: Add API function in `webs/src/api/`, use in views
   - i18n: Add translations for UI text

2. **Post-dev workflow**:
   - Phase 1: Run gofmt, golangci-lint, go test, yarn lint, yarn build
   - Phase 2: Verify frontend ↔ backend sync
   - Phase 3: Update `skill-sublinkpro/reference/api.md`
   - Phase 4: Add API handler tests
   - Phase 5: Prepare commit message

3. **Commit**: `feat(api): add new endpoint for X`

### Fixing a bug

1. **Implementation**:
   - Write a test that reproduces the bug (should fail)
   - Fix the bug
   - Verify the test now passes

2. **Post-dev workflow**:
   - Phase 1: Run validation commands
   - Phase 2: Check if fix affects multiple layers
   - Phase 3: Update docs if behavior was unclear
   - Phase 4: Verify regression test passes
   - Phase 5: Prepare commit message

3. **Commit**: `fix(scope): description of bug fix`

### Updating documentation

1. **Implementation**:
   - Update English `.md` file
   - Update Chinese `.zh-CN.md` file
   - Verify links and examples

2. **Post-dev workflow** (simplified for docs-only):
   - Phase 1: Skip (no code changes)
   - Phase 2: Skip (docs-only)
   - Phase 3: Verify bilingual consistency, test all links
   - Phase 4: Skip (no tests needed)
   - Phase 5: Prepare commit message

3. **Commit**: `docs(scope): description of doc update`

### Adding a new protocol

1. **Implementation**:
   - Create `node/protocol/myprotocol.go`
   - Implement Protocol interface
   - Implement ProxyCapable, SurgeCapable (if applicable)
   - Register in `init()`
   - Add unit tests

2. **Post-dev workflow**:
   - Phase 1: Run gofmt, golangci-lint, protocol tests
   - Phase 2: Usually no frontend changes (dynamically registered)
   - Phase 3: Update docs if public-facing feature
   - Phase 4: Verify protocol tests pass
   - Phase 5: Prepare commit message

3. **Commit**: `feat(protocol): add support for MyProtocol`

---

## For AI Agents: Automation Instructions

### Automatic triggering

After completing code implementation:

```
IF code_implementation_complete:
    AUTOMATICALLY invoke .agents/skills/post-dev-workflow/SKILL.md
    
    RUN Phase 3.1: Code Validation
    IF fails:
        FIX issues
        RETRY validation
    
    IF multi_layer_change:
        RUN Phase 3.2: Cross-Layer Sync
    
    IF behavior_api_config_changed:
        RUN Phase 3.3: Documentation Sync
    
    IF key_logic_changed:
        RUN Phase 3.4: Test Execution
    
    RUN Phase 3.5: Change Summary
    
    REPORT results to user
    
    IF all_phases_pass:
        DECLARE "Work is ready to commit"
    ELSE:
        DECLARE "Issues found, fixing now..."
        FIX issues
        RETRY workflow
```

### Never skip workflow

❌ **DO NOT**:
- Declare work "complete" or "done" without running post-dev workflow
- Skip validation because "it's a small change"
- Update only English docs
- Change backend without checking frontend impact
- Skip tests for "simple" changes

✅ **DO**:
- Always run post-dev workflow after code changes
- Fix all validation errors before proceeding
- Update all impacted layers together
- Maintain bilingual documentation
- Add tests for key logic changes

---

## Related Documentation

- **AGENTS.md**: Architectural principles and mandatory workflow
- **CONTRIBUTING.md**: General contribution guidelines
- **docs/development.md**: Development setup and tech stack
- **.agents/README.md**: Skills system overview
- **.agents/skills/post-dev-workflow/SKILL.md**: Detailed post-dev workflow
- **.agents/skills/pre-commit-check/SKILL.md**: Pre-commit checklist
- **.agents/skills/cross-layer-sync/SKILL.md**: Cross-layer sync guide
- **.agents/skills/doc-sync-check/SKILL.md**: Documentation sync guide
- **.agents/skills/theme-check/SKILL.md**: Theme adaptation checklist

---

## Benefits of This Workflow

1. **Consistency**: Every change goes through the same validation process
2. **Quality**: Catch issues early before they reach production
3. **Completeness**: No half-finished work or missing documentation
4. **Maintainability**: Future developers understand what changed and why
5. **Collaboration**: Clear expectations for both AI and human contributors
6. **Efficiency**: Automated checks reduce manual review burden

---

## Conclusion

This workflow ensures that every code change in SublinkPro meets quality standards, respects architectural boundaries, and maintains complete documentation. By following this workflow consistently, we maintain a high-quality codebase that is easy to understand, maintain, and contribute to.

**Remember**: The post-development workflow (Phase 3) is **MANDATORY** and **AUTOMATIC** after completing any code change. It is not an optional step or a suggestion—it is part of what "done" means.
