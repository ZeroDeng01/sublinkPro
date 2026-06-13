# Contributing to SublinkPro

Thank you for your interest in contributing to SublinkPro! This guide will help you get started.

## Quick Start

1. **Fork and clone** the repository
2. **Set up your environment** - see [Development Guide](docs/development.md)
3. **Create a branch** from `dev` for your changes
4. **Make your changes** following our guidelines below
5. **Test your changes** - see [Testing](#testing)
6. **Submit a pull request** to the `dev` branch

## Branch and Commit Conventions

### Branches

- `main` - Stable release branch
- `dev` - Development branch (target for PRs)
- `feature/*` - Feature branches
- `fix/*` - Bug fix branches

### Commit Messages

Use semantic commit prefixes:

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Test additions or updates
- `chore:` - Build process, dependency updates, or tooling changes

Examples:
```
feat(airports): add batch subscription update support
fix(auth): enhance SSE authentication error handling
docs(i18n): update internationalization requirements
```

## Cross-Layer Synchronization

**Critical rule**: When a change affects multiple layers of the application, all impacted layers must be updated together in the same PR.

### When Backend Changes:

If your backend change affects APIs, fields, permissions, response structures, routes, or task results:

- ✅ Update frontend request layer (`webs/src/api/`)
- ✅ Update frontend views (`webs/src/views/`)
- ✅ Update documentation
- ✅ Update skill API reference (`skill-sublinkpro/reference/api.md`)

### When Frontend Changes:

If your frontend change affects API dependencies, auth behavior, field semantics, page flows, or routing:

- ✅ Verify backend implementation matches
- ✅ Update documentation if user-facing behavior changes

### When Documentation Changes:

- ✅ Update both English and Chinese (`*.zh-CN.md`) versions together
- ✅ Keep language switches and relative links consistent

### When Configuration Changes:

- ✅ Update code, documentation, and example files together
- ✅ Update `skill-sublinkpro/reference/deploy.md` if deployment-related

## Testing

### Frontend

Run from `webs/` directory:

```bash
yarn run lint          # Required for all frontend changes
yarn run build         # Required when affecting build, routing, or assets
yarn run lint:fix      # Auto-fix linting issues
yarn run prettier      # Format code
```

### Backend

Run from repository root:

```bash
gofmt -w <changed-files>    # Format changed Go files
golangci-lint run           # Run linter
go test ./...               # Run all tests
```

### Theme Changes

When modifying frontend UI colors, surfaces, or visual elements:

- ✅ Check both light and dark modes
- ✅ Check desktop and mobile views
- ✅ Check hover, active, disabled, and focus states
- ✅ See [Frontend Theme Guidelines](docs/frontend-theme-guidelines.md) for comprehensive rules

## Internationalization

This project maintains bilingual support (Chinese and English):

- ✅ New user-facing frontend text requires translations in both `zh-CN` and `en-US`
- ✅ Use `useTranslation()` / `Trans` components in React
- ✅ Backend responses for Web UI should include `i18nKey` + `i18nParams`
- ✅ Update both English and Chinese documentation versions

See [Internationalization Guide](docs/internationalization.md) for complete requirements.

## Code Quality Standards

### Commenting

- Add comments for complex business logic, cross-layer contracts, and non-obvious algorithms
- Comments should explain intent and constraints, not restate code
- Update comments when changing nearby code
- See [Development Guide](docs/development.md) for detailed standards

### Testing

- Add or update tests when changing key business logic, APIs, or protocols
- Test names should describe scenario and expected outcome
- Tests must be isolated and not depend on execution order or external state
- See [Development Guide](docs/development.md) for detailed standards

## Pull Request Process

1. **Ensure your PR targets `dev`** (not `main`)
2. **Fill out the PR template** completely
3. **Link related issues** if applicable
4. **Verify all checks pass**:
   - Frontend: `yarn run lint` and `yarn run build`
   - Backend: `golangci-lint run` and `go test ./...`
5. **Request review** from maintainers
6. **Address feedback** promptly

### PR Checks

Automated checks run when PRs are opened, reopened, or marked ready for review:

- Backend: `golangci-lint` and `go test ./...`
- Frontend: `yarn run lint` and `yarn run build`

To manually re-trigger checks, comment `/recheck` on your PR.

## Documentation Updates

When your change affects behavior visible to users or developers:

- ✅ Update relevant files in `docs/`
- ✅ Update `README.md` if appropriate
- ✅ Update feature documentation in `docs/features/` if applicable
- ✅ Keep bilingual consistency (English and `*.zh-CN.md`)

## Architecture Deep Dive

For architectural guidance and detailed contribution requirements, see [AGENTS.md](AGENTS.md), which covers:

- Project structure and boundaries
- Tech stack and mihomo integration
- Configuration rules
- Frontend-backend contracts
- High-value areas to inspect before changes

## Code of Conduct

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.

## Getting Help

- 📖 Read the [Development Guide](docs/development.md)
- 🐛 Report bugs via [GitHub Issues](https://github.com/zerodeng/sublink-pro/issues)
- 💬 Join discussions on [GitHub Discussions](https://github.com/zerodeng/sublink-pro/discussions)
- 📱 Chat on Telegram: [SublinkPro Community](https://t.me/sublinkpro)

## License

By contributing to SublinkPro, you agree that your contributions will be licensed under the same license as the project.
