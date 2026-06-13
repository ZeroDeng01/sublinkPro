# 贡献指南

感谢你对 SublinkPro 的关注!本指南将帮助你开始贡献。

## 快速开始

1. **Fork 并克隆**仓库
2. **设置开发环境** - 参考 [开发指南](docs/development.zh-CN.md)
3. **从 `dev` 分支创建**你的功能分支
4. **进行修改**并遵循下面的规范
5. **测试你的改动** - 参考 [测试](#测试)
6. **提交 PR** 到 `dev` 分支

## 分支与提交规范

### 分支

- `main` - 稳定发布分支
- `dev` - 开发分支(PR 目标分支)
- `feature/*` - 功能分支
- `fix/*` - 修复分支

### 提交信息

使用语义化提交前缀:

- `feat:` - 新功能
- `fix:` - 修复问题
- `docs:` - 文档变更
- `refactor:` - 代码重构
- `test:` - 测试相关
- `chore:` - 构建、依赖或工具变更

示例:
```
feat(airports): 支持批量拉取选中机场订阅
fix(auth): 增强 SSE 认证错误处理
docs(i18n): 更新国际化要求
```

## 跨层同步

**关键规则**:当改动影响应用的多个层次时,所有受影响的层必须在同一个 PR 中一起更新。

### 后端改动时:

如果后端改动影响 API、字段、权限、响应结构、路由或任务结果:

- ✅ 更新前端请求层(`webs/src/api/`)
- ✅ 更新前端页面(`webs/src/views/`)
- ✅ 更新文档
- ✅ 更新 skill API 参考(`skill-sublinkpro/reference/api.md`)

### 前端改动时:

如果前端改动影响 API 依赖、鉴权行为、字段语义、页面流程或路由:

- ✅ 验证后端实现是否匹配
- ✅ 如果用户可见行为变化,更新文档

### 文档改动时:

- ✅ 同时更新英文和中文(`*.zh-CN.md`)版本
- ✅ 保持语言切换和相对链接一致

### 配置改动时:

- ✅ 同时更新代码、文档和示例文件
- ✅ 如果涉及部署,更新 `skill-sublinkpro/reference/deploy.md`

## 测试

### 前端

在 `webs/` 目录执行:

```bash
yarn run lint          # 所有前端改动必须执行
yarn run build         # 影响构建、路由或资源时必须执行
yarn run lint:fix      # 自动修复 lint 问题
yarn run prettier      # 格式化代码
```

### 后端

在仓库根目录执行:

```bash
gofmt -w <改动的文件>     # 格式化改动的 Go 文件
golangci-lint run        # 运行 linter
go test ./...            # 运行所有测试
```

### 主题改动

当修改前端 UI 颜色、表面或视觉元素时:

- ✅ 检查浅色和深色模式
- ✅ 检查桌面端和移动端视图
- ✅ 检查 hover、active、disabled 和 focus 状态
- ✅ 参考 [前端主题适配规范](docs/frontend-theme-guidelines.zh-CN.md) 查看完整规则

## 国际化

本项目维护中英双语支持:

- ✅ 新增用户可见的前端文本需要同时提供 `zh-CN` 和 `en-US` 翻译
- ✅ React 中使用 `useTranslation()` / `Trans` 组件
- ✅ 后端为 Web UI 的响应应包含 `i18nKey` + `i18nParams`
- ✅ 同时更新英文和中文文档版本

参考 [国际化指南](docs/internationalization.zh-CN.md) 查看完整要求。

## 代码质量标准

### 注释规范

- 为复杂业务逻辑、跨层契约和非显而易见的算法添加注释
- 注释应解释意图和约束,而不是重复代码
- 修改代码时同步更新附近注释
- 参考 [开发指南](docs/development.zh-CN.md) 查看详细标准

### 测试规范

- 修改关键业务逻辑、API 或协议时添加或更新测试
- 测试名称应描述场景和期望结果
- 测试必须相互隔离,不依赖执行顺序或外部状态
- 参考 [开发指南](docs/development.zh-CN.md) 查看详细标准

## PR 流程

1. **确保 PR 目标分支是 `dev`**(不是 `main`)
2. **完整填写 PR 模板**
3. **关联相关 issue**(如果有)
4. **验证所有检查通过**:
   - 前端:`yarn run lint` 和 `yarn run build`
   - 后端:`golangci-lint run` 和 `go test ./...`
5. **请求 review**
6. **及时处理反馈**

### PR 检查

当 PR 打开、重新打开或标记为 ready for review 时会自动运行检查:

- 后端:`golangci-lint` 和 `go test ./...`
- 前端:`yarn run lint` 和 `yarn run build`

如需手动重新触发检查,在 PR 中评论 `/recheck`。

## 文档更新

当改动影响用户或开发者可见的行为时:

- ✅ 更新 `docs/` 中的相关文件
- ✅ 适当更新 `README.md`
- ✅ 更新 `docs/features/` 中的功能文档(如适用)
- ✅ 保持双语一致性(英文和 `*.zh-CN.md`)

## 架构深入

关于架构指导和详细贡献要求,参考 [AGENTS.md](AGENTS.md),其中涵盖:

- 项目结构与边界
- 技术栈与 mihomo 集成
- 配置规则
- 前后端契约
- 改动前应检查的关键区域

## 行为准则

贡献前请阅读我们的 [行为准则](CODE_OF_CONDUCT.zh-CN.md)。

## 获取帮助

- 📖 阅读 [开发指南](docs/development.zh-CN.md)
- 🐛 通过 [GitHub Issues](https://github.com/zerodeng/sublink-pro/issues) 报告问题
- 💬 在 [GitHub Discussions](https://github.com/zerodeng/sublink-pro/discussions) 参与讨论
- 📱 加入 Telegram: [SublinkPro 社区](https://t.me/sublinkpro)

## 许可证

向 SublinkPro 贡献即表示你同意你的贡献将采用与项目相同的许可证。
