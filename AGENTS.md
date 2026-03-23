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

## 2. 事实来源 / Source of truth

当多份说明不一致时，按下面顺序信任这些文件：  
When instructions conflict, trust these files first:

1. `webs/package.json`：前端命令与依赖事实来源。  
   `webs/package.json`: source of truth for frontend scripts and dependencies.
2. `docs/development.md`：本地开发流程与目录说明。  
   `docs/development.md`: local development workflow and project structure.
3. `docs/configuration.md`：配置优先级、环境变量与运行语义。  
   `docs/configuration.md`: config precedence, environment variables, and runtime behavior.
4. `.github/workflows/build-release.yml`：CI / release 构建预期。  
   `.github/workflows/build-release.yml`: CI and release build expectations.
5. `Dockerfile`：生产构建路径的权威实现。  
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

### 运行时路径 / Operational and runtime paths

这些目录可能包含运行时状态，应谨慎处理：  
These directories may contain runtime state and should be handled carefully:

- `db/`
- `logs/`
- `cache/`
- `out/`

正常功能开发时，不要随意编辑、删除或格式化这些运行时数据。  
Do not casually edit, delete, or reformat runtime data during normal feature work.

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
yarn run build
yarn run lint
yarn run lint:fix
yarn run prettier
```

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
- `cd webs && yarn run build`
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

## 9. mihomo 核心集成说明 / Mihomo integration guidance

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

## 10. 贡献者工作规范 / Working norms for contributors

### 做最小且边界清晰的改动 / Make minimal, boundary-respecting changes

- 优先在正确层修复问题，而不是在相邻层加补丁。  
  Prefer fixing the correct layer instead of adding workaround logic nearby.
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

### 避免陈旧假设 / Avoid stale assumptions

- 不要向文档写入仓库中不存在的命令。  
  Do not add commands to docs unless they actually exist in the repo or CI.
- 前端测试命令若不存在，就不要假设存在。  
  Do not reference frontend tests unless a real test setup exists.
- 不要把通用脚手架默认行为当成这里的事实。  
  Do not assume generic template behavior applies here.

## 11. 验证建议 / Validation guidance

对于前端改动：  
For frontend changes:

- 至少运行与改动相关的 `yarn` 命令。  
  At minimum, run relevant `yarn` commands in `webs/`.
- 通常应运行 `yarn run lint` 和 `yarn run build`。  
  Usually run `yarn run lint` and `yarn run build`.

对于后端改动：  
For backend changes:

- 至少保证相关包可以成功构建。  
  At minimum, ensure the touched package builds cleanly.
- 如果改动包内已有测试，优先运行相关 `go test`。  
  Use targeted `go test` where tests already exist for touched packages.
- 如果改了启动、构建或配置行为，要跑真实构建命令验证。  
  If you modify startup, build, or config behavior, verify with an actual build command.

注意：  
Notes:

- 仓库里有 Go 测试文件，但 docs / CI 中没有发现单一权威的根级测试命令。  
  The repo contains Go test files, but no single authoritative root test command was found in docs or CI.
- 没有发现权威的全仓 Go lint 命令。  
  No authoritative repo-wide Go lint command was found.
- `webs/package.json` 中没有权威的前端 `test` 或 `typecheck` 脚本。  
  No authoritative frontend `test` or `typecheck` script was found in `webs/package.json`.

因此，写文档或自动化时不要发明不存在的校验流程。  
So when documenting or automating verification, avoid inventing nonexistent commands.

## 12. 改行为前优先查看的区域 / High-value areas to inspect before changing behavior

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

## 13. 文档同步要求 / Documentation expectations

如果你改动了以下内容，应在同一工作中同步更新文档：  
If you change any of the following, update docs in the same work when appropriate:

- 本地开发命令 / Local development commands
- 环境变量或配置优先级 / Environment variables or config precedence
- 部署或构建流程 / Deployment or build flow
- mihomo 相关能力、代理链路、DNS / Host 行为 / Mihomo-related capabilities, proxy flows, or DNS / Host behavior
- 迁移行为 / Migration behavior
- MFA / 认证 / 安全敏感操作 / MFA, auth, or security-sensitive operations
- `docs/features/` 中描述的用户功能流程 / User-facing feature workflows described under `docs/features/`

常见相关文档：  
Relevant docs include:

- `README.md`
- `docs/development.md`
- `docs/configuration.md`
- `docs/installation.md`
- `docs/features/*`

## 14. 安全提示 / Safety notes

- 默认管理员账号是 `admin / 123456`；不要把它重新表述成“安全默认值”，应提醒用户尽快修改。  
  Default admin credentials are `admin / 123456`; do not present them as a safe default, and remind users to change them.
- 根据项目文档，从 SQLite 迁移到 MySQL / PostgreSQL 后需要手动重启实例。  
  According to project docs, migration from SQLite to MySQL/PostgreSQL requires a manual restart afterward.
- 不要假设与上游原项目数据库兼容；README 已明确说明不兼容。  
  Do not assume compatibility with upstream project databases; the README explicitly warns they are not compatible.
- 注意 Docker 挂载的运行时目录：`/app/db`、`/app/template`、`/app/logs`。  
  Be careful with runtime directories mounted by Docker deployments: `/app/db`, `/app/template`, and `/app/logs`.

## 15. 分支与提交约定 / Branch and commit conventions

`docs/development.md` 中记录了以下约定：  
`docs/development.md` documents these conventions:

- `main` 为稳定分支。  
  `main` is the stable branch.
- `dev` 为开发分支。  
  `dev` is the development branch.
- 提交信息优先使用语义化前缀，如 `feat`、`fix`、`docs`、`refactor`。  
  Semantic commit prefixes are preferred, such as `feat`, `fix`, `docs`, and `refactor`.

## 16. 实用变更路径 / Practical change recipes

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
