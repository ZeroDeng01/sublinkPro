English | [简体中文](development-workflow.zh-CN.md)

[Guide index](README.md)

# Development Workflow

## 1. Locate the change

Start from the [development guide](getting-started.md) to find the entry point and normal callers. Keep routes in routers/, HTTP handlers in api/, business rules in services/, and persistence in models/. Update the frontend client and UI together with any changes to their API contract. Check the [configuration reference](../user/configuration.md) before altering defaults or config precedence.

## 2. Implement and exercise the behavior

Reuse related code and cover changes to business logic, permissions, migrations, protocols, or API behavior with tests. For UI changes, check both themes and desktop/mobile layouts, translate user-facing text, and follow the [theme rules](frontend-theme-guidelines.md) and [i18n guide](internationalization.md). Update the relevant English and Chinese docs in the same change.

### Demo mode restrictions

When adding or changing an API, review what an authenticated demo user could trigger, including side effects in background jobs. High-risk routes must register [`middlewares.DemoModeRestrict`](../../middlewares/demo_mode.go) after authentication and before the handler.

- Changes to credentials or core instance data, including backup management, database imports, and restores.
- Changes to instance-wide or deployment settings, such as the public URL or tunnel configuration, and starting or stopping services.
- Exposure of sensitive data or use of the server's network access to reach caller-controlled destinations, such as remote rule conversion and configurable downloads.

Evaluate what the handler actually does, including actions reachable through read endpoints. Test that an authenticated demo request is rejected before it writes data, sends an outbound request, or performs another restricted action. Verify the normal-mode workflow as well.

## 3. Validate locally

| Scope | Commands |
| --- | --- |
| Backend | Format changed Go files with gofmt, then run go build ./..., golangci-lint run, and go test ./... from the root. |
| Frontend | Run yarn run lint under webs/. Run yarn run build for changes affecting production output, routes, or assets. |
| Documentation | Run python3 .github/scripts/check_docs.py from the root, verify new commands, and follow representative links in both languages. |

When the change crosses layers, verify the normal user flow end to end. Coding agents also use the [post-development workflow](../../.agents/skills/post-dev-workflow/SKILL.md) and [pre-commit check](../../.agents/skills/pre-commit-check/SKILL.md).

## 4. Submit for review

Follow [CONTRIBUTING.md](../../CONTRIBUTING.md). Open PRs against dev, describe how the change was verified, and review the automation report. After pushing fixes, use /recheck as described in [PR automated checks](getting-started.md#6-pr-automated-checks).
