[English](README.md) | 简体中文

# 编码 Agent 指南

[文档首页](../README.zh-CN.md) · [仓库指令](../../AGENTS.md)

先读 AGENTS.md，获取命令、代码边界和验证规则，再根据任务定位实际实现。

在宣布完成、暂存、提交，以及创建或更新 PR 之前，先读提交前技能里的[必做检查](../../.agents/skills/pre-commit-check/SKILL.md#required-checks)和[阻断条件](../../.agents/skills/pre-commit-check/SKILL.md#hard-stops)。

| 任务 | 从哪里开始 | 按需阅读 |
| --- | --- | --- |
| 修改 API 或权限 | routers/ 路由、api/ 处理器、webs/src/api/ 请求客户端 | [跨层同步](../../.agents/skills/cross-layer-sync/SKILL.md)和[安全审查](../../.agents/skills/security-review/SKILL.md) |
| 修改 UI 组件 | webs/src/views/ 下的所属页面和 webs/src/components/ 下的共享组件 | [主题规范](../development/frontend-theme-guidelines.zh-CN.md)和[国际化](../development/internationalization.zh-CN.md) |
| 修改调度、路由或代理流量 | services/scheduler/、services/socks5/、services/mihomo/、utils/proxy_client.go | [开发实践](../development/practical-recipes.zh-CN.md) |
| 新增或扩展协议 | node/protocol/ 中的实现，以及 node/ 中的标准化逻辑 | [协议扩展指南](../development/getting-started.zh-CN.md#-新增协议接入指南) |
| 移动或编辑文档 | [用户指南](../user/README.zh-CN.md)或[开发指南](../development/README.zh-CN.md) | [文档同步](../../.agents/skills/doc-sync-check/SKILL.md) |
| 验证改动 | [开发工作流](../development/development-workflow.zh-CN.md) | [开发后验证](../../.agents/skills/post-dev-workflow/SKILL.md)和[提交前检查](../../.agents/skills/pre-commit-check/SKILL.md) |

仓库技能位于 .agents/skills/。如果编码 Agent 未自动加载技能，可直接阅读对应的 SKILL.md。Claude Code 通过 Git 已跟踪的 .claude/skills 软链接发现同一套技能。

可安装的 [SublinkPro Skill](../../skills/sublinkpro/README.zh-CN.md) 用于让 AI 助手操作运行中的实例。下载包附带 [API 参考](../../skills/sublinkpro/reference/api.md)。
