[English](development.md) | 简体中文

# 开发指南

欢迎参与 SublinkPro 的开发。本指南聚焦于：

- 如何在本地跑通前后端开发环境
- 生产构建链路实际是什么
- 哪些文件/目录是高价值入口
- 解锁检测相关扩展点在哪里

---

## 📁 项目结构

```text
sublinkPro/
├── api/                     # HTTP API / controller
├── models/                  # 数据模型、持久化、迁移
├── services/                # 业务服务与后台子系统
│   ├── scheduler/           # 定时任务与任务调度
│   ├── mihomo/              # mihomo 集成（测速、DNS、Host、代理出站等）
│   └── unlock/              # 解锁检测注册表、运行时、checker
├── routers/                 # 路由注册
├── node/                    # 订阅与协议解析/转换逻辑
├── utils/                   # 通用工具函数
├── database/                # 数据库连接与方言支持
├── cache/                   # 缓存层
├── dto/                     # DTO / 表单结构
├── webs/                    # React + Vite 前端
│   └── src/
│       ├── api/            # 前端请求边界
│       ├── views/          # 页面级功能
│       ├── components/     # 公共组件
│       ├── utils/          # 前端通用工具
│       ├── themes/         # 主题与 MUI overrides
│       └── routes/         # 路由定义
├── template/                # 模板文件
├── docs/                    # 文档
├── skill-sublinkpro/        # AI agent 技能：基于 REST API 的 AI 操作接口（可移植 SKILL.md 格式、面向用户、可分发）
├── static/                  # 生产构建时前端产物放置目录
├── main.go                  # 应用入口
├── Dockerfile               # Docker 构建
└── README.md
```

---

## 🧩 前端组件放置规则

前端组件必须放在能体现其归属的位置。不要因为某个 import 路径方便，就把组件随意放到不相关目录里。

- 跨页面复用、与具体页面无关的组件放在 `webs/src/components/`。
- 只服务于某个功能或页面的组件，放在 `webs/src/views/<feature>/component/` 或 `webs/src/views/<feature>/components/` 下，并沿用该功能已有的目录命名。
- `webs/src/views/<feature>/index.jsx` 这类页面入口应主要负责页面组合、数据加载和功能编排，不要堆放可独立拆分的复杂组件。
- 不要从无关页面导出 feature-local 组件给其他页面使用。如果另一个页面也需要它，应先移动到 `webs/src/components/` 或更合适的共享模块。
- 新增共享组件前，先检查附近功能目录、`webs/src/components/` 和 `webs/src/ui-component/` 是否已有类似模式。

示例：

- 多个页面都会用到的通用状态面板，应放在 `webs/src/components/`。
- 只服务于机场管理的弹窗，应放在 `webs/src/views/airports/component/`。
- 只服务于新的 `system-updates` 页面时，面板应放在 `webs/src/views/system-updates/components/`；只有明确需要跨页面复用时，才上移到 `webs/src/components/`。

---

## 🔧 技术栈

| 层级 | 技术 |
|:---|:---|
| 后端框架 | Go + Gin |
| ORM | GORM |
| 数据库 | SQLite（默认）/ MySQL / PostgreSQL |
| 前端框架 | React 19 + Vite |
| UI | Material UI |
| 前端包管理 | Yarn 4 |
| 调度 | robfig/cron |

---

## 💻 本地开发

### 1. 克隆项目

```bash
git clone https://github.com/ZeroDeng01/sublinkPro.git
cd sublinkPro
```

### 2. 后端开发

建议使用 **Go 1.26.4** 或更高版本，与仓库、Docker 和 CI 保持一致。

```bash
go mod download
go run main.go
```

默认后端监听 `:8000`。

### 3. 前端开发

在 `webs/` 下执行：

```bash
yarn install
yarn run start
```

Vite 默认开发端口为 `3000`，并通过 `/api` 代理后端请求。

### 4. 前端校验

在 `webs/` 下执行：

```bash
yarn run lint
yarn run build
yarn run lint:fix
yarn run prettier
```

前端开发完成后必须至少运行 `yarn run lint`；涉及构建产物、资源路径、路由、base-path 或生产集成时，还必须运行 `yarn run build`。

> [!NOTE]
> 当前仓库**没有权威的前端 `test` 或 `typecheck` 脚本**。不要在文档或自动化里发明不存在的校验流程。

### 5. 后端校验

后端开发完成后，Go 文件必须通过 `gofmt` 格式化，并运行 `golangci-lint` 静态检查与相关测试：

```bash
gofmt -w <changed-go-files>
golangci-lint run
go test ./...
```

新增或修改关键业务逻辑、接口契约、权限判断、配置语义、迁移、调度任务、mihomo 集成、协议解析或数据转换时，应同步补充或更新对应 Go 测试。GitHub 发布构建会在构建二进制前运行 `golangci-lint` 和全仓 `go test ./...`。

### 6. PR 自动检查

`.github/workflows/pr-checks.yml` 会在 PR 打开、重新打开或标记 ready for review 时自动运行，用于快速判断 PR 是否满足基础质量要求。后续修复提交不会自动重复消耗 Actions；PR 作者本人或仓库管理员确认修复完成后，可在 PR 评论 `/recheck` 手动触发新一轮检查。

- 后端：`golangci-lint`、`go test ./...`
- 前端：`yarn run lint`、`yarn run build`

每个检查 job 会写入 GitHub Step Summary，展示对应检查项是通过还是失败，方便作者和 reviewer 快速判断哪些工作已经完成、哪些还需要修复。

### 7. 普通后端构建

```bash
go build -o sublinkpro main.go
```

这适合开发环境或快速本地编译，不代表生产嵌入构建。

### 8. 生产构建（实际流程）

生产构建是两阶段；本地执行生产风格构建前，应先完成前端 lint / build 和后端格式、lint、测试校验：

```bash
# 1) 前端 lint 与构建
cd webs
yarn run lint
yarn run build

# 2) 后端格式、lint 与测试
cd ..
gofmt -w <changed-go-files>
golangci-lint run
go test ./...

# 3) 准备生产静态资源
rm -rf static && mkdir -p static
cp -R webs/dist/. static/

# 4) 构建生产后端（嵌入前端资源）
CGO_ENABLED=0 go build -tags=prod -ldflags="-s -w" -o sublinkPro
```

> [!IMPORTANT]
> 如果你修改了前端资源路径、PWA 资产、base-path、嵌入逻辑或静态资源服务方式，必须同时验证：
> 
> - `webs` 本地开发模式
> - 前端 build 产物
> - `static/` 复制后的生产嵌入构建

---

## 🧭 关键运行时约定

### 路径边界

- 前端 UI：`/` 或 `SUBLINK_WEB_BASE_PATH` 指定的路径
- API：始终在 `/api/*`
- 订阅/分享：始终在 `/c/*`

`SUBLINK_WEB_BASE_PATH` 只影响 Web UI，不影响 API 和订阅获取路径。

### 运行时目录

这些目录属于运行时状态，请谨慎处理：

- `db/`
- `logs/`
- `template/`
- `out/`

其中：

- `db/`：数据库、配置文件、GeoIP 等本地数据
- `template/`：模板文件
- `logs/`：运行日志

---

## 🔍 高价值入口文件

| 模块 | 文件 | 说明 |
|:---|:---|:---|
| 节点测速 | `services/scheduler/speedtest_task.go` | 延迟、速度、质量、解锁检测主流程 |
| 解锁检测 | `services/unlock/*.go` | Provider registry / runtime / orchestrator / checkers |
| 标签规则 | `services/tag_service.go` | 自动标签规则执行 |
| 订阅生成 | `api/clients.go` | 订阅输出与节点筛选、rename |
| 链式代理 | `api/subscription_chain.go` / `models/subscription_chain_rule.go` | 订阅链式代理规则与条件选节点 |
| Host 管理 | `models/host.go` | Host 映射、批量写入、缓存管理 |
| DNS 解析 | `services/mihomo/dns_resolver.go` | 自定义 DNS 与代理解析 |
| 数据迁移 | `models/db_migrate.go` | 数据库迁移脚本 |

---

## 🔌 新增协议接入指南

当前协议系统已经重构为**自注册 + 能力接口**模式。目标是：

> 新增一种协议时，开发者只需要在 `node/protocol/` 下增加一个协议文件，在这个文件里实现协议本身、导出能力，并完成注册。

### 协议扩展入口

建议直接参考：

- `node/protocol/protocol_demo.go`：标准示例协议
- 任意真实协议文件，如：
  - `node/protocol/vmess.go`
  - `node/protocol/ss.go`
  - `node/protocol/http.go`

### 当前协议体系结构

中心能力位于 `node/protocol/protocol_meta.go`：

- `Protocol`：核心协议规范
- `ProxyCapable`：支持 Clash Proxy 结构体转换
- `SurgeCapable`：支持 Surge 行导出
- `SupportsClient(...)`：声明协议对 Clash / mihomo / v2ray / Surge 等客户端的订阅输出兼容性
- `MustRegisterProtocol(...)`：协议注册入口

新增协议后，以下链路会自动接入，不需要再去额外补 switch：

- 协议识别（alias / scheme）
- 节点 raw 解析
- 节点 raw 字段更新
- 节点 identity 提取（名称 / host / port / address）
- 去重字段读取
- 节点链接重命名
- `LinkToProxy` 分发
- `EncodeSurge` 分发
- `EncodeProxyLink` 分发
- v2ray raw 输出兼容性过滤（通过协议文件内声明的客户端支持能力）
- 协议 UI 元数据输出

### 新增协议的推荐步骤

1. 在 `node/protocol/` 新增一个协议文件，例如：

   ```text
   node/protocol/myprotocol.go
   ```

2. 定义协议结构体。

   结构体字段会作为默认 UI 字段元数据来源，因此命名要稳定、清晰。

3. 实现协议链接的 `Decode` / `Encode`。

   至少要保证：

   - `DecodeXxxURL(string) (Xxx, error)`
   - `EncodeXxxURL(Xxx) string`

4. 如果需要从 Clash Proxy 反推链接，补 `ConvertProxyToXxx(proxy Proxy) Xxx`。

5. 如果协议支持 Clash 导出，在同一文件中实现 `buildXxxProxy(link Urls, config OutputConfig)`。

6. 如果协议支持 Surge 导出，在同一文件中实现 `buildXxxSurgeLine(link string, config OutputConfig)`。

7. 在同一个文件里 `init()` 自注册。

8. 在协议文件内声明客户端兼容性；默认规则如下：

   - `newProtocolSpec(...)` 默认支持 `ClientV2ray`。
   - `newProxyProtocolSpec(...)` 默认支持 `ClientClash`、`ClientMihomo`、`ClientV2ray`。
   - `newProxySurgeProtocolSpec(...)` 默认支持 `ClientClash`、`ClientMihomo`、`ClientV2ray`、`ClientSurge`。
   - 如果协议只适合部分客户端，在 `base` 上调用 `WithClientSupport(...)` 覆盖默认值，例如 Mieru 只声明 `ClientClash` 与 `ClientMihomo`，因此 v2ray / Surge 输出会跳过它。

   可用客户端常量当前包括：`ClientClash`、`ClientMihomo`、`ClientV2ray`、`ClientSurge`。新增客户端渲染器前，不要只在 `api/clients.go` 写协议特例，应先让协议注册文件声明支持关系。

### 标准注册模板

```go
func init() {
    base := newProtocolSpec(
        "myprotocol",
        []string{"myprotocol://"},
        "MyProtocol",
        "#1976d2",
        "M",
        MyProtocol{},
        "Name",
        DecodeMyProtocolURL,
        EncodeMyProtocolURL,
        func(p MyProtocol) LinkIdentity {
            return buildIdentity("myprotocol", p.Name, p.Server, utils.GetPortString(p.Port))
        },
        // 可选：手工字段 schema，若不传则从结构体反射生成
    )

    // 可选：覆盖客户端兼容性；不写则使用构造器默认值。
    // base = base.WithClientSupport(ClientClash, ClientMihomo)

    MustRegisterProtocol(newProxySurgeProtocolSpec(
        base,
        buildMyProtocolProxy,
        func(proxy Proxy) bool {
            return proxyTypeMatches(proxy, "myprotocol")
        },
        ConvertProxyToMyProtocol,
        EncodeMyProtocolURL,
        buildMyProtocolSurgeLine,
    ))
}
```

如果协议只支持 Clash，不支持 Surge，可以使用：

```go
MustRegisterProtocol(newProxyProtocolSpec(...))
```

如果协议只是演示协议、只需要解析和 UI 元数据，也可以只注册 `newProtocolSpec(...)`。

如果协议有不适合完整 Decode / Import 的额外分享链接前缀，但仍需要参与客户端兼容性判断，可以使用 `WithClientSupportAliases(...)` 增加“仅兼容性检测”的 alias。这样不会把该前缀注册成完整协议解析入口；例如 Mieru 对 `mierus://` 只用于判断 v2ray 不应输出，并不会声明已支持官方 `mierus://` 分享链接的逐字段解析。

### VLESS XHTTP 映射约定

当前仓库对 `vless + xhttp` 的处理遵循以下约定：

- URL 顶层字段：
  - `type=xhttp` → Clash / mihomo `network: xhttp`
  - `encryption` → Clash / mihomo 顶层 `encryption`
  - `path` → `xhttp-opts.path`
  - `host` → `xhttp-opts.host`
  - `mode` → `xhttp-opts.mode`
  - `extra` → 先解码 JSON，再映射到 `xhttp-opts`
- `extra` 当前已支持的字段：
  - `headers` → `xhttp-opts.headers`
  - `noGRPCHeader` → `xhttp-opts.no-grpc-header`
  - `xPaddingBytes` → `xhttp-opts.x-padding-bytes`
  - `downloadSettings` → `xhttp-opts.download-settings`
- `downloadSettings` 中当前已支持的常见子字段包括：
  - `path`、`host`、`headers`、`server`、`port`、`tls`、`alpn`
  - `skipCertVerify` → `skip-cert-verify`
  - `clientFingerprint` → `client-fingerprint`
  - `privateKey` → `private-key`
  - `realityOpts` → `reality-opts`
  - `echOpts` → `ech-opts`

需要特别区分两类 ECH 语义：

- VLESS URL 顶层 `ech=...` 对应的是 Xray/VLESS 的 `echConfigList` 语义，不是所有写法都能与 mihomo `ech-opts` 无损互转。
- 当 `ech` 是固定 **base64 ECHConfig** 时，会映射到顶层 `ech-opts.enable: true` + `ech-opts.config`。
- 当 `ech` 是 Xray 的 DNS / URI 风格（如 `domain+https://...`）时，只会按 mihomo 可表达的能力做最佳努力映射：保留 `enable: true`，并在可识别时写入 `query-server-name`；其中 resolver URI 本身不会被保留。
- 当节点来自 **Clash/mihomo YAML 导入链路**（包括机场订阅中的 Clash YAML 与手动 Clash YAML 导入）时，如果只有顶层 `ech-opts.query-server-name` 可恢复，系统会在写入 `Node.Link` 前按本地兼容规则重建为 `ech=<query-server-name>+https://dns.alidns.com/dns-query`。
- `extra.downloadSettings.echOpts` 只用于 `xhttp` 的嵌套下载设置映射，会转换成 mihomo `xhttp-opts.download-settings.ech-opts`。
- 顶层 `ech` 与 `extra.downloadSettings.echOpts` 都会映射到各自层级的 `ech-opts`，但两者不会互相合并或覆盖。

实现时需要注意：

- `xhttp` 只允许出现在 VLESS 上，不要复用到其他协议。
- 不要把 `xhttp` 静默降级成 `http`、`h2`、`grpc`。
- 用户在订阅设置中勾选“跳过证书验证”后，会通过 `OutputConfig.Cert` 强制覆盖输出配置；对于 `xhttp`，这条规则同时作用于顶层 `skip-cert-verify` 和 `download-settings.skip-cert-verify`。

### 字段元数据说明

`newProtocolSpec(...)` 最后可以追加 `FieldMeta`，用于驱动前端字段展示：

- `Name`：字段名
- `Label`：显示名称
- `Type`：`string` / `int` / `bool`
- `Group`：分组，如 `basic` / `auth` / `transport` / `tls` / `advanced`
- `Description`：字段说明
- `Placeholder`：输入占位提示
- `Options`：枚举选项
- `Advanced`：是否为高级字段
- `Secret`：是否为敏感字段
- `Multiline`：是否建议多行显示

如果不传 `FieldMeta`，系统会回退到结构体反射元数据，这样可以做到“最少接入”。

### 什么时候还需要改协议文件之外的地方？

理想目标是：**只改协议文件并注册即可。**

当前还保留少量“协议外工作”，但它们不属于协议核心接入：

- 补该协议的单元测试
- 如需对外说明，更新 README / docs 的支持矩阵
- 如需更好的前端交互，再补字段元数据（仍可写在协议文件内）

正常情况下，你不应该再去改：

- `node/protocol/clash.go` 的协议分发
- `node/protocol/surge.go` 的协议分发
- `node/sub.go` 的链接生成 switch
- `api/node.go` 的协议判断
- `api/node_raw.go` 的名称提取 switch

如果你发现新增协议还需要改这些地方，说明抽象出现了倒退，应优先修抽象而不是继续补 case。

### ProtocolDemo 的用途

`node/protocol/protocol_demo.go` 不是生产协议，而是协议扩展模板。

它展示了：

- 如何定义协议结构体
- 如何实现 Decode / Encode
- 如何补 `LinkIdentity`
- 如何声明字段元数据
- 如何实现 Clash / Surge 导出能力
- 如何在一个文件里完成注册

新增真实协议时，建议直接复制 `ProtocolDemo` 的结构再改造成你的协议，而不是从零拼装。

---

## ⏰ 定时任务开发指南

SublinkPro 使用模块化定时任务系统，基于 `robfig/cron`。

### 目录结构

```text
services/scheduler/
├── manager.go
├── job_ids.go
├── subscription_task.go
├── speedtest_task.go
├── host_cleanup_task.go
├── reporter.go
├── utils.go
└── bridge.go
```

### 添加新任务的基本步骤

1. 在 `job_ids.go` 定义任务 ID
2. 在 `services/scheduler/` 新增任务文件
3. 在 `manager.go` 的加载逻辑里接入
4. 如有前端任务进度需求，接入 `TaskManager`

### 带进度报告的任务

```go
func ExecuteYourTaskWithProgress() {
    tm := getTaskManager()

    task, ctx, err := tm.CreateTask(
        models.TaskTypeYourType,
        "你的任务名称",
        models.TaskTriggerScheduled,
        100,
    )
    if err != nil {
        utils.Error("创建任务失败: %v", err)
        return
    }

    taskID := task.ID

    for i := 1; i <= 100; i++ {
        select {
        case <-ctx.Done():
            utils.Info("任务被取消")
            return
        default:
        }

        tm.UpdateProgress(taskID, i, "当前处理项", map[string]interface{}{
            "status": "success",
        })
    }

    tm.CompleteTask(taskID, "任务完成", map[string]interface{}{
        "total": 100,
    })
}
```

---

## 🌍 解锁检测扩展指南

解锁检测沿用节点检测 / 测速策略链路，不额外起一套独立任务系统。

### 关键文件

- `api/node_check.go`
- `models/node_check_profile.go`
- `models/node.go`
- `models/unlock.go`
- `services/scheduler/speedtest_config.go`
- `services/scheduler/speedtest_task.go`
- `services/unlock/registry.go`
- `services/unlock/runtime.go`
- `services/unlock/orchestrator.go`
- `services/unlock/checker_*.go`

### 设计原则

- 每个 Provider 一个独立 Checker
- 统一 registry / orchestrator
- 共享 runtime（代理 HTTP client、timeout、落地国家）
- 统一结果结构：`models.UnlockProviderResult`

### 新增一个 Provider

1. 新增 `services/unlock/checker_<provider>.go`
2. 实现：

```go
type UnlockChecker interface {
    Key() string
    Aliases() []string
    Check(runtime UnlockRuntime) models.UnlockProviderResult
}
```

3. 在 `init()` 中注册 `RegisterUnlockChecker(...)`
4. 在 checker 内同时声明 Provider 元数据（展示名、分类、rename 变量等）
5. 如新增了新的状态语义，在 `services/unlock/meta.go` 中补充状态元数据
6. 更新 `docs/features/unlock-check.md`（仅在文档需要列举当前内置 Provider 时）

> [!IMPORTANT]
> 当前前端的节点筛选、标签规则、链式代理条件、订阅编辑中的 unlock 选项都通过后端元数据动态消费。
> 正常情况下，新增一个 checker **不需要**再去前端补 Provider / 状态枚举，也不需要手动同步多个页面的选项列表。

### 命名构建器变量

推荐使用 provider-specific 形式：

- `$Unlock(gemini)`
- `$Unlock(openai)`
- `$Unlock(netflix)`

这些变量通过后端元数据动态下发。

### 多条件解锁筛选

当前节点列表与订阅过滤都支持多条规则。

- 一条规则内部：AND
- 多条规则之间：OR / AND 可选
- 没有规则：表示不启用解锁筛选

### Tag / Chain 规则中的解锁条件

当前 Tag 自动规则和 Chain 规则都已支持：

- `unlock_provider`
- `unlock_status`
- `unlock_keyword`
- `unlock_result`

推荐优先使用 `unlock_provider` 和 `unlock_status` 做精确匹配；`unlock_keyword` 适合做模糊搜索。

这些字段的 schema、可用操作符、枚举值来源现在都由后端统一下发：

- `unlock_provider` → 动态读取已注册 checker 的 Provider 列表
- `unlock_status` → 动态读取后端状态元数据
- `unlock_keyword` / `unlock_result` → 作为文本字段处理

### 解锁检测并行执行

当前单个节点内多个 Provider 检测由 `services/unlock/orchestrator.go` 做**受控并行**。

- 每个节点内部：多 Provider 并行
- 使用小规模并发上限
- 结果顺序保持稳定

---

## 🕐 Cron 表达式格式

本项目使用 5 字段 Cron 格式（不含秒）：

| 字段 | 取值范围 | 说明 |
|:---|:---|:---|
| 分钟 | 0-59 | 每小时的第几分钟 |
| 小时 | 0-23 | 每天的第几小时 |
| 日 | 1-31 | 每月的第几天 |
| 月 | 1-12 | 每年的第几月 |
| 周 | 0-6 | 每周的第几天（0=周日） |

常用示例：

| 表达式 | 说明 |
|:---|:---|
| `*/5 * * * *` | 每 5 分钟 |
| `0 */2 * * *` | 每 2 小时 |
| `30 8 * * *` | 每天 8:30 |
| `0 0 * * 0` | 每周日 00:00 |
| `0 2 1 * *` | 每月 1 日 02:00 |

---

## 💡 开发建议

1. 任务应尽量幂等
2. 长任务支持取消 (`ctx.Done()`)
3. 修改配置语义时同步更新文档
4. 前端命令、生产构建流程优先以 `webs/package.json`、CI、Dockerfile 为准
5. 不要在文档中发明仓库里不存在的命令

---

## 📝 注释规范

合理、规范、可维护的注释是优秀开发习惯的一部分,但注释应解释意图、约束和边界,而不是重复代码本身。

### 何时添加注释

在新增或修改以下内容时必须补充必要注释:

- 关键业务逻辑
- 跨层契约
- 复杂条件分支
- 调度任务
- 迁移
- 并发流程
- 缓存策略
- mihomo 集成
- 认证/安全逻辑
- 配置优先级
- 非显而易见的算法

### 注释语言

注释以中文为主。如涉及以下内容,可保留英文术语或在中文后补充英文:

- 上游库
- 协议字段
- 标准术语
- 公开 API 名称
- 必须与英文文档一致的概念

### 注释维护

- 修改代码时要同步更新附近已有注释
- 避免注释描述旧行为、旧字段、旧限制或旧流程
- 不要为了"有注释"而添加噪音注释
- 注释不得掩盖设计问题
- 对 TODO、FIXME、临时兼容逻辑或已知限制,必须写清楚原因、触发条件和后续处理方向

### Go 注释要求

- 导出的包、类型、函数、方法、接口、常量和变量应遵循 Go 文档注释习惯
- 包注释应放在 `doc.go` 或包内合适文件顶部
- 对非导出但业务关键的函数、结构体字段、状态枚举、迁移步骤和 goroutine 生命周期,也应添加简短中文注释
- 错误处理注释应说明业务语义或恢复策略
- 注释内容应能通过 `gofmt` 保持整洁

### 前端注释要求

- React 组件、Hook、API 请求封装、复杂 `useMemo` / `useEffect`、权限判断、响应式布局分支、主题适配 helper 和跨组件数据流,应在非显而易见时补充中文注释说明设计意图
- JSX 中避免堆叠注释;优先把复杂条件提取为命名变量或小组件
- 主题和样式注释应说明复用的语义层级、light / dark 差异或与现有样板的关系
- 前端请求层注释必须与后端接口语义保持一致
- 临时 UI 限制、兼容浏览器行为、移动端特殊处理和可访问性折中必须写明原因

---

## 🧪 测试规范

规范、可维护的测试用例和清晰的测试文件组织是保证代码质量的基本要求。测试应验证真实业务行为和边界条件,而不是只覆盖实现细节或为了提高覆盖率而堆砌脆弱用例。

### 何时添加测试

在新增或修改以下内容时应同步补充或更新对应 Go 测试:

- 后端关键业务逻辑
- 接口契约
- 权限判断
- 配置语义
- 迁移
- 调度任务
- mihomo 集成
- 协议解析
- 数据转换

### 测试命名

测试名称应描述"场景 + 期望结果"。避免只写 `TestSuccess`、`TestError`、`should work` 这类无法表达业务意图的名称。

### 测试覆盖

- 测试应覆盖正常路径、边界条件、错误路径和权限/配置差异
- 不要只验证最容易通过的 happy path
- 测试数据应尽量局部、可读、最小化,并通过 helper / fixture 表达业务含义
- 测试之间必须相互隔离
- 断言应明确验证关键输出、状态变化、副作用和错误语义

### 测试维护

- 修复缺陷时,优先先写能复现问题的回归测试,再修复实现
- 不得删除、跳过或弱化失败测试来掩盖问题
- 如果当前模块已有测试风格、fixture、mock 或 helper,应优先复用并保持一致
- 不要为未接入的测试框架、脚本或 CI 流程编写假设性测试说明

### Go 测试要求

- Go 测试文件必须使用 `_test.go` 后缀,并与被测文件或被测包放在同一包目录下
- Go 测试函数命名应遵循标准格式:`TestXxx`、`BenchmarkXxx`、`FuzzXxx`、`ExampleXxx`
- 优先使用表驱动测试覆盖同一函数的多种输入、边界和错误情况
- 测试 helper 应调用 `t.Helper()`
- 临时目录和临时文件应优先使用 `t.TempDir()`
- 清理动作应通过 `t.Cleanup()` 注册
- 涉及时间、随机数、网络、数据库、文件系统或 goroutine 的测试,应显式控制依赖
- HTTP handler 测试应优先使用 `httptest` 和明确的请求/响应断言
- 数据库相关测试应使用隔离的测试库、事务或临时存储
- 并发测试不要依赖 `time.Sleep` 猜测时序

### 前端测试边界

本仓库不默认要求为前端改动新增测试。前端改动的常规质量保障以 lint、build、人工交互验证和 light / dark / 响应式检查为主。

只有在以下情况才补充或更新前端测试:

- 用户明确要求
- 仓库后续接入前端测试框架
- 当前模块已经存在前端测试

如果确实编写前端测试:

- 文件应与被测文件保持可追溯的命名关系
- 优先使用 `*.test.jsx`、`*.test.js`、`*.spec.jsx` 或 `*.spec.js`
- 靠近被测组件/hook/helper 放置
- 应从用户可观察行为出发
- 验证文本、角色、状态变化、交互结果、错误提示和 loading / empty / disabled 等状态
- 不要断言内部 state 或 MUI 生成的脆弱 class 名
