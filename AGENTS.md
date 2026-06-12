# AGENTS.md

本文件为 `SublinkPro` 的人类与 AI 贡献者提供仓库级工作说明。  
This file gives human and AI contributors a repo-specific working guide for `SublinkPro`.

## 1. 项目形态 / Project shape

- 这是一个**单仓库全栈应用**，不是 monorepo。  
  This repository is a **single full-stack application**, not a monorepo.
- **后端**位于仓库根目录，使用 Go。  
  The **backend** lives at the repo root and is written in Go.
- **前端**位于 `webs/`，使用 React + Vite。  
  The **frontend** lives in `webs/` and uses React + Vite.
- **生产环境集成方式**是先构建前端资源，再由 Go 产物在生产构建中提供并嵌入这些静态资源。  
  **Production integration** works by building the frontend first, then serving and embedding those built assets from the Go production build.

主要入口文件：  
Primary entry points:

- 后端：`main.go`  
  Backend: `main.go`
- 前端：`webs/src/index.jsx`  
  Frontend: `webs/src/index.jsx`

关键架构边界：  
Important architectural boundaries:

- `routers/`：注册路由。  
  `routers/`: route registration.
- `api/`：HTTP handler / controller 逻辑。  
  `api/`: HTTP handlers / controller logic.
- `services/`：业务逻辑与长生命周期后台流程。  
  `services/`: business logic and long-running/background workflows.
- `models/`：持久化、模型与迁移相关逻辑。  
  `models/`: persistence, data models, and migration-related logic.
- `middlewares/`：认证与请求链路中间件。  
  `middlewares/`: auth and request middleware.
- `node/`：订阅与协议解析/转换逻辑。  
  `node/`: subscription and protocol parsing/conversion logic.
- `webs/src/api/`：前端请求边界。  
  `webs/src/api/`: the frontend API boundary.
- `webs/src/views/`：前端页面级功能。  
  `webs/src/views/`: page-level frontend features.
- `skill-sublinkpro/`：面向用户的 AI agent 技能（可移植 `SKILL.md` 格式，不绑定特定厂商，任何兼容的 AI agent / 客户端均可使用），通过 `X-API-Key` 调用 `/api/v1/*` REST API；它是 API 契约的外部消费者，其 `reference/api.md` 是 REST API 的镜像文档。  
  `skill-sublinkpro/`: a user-facing AI agent skill (portable `SKILL.md` format, vendor-neutral and usable by any compatible AI agent/client), calling the `/api/v1/*` REST API via `X-API-Key`; it is an external consumer of the API contract, and its `reference/api.md` mirrors the REST API.

## 2. 事实来源 / Source of truth

当多份说明不一致时，按下面顺序信任这些文件：  
When instructions conflict, trust these files first:

1. `webs/package.json`：前端命令与依赖事实来源。  
   `webs/package.json`: source of truth for frontend scripts and dependencies.
2. `docs/development.md`：本地开发流程与目录说明。  
   `docs/development.md`: local development workflow and project structure.
3. `docs/configuration.md`：配置优先级、环境变量与运行语义。  
   `docs/configuration.md`: config precedence, environment variables, and runtime behavior.
4. `.github/workflows/pr-checks.yml`：PR 自动检查预期。
   `.github/workflows/pr-checks.yml`: PR automated check expectations.
5. `.github/workflows/build-release.yml`：CI / release 构建预期。
   `.github/workflows/build-release.yml`: CI and release build expectations.
6. `Dockerfile`：生产构建路径的权威实现。
   `Dockerfile`: canonical production build sequence.

对于前端命令、输出目录和工具链，请优先依赖上面的仓库文件，不要套用通用模板或旧脚手架默认值。  
For frontend commands, output paths, and toolchain behavior, rely on the repository files above rather than generic template defaults or framework assumptions.

## 3. 技术栈 / Tech stack

- 后端：Go + Gin + GORM  
  Backend: Go + Gin + GORM
- 核心网络/代理内核：mihomo（MetaCubeX）  
  Core network/proxy engine: mihomo (MetaCubeX)
- 默认数据库：SQLite  
  Default database: SQLite
- 可选数据库：MySQL / PostgreSQL（通过 DSN）  
  Optional databases: MySQL / PostgreSQL via DSN
- 前端：React 19 + Vite  
  Frontend: React 19 + Vite
- UI：Material UI  
  UI: Material UI
- 前端包管理器：Yarn 4（`packageManager: yarn@4.10.3`）  
  Frontend package manager: Yarn 4 (`packageManager: yarn@4.10.3`)
- 定时调度：`robfig/cron`  
  Scheduler: `robfig/cron`

### mihomo 说明 / About mihomo

- `mihomo` 在这个仓库里不是普通第三方依赖，而是多个核心网络能力的底层支撑。  
  In this repository, `mihomo` is not a minor third-party dependency; it is the underlying engine for several core networking capabilities.
- 从上游语义上看，mihomo 更接近**代理 / 路由 / DNS / 网络出站控制内核**，而不是单纯的 GUI、单一协议实现，或“只是一个客户端”。  
  Upstream, mihomo is better described as a **proxy / routing / DNS / outbound networking core**, not merely a GUI, a single protocol implementation, or “just a client.”
- 在本仓库里，contributors 应把 mihomo 理解为“代理适配器创建、代理拨号、延迟/速度测试、DNS 解析、Host 注入与部分代理外呼能力”的关键集成点。  
  In this repo, contributors should treat mihomo as the key integration point for proxy adapter creation, proxied dialing, delay/speed testing, DNS resolution, Host injection, and some outbound proxy-backed requests.

## 4. 目录指南 / Directory guide

### 后端 / Backend

- `main.go`：应用入口、启动编排、配置/数据库初始化、服务启动与静态资源服务。  
  `main.go`: application entry, startup orchestration, bootstrap, and serving behavior.
- `api/`：后端 HTTP handler。  
  `api/`: backend HTTP handlers.
- `routers/`：路由注册。  
  `routers/`: route registration.
- `services/`：业务服务与后台子系统。  
  `services/`: business logic and long-running subsystems.
- `services/mihomo/`：mihomo 集成核心层；处理代理适配器、测速、DNS 解析和 Host 同步。  
  `services/mihomo/`: core mihomo integration layer for proxy adapters, testing, DNS resolution, and Host synchronization.
- `services/scheduler/`：定时任务与调度管理；涉及 cron / 任务中心时优先从这里入手。  
  `services/scheduler/`: scheduled jobs and scheduler management; start here for cron/task work.
- `models/`：数据库模型、迁移和持久化逻辑。  
  `models/`: database models, migrations, and persistence logic.
- `middlewares/`：认证和请求链路中间件。  
  `middlewares/`: auth and request pipeline logic.
- `utils/`：共享工具函数。  
  `utils/`: shared helpers.
- `template/`：运行时模板资源。  
  `template/`: runtime template assets.
- `config/`、`settings/`、`database/`、`constants/`、`dto/`：后端支持层。  
  `config/`, `settings/`, `database/`, `constants/`, and `dto/`: backend support layers.

### 前端 / Frontend

- `webs/src/views/`：主要页面与功能入口。  
  `webs/src/views/`: page-level features.
- `webs/src/api/`：前端请求层。  
  `webs/src/api/`: request layer.
- `webs/src/routes/`：路由与导航结构。  
  `webs/src/routes/`: route setup and navigation structure.
- `webs/src/layout/`、`webs/src/themes/`、`webs/src/components/`：共享 UI 外壳与组件。  
  `webs/src/layout/`, `webs/src/themes/`, and `webs/src/components/`: shared UI shell and components.

## 4.1 前端主题适配规范 / Frontend theme adaptation rules

凡是前端 UI 改动涉及颜色、表面、阴影、边框、图标、弹窗、面板、预览卡、悬浮层、提示块、标签、富文本或状态色，都必须同时检查 light / dark 两种模式，而不是只在当前模式下“看起来正常”。  
Any frontend UI change involving colors, surfaces, shadows, borders, icons, dialogs, panels, preview cards, overlays, notices, chips, rich text, or status colors must be checked in both light and dark modes, not only in the mode currently visible.

### 主题事实来源 / Theme source of truth

- 优先使用 `theme.vars?.palette` 与 `theme.palette` 中现有的 token。  
  Prefer existing tokens from `theme.vars?.palette` and `theme.palette`.
- 表面层优先使用：`background.default`、`background.paper`、`divider`、`text.primary`、`text.secondary`。  
  Prefer `background.default`, `background.paper`, `divider`, `text.primary`, and `text.secondary` for surfaces and text hierarchy.
- 状态或强调色优先使用现有 palette 语义：`primary` / `secondary` / `info` / `success` / `warning` / `error`。  
  Use existing semantic palette colors for emphasis or state: `primary`, `secondary`, `info`, `success`, `warning`, and `error`.
- 透明度和层级优先通过 `alpha()` 或仓库已有 helper（如 `withAlpha()`）表达，不要硬写一组新的浅深色体系。  
  Express opacity and elevation with `alpha()` or existing helpers like `withAlpha()` instead of inventing a parallel color system.

### 实施规则 / Implementation rules

- 外层卡片、弹窗主体、页面主 panel：优先使用 `background.paper` 语义；暗色模式下允许使用轻量渐变、弱边框和轻微内高光，但不要做成大块发白表面。  
  Outer cards, dialog bodies, and main panels should generally use `background.paper`; in dark mode, light gradients, soft borders, and subtle inner highlights are acceptable, but avoid large bright slabs.
- 内层嵌套 panel、底栏、预览 footer、提示块：优先用 `background.default` 或基于语义色的低透明度染色表面，不要“白底嵌白底”。  
  Nested panels, bottom strips, preview footers, and notice blocks should generally use `background.default` or lightly tinted semantic surfaces; do not create white-on-white layering.
- 主值 / 标题 / 可点击重点信息：优先用 `text.primary` 及字重体现层级；暗色模式可以提高 alpha，但不要默认全改成纯白。  
  Main values, titles, and key interactive text should rely on `text.primary` and font weight; in dark mode, a higher alpha is acceptable, but do not default everything to pure white.
- 次级说明 / 辅助信息 / 百分比 / 提示文案：优先用 `text.secondary` 或基于 `text.primary` 的较低 alpha，而不是继续叠一个新的灰色常量。  
  Secondary descriptions, helper text, percentages, and hints should use `text.secondary` or a lower-alpha `text.primary`, rather than introducing ad-hoc gray constants.
- 图标按钮、header action、通知入口、theme toggle 等高频交互元素：必须同时检查静态、hover、active/selected、disabled 四种状态在 light / dark 下的可见性。  
  High-frequency interactive elements such as icon buttons, header actions, notification entry points, and theme toggles must be checked in light/dark across default, hover, active/selected, and disabled states.
- 富文本、markdown、代码块、列表和 footer overlay：不要只修容器背景，必须同步检查正文、列表项、code、link、marker 和底栏文字对比度。  
  For rich text, markdown, code blocks, lists, and footer overlays, do not stop at fixing the container background; also verify body text, list items, code, links, markers, and footer text contrast.

### 明确禁区 / Explicit anti-patterns

- 不要在组件里直接硬编码 `#fff`、`white`、`#000`、`black` 作为大面积表面颜色，除非确实是品牌/图标语义且无法用 token 表达。  
  Do not hardcode `#fff`, `white`, `#000`, or `black` for large surfaces unless it is truly brand/icon-specific and cannot be expressed with theme tokens.
- 不要通过“把 dark mode 所有字都调成纯白”来解决可读性问题。  
  Do not solve dark-mode readability by turning all text pure white.
- 不要在一个页面里为同类 surface 写多套互相无关的 `isDark ? ... : ...` magic number。  
  Do not create multiple unrelated `isDark ? ... : ...` magic-number surface styles for equivalent UI layers within one page.
- 不要只修一个组件实例而放任同类 preview / dialog / footer / chip 在其他文件继续使用旧语义。  
  Do not fix only one component instance while equivalent previews, dialogs, footers, or chips elsewhere still use outdated semantics.

### 推荐工作方式 / Recommended workflow

- 先在相关文件附近寻找现有 helper、复用组件或同类 `sx` 模式，再决定是否新增小型共享 helper。  
  First look for nearby helpers, reusable components, or similar `sx` patterns before deciding to add a small shared helper.
- 如果同类问题出现在 2 个以上文件中，优先抽一个小而明确的 helper，而不是复制粘贴一组颜色表达式。  
  If the same theme problem appears in 2 or more files, prefer a small focused helper over copy-pasting color expressions.
- 交付说明中必须写明：检查了哪些 light / dark 场景、改了哪些 surface / text / icon 层级、哪些相关层已检查但无需修改。  
  In your delivery summary, explicitly state which light/dark scenarios were checked, which surface/text/icon layers changed, and which related layers were inspected but needed no modification.

### 主题覆盖范围 / Theme coverage expectations

- 主题适配不是“修当前看到的那个组件”就结束；只要改动涉及某个页面、模块、共享 helper、theme override 或公共组件的主题语义，就必须把该功能链路中的同类页面、弹窗、抽屉、预览卡、嵌套 panel、底栏、overlay 与移动端入口一并检查。  
  Theme adaptation does not stop after fixing the one visible component; if the change affects a page, module, shared helper, theme override, or reusable component, inspect all related pages, dialogs, drawers, preview cards, nested panels, footers, overlays, and mobile entry points in that feature flow.
- 如果你改的是公共主题基础设施，例如 `webs/src/themes/*`、`webs/src/utils/colorUtils.js`、共享 `sx` helper、通用 dialog/card/panel 样式，那么默认影响范围应视为“所有复用该模式的页面”，不能只验证当前需求触达的单一路径。  
  If you change shared theme infrastructure such as `webs/src/themes/*`, `webs/src/utils/colorUtils.js`, shared `sx` helpers, or common dialog/card/panel styles, treat the default blast radius as every page reusing that pattern rather than only the page touched by the current task.
- 如果一个模块同时存在桌面版与移动版、dialog 与 fullscreen mobile、drawer 与页面内 panel、卡片列表与详情面板两套入口，主题适配时必须成套覆盖，不能只修桌面或只修主入口。  
  If a module has desktop and mobile variants, dialog and fullscreen mobile modes, drawer and inline-panel variants, or both list-card and detail-panel entry points, theme work must cover the full set rather than only the desktop or primary path.
- 凡是带有 `useMediaQuery`、`fullScreen`、`Drawer`、`Popover`、`Menu`、`Tooltip`、`Collapse`、`Tabs`、`Stepper`、`ReactMarkdown`、`code` / `pre`、自定义 footer overlay、sticky action bar 的组件，都应默认视为主题适配高风险区域。  
  Components using `useMediaQuery`, `fullScreen`, `Drawer`, `Popover`, `Menu`, `Tooltip`, `Collapse`, `Tabs`, `Stepper`, `ReactMarkdown`, `code` / `pre`, custom footer overlays, or sticky action bars should be treated as high-risk theme-adaptation areas by default.
- 对话框和弹层必须按完整结构检查：`Dialog` / `Drawer` / `Popover` / `Menu` / `Tooltip` 的 trigger、paper、title、content、actions、backdrop、scroll container、内部嵌套 panel、关闭按钮、底部操作区与 fullscreen mobile 形态都要核对，不允许只改容器背景。  
  Dialogs and overlays must be checked as full structures: triggers, paper, title, content, actions, backdrop, scroll containers, nested panels, close buttons, footer actions, and fullscreen mobile variants for `Dialog` / `Drawer` / `Popover` / `Menu` / `Tooltip`; changing only the container background is not enough.
- 当某个模块已经在仓库中存在成熟的 dark/light 适配样板时，优先复用其 surface 分层、边框透明度、状态色和移动端处理，而不是重新发明一套局部规则。当前优先参考簇包括 `NodePreviewDialog.jsx`、`NodePreviewCard.jsx`、`NodePreviewDetailsPanel.jsx`、`ChainProxyDialog.jsx`、`ChainPreviewDialog.jsx`、`MobileChainBuilder.jsx`、`ConditionBuilder.jsx`、`layout/MainLayout/Header/NotificationSection/index.jsx`、`views/dashboard/Default/index.jsx`，以及机场管理链路中的 `views/airports/index.jsx`、`AirportMobileList.jsx`、`AirportFormDialog.jsx`、`AirportBatchEditDialog.jsx` 与共享的 `components/CronExpressionGenerator.jsx`。  
  When a mature dark/light adaptation pattern already exists in the repo, reuse its surface layering, border opacity, state colors, and mobile handling instead of inventing a new local rule set. Current primary reference clusters include `NodePreviewDialog.jsx`, `NodePreviewCard.jsx`, `NodePreviewDetailsPanel.jsx`, `ChainProxyDialog.jsx`, `ChainPreviewDialog.jsx`, `MobileChainBuilder.jsx`, `ConditionBuilder.jsx`, `layout/MainLayout/Header/NotificationSection/index.jsx`, `views/dashboard/Default/index.jsx`, plus the airport-management theme cluster: `views/airports/index.jsx`, `AirportMobileList.jsx`, `AirportFormDialog.jsx`, `AirportBatchEditDialog.jsx`, and the shared `components/CronExpressionGenerator.jsx`.

### 浅色迁移到深色时的高风险遗漏点 / High-risk misses when migrating from light to dark

- 只改外层页面容器，没有同步处理内层 header strip、footer strip、nested panel、统计块、详情块、空状态、skeleton 或 preview footer。  
  Only the outer page container was updated, while inner header strips, footer strips, nested panels, stat blocks, detail blocks, empty states, skeletons, or preview footers were left behind.
- 只改文本颜色，没有同步处理边框、阴影、divider、hover、active、selected、disabled、focus-visible 与滚动容器边界。  
  Text color was updated, but borders, shadows, dividers, hover, active, selected, disabled, focus-visible states, and scroll-container edges were not.
- 只修主页面，没有同步修同模块的 dialog、drawer、popover、menu、tooltip、fullscreen mobile dialog、底部菜单或二级详情入口。  
  The main page was fixed, but the same module's dialogs, drawers, popovers, menus, tooltips, fullscreen mobile dialogs, bottom menus, or secondary detail entry points were not.
- 只修桌面端，没有同步检查移动端卡片、响应式折叠布局、safe-area、sticky footer、横向滚动、点击热区与 icon button 可见性。  
  Desktop UI was fixed, but mobile cards, responsive collapsed layouts, safe areas, sticky footers, horizontal scrolling, tap targets, and icon-button visibility were not checked.
- 只修容器，没有同步修 `ReactMarkdown`、`code`、`pre`、列表 marker、链接、表格、chip、badge、tag、状态提示块与通知浮层。  
  Containers were updated without also fixing `ReactMarkdown`, `code`, `pre`, list markers, links, tables, chips, badges, tags, status notices, and notification overlays.
- 状态色仍沿用浅色主题的 tint，导致 success / warning / error / info 在暗色模式下背景不够区分、文字发灰、边框发脏。  
  State colors kept light-mode tints, making success / warning / error / info backgrounds indistinct, text muddy, or borders dirty in dark mode.
- 图标、次级文字、百分比、caption、placeholder、helper text 与 disabled 文案没有跟随正文层级一起调整，导致“主标题能看，辅文全部发灰”。  
  Icons, secondary text, percentages, captions, placeholders, helper text, and disabled copy were not adjusted with the main text hierarchy, resulting in readable titles but washed-out supporting content.
- DataGrid、Autocomplete、Select、ToggleButtonGroup、Tabs、列表 hover 行、搜索框、过滤条和表单控件只修了默认态，没有连同 popup、选中态、聚焦态一起检查。  
  DataGrid, Autocomplete, Select, ToggleButtonGroup, Tabs, hovered list rows, search bars, filter bars, and form controls were only fixed in their default state without validating popups, selected states, and focus states.

### 深色主题迁移修改原则 / Dark-mode migration principles

- 优先复用仓库既有模式：`const palette = theme.vars?.palette || theme.palette;`，再抽局部的 `dialogSurface`、`mutedPanelSurface`、`nestedPanelSurface`、`panelBorder` 等语义变量，避免在 JSX 中散落大量 `isDark ? ... : ...`。  
  Prefer the repo's established pattern: start with `const palette = theme.vars?.palette || theme.palette;`, then derive local semantic variables such as `dialogSurface`, `mutedPanelSurface`, `nestedPanelSurface`, and `panelBorder` instead of scattering many `isDark ? ... : ...` expressions across JSX.
- 深色模式优先做“语义表面分层”，不是简单把浅色背景替换成黑色。外层通常使用 `background.paper`，内层面板、页脚、提示块与弱化区域优先使用 `background.default` 或基于语义色的低透明度染色。  
  In dark mode, prioritize semantic surface layering rather than replacing light backgrounds with black. Outer containers should usually use `background.paper`, while inner panels, footers, notices, and muted sections should prefer `background.default` or low-opacity semantic tinting.
- 边框、分隔线和层级优先通过 `divider`、`alpha()`、`withAlpha()` 与轻量 inset highlight 表达，不要在每个组件里手工发明一套新的灰阶。  
  Express borders, dividers, and elevation with `divider`, `alpha()`, `withAlpha()`, and subtle inset highlights instead of inventing a new gray scale in each component.
- 主文本、强调文本、次级文本、占位文案、禁用文案、图标与状态色必须保持层级关系稳定；不要用“全部调亮”掩盖层级混乱。  
  Main text, emphasis text, secondary text, placeholders, disabled copy, icons, and state colors must preserve stable hierarchy; do not hide hierarchy problems by simply brightening everything.
- 状态色优先保持“语义文字 + 低透明度背景 + 轻量边框”的组合，而不是在暗色模式下直接铺满高饱和纯色块。  
  State colors should generally use the combination of semantic text, low-opacity backgrounds, and light borders rather than large high-saturation solid fills in dark mode.
- 当同类问题在多个文件重复出现时，优先抽小型共享 helper 或复用现有 override；不要把相同的 dark-mode magic number 复制到多个模块。  
  When the same issue appears in multiple files, extract a small shared helper or reuse an existing override instead of copying the same dark-mode magic numbers into multiple modules.

### 主题迁移执行手册 / Theme migration playbook

当你在本仓库做主题适配时，默认按下面顺序执行，不要跳步：  
When performing theme adaptation in this repo, follow this sequence by default and do not skip steps:

1. 先界定影响范围：当前改动触及的是页面、模块、共享组件、theme override，还是全局 theme helper；如果是后两者，默认把所有复用该模式的页面都列入检查范围。  
   First define the blast radius: determine whether the change affects a page, a module, a shared component, a theme override, or a global theme helper; if it is one of the latter two, default to checking every page reusing that pattern.
2. 先找仓库内现成样板，再写代码：优先参考 `NodePreview*`、`Chain*`、`MobileChainBuilder.jsx`、`ConditionBuilder.jsx`、`NotificationSection`、`dashboard/Default`，以及机场管理链路中的 `airports/index.jsx`、`AirportMobileList.jsx`、`AirportFormDialog.jsx`、`AirportBatchEditDialog.jsx`、`CronExpressionGenerator.jsx`，先复用它们的 surface 分层和状态表达，再决定是否抽新的 helper。  
   Find an internal reference implementation before writing code: first reuse the surface layering and state treatment from `NodePreview*`, `Chain*`, `MobileChainBuilder.jsx`, `ConditionBuilder.jsx`, `NotificationSection`, `dashboard/Default`, and the airport-management cluster (`airports/index.jsx`, `AirportMobileList.jsx`, `AirportFormDialog.jsx`, `AirportBatchEditDialog.jsx`, `CronExpressionGenerator.jsx`), then decide whether a new helper is actually needed.
3. 先修语义分层，再修局部颜色：优先处理外层 surface、内层 panel、footer/header strip、边框、divider、文本层级、状态色和 icon，再处理局部渐变、阴影或装饰。  
   Fix semantic layering before tweaking local colors: handle outer surfaces, inner panels, footer/header strips, borders, dividers, text hierarchy, state colors, and icons before touching local gradients, shadows, or decorative polish.
4. 改主路径时必须同步检查分支路径：桌面端改完后，继续检查移动端、fullscreen dialog、drawer、popover、menu、tooltip、collapse、empty state、skeleton、footer overlay、markdown/code/pre、列表 hover 与表单 focus。  
   Whenever you change the main path, immediately check the branch paths as well: after desktop, continue through mobile, fullscreen dialogs, drawers, popovers, menus, tooltips, collapse states, empty states, skeletons, footer overlays, markdown/code/pre, hovered lists, and focused form controls.
5. 交付前必须写清覆盖面：说明本次适配覆盖了哪些页面/模块/弹层/移动端入口，复用了哪些仓库现有样板，哪些同类实例已检查但无需修改。  
   Before delivery, explicitly state which pages/modules/overlays/mobile entry points were covered, which existing repo patterns were reused, and which similar instances were checked but did not need changes.

如果你没有时间把一个模块的桌面端、移动端、dialog/drawer、overlay、details panel 一起检查完，那就不要宣称该模块“已完成主题适配”；应明确标记为部分覆盖。  
If you do not have time to inspect a module's desktop, mobile, dialog/drawer, overlay, and detail-panel variants together, do not claim the module is “theme-adapted”; mark it explicitly as partial coverage instead.

### 主题改动完成条件 / Theme-related definition of done

- 至少核对 light / dark 两种模式。  
  Check both light and dark modes.
- 至少核对相关 hover / active / disabled 或展开态。  
  Check relevant hover / active / disabled or expanded states.
- 不允许出现“容器变暗了，但内层 panel、footer、markdown、percent、icon 仍然发白或发灰不可读”的交付状态。  
  Do not ship a state where the container is fixed but inner panels, footers, markdown, percentages, or icons remain too bright or too dim to read.
- 不允许把主题适配交付成“仅当前页面修好、同模块弹窗/移动端/浮层仍然沿用旧语义”的半完成状态。  
  Do not ship theme work in a half-finished state where the current page looks correct but dialogs, mobile variants, or overlays in the same module still use the old semantics.
- 只要本次改动触及共享主题基础设施、公共组件或模块级视觉语义，交付说明中必须明确列出已覆盖的页面、弹层、移动端入口和已检查但无需修改的相关实例。  
  Whenever the change touches shared theme infrastructure, reusable components, or module-level visual semantics, the delivery summary must explicitly list the covered pages, overlays, mobile entry points, and the related instances that were checked but did not require changes.

### 运行时路径 / Operational and runtime paths

这些目录可能包含运行时状态，应谨慎处理：  
These directories may contain runtime state and should be handled carefully:

- `db/`
- `logs/`
- `cache/`
- `out/`

正常功能开发时，不要随意编辑、删除或格式化这些运行时数据。  
Do not casually edit, delete, or reformat runtime data during normal feature work.

## 4.2 国际化规范 / Internationalization requirements

本仓库已经具备前端国际化基础设施和成对维护的国际化实施契约。凡是新增或修改用户可见文本、API 展示语义、任务结果、通知内容、导出内容、文档说明或语言选择行为，都必须先阅读并遵守 `docs/internationalization.md` 与 `docs/internationalization.zh-CN.md`。
This repository already has frontend i18n infrastructure and paired internationalization contract documents. Whenever you add or change user-visible text, API display semantics, task results, notification content, exported content, documentation, or language-selection behavior, read and follow `docs/internationalization.md` and `docs/internationalization.zh-CN.md` first.

### 当前国际化形态 / Current i18n shape

- 首批支持 locale 是 `zh-CN` 和 `en-US`；新增用户可见前端文本时，必须在同一次变更中补齐这两种语言。
  The initial supported locales are `zh-CN` and `en-US`; new user-visible frontend text must include both translations in the same change.
- 前端使用 `i18next`、`react-i18next` 和 `i18next-browser-languagedetector`，入口在 `webs/src/i18n/index.js`，locale、MUI 映射与格式化 helper 在 `webs/src/i18n/locales.js`。
  The frontend uses `i18next`, `react-i18next`, and `i18next-browser-languagedetector`; the entrypoint is `webs/src/i18n/index.js`, with locale normalization, MUI mapping, and formatting helpers in `webs/src/i18n/locales.js`.
- 翻译资源当前按 locale 放在 `webs/src/i18n/locales/`，例如 `zh-CN.json` 和 `en-US.json`；不要把前端翻译资源放进后端目录、运行时目录或数据库。
  Translation resources currently live under `webs/src/i18n/locales/`, such as `zh-CN.json` and `en-US.json`; do not place frontend translation resources in backend directories, runtime directories, or the database.
- 语言选择不使用 URL path 或 subdomain；`SUBLINK_WEB_BASE_PATH` 已经是 Web UI 路由关注点，国际化改动不得额外引入语言路径前缀或域名路由层。
  Language selection does not use URL paths or subdomains; `SUBLINK_WEB_BASE_PATH` is already the Web UI routing concern, so i18n changes must not add locale path prefixes or host-based routing.

### 前端国际化规则 / Frontend i18n rules

- React 组件中的用户可见文本必须通过 `useTranslation()`、`Trans` 或已有 i18n helper 输出；不要在 JSX、toast、dialog、菜单、表单校验、空状态、tooltip、图表标签或移动端入口中继续硬编码中文或英文文案。
  User-visible text in React components must go through `useTranslation()`, `Trans`, or existing i18n helpers; do not keep hardcoded Chinese or English copy in JSX, toasts, dialogs, menus, form validation, empty states, tooltips, chart labels, or mobile entry points.
- Key 必须表达稳定语义而不是当前文案；使用 lower camel case 的点路径，按功能或稳定 UI 边界组织 namespace，优先复用 `common`、`airports`、`subscriptions`、`dashboard`、`settings`、`tasks` 等既有结构。
  Keys must express stable meaning rather than current wording; use lower-camel-case dot paths, organize namespaces by feature or stable UI boundary, and prefer existing structures such as `common`, `airports`, `subscriptions`, `dashboard`, `settings`, and `tasks`.
- 不要在 JSX 中拼接多个翻译片段来组成一句话；完整短语、句子、复数、插值和富文本应作为完整翻译项处理，富文本使用 `Trans` 并显式传入允许的组件。
  Do not concatenate translated fragments in JSX to build a sentence; full phrases, sentences, plurals, interpolation, and rich text should be represented as complete translation entries, with rich text rendered through `Trans` and explicit allowed components.
- 日期、时间、相对时间、数字、百分比、字节、流量和统计值必须通过接收当前 locale 的共享 helper 或 `Intl` 格式化；不要在功能组件里硬编码中文或英文日期顺序、分隔符或单位。
  Dates, times, relative time, numbers, percentages, bytes, traffic values, and statistics must be formatted through shared helpers that receive the active locale or through `Intl`; do not hardcode Chinese or English date order, separators, or units in feature components.
- MUI locale 必须由当前应用 locale 推导，新增语言时同步更新 `getMuiLocale()`、语言归一化、语言切换入口和完整翻译资源。
  The MUI locale must be derived from the active app locale; when adding a language, update `getMuiLocale()`, language normalization, language switchers, and complete translation resources together.
- 协议值、节点名称、域名、文件路径、API key、token、模板语法、订阅输出和外部客户端消费的字段值必须保持原样；只翻译周边 UI 标签和说明。
  Protocol values, node names, domains, file paths, API keys, tokens, template syntax, subscription output, and field values consumed by external clients must remain unchanged; translate only surrounding UI labels and descriptions.

### 后端国际化规则 / Backend i18n rules

- 后端必须保持现有 API 客户端兼容。已经暴露的 `msg`、`code`、字段名、路由、状态码和业务码不得为了国际化随意改名或改语义。
  The backend must preserve compatibility with existing API clients. Existing `msg`, `code`, field names, routes, status codes, and business codes must not be renamed or semantically changed just for i18n.
- 新增或改造会在 Web UI 展示的错误、任务结果、通知结果或操作反馈时，优先在响应中补充可选的 `i18nKey` 和 `i18nParams`，让前端按当前 locale 渲染；`msg` 保留为旧客户端和诊断回退。
  When adding or revising errors, task results, notification results, or operation feedback shown in the Web UI, prefer adding optional `i18nKey` and `i18nParams` so the frontend can render with the active locale; keep `msg` as the fallback for older clients and diagnostics.
- `i18nParams` 只能包含普通 JSON 值，不能塞入已本地化句子、HTML、React 结构、错误对象、数据库模型或敏感数据。
  `i18nParams` may contain only plain JSON values, not localized sentences, HTML, React structures, error objects, database models, or sensitive data.
- 日志、调试标识、协议输出、订阅内容、模板变量、webhook payload、Telegram Bot 回复、导出文件和外部客户端直接消费的内容，应按其产品边界处理；如果内容会进入 Web UI 且需要本地化，使用 `i18nKey` 加参数而不是让后端按浏览器语言拼接文案。
  Logs, debug identifiers, protocol output, subscription content, template variables, webhook payloads, Telegram Bot replies, exported files, and content consumed directly by external clients should be handled according to their product boundary; if the content enters the Web UI and needs localization, use `i18nKey` plus params instead of having the backend assemble browser-language text.

### 跨层同步 / Cross-layer synchronization

- 任何影响前端展示文案、API 错误语义、任务中心结果、通知内容、设置说明、认证/MFA 提示、配置语义、订阅分享提示或导出展示的改动，都必须同步检查后端 handler/service、`webs/src/api/`、相关 `webs/src/views/` 页面、翻译资源和文档。
  Any change affecting frontend copy, API error semantics, task-center results, notification content, settings descriptions, auth/MFA prompts, configuration semantics, subscription-share prompts, or exported display must check backend handlers/services, `webs/src/api/`, related `webs/src/views/` pages, translation resources, and docs together.
- 如果只改一层后确认另一层无需修改，交付说明必须写明已检查哪些前端、后端或文档入口，以及为什么无需同步调整。
  If only one layer changes and the other layer truly needs no edit, the delivery summary must state which frontend, backend, or documentation entry points were checked and why no synchronized change was required.
- i18n 行为或贡献者契约变化时，必须成对更新 `docs/internationalization.md` 与 `docs/internationalization.zh-CN.md`；涉及 README 或功能文档时，也必须同步更新英文 canonical 文件和对应 `*.zh-CN.md` 文件。
  When i18n behavior or contributor contracts change, update `docs/internationalization.md` and `docs/internationalization.zh-CN.md` as a pair; if README or feature docs are affected, also update the English canonical file and the matching `*.zh-CN.md` file.

### 国际化完成条件 / I18n definition of done

- 至少确认 `zh-CN` 和 `en-US` 两种语言下新增或修改的页面、弹窗、抽屉、移动端入口、空状态、错误状态、任务结果和通知文案都可读且语义一致。
  Verify at least that the changed pages, dialogs, drawers, mobile entry points, empty states, error states, task results, and notification copy are readable and semantically aligned in both `zh-CN` and `en-US`.
- 前端 i18n 代码改动必须运行 `yarn run lint`；如果影响路由、资源、base-path、生产集成或构建行为，还必须运行 `yarn run build`。不要发明仓库中不存在的前端 `test` 或 `typecheck` 命令。
  Frontend i18n code changes must run `yarn run lint`; if routing, assets, base-path, production integration, or build behavior is affected, also run `yarn run build`. Do not invent frontend `test` or `typecheck` commands that do not exist in this repository.
- 后端 i18n 行为改动必须对改动的 Go 文件运行 `gofmt`，并运行 `golangci-lint run` 和相关 `go test`；涉及 API 响应契约时，优先补充或更新 handler/response 回归测试。
  Backend i18n behavior changes must run `gofmt` on changed Go files plus `golangci-lint run` and relevant `go test`; when API response contracts are affected, prefer adding or updating handler/response regression tests.
- 文档-only 的 AGENTS 或 i18n 规范改动不需要运行前后端构建，但必须人工核对链接、双语一致性、现有命令真实性，以及是否与本文件的跨层同步、测试和文档规则冲突。
  Documentation-only AGENTS or i18n contract changes do not require frontend/backend builds, but must be manually checked for links, bilingual consistency, real command names, and consistency with this file's cross-layer, testing, and documentation rules.

## 5. 本地开发命令 / Local development commands

### 后端 / Backend

在仓库根目录执行：  
Run from the repo root:

```bash
go mod download
go run main.go
```

### 前端 / Frontend

在 `webs/` 目录执行：  
Run from `webs/`:

```bash
yarn install
yarn run start
```

### 前端校验 / Frontend validation

在 `webs/` 目录执行：  
Run from `webs/`:

```bash
yarn run lint
yarn run build
yarn run lint:fix
yarn run prettier
```

前端改动完成后必须至少运行 `yarn run lint`；涉及构建、路由、资源或生产集成时还必须运行 `yarn run build`。
After frontend changes, run at least `yarn run lint`; also run `yarn run build` when build output, routing, assets, or production integration are affected.

### 后端校验 / Backend validation

在仓库根目录执行：
Run from the repo root:

```bash
gofmt -w <changed-go-files>
golangci-lint run
go test ./...
```

后端开发完成后必须保持 Go 文件通过 `gofmt` 格式化，并运行 `golangci-lint run` 和相关 Go 测试；发布前的 GitHub 工作流会运行 `golangci-lint` 和全仓 `go test ./...`。
After backend changes, keep Go files formatted with `gofmt`, and run `golangci-lint run` plus relevant Go tests; the GitHub release workflow runs `golangci-lint` and repository-wide `go test ./...` before publishing.

### PR 自动检查 / PR automated checks

`.github/workflows/pr-checks.yml` 会在 PR 打开、重新打开或标记 ready for review 时自动运行；后续修复提交不会自动重复消耗 Actions，PR 作者本人或仓库管理员可在 PR 评论 `/recheck` 手动触发新一轮检查。
`.github/workflows/pr-checks.yml` runs automatically when a PR is opened, reopened, or marked ready for review; later fix commits do not auto-consume Actions, and the PR author or repository admin can comment `/recheck` on the PR to trigger another run.

- 后端检查：`golangci-lint` 和 `go test ./...`。
  Backend checks: `golangci-lint` and `go test ./...`.
- 前端检查：`yarn run lint` 和 `yarn run build`。
  Frontend checks: `yarn run lint` and `yarn run build`.

### 后端构建 / Backend build

在仓库根目录执行：  
Run from the repo root:

```bash
go build -o sublinkpro main.go
```

### 生产风格后端构建 / Production-style backend build

CI 和 Docker 使用会嵌入前端资源的生产构建命令：  
CI and Docker use a production build that embeds frontend assets:

```bash
CGO_ENABLED=0 go build -tags=prod -ldflags="-s -w" -o sublinkPro
```

## 6. 构建与发布预期 / Build and release expectations

这个仓库的生产构建是**两阶段流程**：  
This repo’s production build is a **two-step process**:

1. 先在 `webs/` 中构建前端，生成 `dist/`。  
   Build the frontend in `webs/` so `dist/` exists.
2. 再把前端产物放到 `static/`，供 Go 生产构建嵌入。  
   Make those built assets available under `static/` for the Go production build.
3. 最后使用 `-tags=prod` 构建 Go 二进制。  
   Then build the Go binary with `-tags=prod`.

在 `.github/workflows/build-release.yml` 中观察到的 CI 流程：  
Observed CI flow in `.github/workflows/build-release.yml`:

- Node 22  
  Node 22
- `corepack enable`
- `cd webs && yarn install --immutable`
- `cd webs && yarn run lint`
- `cd webs && yarn run build`
- Go 1.26.3
- `golangci-lint`
- `go test ./...`
- 下载前端产物到 `static/`  
  download frontend artifacts into `static/`
- `go build -tags=prod ...`

在 `Dockerfile` 中观察到的流程：  
Observed Docker flow in `Dockerfile`:

- 在 `webs/` 中安装并构建前端  
  install and build the frontend in `webs/`
- 将前端产物复制到 `./static`  
  copy the frontend build into `./static`
- 使用 `go build -tags=prod` 构建后端  
  run `go build -tags=prod`

如果你修改前端资源布局、生产嵌入逻辑或路由/base-path 行为，必须同时验证本地开发和生产构建。  
If you change frontend asset layout, production embedding, or routing/base-path behavior, verify both local development and production build assumptions.

## 7. 配置规则 / Configuration rules

这个仓库的配置优先级是明确且重要的：  
Config precedence is explicit and important in this repo:

1. 命令行参数 / Command-line flags
2. 环境变量 / Environment variables
3. 配置文件 `db/config.yaml` / Config file `db/config.yaml`
4. 数据库存储设置 / Database-stored settings
5. 默认值 / Defaults

`docs/configuration.md` 中的常见默认值：  
Common runtime defaults documented in `docs/configuration.md`:

- 端口：`8000` / Port: `8000`
- 数据目录：`./db` / DB path: `./db`
- 日志目录：`./logs` / Log path: `./logs`
- 默认 SQLite DSN：`sqlite://./db/sublink.db`  
  Default SQLite DSN: `sqlite://./db/sublink.db`

修改配置行为时请遵守以下约束：  
When changing configuration behavior:

- 如果语义变化，必须同步更新文档。  
  Update docs if semantics change.
- 除非任务明确要求，否则不要改变优先级规则。  
  Preserve precedence unless the task explicitly requires changing it.
- 对 JWT / API 加密密钥等敏感配置保持谨慎。  
  Be careful with sensitive settings such as JWT and API encryption keys.

多实例部署时，下面这些配置的语义尤其不能随意更改：  
For multi-instance deployments, do not casually change semantics around:

- `SUBLINK_JWT_SECRET`
- `SUBLINK_API_ENCRYPTION_KEY`
- `SUBLINK_MFA_RESET_SECRET`

## 8. 前后端契约 / Frontend-backend contract

- 前端请求统一经过 `webs/src/api/request.js`。  
  Frontend requests go through `webs/src/api/request.js`.
- 前端以 `/api` 作为 API 边界。  
  The frontend uses `/api` as the API boundary.
- base-path 行为是该仓库特有的前后端集成点。  
  Base-path behavior is a repo-specific frontend/backend integration point.
- `SUBLINK_WEB_BASE_PATH` 会影响 Web UI 路由，但不会影响 `/api/*` 与 `/c/*`。  
  `SUBLINK_WEB_BASE_PATH` affects web UI routing but not `/api/*` or `/c/*`.

如果你改动路由、认证、静态资源路径或页面刷新行为，请先同时检查前端和后端实现。  
If you touch routing, auth, asset paths, or page refresh behavior, inspect both backend and frontend sides before changing code.

- 任何影响前后端契约、接口字段、路由、鉴权、页面流程、base-path、静态资源、任务结果展示或配置语义的改动，都必须在**同一工作**中同步检查并更新相关前端、后端与文档，不能只改其中一层。  
  Any change affecting the frontend-backend contract, API fields, routing, auth, page flows, base-path, static assets, task result presentation, or configuration semantics must be checked and updated across the relevant frontend, backend, and documentation **within the same piece of work**; do not change only one layer.
- Web 前端不是 `/api/v1/*` 的唯一消费者：`skill-sublinkpro/` 也直接消费同一套 REST API。任何新增、删除、改名或修改 `/api/v1/*`（及 `/c/*`）路由、请求字段、内容类型（form vs JSON）、鉴权方式、响应结构或业务码的改动，都必须同步更新 `skill-sublinkpro/reference/api.md`，必要时更新 `skill-sublinkpro/SKILL.md` 与 `skill-sublinkpro/scripts/sublink.py`，保证 skill 与系统接口同步。  
  The web frontend is not the only consumer of `/api/v1/*`: `skill-sublinkpro/` also consumes the same REST API directly. Any change that adds, removes, renames, or modifies `/api/v1/*` (and `/c/*`) routes, request fields, content type (form vs JSON), auth behavior, response shapes, or business codes must also update `skill-sublinkpro/reference/api.md`, and when relevant `skill-sublinkpro/SKILL.md` and `skill-sublinkpro/scripts/sublink.py`, so the skill stays in sync with the system's API.
- 严禁出现“只改后端但前端请求/展示未同步”或“只改前端但后端接口/权限/数据结构未核对”的情况；如果确认另一层无需修改，也应在变更说明中明确写明已检查且无需同步调整。  
  Do not leave the repo in a state where backend changes are not reflected in frontend requests/UI, or frontend changes are made without checking backend APIs, permissions, and data structures; if another layer truly needs no changes, explicitly state that it was checked and no sync update was required.

## 9. 跨层同步为强制要求 / Cross-layer synchronization is mandatory

本仓库的默认工作方式不是“改到哪层算哪层”，而是“凡是行为受影响的层都必须一起检查并同步”。任何会影响接口、数据结构、页面展示、用户流程、配置语义、部署方式或文档含义的改动，都必须把相关层作为一个整体交付。  
The default workflow in this repo is not “change only the layer you touched,” but “inspect and synchronize every impacted layer.” Any change that affects APIs, data structures, UI behavior, user flows, configuration semantics, deployment behavior, or documentation meaning must be delivered as a coordinated cross-layer change.

### 完成条件 / Definition of done

- 后端改动如果影响接口、字段、权限、返回结构、任务结果、静态资源路径或路由行为，必须同步检查并更新前端请求层、页面层和相关文档。  
  If a backend change affects APIs, fields, permissions, response shapes, task results, static asset paths, or routing behavior, update the frontend request layer, frontend views, and related documentation in the same work.
- 前端改动如果影响接口依赖、鉴权方式、字段语义、页面流程、base-path、资源路径或刷新行为，必须同步核对后端实现和相关文档。  
  If a frontend change affects API dependencies, auth behavior, field semantics, page flows, base-path handling, asset paths, or refresh behavior, verify and update the backend implementation and related docs in the same work.
- 配置、部署、构建、迁移、MFA、安全语义、mihomo 相关能力的改动，必须同步更新 `README.md`、`docs/` 和必要的功能说明，不得只改代码。  
  Changes to configuration, deployment, build flow, migration, MFA, security semantics, or mihomo-related capabilities must be reflected in `README.md`, `docs/`, and relevant feature documentation; code-only delivery is not acceptable.

### 交付禁区 / Unacceptable delivery states

- 只改后端，但前端请求、展示、交互或提示文案仍停留在旧语义。  
  Backend changed, but frontend requests, UI behavior, or copy still reflect the old contract.
- 只改前端，但后端接口、权限、字段含义、返回结构或运行时行为未核对。  
  Frontend changed, but backend APIs, permissions, field meanings, response shapes, or runtime behavior were not checked.
- 代码行为已变更，但 `README.md`、`docs/` 或 `docs/features/*` 仍描述旧流程、旧字段、旧命令或旧配置语义。  
  Code behavior changed, but `README.md`, `docs/`, or `docs/features/*` still describe old flows, fields, commands, or configuration semantics.

### 允许不修改另一层的前提 / When it is acceptable not to change another layer

- 只有在你**已经检查**受影响的前端、后端或文档后，确认它们无需变更时，才可以不修改。  
  You may leave another layer untouched only after you have actually checked it and confirmed no update is required.
- 这种情况必须在变更说明中明确写出“已检查哪些层，以及为什么无需同步修改”。  
  In that case, explicitly state in your change summary which layers were checked and why no synchronized update was needed.

### 交付前检查清单 / Pre-delivery checklist

在宣称工作完成之前，至少逐项核对以下内容：  
Before claiming the work is complete, check each of the following:

- 后端是否有任何接口、字段、权限、路由、返回结构或运行语义变化；如果有，前端请求层、页面层和相关提示文案是否已经同步。  
  Whether any backend API, field, permission, route, response shape, or runtime semantic changed; if so, whether the frontend request layer, views, and user-facing copy were synchronized.
- 前端是否有任何交互流程、鉴权方式、字段依赖、base-path、资源路径或刷新行为变化；如果有，后端实现和相关文档是否已经核对。  
  Whether any frontend interaction flow, auth behavior, field dependency, base-path, asset path, or refresh behavior changed; if so, whether the backend implementation and related documentation were checked.
- `README.md`、`docs/`、`docs/features/*` 中是否存在仍然描述旧行为、旧字段、旧命令或旧配置语义的内容。  
  Whether `README.md`, `docs/`, or `docs/features/*` still describe outdated behavior, fields, commands, or configuration semantics.
- 是否已经运行与改动匹配的验证命令，并确认结果覆盖了受影响的前端、后端或构建链路。  
  Whether the relevant validation commands were run and their results cover the affected frontend, backend, or build pipeline.
- 变更说明中是否明确写出：改了哪些层、同步了哪些层、哪些层已检查但无需修改，以及为什么。  
  Whether the change summary explicitly states which layers changed, which layers were synchronized, and which layers were checked but required no modification, including why.

### 主题改动专项检查清单 / Theme-specific pre-delivery checklist

在宣称主题适配完成之前，至少逐项核对以下内容：  
Before claiming theme adaptation work is complete, check each of the following:

- 是否已经同时检查 light / dark 两种模式，而不是只在当前模式下截图或肉眼确认。  
  Whether both light and dark modes were checked rather than only visually validating the currently active one.
- 是否已经同时检查桌面端与移动端，尤其是 `useMediaQuery`、`fullScreen`、drawer、responsive collapse、sticky footer、bottom action bar、safe-area 相关路径。  
  Whether both desktop and mobile variants were checked, especially paths involving `useMediaQuery`, `fullScreen`, drawers, responsive collapse, sticky footers, bottom action bars, and safe areas.
- 是否已经检查同模块的页面主体、dialog、drawer、popover、menu、tooltip、preview card、details panel、footer overlay、empty state、skeleton 与通知/浮层入口。  
  Whether the module's page body, dialogs, drawers, popovers, menus, tooltips, preview cards, detail panels, footer overlays, empty states, skeletons, and notification/overlay entry points were checked.
- 是否已经检查主文字、次级文字、icon、chip、badge、link、code、pre、placeholder、helper text、disabled 文案和状态色在 light / dark 下都保持可读且层级一致。  
  Whether primary text, secondary text, icons, chips, badges, links, code, pre, placeholders, helper text, disabled copy, and state colors remain readable and hierarchically consistent in both light and dark modes.
- 是否已经检查 hover / active / selected / disabled / focus-visible / expanded 等交互态，而不是只看静态截图。  
  Whether hover / active / selected / disabled / focus-visible / expanded states were verified instead of relying only on static screenshots.
- 是否已经确认没有引入新的大面积硬编码 `#fff`、`white`、`#000`、`black`、脱离 token 的灰阶或只在单文件生效的 dark-mode magic number。  
  Whether no new large hardcoded `#fff`, `white`, `#000`, `black`, ad-hoc grays, or single-file dark-mode magic numbers were introduced.
- 是否已经在变更说明中明确写出：本次主题适配覆盖了哪些页面/模块/弹层/移动端入口，哪些同类实例已检查但无需修改，以及参考了哪些现有主题样板。  
  Whether the change summary clearly states which pages/modules/overlays/mobile entry points were covered, which similar instances were checked but did not need changes, and which existing theme reference patterns were followed.

## 10. mihomo 核心集成说明 / Mihomo integration guidance

本仓库有一批核心功能直接依赖 mihomo 相关能力，因此任何涉及代理、测速、DNS、Host 映射、链式代理兼容性、代理下载或代理外呼的改动，都不应把 mihomo 视为可忽略细节。  
Several core features in this repo directly depend on mihomo-related capabilities, so any change involving proxying, speed testing, DNS, Host mapping, chain-proxy compatibility, proxied downloads, or proxied outbound requests should not treat mihomo as an incidental detail.

### 关键依赖能力 / Key dependent capabilities

- 节点延迟测试与下载测速。  
  Node delay testing and download speed testing.
- 代理节点适配器构建与代理拨号。  
  Proxy adapter construction and proxied dialing.
- 代理服务器 DNS 解析与自定义 DNS / DoH 路径。  
  Proxy server DNS resolution and custom DNS / DoH flows.
- Host 映射同步到运行时解析器。  
  Host mapping synchronization into the runtime resolver.
- 通过代理执行的订阅下载、机场流量查询和部分外部请求。  
  Proxied subscription downloads, airport usage retrieval, and some outbound external requests.

### 优先查看的文件 / Files to inspect first

- `services/mihomo/mihomo.go`：mihomo 适配器、延迟测试、下载测速、落地 IP / IP 质量相关能力的核心封装。  
  `services/mihomo/mihomo.go`: primary wrapper for mihomo adapters, delay testing, speed testing, and landing-IP / IP-quality related flows.
- `services/mihomo/dns_resolver.go`：代理服务器域名解析、自定义 DNS、代理 DoH 等逻辑入口。  
  `services/mihomo/dns_resolver.go`: entry point for proxy host resolution, custom DNS, and proxy-assisted DoH.
- `services/mihomo/host_resolver.go`：将数据库中的 Host 映射同步到 mihomo 解析器。  
  `services/mihomo/host_resolver.go`: synchronizes DB Host mappings into the mihomo resolver.
- `services/scheduler/speedtest_task.go`：测速系统与 mihomo 集成的主入口。  
  `services/scheduler/speedtest_task.go`: main integration point between the speedtest system and mihomo.
- `utils/proxy_client.go`：共享代理 HTTP 客户端封装，许多代理外呼能力通过这里走 mihomo 适配器。  
  `utils/proxy_client.go`: shared proxy HTTP client wrapper; many proxied outbound flows route through mihomo adapters here.
- `node/sub.go`、`node/usage.go`：订阅下载、机场流量获取等代理请求路径。  
  `node/sub.go` and `node/usage.go`: proxied request paths for subscription downloads and airport usage retrieval.
- `main.go`：启动时注入 mihomo 相关适配与 Host 同步逻辑。  
  `main.go`: startup wiring for mihomo-related adapters and Host synchronization.

### 改动建议 / Change guidance

- 如果你修改**测速**逻辑，请同时检查 `services/scheduler/speedtest_task.go` 和 `services/mihomo/mihomo.go`。  
  If you change **speedtest** logic, inspect both `services/scheduler/speedtest_task.go` and `services/mihomo/mihomo.go`.
- 如果你修改**DNS / Host** 行为，请同时检查 `services/mihomo/dns_resolver.go`、`services/mihomo/host_resolver.go`、`api/host.go` 与 `models/host.go`。  
  If you change **DNS / Host** behavior, inspect `services/mihomo/dns_resolver.go`, `services/mihomo/host_resolver.go`, `api/host.go`, and `models/host.go` together.
- 如果你修改**代理下载 / 代理外呼**行为，请同时检查 `utils/proxy_client.go`、`node/sub.go`、`node/usage.go` 以及相关配置说明。  
  If you change **proxied downloads / outbound proxy** behavior, inspect `utils/proxy_client.go`, `node/sub.go`, `node/usage.go`, and the related configuration docs together.
- 如果你修改**启动 / 数据迁移 / Host 持久化**逻辑，请确认 `main.go` 与迁移后的 mihomo Host 同步仍然成立。  
  If you change **startup / migration / Host persistence** behavior, verify that `main.go` and post-migration mihomo Host synchronization still hold.
- 若某个特性只是“兼容 mihomo 生态”而不是“直接调用 mihomo Go 包”，在文档中要写清楚，不要把生态兼容误写成直接运行时依赖。  
  If a feature is only “mihomo ecosystem compatible” rather than “directly importing the mihomo Go package,” document that distinction clearly instead of overstating direct runtime dependency.

## 11. 贡献者工作规范 / Working norms for contributors

### 做最小且边界清晰的改动 / Make minimal, boundary-respecting changes

- 优先在正确层修复问题，而不是在相邻层加补丁。  
  Prefer fixing the correct layer instead of adding workaround logic nearby.
- 但“在正确层修复”不等于“只改一层就结束”；凡是改动会影响其他层的输入、输出、展示、文案、配置或操作路径，必须把受影响的层一并同步到位。  
  But “fixing the correct layer” does not mean “changing only one layer and stopping”; whenever a change affects another layer’s inputs, outputs, presentation, copy, configuration, or operational flow, update all impacted layers together.
- 不要把业务逻辑塞进 `routers/`。  
  Do not move business logic into `routers/`.
- 不要把持久化细节放进前端 UI 代码。  
  Do not put persistence details into UI code.
- 调度/定时任务相关逻辑应保留在 `services/scheduler/`。  
  Keep scheduler-specific logic in `services/scheduler/`.

### 遵循现有模式 / Follow existing patterns

- 后端遵循标准 Go 风格和现有包结构。  
  Backend work should follow standard Go style and the existing package structure.
- 前端遵循当前的 Vite + MUI + ESLint/Prettier 体系。  
  Frontend work should follow the current Vite + MUI + ESLint/Prettier setup.
- 引入新模式前，先匹配周围文件的命名与组织方式。  
  Match surrounding naming and file organization before introducing new patterns.

### 注释规范 / Commenting standards

合理、规范、可维护的注释是优秀开发习惯的一部分，但注释应解释意图、约束和边界，而不是重复代码本身。  
Clear, maintainable comments are part of good engineering practice, but comments should explain intent, constraints, and boundaries rather than restating the code.

- 新增或修改关键业务逻辑、跨层契约、复杂条件分支、调度任务、迁移、并发流程、缓存策略、mihomo 集成、认证 / 安全逻辑、配置优先级和非显而易见的算法时，必须补充必要注释。  
  Add necessary comments when adding or changing key business logic, cross-layer contracts, complex branches, scheduled tasks, migrations, concurrency flows, caching strategies, mihomo integrations, auth/security logic, configuration precedence, or non-obvious algorithms.
- 注释以中文为主；如涉及上游库、协议字段、标准术语、公开 API 名称或必须与英文文档一致的概念，可保留英文术语或在中文后补充英文。  
  Comments should primarily use Chinese; keep English terms when they refer to upstream libraries, protocol fields, standard terminology, public API names, or concepts that must match English documentation.
- 修改代码时要同步更新附近已有注释，避免注释描述旧行为、旧字段、旧限制或旧流程。  
  When changing code, update nearby comments so they do not describe old behavior, fields, constraints, or flows.
- 不要为了“有注释”而添加噪音注释；简单 getter、直观赋值、普通 JSX 结构、显而易见的错误返回和自解释变量名通常不需要注释。  
  Do not add noisy comments just to increase comment count; simple getters, obvious assignments, ordinary JSX structure, straightforward error returns, and self-explanatory variable names usually do not need comments.
- 注释不得掩盖设计问题；如果逻辑复杂到必须用大量注释解释，优先考虑拆分函数、提取语义变量或重命名，使代码本身更清晰。  
  Comments must not hide design problems; if logic needs extensive explanation, first consider splitting functions, extracting semantic variables, or renaming so the code is clearer.
- 对 TODO、FIXME、临时兼容逻辑或已知限制，必须写清楚原因、触发条件和后续处理方向；不要留下无上下文的占位注释。  
  For TODOs, FIXMEs, temporary compatibility logic, or known limitations, document the reason, trigger condition, and follow-up direction; do not leave context-free placeholder comments.

#### Go 注释要求 / Go comment requirements

- 导出的包、类型、函数、方法、接口、常量和变量应遵循 Go 文档注释习惯：注释以被注释标识符开头，并用完整句子说明职责、输入输出约束或副作用。  
  Exported packages, types, functions, methods, interfaces, constants, and variables should follow Go documentation conventions: start the comment with the identifier and describe responsibility, input/output constraints, or side effects in complete sentences.
- 包注释应放在 `doc.go` 或包内合适文件顶部，用于说明包职责和边界；不要在包注释中堆砌实现细节。  
  Package comments should live in `doc.go` or an appropriate package file header and explain package responsibility and boundaries without dumping implementation details.
- 对非导出但业务关键的函数、结构体字段、状态枚举、迁移步骤和 goroutine 生命周期，也应添加简短中文注释说明为什么存在、何时调用、有什么副作用。  
  Non-exported but business-critical functions, struct fields, state enums, migration steps, and goroutine lifecycles should also have concise Chinese comments explaining why they exist, when they run, and what side effects they have.
- 错误处理注释应说明业务语义或恢复策略，不要写成 `// 返回错误` 这类重复代码的注释。  
  Error-handling comments should explain business semantics or recovery strategy, not restate code such as `// return error`.
- 注释内容应能通过 `gofmt` 保持整洁；不要使用大段 ASCII 图、过长行或和 Go doc 风格冲突的装饰性注释。  
  Comments should remain tidy under `gofmt`; avoid large ASCII diagrams, overly long lines, or decorative comments that conflict with Go doc style.

#### 前端注释要求 / Frontend comment requirements

- React 组件、Hook、API 请求封装、复杂 `useMemo` / `useEffect`、权限判断、响应式布局分支、主题适配 helper 和跨组件数据流，应在非显而易见时补充中文注释说明设计意图。  
  React components, hooks, API wrappers, complex `useMemo` / `useEffect` logic, permission checks, responsive layout branches, theme helpers, and cross-component data flow should have Chinese comments when their intent is not obvious.
- JSX 中避免堆叠注释；优先把复杂条件提取为命名变量或小组件，再在变量或 helper 附近说明原因。  
  Avoid piling comments inside JSX; prefer extracting complex conditions into named variables or small components, then document the reason near the variable or helper.
- 主题和样式注释应说明复用的语义层级、light / dark 差异或与现有样板的关系，不要只描述颜色本身。  
  Theme and style comments should explain semantic layering, light/dark differences, or relationship to existing patterns rather than merely describing colors.
- 前端请求层注释必须与后端接口语义保持一致；如果接口字段、权限或错误结构变化，要同步更新请求层、页面层和相关注释。  
  Comments in the frontend API layer must stay aligned with backend API semantics; if fields, permissions, or error shapes change, update request-layer comments, view-layer comments, and related code together.
- 临时 UI 限制、兼容浏览器行为、移动端特殊处理和可访问性折中必须写明原因，避免后续维护者误删。  
  Temporary UI constraints, browser compatibility behavior, mobile-specific handling, and accessibility tradeoffs must explain their reason so future maintainers do not remove them accidentally.

### 测试规范 / Testing standards

规范、可维护的测试用例和清晰的测试文件组织是保证代码质量的基本要求。测试应验证真实业务行为和边界条件，而不是只覆盖实现细节或为了提高覆盖率而堆砌脆弱用例。  
Clear, maintainable tests and predictable test file organization are basic requirements for code quality. Tests should verify real behavior and boundaries rather than implementation details or fragile coverage-only cases.

- 新增或修改后端关键业务逻辑、接口契约、权限判断、配置语义、迁移、调度任务、mihomo 集成、协议解析和数据转换时，应同步补充或更新对应 Go 测试。  
  Add or update Go tests when adding or changing backend key business logic, API contracts, permission checks, configuration semantics, migrations, scheduled jobs, mihomo integrations, protocol parsing, or data transformations.
- 测试名称应描述“场景 + 期望结果”，避免只写 `TestSuccess`、`TestError`、`should work` 这类无法表达业务意图的名称。  
  Test names should describe the scenario and expected outcome; avoid names like `TestSuccess`, `TestError`, or `should work` that do not express business intent.
- 测试应覆盖正常路径、边界条件、错误路径和权限 / 配置差异；不要只验证最容易通过的 happy path。  
  Tests should cover happy paths, boundaries, error paths, and permission/configuration differences; do not test only the easiest happy path.
- 测试数据应尽量局部、可读、最小化，并通过 helper / fixture 表达业务含义；不要复制大量无关字段或依赖生产运行时数据。  
  Test data should be local, readable, minimal, and expressed through helpers or fixtures with business meaning; do not copy large irrelevant payloads or depend on production runtime data.
- 测试之间必须相互隔离，不能依赖执行顺序、共享可变全局状态、真实外部网络、用户本机环境或 `db/`、`logs/`、`cache/`、`out/` 中的运行时状态。  
  Tests must be isolated from each other and must not depend on execution order, shared mutable global state, real external networks, the user's local environment, or runtime state under `db/`, `logs/`, `cache/`, or `out/`.
- 断言应明确验证关键输出、状态变化、副作用和错误语义；不要只断言“没有报错”，除非该函数的唯一契约就是不产生错误。  
  Assertions should explicitly verify important outputs, state changes, side effects, and error semantics; do not only assert “no error” unless that is the function's only contract.
- 测试失败信息应能帮助定位问题，必要时包含输入场景、期望值和实际值；不要留下难以判断原因的裸断言。  
  Failure messages should help diagnose the issue, including scenario, expected value, and actual value when useful; avoid bare assertions that make failures hard to interpret.
- 修复缺陷时，优先先写能复现问题的回归测试，再修复实现；不得删除、跳过或弱化失败测试来掩盖问题。  
  When fixing a bug, prefer writing a regression test that reproduces the issue before changing implementation; never delete, skip, or weaken failing tests to hide problems.
- 如果当前模块已有测试风格、fixture、mock 或 helper，应优先复用并保持一致；只有在现有模式明显不足时才新增更清晰的小型测试 helper。  
  If the module already has test style, fixtures, mocks, or helpers, reuse and match them first; introduce a clearer small test helper only when existing patterns are insufficient.
- 不要为未接入的测试框架、脚本或 CI 流程编写假设性测试说明；测试命令必须来自仓库事实来源或本次同步新增的真实脚本。  
  Do not document hypothetical test frameworks, scripts, or CI flows that are not wired into the repo; test commands must come from repository sources of truth or real scripts added in the same work.

#### Go 测试要求 / Go test requirements

- Go 测试文件必须使用 `_test.go` 后缀，并与被测文件或被测包放在同一包目录下；例如 `a.go` 的测试文件应命名为 `a_test.go`。  
  Go test files must use the `_test.go` suffix and live in the same package directory as the code under test; for example, tests for `a.go` should be named `a_test.go`.
- Go 测试函数命名应遵循标准格式：`TestXxx`、`BenchmarkXxx`、`FuzzXxx`、`ExampleXxx`；表驱动子用例应通过 `t.Run` 命名清楚具体场景。  
  Go test functions should follow standard names: `TestXxx`, `BenchmarkXxx`, `FuzzXxx`, and `ExampleXxx`; table-driven subtests should use `t.Run` with clear scenario names.
- 优先使用表驱动测试覆盖同一函数的多种输入、边界和错误情况；每个 case 应包含清晰的名称、输入、期望输出和期望错误语义。  
  Prefer table-driven tests for multiple inputs, boundaries, and error cases of the same function; each case should include a clear name, input, expected output, and expected error semantics.
- 测试 helper 应调用 `t.Helper()`，临时目录和临时文件应优先使用 `t.TempDir()`，清理动作应通过 `t.Cleanup()` 注册。  
  Test helpers should call `t.Helper()`, temporary directories and files should prefer `t.TempDir()`, and cleanup should be registered with `t.Cleanup()`.
- 涉及时间、随机数、网络、数据库、文件系统或 goroutine 的测试，应显式控制依赖或使用可替换接口，避免不稳定、慢速或依赖真实环境的 flaky test。  
  Tests involving time, randomness, network, databases, filesystems, or goroutines should explicitly control dependencies or use replaceable interfaces to avoid flaky, slow, or environment-dependent tests.
- HTTP handler 测试应优先使用 `httptest` 和明确的请求 / 响应断言；数据库相关测试应使用隔离的测试库、事务或临时存储，不能读写生产默认数据目录。  
  HTTP handler tests should prefer `httptest` with explicit request/response assertions; database tests should use isolated test databases, transactions, or temporary storage and must not read or write production default data directories.
- 并发测试不要依赖 `time.Sleep` 猜测时序；优先使用 channel、context、WaitGroup、可控 clock 或明确的同步点。  
  Concurrency tests should not rely on `time.Sleep` to guess timing; prefer channels, contexts, WaitGroups, controllable clocks, or explicit synchronization points.
- 如果使用外部包测试形式（`package xxx_test`），应只验证公开 API；如果需要访问包内细节，应使用同包测试并保持测试关注行为而不是私有实现。  
  If using external package tests (`package xxx_test`), test only public APIs; if package internals must be accessed, use same-package tests while keeping focus on behavior rather than private implementation.

#### 前端测试边界 / Frontend test boundaries

- 本仓库不默认要求为前端改动新增测试；前端改动的常规质量保障以 lint、build、人工交互验证和 light / dark / 响应式检查为主。  
  This repository does not require frontend tests by default; frontend quality is primarily guarded by lint, build, manual interaction checks, and light/dark/responsive verification.
- 只有在用户明确要求、仓库后续接入前端测试框架，或当前模块已经存在前端测试时，才补充或更新前端测试。  
  Add or update frontend tests only when explicitly requested, when a frontend test framework is later wired into the repo, or when the touched module already has frontend tests.
- 如果确实编写前端测试，文件应与被测文件保持可追溯的命名关系，优先使用 `*.test.jsx`、`*.test.js`、`*.spec.jsx` 或 `*.spec.js`，并靠近被测组件 / hook / helper 放置，除非仓库已有集中测试目录约定。  
  If frontend tests are written, test files should keep traceable names to the code under test, preferably `*.test.jsx`, `*.test.js`, `*.spec.jsx`, or `*.spec.js`, placed near the component/hook/helper unless the repo already has a centralized test directory convention.
- 如果确实编写前端测试，应从用户可观察行为出发，验证文本、角色、状态变化、交互结果、错误提示和 loading / empty / disabled 等状态，而不是断言内部 state 或 MUI 生成的脆弱 class 名。  
  If frontend tests are written, they should start from user-observable behavior, verifying text, roles, state changes, interaction results, error messages, and loading/empty/disabled states rather than internal state or fragile MUI-generated class names.
- 不要为了满足后端测试规范而发明前端测试脚本、测试框架或 CI 流程；前端测试命令必须来自 `webs/package.json`、CI 或本次真实接入的脚本。  
  Do not invent frontend test scripts, frameworks, or CI flows to satisfy backend testing standards; frontend test commands must come from `webs/package.json`, CI, or scripts genuinely added in the same work.

### 避免陈旧假设 / Avoid stale assumptions

- 不要向文档写入仓库中不存在的命令。  
  Do not add commands to docs unless they actually exist in the repo or CI.
- 前端测试命令若不存在，就不要假设存在。  
  Do not reference frontend tests unless a real test setup exists.
- 不要把通用脚手架默认行为当成这里的事实。  
  Do not assume generic template behavior applies here.

## 12. 验证建议 / Validation guidance

对于前端改动：  
For frontend changes:

- 必须运行 `yarn run lint`。
  Run `yarn run lint`.
- 涉及构建产物、资源路径、路由、base-path 或生产集成时，必须同时运行 `yarn run build`。
  Also run `yarn run build` when build output, asset paths, routing, base-path, or production integration are affected.

对于后端改动：  
For backend changes:

- Go 文件必须通过 `gofmt` 格式化；本地开发可对改动文件运行 `gofmt -w <changed-go-files>`。
  Go files must be formatted with `gofmt`; locally, run `gofmt -w <changed-go-files>` for changed Go files.
- 必须运行 `golangci-lint run` 做基础 Go lint / 静态检查。
  Run `golangci-lint run` for baseline Go lint/static checks.
- 新增或修改后端逻辑时必须运行相关 `go test`；发布前 GitHub 工作流运行全仓 `go test ./...`。
  Run relevant `go test` for backend logic changes; the GitHub release workflow runs repository-wide `go test ./...` before publishing.
- 至少保证相关包可以成功构建。  
  At minimum, ensure the touched package builds cleanly.
- 如果改了启动、构建或配置行为，要跑真实构建命令验证。  
  If you modify startup, build, or config behavior, verify with an actual build command.

注意：  
Notes:

- GitHub 发布构建会在构建二进制前运行 `golangci-lint` 和 `go test ./...`。
  The GitHub release build runs `golangci-lint` and `go test ./...` before building binaries.
- `webs/package.json` 中没有权威的前端 `test` 或 `typecheck` 脚本。  
  No authoritative frontend `test` or `typecheck` script was found in `webs/package.json`.

因此，写文档或自动化时不要发明不存在的前端测试 / typecheck 流程。
So when documenting or automating verification, avoid inventing nonexistent frontend test/typecheck workflows.

## 13. 改行为前优先查看的区域 / High-value areas to inspect before changing behavior

以下文件/目录通常是高价值起点：  
Use these files/directories as likely starting points:

- 认证 / MFA：`api/auth.go`、`api/auth_mfa.go`、`middlewares/`  
  Auth / MFA: `api/auth.go`, `api/auth_mfa.go`, `middlewares/`
- 订阅生成：`api/clients.go`  
  Subscription generation: `api/clients.go`
- mihomo 集成核心：`services/mihomo/`  
  Mihomo integration core: `services/mihomo/`
- 定时任务：`services/scheduler/`  
  Scheduled tasks: `services/scheduler/`
- 标签逻辑：`services/tag_service.go`  
  Tag logic: `services/tag_service.go`
- 协议解析：`node/protocol/`  
  Protocol parsing: `node/protocol/`
- 代理 HTTP 客户端：`utils/proxy_client.go`  
  Proxy HTTP client: `utils/proxy_client.go`
- Host 管理：`models/host.go`  
  Host management: `models/host.go`
- 数据迁移：`models/db_migrate.go`  
  DB migration: `models/db_migrate.go`
- 前端页面：`webs/src/views/`  
  Frontend feature pages: `webs/src/views/`
- 前端路由：`webs/src/routes/`  
  Frontend routing: `webs/src/routes/`
- 前端 API 层：`webs/src/api/`  
  Frontend API layer: `webs/src/api/`
- 前端主题与视觉语义：`webs/src/themes/`、`webs/src/utils/colorUtils.js`  
  Frontend theme and visual semantics: `webs/src/themes/`, `webs/src/utils/colorUtils.js`
- 节点预览相关主题链路：`webs/src/views/subscriptions/component/NodePreviewDialog.jsx`、`NodePreviewCard.jsx`、`NodePreviewDetailsPanel.jsx`  
  Node preview theme cluster: `webs/src/views/subscriptions/component/NodePreviewDialog.jsx`, `NodePreviewCard.jsx`, `NodePreviewDetailsPanel.jsx`
- 链式代理相关主题链路：`webs/src/views/subscriptions/component/ChainProxyDialog.jsx`、`ChainPreviewDialog.jsx`、`ChainFlowBuilder.jsx`、`ChainCanvasView.jsx`、`MobileChainBuilder.jsx`、`ConditionBuilder.jsx`  
  Chain proxy theme cluster: `webs/src/views/subscriptions/component/ChainProxyDialog.jsx`, `ChainPreviewDialog.jsx`, `ChainFlowBuilder.jsx`, `ChainCanvasView.jsx`, `MobileChainBuilder.jsx`, `ConditionBuilder.jsx`
- 机场管理相关主题链路：`webs/src/views/airports/index.jsx`、`AirportMobileList.jsx`、`AirportFormDialog.jsx`、`AirportBatchEditDialog.jsx`、`AirportDeduplicationConfig.jsx`、`webs/src/components/CronExpressionGenerator.jsx`  
  Airport management theme cluster: `webs/src/views/airports/index.jsx`, `AirportMobileList.jsx`, `AirportFormDialog.jsx`, `AirportBatchEditDialog.jsx`, `AirportDeduplicationConfig.jsx`, and `webs/src/components/CronExpressionGenerator.jsx`

## 14. 文档同步要求 / Documentation expectations

如果你改动了以下内容，应在同一工作中同步更新文档：  
If you change any of the following, update docs in the same work when appropriate:

- 本地开发命令 / Local development commands
- 环境变量或配置优先级 / Environment variables or config precedence
- 部署或构建流程 / Deployment or build flow
- mihomo 相关能力、代理链路、DNS / Host 行为 / Mihomo-related capabilities, proxy flows, or DNS / Host behavior
- 迁移行为 / Migration behavior
- MFA / 认证 / 安全敏感操作 / MFA, auth, or security-sensitive operations
- `docs/features/` 中描述的用户功能流程 / User-facing feature workflows described under `docs/features/`
- `/api/v1/*` 或 `/c/*` 的路由、请求字段、内容类型、鉴权、响应结构或业务码 / `/api/v1/*` or `/c/*` routes, request fields, content type, auth, response shapes, or business codes

常见相关文档：  
Relevant docs include:

- `README.md`
- `README.zh-CN.md`
- `docs/development.md`
- `docs/development.zh-CN.md`
- `docs/configuration.md`
- `docs/configuration.zh-CN.md`
- `docs/installation.md`
- `docs/installation.zh-CN.md`
- `docs/features/*`
- `docs/features/*.zh-CN.md`
- `skill-sublinkpro/SKILL.md`、`skill-sublinkpro/reference/api.md`、`skill-sublinkpro/reference/deploy.md`、`skill-sublinkpro/README.md`、`skill-sublinkpro/README.zh-CN.md`

- 文档同步不是可选收尾步骤，而是变更完成条件的一部分。只要代码改动改变了行为、接口、字段、页面文案、用户流程、命令、配置、部署、迁移或安全语义，就必须在同一工作中同步更新相应文档。  
  Documentation sync is not an optional finishing step; it is part of the definition of done. If a code change alters behavior, APIs, fields, page copy, user flows, commands, configuration, deployment, migration, or security semantics, the relevant documentation must be updated in the same piece of work.
- 文档现在同时维护英文 canonical 版本和简体中文 `*.zh-CN.md` 版本；凡是需要更新 `README.md`、`docs/*.md` 或 `docs/features/*.md` 的变更，必须同步更新对应的 `README.zh-CN.md`、`docs/*.zh-CN.md` 或 `docs/features/*.zh-CN.md`，并保持语言切换和相对链接一致。
  Documentation is now maintained in both canonical English files and Simplified Chinese `*.zh-CN.md` files; any change that updates `README.md`, `docs/*.md`, or `docs/features/*.md` must also update the matching `README.zh-CN.md`, `docs/*.zh-CN.md`, or `docs/features/*.zh-CN.md`, keeping language switches and relative links consistent.
- 不允许出现“代码已经修改，但 `README.md`、`docs/` 或功能说明仍保留旧语义”的交付状态；如果确认无需更新文档，也应在变更说明中明确写明原因。  
  Do not leave the project in a state where code has changed but `README.md`, `docs/`, or feature documentation still describes the old behavior; if no doc update is needed, explicitly state why.
- 不允许只更新一种语言的文档后交付；如果某个变更确实只影响其中一种语言（例如修正翻译措辞），交付说明必须明确写出另一种语言已检查且无需修改。
  Do not ship documentation updates in only one language; if a change truly affects only one language, such as wording in a translation, the delivery summary must state that the other language was checked and does not need changes.
- `skill-sublinkpro/` 是面向用户的 AI agent 技能（采用可移植的 `SKILL.md` 格式，不绑定特定厂商，任何兼容的 AI agent / 客户端均可使用），通过 `X-API-Key` 调用本系统 REST API，是 API 契约在前端之外的第三个消费者。凡是改动 `/api/v1/*` 或 `/c/*` 的路由、请求字段、内容类型（form vs JSON）、鉴权方式、响应结构或业务码，都必须在同一工作中同步更新 `skill-sublinkpro/reference/api.md`（端点目录）以及受影响的 `SKILL.md` 工作流说明；新增端点或功能时，也应评估是否需要在 skill 中补充对应用法。skill README 改动需同步英文与 `*.zh-CN.md` 两个版本。
  `skill-sublinkpro/` is a user-facing AI agent skill (portable, vendor-neutral `SKILL.md` format usable by any compatible AI agent/client) that drives this system's REST API via `X-API-Key`; it is the third consumer of the API contract beyond the frontend. Whenever you change `/api/v1/*` or `/c/*` routes, request fields, content type (form vs JSON), auth, response shapes, or business codes, you must update `skill-sublinkpro/reference/api.md` (the endpoint catalog) and any affected `SKILL.md` workflow guidance in the same work; when adding endpoints or features, evaluate whether the skill needs matching usage. Skill README changes must update both the English and `*.zh-CN.md` versions.
- 除 API 契约外，skill 还内置了**部署/安装与配置引导**（`skill-sublinkpro/reference/deploy.md`）。凡是改动部署方式或配置语义，都必须在同一工作中同步更新该文件，使其与真实行为一致。具体包括：`install.sh` / `uninstall.sh` 的流程或命令、`Dockerfile`、`docker-compose.example.yml`、`config.example.yaml`、镜像名 `zerodeng/sublink-pro`、默认端口、数据目录（`./db ./template ./logs`）、默认账号、以及任何 `SUBLINK_*` 环境变量或 `config.yaml` 字段的新增/改名/删除/默认值变化。deploy.md 中的命令与变量必须能逐条对应到这些事实来源，不得保留已过时的安装步骤或配置项。
  Beyond the API contract, the skill also embeds **deployment/install and configuration guidance** (`skill-sublinkpro/reference/deploy.md`). Whenever you change a deployment method or configuration semantics, you must update that file in the same work so it matches real behavior. This includes: the flow or commands in `install.sh` / `uninstall.sh`, `Dockerfile`, `docker-compose.example.yml`, `config.example.yaml`, the image name `zerodeng/sublink-pro`, the default port, the data directories (`./db ./template ./logs`), the default credentials, and any addition/rename/removal/default change of a `SUBLINK_*` environment variable or `config.yaml` field. Every command and variable in deploy.md must trace back to these sources of truth; do not leave outdated install steps or config options behind.
- `skill-sublinkpro/` 整个目录会通过 `//go:embed skill-sublinkpro`（声明于 `main.go`）打包进 Go 二进制，并由登录受限的 `/api/v1/skill/download`（`routers/skill.go` + `api/skill.go`，运行时即时打 zip）下发，Web UI 顶栏的下载图标（`webs/src/layout/MainLayout/Header/SkillDownloadSection`）触发下载。因此：① 该目录必须始终保持可被 `//go:embed` 嵌入（不要引入嵌入无法处理的内容，注意默认 `//go:embed` 会跳过点文件/下划线开头文件）；② 下载端点必须保持登录受限（`middlewares.AuthToken`），不要改成匿名开放；③ 顶栏图标的提示文案是用户可见文本，新增/修改须同步 `en-US.json` 与 `zh-CN.json` 两个 locale。skill 内容更新后无需改构建流程——`go build` 会自动嵌入最新源文件，zip 在运行时生成，始终与嵌入内容一致。
  The entire `skill-sublinkpro/` directory is bundled into the Go binary via `//go:embed skill-sublinkpro` (declared in `main.go`) and served by the login-gated `/api/v1/skill/download` endpoint (`routers/skill.go` + `api/skill.go`, zipped on-the-fly at runtime), triggered by the download icon in the web UI top bar (`webs/src/layout/MainLayout/Header/SkillDownloadSection`). Therefore: (1) the directory must remain embeddable by `//go:embed` (don't add content embed can't handle; note plain `//go:embed` skips dotfiles and underscore-prefixed files); (2) the download endpoint must stay login-gated (`middlewares.AuthToken`) — do not make it anonymous; (3) the top-bar icon's tooltip is user-visible text, so additions/changes must update both `en-US.json` and `zh-CN.json`. No build change is needed when skill content changes — `go build` re-embeds the latest source files automatically and the zip is generated at runtime, always matching the embedded content.
- `skill-sublinkpro/reference/docs.md` 维护了项目文档的主题→路径映射表，用于将 skill 变成"在线使用手册"——当用户询问某个功能如何工作、如何配置、如何排查问题时，AI agent 会从 GitHub fetch 对应文档后回答，保证答案始终与最新发布的文档一致。因此：当项目文档（`docs/*.md`、`README*.md`）有新增、重命名、路径变化或主题变化时，必须同步更新 `docs.md` 中的映射表，确保主题描述与 doc path 正确对应；当文档无法解决用户问题时，skill 会引导用户去 GitHub issues 或 Telegram 联系作者，这部分引导逻辑位于 `SKILL.md` "Documentation & help" 一节，链接和措辞变化时需同步修改。
  `skill-sublinkpro/reference/docs.md` maintains a topic-to-path mapping of the project's documentation, turning the skill into an "online manual"—when users ask how a feature works, how to configure it, or how to troubleshoot, the AI agent fetches the relevant doc from GitHub and answers, ensuring responses always match the latest published docs. Therefore: when project documentation (`docs/*.md`, `README*.md`) is added, renamed, moved, or changes topic coverage, you must synchronously update the mapping table in `docs.md` to keep topic descriptions and doc paths correct; when docs cannot solve a user's problem, the skill guides them to GitHub issues or Telegram to contact the author—this guidance is in the "Documentation & help" section of `SKILL.md`, and link or wording changes must be updated there as well.

## 15. 安全提示 / Safety notes

- 默认管理员账号是 `admin / 123456`；不要把它重新表述成“安全默认值”，应提醒用户尽快修改。  
  Default admin credentials are `admin / 123456`; do not present them as a safe default, and remind users to change them.
- 根据项目文档，从 SQLite 迁移到 MySQL / PostgreSQL 后需要手动重启实例。  
  According to project docs, migration from SQLite to MySQL/PostgreSQL requires a manual restart afterward.
- 不要假设与上游原项目数据库兼容；README 已明确说明不兼容。  
  Do not assume compatibility with upstream project databases; the README explicitly warns they are not compatible.
- 注意 Docker 挂载的运行时目录：`/app/db`、`/app/template`、`/app/logs`。  
  Be careful with runtime directories mounted by Docker deployments: `/app/db`, `/app/template`, and `/app/logs`.

## 16. 分支与提交约定 / Branch and commit conventions

`docs/development.md` 中记录了以下约定：  
`docs/development.md` documents these conventions:

- `main` 为稳定分支。  
  `main` is the stable branch.
- `dev` 为开发分支。  
  `dev` is the development branch.
- 提交信息优先使用语义化前缀，如 `feat`、`fix`、`docs`、`refactor`。  
  Semantic commit prefixes are preferred, such as `feat`, `fix`, `docs`, and `refactor`.

## 17. 实用变更路径 / Practical change recipes

### 添加后端功能 / Add a backend feature

通常会涉及：  
Usually touch:

1. `models/`：需要持久化改动时。  
   `models/` if persistence changes are needed.
2. `services/`：业务逻辑。  
   `services/` for business logic.
3. `api/`：handler 层。  
   `api/` for handler logic.
4. `routers/`：路由暴露。  
   `routers/` for route exposure.
5. `webs/src/api/` + `webs/src/views/`：如果需要前端支持。  
   `webs/src/api/` + `webs/src/views/` if UI support is needed.

### 添加或修改定时任务 / Add or change a scheduled task

优先阅读：  
Start by reading:

- `services/scheduler/manager.go`
- `services/scheduler/job_ids.go`
- 同类现有任务文件  
  similar existing task files

遵循 `docs/development.md` 中已有的系统任务负数 ID 约定。  
Follow the existing negative-ID convention for system jobs documented in `docs/development.md`.

### 修改 mihomo 相关能力 / Change mihomo-related behavior

优先阅读：  
Start by reading:

- `services/mihomo/mihomo.go`
- `services/mihomo/dns_resolver.go`
- `services/mihomo/host_resolver.go`
- `services/scheduler/speedtest_task.go`
- `utils/proxy_client.go`
- `node/sub.go`
- `node/usage.go`
- `docs/features/speedtest.md`
- `docs/features/host.md`
- `docs/features/chain-proxy.md`

这类改动通常会同时影响测速、代理请求路径、Host 解析、DNS 行为和文档描述；不要只改其中一层。  
These changes often affect speed testing, proxied request flows, Host resolution, DNS behavior, and documentation at the same time; do not change only one layer in isolation.

### 修改部署或构建行为 / Change deployment or build behavior

修改前请一起检查：  
Inspect these together before editing:

- `Dockerfile`
- `embed_prod.go`
- `.github/workflows/build-release.yml`
- `main.go`

这些文件共同组成生产构建链路，只改其中一个通常会破坏发布流程。  
These files form one production pipeline; changing only one of them is likely to break releases.
