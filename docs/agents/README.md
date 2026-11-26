English | [简体中文](README.zh-CN.md)

# Coding Agent Guide

[Documentation home](../README.md) · [Repository instructions](../../AGENTS.md)

Start with AGENTS.md for commands, boundaries, and the validation policy. Use the paths below to find the relevant implementation before editing.

Before declaring work complete, staging, committing, or opening or updating a PR, read the pre-commit skill's [required checks](../../.agents/skills/pre-commit-check/SKILL.md#required-checks) and [hard stops](../../.agents/skills/pre-commit-check/SKILL.md#hard-stops).

| Task | Start here | Read when needed |
| --- | --- | --- |
| Change an API or permission | Route registration in routers/, handlers in api/, request clients in webs/src/api/ | [Cross-layer sync](../../.agents/skills/cross-layer-sync/SKILL.md) and [security review](../../.agents/skills/security-review/SKILL.md) |
| Change UI components | Owning page in webs/src/views/ and reusable components in webs/src/components/ | [Theme rules](../development/frontend-theme-guidelines.md) and [translations](../development/internationalization.md) |
| Change scheduling, routing, or proxy traffic | services/scheduler/, services/socks5/, services/mihomo/, utils/proxy_client.go | [Development recipes](../development/practical-recipes.md) |
| Add or extend a protocol | Implementations in node/protocol/ and normalization in node/ | [Protocol extension guide](../development/getting-started.md#-protocol-extension-guide) |
| Move or edit documentation | [User guide](../user/README.md) or [developer guide](../development/README.md) | [Documentation sync](../../.agents/skills/doc-sync-check/SKILL.md) |
| Validate a change | [Development workflow](../development/development-workflow.md) | [Post-development workflow](../../.agents/skills/post-dev-workflow/SKILL.md) and [pre-commit check](../../.agents/skills/pre-commit-check/SKILL.md) |

Repository skills live in .agents/skills/. Open the relevant SKILL.md directly if the coding agent does not discover it. Claude Code discovers these same files through the tracked .claude/skills symlink.

The [SublinkPro Skill](../../skills/sublinkpro/README.md) is the installable tool for users who want an AI assistant to operate their instance. Its [API reference](../../skills/sublinkpro/reference/api.md) is bundled with the download.
