# Pre-Commit Check Skill

Validation checklist to run before committing code changes.

## When to use this skill

Use this skill:
- Before every commit
- Before creating or updating a pull request
- After completing any code changes
- When PR checks fail and you need to know what to fix

## Prerequisites

- Changes are complete and ready to commit
- You have tested the changes locally
- You understand what layers your changes affect

## Quick Check Matrix

| Changed Files | Must Run |
|---|---|
| Any Go files | `gofmt`, `golangci-lint`, relevant `go test` |
| Any `webs/` frontend files | `yarn run lint` |
| Routing, assets, build config | `yarn run build` |
| i18n, theme, major UI | `yarn run lint` + `yarn run build` |
| Documentation only | Manual link/consistency check |

## Backend Validation Checklist

### When Go files changed:

#### Format check
```bash
# From repo root
gofmt -w <changed-go-files>
```
- [ ] All changed Go files formatted with `gofmt`
- [ ] No format changes remain after running `gofmt -w`

#### Linter
```bash
# From repo root
golangci-lint run
```
- [ ] `golangci-lint run` passes with no errors
- [ ] Fixed any warnings if applicable

#### Tests
```bash
# From repo root
go test ./...
```
- [ ] Relevant package tests pass
- [ ] Full `go test ./...` passes if time permits
- [ ] Added/updated tests for changed business logic

### When to add/update Go tests:
- [ ] Changed key business logic
- [ ] Changed API contracts
- [ ] Changed permission checks
- [ ] Changed configuration semantics
- [ ] Changed migrations
- [ ] Changed scheduled jobs
- [ ] Changed mihomo integrations
- [ ] Changed protocol parsing
- [ ] Changed data transformations

## Frontend Validation Checklist

### Always run for frontend changes:

#### Lint
```bash
# From webs/ directory
cd webs
yarn run lint
```
- [ ] `yarn run lint` passes with no errors
- [ ] No ESLint errors remain

#### Auto-fix (optional)
```bash
# From webs/ directory
yarn run lint:fix
yarn run prettier
```

### Run build when these are affected:

#### Build validation
```bash
# From webs/ directory
yarn run build
```

Run build when changes affect:
- [ ] Routing configuration
- [ ] Asset paths or imports
- [ ] Base path behavior (`SUBLINK_WEB_BASE_PATH`)
- [ ] Production integration
- [ ] Build configuration (vite.config.js)
- [ ] Static file structure

## i18n Validation Checklist

### When i18n changes:

#### Frontend i18n
- [ ] Added translations for both `zh-CN` and `en-US`
- [ ] Translation keys use stable semantic names (not current wording)
- [ ] Used `useTranslation()` or `Trans` components
- [ ] No hardcoded Chinese or English text in JSX
- [ ] Ran `yarn run lint`
- [ ] Ran `yarn run build` (if routing/assets affected)

#### Backend i18n
- [ ] Added `i18nKey` + `i18nParams` for Web UI display
- [ ] Kept `msg` for backward compatibility
- [ ] Ran `gofmt -w <changed-files>`
- [ ] Ran `golangci-lint run`
- [ ] Ran relevant `go test`

#### Documentation i18n
- [ ] Updated both `.md` and `.zh-CN.md` versions
- [ ] Language switch links work
- [ ] Relative links consistent
- [ ] Manually verified markdown formatting

## Theme Changes Validation Checklist

### When theme/UI changes:

#### Required checks
- [ ] Tested in both light mode and dark mode
- [ ] Tested on desktop and mobile viewports
- [ ] Verified hover, active, disabled states
- [ ] Checked dialogs, drawers, popovers
- [ ] Ran `yarn run lint`
- [ ] Ran `yarn run build`

#### Use theme-check skill
- [ ] If major theme work, run `.agents/skills/theme-check/SKILL.md`

## Cross-Layer Changes Validation Checklist

### When changes span multiple layers:

#### Required checks
- [ ] Verified all impacted layers are synchronized
- [ ] Updated frontend if backend API changed
- [ ] Updated backend if frontend contract changed
- [ ] Updated documentation if behavior changed
- [ ] Updated `skill-sublinkpro/reference/api.md` if `/api/v1/*` or `/c/*` changed

#### Use cross-layer-sync skill
- [ ] If complex cross-layer work, run `.agents/skills/cross-layer-sync/SKILL.md`

## Documentation-Only Changes Validation Checklist

### When only docs changed:

#### Manual checks
- [ ] All markdown links resolve correctly
- [ ] Both English and Chinese versions updated (if applicable)
- [ ] Language switch links work
- [ ] Code examples are accurate
- [ ] Command examples match actual repo commands
- [ ] No references to non-existent files or features

#### No build needed
- ✅ No need to run `yarn run lint` or `yarn run build` for pure documentation changes
- ✅ No need to run Go validation for pure documentation changes

## Git Commit Checklist

### Before committing:

#### Commit scope
- [ ] Staging only related changes (avoid `git add .`)
- [ ] Not committing sensitive files (`.env`, credentials, etc.)
- [ ] Not committing large binary files unintentionally
- [ ] Not committing `db/`, `logs/`, `cache/`, `out/` runtime data

#### Commit message
- [ ] Used semantic prefix (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`)
- [ ] Message describes what changed and why
- [ ] Message is under 72 characters for subject line (if possible)

### Examples:
```bash
feat(airports): support batch subscription updates
fix(auth): enhance SSE authentication error handling
docs(i18n): update internationalization requirements
refactor(theme): extract theme adaptation rules to separate doc
test(api): add handler tests for subscription endpoints
chore(deps): update frontend dependencies
```

## PR Validation Checklist

### Before creating/updating PR:

#### Validation
- [ ] All pre-commit checks passed
- [ ] Tested changes end-to-end locally
- [ ] Verified no console errors
- [ ] Branch is up to date with target branch (`dev`)

#### Documentation
- [ ] PR description explains what and why
- [ ] Referenced related issues (if any)
- [ ] Listed what was tested
- [ ] Noted any breaking changes
- [ ] Documented cross-layer sync (if applicable)

#### Ready for review
- [ ] Target branch is `dev` (not `main`)
- [ ] No WIP/debug commits included
- [ ] Commit history is clean and logical

## Common Issues and Fixes

### Frontend lint fails
```bash
cd webs
yarn run lint:fix
yarn run prettier
```

### Go format issues
```bash
gofmt -w $(find . -name "*.go" | grep -v vendor)
```

### golangci-lint errors
- Read the error messages carefully
- Fix issues one by one
- Common: unused variables, error handling, imports

### Build fails
- Check for import errors
- Verify file paths are correct
- Check vite.config.js if route-related

### Tests fail
- Read test output
- Fix the actual issue, don't skip tests
- Add regression test if fixing a bug

## Automated PR Checks

The repository runs these checks automatically when PR is opened/updated:

- Backend: `golangci-lint run`, `go test ./...`
- Frontend: `yarn run lint`, `yarn run build`

You can manually re-trigger checks by commenting `/recheck` on the PR.

## Exit Criteria

✅ Can commit/push when:
- All relevant validation commands pass
- No linter errors or test failures
- Documentation updated if behavior changed
- Git staging includes only intended changes

❌ Cannot commit/push when:
- Validation commands fail
- Tests are broken
- Lint errors remain
- Sensitive files are staged
- Runtime data directories are staged
