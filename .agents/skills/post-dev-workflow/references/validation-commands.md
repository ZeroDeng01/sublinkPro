# Validation Commands Reference

Complete reference for validation commands used in Phase 1 of the post-development workflow.

---

## Backend Validation

Run all commands from repository root.

### Format Check

```bash
# Format changed Go files
gofmt -w <changed-files>

# Format all Go files (if unsure which files changed)
gofmt -w .
```

**Expected result**: Files are formatted according to Go standards. Running `gofmt -w` again produces no changes.

### Lint

```bash
golangci-lint run
```

**Expected result**: Exit status 0, no errors reported.

**Common lint errors**:
- Unused variables/imports → Remove them
- Error handling issues → Add proper error checks
- Shadowed variables → Rename to avoid confusion
- Cyclomatic complexity → Refactor complex functions

### Test

```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./services/scheduler/...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific test
go test -run TestBatchUpdateAirports ./api/...
```

**Expected result**: All tests pass. No race conditions detected.

**Test flags**:
- `-v` - Verbose output
- `-short` - Skip long-running tests
- `-timeout 30s` - Set timeout (default 10m)
- `-count=1` - Disable test caching

---

## Frontend Validation

Run all commands from `webs/` directory.

### Lint

```bash
cd webs

# Run ESLint
yarn run lint

# Auto-fix lint issues
yarn run lint:fix

# Run Prettier
yarn run prettier
```

**Expected result**: Exit status 0, no errors reported.

**Common lint errors**:
- Missing key prop in lists → Add unique `key` prop
- Unused variables → Remove them
- Console logs left in code → Remove or guard with environment check
- Missing prop types → Add prop validation
- Accessibility issues → Add ARIA labels, alt text

### Build

```bash
cd webs

# Production build
yarn run build

# Build with source maps (for debugging)
yarn run build --sourcemap
```

**Expected result**: Build completes without errors. Output in `webs/dist/`.

**When build is required**:
- Routing changes (new pages, route changes)
- Asset changes (images, fonts, static files)
- Build config changes (vite.config.js, package.json)
- Production integration changes (embedded in Go binary)
- Import/export structure changes

**When build can be skipped**:
- Pure logic changes (no imports/exports changed)
- Component internal changes (no props/exports changed)
- Styling changes (CSS/MUI changes)

### Test

```bash
cd webs

# Run all tests
yarn test

# Run tests without watch mode
yarn test --run

# Run tests with coverage
yarn test --coverage

# Run specific test file
yarn test src/utils/colorUtils.test.js
```

**Expected result**: All tests pass. Reasonable coverage for changed code.

**Test patterns**:
- Unit tests for utilities: `src/utils/*.test.js`
- Component tests: `src/components/*.test.jsx`
- API tests: `src/api/*.test.js`

---

## Validation Checklist

### Backend

- [ ] `gofmt -w <files>` produces no further changes
- [ ] `golangci-lint run` exits with status 0
- [ ] `go test ./...` passes (or relevant subset)
- [ ] No new TODO/FIXME comments without tracking issues
- [ ] Error messages use `i18nKey` + `i18nParams` (not hardcoded strings)

### Frontend

- [ ] `yarn run lint` exits with status 0
- [ ] `yarn run build` succeeds (if applicable)
- [ ] `yarn test` passes (if tests exist)
- [ ] No console.log/console.error left in code
- [ ] No hardcoded strings (use i18next `t()` function)
- [ ] Accessibility: ARIA labels, alt text, keyboard navigation

---

## When Validation Fails

### Format Failures

**Backend**:
```bash
# Auto-fix format issues
gofmt -w .
```

**Frontend**:
```bash
cd webs
yarn run lint:fix
yarn run prettier
```

### Lint Failures

**Review each error**:
1. Is the lint rule correct? (Usually yes)
2. Can the code be fixed? (Usually yes)
3. Should the rule be adjusted? (Rare, needs justification)

**Do not disable lint rules without documenting why in the PR.**

### Build Failures

**Common causes**:
- Import path errors → Fix import statements
- Missing dependencies → Run `yarn install` or `go mod tidy`
- Type errors → Fix TypeScript/prop types
- Asset path errors → Fix static file references
- Config errors → Check vite.config.js, package.json

**Debugging**:
```bash
# Frontend: check import paths
cd webs
yarn run build --debug

# Backend: check module dependencies
go mod verify
go mod tidy
```

### Test Failures

**Do not skip failing tests.** Fix the actual issue:

1. **Is the test correct?**
   - Does it test the right behavior?
   - Are expectations accurate?

2. **Is the code correct?**
   - Does the code match the intended behavior?
   - Did the change break existing functionality?

3. **Is the test outdated?**
   - Did the change intentionally alter behavior?
   - Should the test be updated to match new behavior?

**Update tests only when behavior intentionally changes. Document why in the commit message.**

---

## Performance Validation

For performance-critical changes, also run:

```bash
# Backend: benchmark tests
go test -bench=. -benchmem ./...

# Frontend: build size analysis
cd webs
yarn run build --analyze
```

See `.agents/skills/performance-check/SKILL.md` for detailed performance validation.

---

## Security Validation

For security-critical changes (auth, permissions, sensitive data), also run:

See `.agents/skills/security-review/SKILL.md` for detailed security validation.

---

## Validation Before Commit

Minimum validation before committing:

```bash
# Backend changes
gofmt -w <changed-files>
golangci-lint run
go test ./...

# Frontend changes
cd webs
yarn run lint
yarn run build  # if applicable
yarn test       # if tests exist

# Both
git diff        # Review changes
git status      # Check staged files
```

**Do not commit if any validation fails.**
