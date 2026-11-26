# AGENTS.md

SublinkPro is a single Go application with a React/Vite frontend. Start with the [coding agent guide](docs/agents/README.md) for task navigation and the [developer guide](docs/development/README.md) for setup.

## Required workflow for agents

1. Run the [post-development workflow](.agents/skills/post-dev-workflow/SKILL.md) after every code change. If skill discovery is unavailable, open the `SKILL.md` directly. The tracked `.claude/skills` symlink points to `.agents/skills`.
2. Run the [pre-commit checklist](.agents/skills/pre-commit-check/SKILL.md) before declaring work complete, running `git add` or `git commit`, or creating or updating a PR. This applies to code, documentation, and configuration. Small changes and requests to skip checks still require validation.
3. Run all applicable checks, fix failures, and rerun the affected checks. A failure or an unresolved required check blocks staging, committing, opening or updating a PR, and declaring the change complete. Report what ran and what remains blocked.
4. When preparing a Git submission, stage only intended, validated changes. Inspect `git status --short` and `git diff --cached`, then present the result for user verification. Never auto-commit. Prepare a [semantic commit message](.agents/skills/post-dev-workflow/references/commit-message-guide.md) with a subject of at most 72 characters, the reason for the change, and validation results. Link relevant issues when available. See the [contribution guide](CONTRIBUTING.md).

PRs target `dev`. The `main` branch is the release line and publishes previews. Stable releases use `v*` tags.

## Temporary files and staged changes

- Use the ignored project-local `.agents/.tmp/` directory for **all** agent scratch files, temporary tests, debug scripts, logs, screenshots, generated reports, and intermediate outputs. Create it if needed and direct tools' scratch output there. Keep agent work files out of system temp directories and other repository paths.
- Keep shipping source files, configuration, assets, and documentation deliverables in their normal locations. Remove disposable files you created when done.
- Exclude credentials, secrets, runtime data under `db/`, `logs/`, `cache/` or `out/`, agent sessions, temporary summaries, and unrelated large binaries from staged changes. Preserve the user's staged work.

## Project boundaries

The Go entry point is [main.go](main.go). The frontend entry point is [webs/src/index.jsx](webs/src/index.jsx). The production build compiles the frontend first and embeds the output in the Go binary.

| Area | Location |
| --- | --- |
| Routes, auth and MFA | `routers/`, `middlewares/`, [api/auth.go](api/auth.go), [api/auth_mfa.go](api/auth_mfa.go) |
| HTTP handlers and request types | `api/`, `dto/` |
| Business logic, outbound proxying, and background work | `services/`, especially `services/scheduler/`, `services/socks5/`, and `services/mihomo/` |
| Persistence, settings, and migrations | `config/`, `models/`, `database/`, [models/db_migrate.go](models/db_migrate.go) |
| Subscriptions and protocols | [api/clients.go](api/clients.go), `node/`, `node/protocol/` |
| Frontend API, routes, and pages | `webs/src/api/`, `webs/src/routes/`, `webs/src/views/` |
| Installable SublinkPro Skill | [skills/sublinkpro](skills/sublinkpro/) |

- Keep route registration in `routers/`, business logic in `services/`, persistence in backend models, and scheduled tasks in `services/scheduler/`. Prefer existing patterns and abstractions in the owning layer.
- When adding or changing an API that modifies core data, deployment settings, or service state, exposes sensitive data, or uses server network access to reach a caller-controlled destination, register `middlewares.DemoModeRestrict` after authentication and before the handler. Verify both demo and normal mode against the [demo mode restrictions](docs/development/development-workflow.md#demo-mode-restrictions).
- Keep feature-local components under `webs/src/views/<feature>/component(s)/`. Put intentionally shared components in `webs/src/components/`. See the [theme rules](docs/development/frontend-theme-guidelines.md) for light and dark surfaces, the [reference components](docs/development/frontend-theme-guidelines.md#reference-patterns) for dialogs and mobile layouts, and the [i18n guide](docs/development/internationalization.md) for `zh-CN`, `en-US`, and `i18nKey`/`i18nParams`.
- Check `services/mihomo/` for adapter dialing, delay and speed tests, DNS and DoH, Host mappings, chain proxies, and proxied downloads. Nearby entry points include the [speed-test scheduler](services/scheduler/speedtest_task.go), [shared proxy client](utils/proxy_client.go), [Host storage](models/host.go), [subscription fetching](node/sub.go), and [subscription usage](node/usage.go).

## Run and validate

Use [go.mod](go.mod) and [webs/package.json](webs/package.json) for toolchains and scripts. Check the [PR workflow](.github/workflows/pr-checks.yml) for CI validation and the [release workflow](.github/workflows/build-release.yml) and [Dockerfile](Dockerfile) for builds, output paths, and packaging. Use the [development setup](docs/development/getting-started.md) for local setup and the [build guide](docs/development/build-and-deployment.md) for asset embedding.

| Change | Required checks |
| --- | --- |
| Go | Use `gofmt -w` on changed Go files and confirm `gofmt -l` lists none. Run `go build ./...`, `golangci-lint run`, and `go test ./...`. If an environment limitation prevents the full suite from running, run at least the relevant packages and report the limitation. Investigate failures in any tests that ran. |
| Frontend | Run `yarn run lint` in `webs/` for every frontend change. Also run `yarn run build` for changes affecting build configuration, routing, assets, or production integration. Resolve missing imports and broken dependencies. |
| Documentation | Run `python3 .github/scripts/check_docs.py` from the repository root. Verify both language versions, links, and affected examples against the implementation. |
| Key logic or defects | Add tests for new behavior, update tests for changed behavior, and add regression coverage for bug fixes. |
| Multi-layer | Verify the backend, frontend, and docs together using the [cross-layer checklist](.agents/skills/cross-layer-sync/SKILL.md). |

Open the relevant specialist skill for [theme changes](.agents/skills/theme-check/SKILL.md), [auth or sensitive data](.agents/skills/security-review/SKILL.md), and [performance-sensitive changes](.agents/skills/performance-check/SKILL.md).

## Keep documentation and contracts in sync

- Frontend requests go through `webs/src/api/request.js`. `SUBLINK_WEB_BASE_PATH` changes only the Web UI path. Keep `/api/*` and `/c/*` independent of it.
- In the same change, update affected frontend clients and views for backend API changes, verify frontend contract changes against the backend, and update the [Skill API reference](skills/sublinkpro/reference/api.md) for `/api/v1/*` or `/c/*` changes.
- Synchronize configuration and deployment changes across implementation, examples, both documentation languages, and the [Skill deployment guide](skills/sublinkpro/reference/deploy.md).
- Update affected English and Simplified Chinese guides together. For new or moved guides, update indexes, incoming links, and the [Skill documentation map](skills/sublinkpro/reference/docs.md). Use the [documentation checklist](.agents/skills/doc-sync-check/SKILL.md).

## Configuration and operations

- Config precedence is flags, environment, `db/config.yaml`, database settings, then defaults. Common defaults are port `8000`, `sqlite://./db/sublink.db`, `./db` for data, and `./logs` for logs. Docker uses `/app/db`, `/app/template`, and `/app/logs`. See the [configuration reference](docs/user/configuration.md).
- In deployment guidance, remind users to change the default admin password. SQLite-to-MySQL/PostgreSQL migration requires a manual restart. SublinkPro databases are incompatible with upstream project databases. See the [security guide](docs/user/security-guidelines.md).
