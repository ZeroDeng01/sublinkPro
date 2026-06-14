# Commit Message Guide

Complete reference for writing high-quality commit messages in Phase 5 of the post-development workflow.

---

## Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

---

## Type

Choose one that best describes the change:

| Type | When to Use | Examples |
|---|---|---|
| `feat:` | New feature for users | feat(airports): add batch update dialog |
| `fix:` | Bug fix | fix(auth): handle SSE authentication timeout |
| `docs:` | Documentation only | docs(api): update endpoint reference |
| `refactor:` | Code restructuring, no behavior change | refactor(services): extract common validation logic |
| `test:` | Test additions or updates | test(scheduler): add cron expression tests |
| `chore:` | Build, deps, tooling, configs | chore(deps): upgrade sass to 1.101.0 |
| `style:` | Code style/format only (not UI) | style(api): run gofmt on all files |
| `perf:` | Performance improvement | perf(query): optimize database index usage |

---

## Scope

The component or area affected by the change:

### Common Scopes

| Scope | Area |
|---|---|
| `airports` | Airport/subscription management |
| `auth` | Authentication, MFA |
| `nodes` | Node management, preview |
| `chain` | Chain proxy |
| `unlock` | Unlock/region detection |
| `scheduler` | Scheduled tasks, cron |
| `theme` | UI theme, colors, surfaces |
| `i18n` | Internationalization |
| `config` | Configuration management |
| `api` | API endpoints, handlers |
| `ui` | General UI components |
| `db` | Database, migrations |
| `deps` | Dependencies |
| `docker` | Docker setup, deployment |
| `ci` | CI/CD workflows |

### Multiple Scopes

When change affects multiple areas, choose the primary scope or use format:

```
feat(airports,nodes): add batch operations
```

Or use broader scope:

```
feat(ui): add batch operations for airports and nodes
```

---

## Subject

**Rules**:
- Concise summary of the change
- Imperative mood ("add feature", not "added feature")
- No period at the end
- Lowercase after colon
- Aim for ≤72 characters

**Good**:
```
feat(airports): add batch subscription update
fix(auth): handle expired tokens correctly
docs(configuration): clarify database DSN format
```

**Bad**:
```
feat(airports): Added a new batch subscription update feature.  # Past tense, period, too long
fix(auth): fixed bugs  # Vague, not specific
docs: updates  # Missing scope, too vague
```

---

## Body

**When to include body**:
- Change needs context or explanation
- Non-obvious implementation choices
- Breaking changes requiring migration
- Multiple related changes in one commit

**Format**:
- Blank line after subject
- Wrap at 72 characters
- Explain **what** and **why**, not **how** (code shows how)
- Use bullet points for multiple items
- Reference related issues/PRs

**Example**:
```
feat(airports): add batch subscription update

Added batch update dialog with progress tracking.
Users can now select multiple airports and update subscriptions
in parallel, significantly reducing time for large updates.

Implementation:
- Backend: Added POST /api/v1/airports/batch-update endpoint
- Frontend: Added AirportBatchUpdateDialog with progress UI
- Added concurrent update with error collection

Closes #789
```

---

## Footer

**When to include footer**:
- Related issues/PRs
- Breaking changes
- Co-authors

**Format**:

### Issue References

```
Closes #123
Fixes #456
Relates to #789
```

### Breaking Changes

```
BREAKING CHANGE: Database schema changed, requires migration.

Run `docker-compose down && docker-compose up` to apply migration.
```

### Co-authors

```
Co-authored-by: Name <email@example.com>
```

---

## Complete Examples

### Simple Feature

```
feat(theme): add surface elevation system

Implemented Material Design surface elevation for light/dark modes.
All dialogs and cards now properly layer according to theme guidelines.
```

### Bug Fix with Context

```
fix(auth): enhance SSE authentication error handling

SSE authentication failures now return proper i18n error messages
instead of generic "connection failed" errors. Added retry logic
for transient network errors.

Changes:
- Added i18nKey support in SSE error responses
- Added exponential backoff retry (max 3 attempts)
- Improved error logging for debugging

Fixes #456
```

### Multi-Layer Change

```
feat(config): add database connection pooling

Added support for configuring database connection pool size.

Backend:
- Added SUBLINK_DB_MAX_CONNECTIONS config option
- Default: 25 connections (SQLite), 100 (MySQL/PostgreSQL)
- Applied to GORM connection pool configuration

Frontend:
- No changes (internal optimization)

Documentation:
- Updated docs/configuration.md + .zh-CN.md
- Updated skill-sublinkpro/reference/deploy.md
- Added example in docker-compose.yml

Tests:
- Added connection pool tests in models/db_test.go
- Verified pool limits under load

Closes #234
```

### Refactoring

```
refactor(services): extract common validation logic

Extracted repeated validation patterns into shared validator functions.
No behavior changes, purely structural improvement.

Changes:
- Created services/validators/ package
- Moved email, URL, cron validation to validators
- Updated all callers to use shared functions
- Added validator tests

All existing tests pass without modification.
```

### Documentation Only

```
docs(i18n): clarify backend i18n usage

Updated internationalization guide with backend i18nKey usage patterns.
Added examples for:
- Error messages with i18nKey + i18nParams
- Validation errors
- Dynamic message parameters

No code changes.
```

### Breaking Change

```
feat(config)!: change default database path

Changed default database path from ./sublink.db to ./db/sublink.db
for consistency with other data directories.

BREAKING CHANGE: Existing installations need to migrate database location.

Migration steps:
1. Stop SublinkPro
2. Move ./sublink.db to ./db/sublink.db
3. Restart SublinkPro

Or set SUBLINK_DB_DSN explicitly to preserve old location.

Closes #567
```

---

## Validation Checklist

Before committing, verify commit message:

- [ ] Type is correct and from standard list
- [ ] Scope clearly identifies affected area
- [ ] Subject is concise, imperative, no period
- [ ] Body explains what and why (if needed)
- [ ] Body wraps at 72 characters
- [ ] Footer references related issues
- [ ] Breaking changes clearly documented
- [ ] Commit message is readable and informative

---

## When to Split Commits

**Split into multiple commits when**:
- Changes affect unrelated areas
- Change includes refactoring + new feature
- Large feature with logical stages
- Documentation and code can be separated

**Keep in single commit when**:
- Changes are tightly coupled
- All changes needed for feature to work
- Cross-layer synchronization (backend + frontend for same feature)

---

## Commit Message Anti-Patterns

❌ **Vague subjects**:
```
fix: fix bugs
chore: updates
feat: improvements
```

✅ **Specific subjects**:
```
fix(auth): handle expired JWT tokens
chore(deps): upgrade sass to 1.101.0
feat(airports): add batch subscription update
```

---

❌ **Past tense**:
```
feat(nodes): added preview dialog
fix(auth): fixed login issue
```

✅ **Imperative mood**:
```
feat(nodes): add preview dialog
fix(auth): fix login issue
```

---

❌ **Code-level details in subject**:
```
feat(airports): change line 45 in AirportDialog.jsx
fix(auth): update if statement in auth.go
```

✅ **User/feature-level description**:
```
feat(airports): add batch subscription update
fix(auth): handle concurrent login attempts
```

---

❌ **No context for non-obvious changes**:
```
refactor(services): change validation
```

✅ **Context provided**:
```
refactor(services): extract common validation logic

Extracted repeated validation patterns into shared validator functions.
No behavior changes, purely structural improvement.
```

---

❌ **Missing scope**:
```
feat: add batch update
fix: fix error
```

✅ **Clear scope**:
```
feat(airports): add batch subscription update
fix(auth): handle SSE authentication errors
```

---

## Special Cases

### Merge Commits

When merging branches (usually automated):
```
Merge pull request #123 from user/feature-branch

feat(airports): add batch subscription update
```

### Revert Commits

When reverting previous commits:
```
revert: feat(airports): add batch subscription update

This reverts commit abc123def456.

Reason: Found critical performance issue in batch update logic.
Will fix and re-introduce in next PR.
```

### Initial Commits

For first commit in repository:
```
chore: initial commit

SublinkPro - Subscription management platform
```

---

## Commit Message Template

Use this template for your commits:

```
<type>(<scope>): <subject>

## What changed
- 

## Why
- 

## How validated
- Backend: 
- Frontend: 
- Cross-layer sync: 
- Documentation: 
- Tests: 

## Related
Closes #
```

Save this template as `assets/commit-template.txt` and use:

```bash
git config commit.template .agents/skills/post-dev-workflow/assets/commit-template.txt
```

---

## Tools

### Commitlint (Optional)

Configure commitlint to enforce commit message format:

```javascript
// commitlint.config.js
module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'type-enum': [
      2,
      'always',
      ['feat', 'fix', 'docs', 'refactor', 'test', 'chore', 'style', 'perf']
    ],
    'subject-case': [2, 'always', 'lower-case'],
    'subject-full-stop': [2, 'never', '.'],
    'body-max-line-length': [2, 'always', 72]
  }
};
```

### Git Hooks

Use pre-commit hook to validate commit messages:

```bash
#!/bin/bash
# .git/hooks/commit-msg

commit_msg=$(cat "$1")

# Check format
if ! echo "$commit_msg" | grep -qE '^(feat|fix|docs|refactor|test|chore|style|perf)\(.+\): .+'; then
  echo "Error: Commit message does not match format"
  echo "Expected: <type>(<scope>): <subject>"
  exit 1
fi
```

---

## Related Skills

- `.agents/skills/pre-commit-check/SKILL.md` - Pre-commit validation
- `.agents/skills/cross-layer-sync/SKILL.md` - Cross-layer synchronization context
- `.agents/skills/doc-sync-check/SKILL.md` - Documentation context
