[English](development-workflow.md) | 简体中文

[返回指南目录](README.zh-CN.md)

# 开发工作流

## 1. 定位修改

从[开发指南](getting-started.zh-CN.md)查找入口和正常调用链。路由放在 routers/，HTTP 处理器放在 api/，业务规则放在 services/，持久化逻辑放在 models/。修改 API 契约时，同步修改相关的前端请求和界面。修改默认值或配置优先级前，核对[配置说明](../user/configuration.zh-CN.md)。

## 2. 实现并验证行为

复用相关代码，为业务逻辑、权限、迁移、协议或 API 行为变更补充测试。修改界面时检查亮暗主题和桌面、移动端布局，翻译用户可见文案，并遵守[主题规范](frontend-theme-guidelines.zh-CN.md)和[国际化指南](internationalization.zh-CN.md)。在同一改动中更新受影响的中英文文档。

### 演示模式限制

新增或修改 API 时，检查已登录的演示用户能触发哪些行为，包括后台任务的副作用。高风险路由必须在认证之后、处理器之前注册 [`middlewares.DemoModeRestrict`](../../middlewares/demo_mode.go)。

- 改动凭据或实例核心数据，包括管理备份、导入或恢复数据库。
- 改动实例级配置或部署环境设置，例如公网访问地址和隧道配置，以及启动或停止服务。
- 暴露敏感数据，或使用服务器的网络权限访问由调用方控制的目标，例如远程规则转换和可配置地址的下载。

按处理器的实际行为判断风险，同时检查读取接口可能触发的操作。测试应验证已登录的演示请求在写入数据、发起出站请求或执行其他受限操作前被拒绝，并覆盖普通模式的正常流程。

## 3. 本地验证

| 范围 | 命令 |
| --- | --- |
| 后端 | 用 gofmt 格式化修改过的 Go 文件，再在仓库根目录运行 go build ./...、golangci-lint run 和 go test ./...。 |
| 前端 | 在 webs/ 运行 yarn run lint。改动涉及生产产物、路由或静态资源时，再运行 yarn run build。 |
| 文档 | 在仓库根目录运行 python3 .github/scripts/check_docs.py，验证新命令并检查两种语言中的代表性链接。 |

跨层改动需要端到端验证正常用户流程。编码 Agent 还需执行[开发后验证](../../.agents/skills/post-dev-workflow/SKILL.md)和[提交前检查](../../.agents/skills/pre-commit-check/SKILL.md)。

## 4. 提交评审

遵循[贡献指南](../../CONTRIBUTING.zh-CN.md)。PR 以 dev 为目标分支，说明验证结果并查看自动检查报告。推送修复后，按[PR 自动检查](getting-started.zh-CN.md#6-pr-自动检查)使用 /recheck。
