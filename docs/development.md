English | [简体中文](development.zh-CN.md)

# Development Guide

Welcome to SublinkPro development. This guide focuses on:

- How to run the backend and frontend locally
- What the production build flow actually does
- Which files and directories are high value entry points
- Where unlock check extension points live

---

## 📁 Project Structure

```text
sublinkPro/
├── api/                     # HTTP API / controller
├── models/                  # Data models, persistence, migrations
├── services/                # Business services and background subsystems
│   ├── scheduler/           # Scheduled tasks and task scheduler
│   ├── mihomo/              # mihomo integration, speed test, DNS, Host, proxied outbound
│   └── unlock/              # Unlock registry, runtime, checker modules
├── routers/                 # Route registration
├── node/                    # Subscription and protocol parsing/conversion
├── utils/                   # Shared helpers
├── database/                # Database connection and dialect support
├── cache/                   # Cache layer
├── dto/                     # DTO / form structures
├── webs/                    # React + Vite frontend
│   └── src/
│       ├── api/            # Frontend request boundary
│       ├── views/          # Page level features
│       ├── components/     # Shared components
│       ├── utils/          # Frontend helpers
│       ├── themes/         # Theme and MUI overrides
│       └── routes/         # Route definitions
├── template/                # Template files
├── docs/                    # Documentation
├── static/                  # Frontend build assets for production builds
├── main.go                  # Application entry
├── Dockerfile               # Docker build
└── README.md
```

---

## 🔧 Tech Stack

| Layer | Technology |
|:---|:---|
| Backend framework | Go + Gin |
| ORM | GORM |
| Database | SQLite, default / MySQL / PostgreSQL |
| Frontend framework | React 19 + Vite |
| UI | Material UI |
| Frontend package manager | Yarn 4 |
| Scheduler | robfig/cron |

---

## 💻 Local Development

### 1. Clone the project

```bash
git clone https://github.com/ZeroDeng01/sublinkPro.git
cd sublinkPro
```

### 2. Backend development

Go **1.26.3** or newer is recommended, matching the repository, Docker, and CI.

```bash
go mod download
go run main.go
```

The backend listens on `:8000` by default.

### 3. Frontend development

Run under `webs/`:

```bash
yarn install
yarn run start
```

The default Vite dev port is `3000`, with `/api` proxied to the backend.

### 4. Frontend validation

Run under `webs/`:

```bash
yarn run lint
yarn run build
yarn run lint:fix
yarn run prettier
```

After frontend changes, run at least `yarn run lint`. Also run `yarn run build` when build output, asset paths, routing, base path behavior, or production integration is affected.

> [!NOTE]
> This repository has **no authoritative frontend `test` or `typecheck` script**. Don't invent validation flows in docs or automation.

### 5. Backend validation

After backend changes, Go files must be formatted with `gofmt`, and `golangci-lint` plus related tests must run:

```bash
gofmt -w <changed-go-files>
golangci-lint run
go test ./...
```

When adding or changing key business logic, API contracts, permission checks, configuration semantics, migrations, scheduled tasks, mihomo integration, protocol parsing, or data conversion, add or update matching Go tests. The GitHub release build runs `golangci-lint` and full repository `go test ./...` before building binaries.

### 6. PR automated checks

`.github/workflows/pr-checks.yml` runs automatically when a PR is opened, reopened, or marked ready for review. It gives a quick baseline quality signal. Later fix commits do not automatically consume Actions again. After fixes are ready, the PR author or a repository admin can comment `/recheck` on the PR to trigger another round.

- Backend: `golangci-lint`, `go test ./...`
- Frontend: `yarn run lint`, `yarn run build`

Each check job writes to the GitHub Step Summary, showing pass or failure status for each item so authors and reviewers can see what is done and what still needs work.

### 7. Normal backend build

```bash
go build -o sublinkpro main.go
```

This is useful for development or a quick local compile. It is not the production embedded build.

### 8. Production build, actual flow

Production build has two stages. Before running a production style local build, finish frontend lint/build and backend format/lint/tests first:

```bash
# 1) Frontend lint and build
cd webs
yarn run lint
yarn run build

# 2) Backend format, lint, and tests
cd ..
gofmt -w <changed-go-files>
golangci-lint run
go test ./...

# 3) Prepare production static assets
rm -rf static && mkdir -p static
cp -R webs/dist/. static/

# 4) Build production backend with embedded frontend assets
CGO_ENABLED=0 go build -tags=prod -ldflags="-s -w" -o sublinkPro
```

> [!IMPORTANT]
> If you change frontend asset paths, PWA assets, base path behavior, embedding logic, or static file serving, verify all of these:
>
> - `webs` local development mode
> - Frontend build output
> - Production embedded build after copying assets into `static/`

---

## 🧭 Key Runtime Conventions

### Path boundaries

- Frontend UI: `/` or the path set by `SUBLINK_WEB_BASE_PATH`
- API: always under `/api/*`
- Subscription/share access: always under `/c/*`

`SUBLINK_WEB_BASE_PATH` affects only the Web UI. It does not affect API or subscription fetch paths.

### Runtime directories

These directories contain runtime state. Handle them carefully:

- `db/`
- `logs/`
- `template/`
- `out/`

Where:

- `db/`: database, config files, GeoIP, and other local data
- `template/`: template files
- `logs/`: runtime logs

---

## 🔍 High Value Entry Files

| Module | File | Notes |
|:---|:---|:---|
| Node speed tests | `services/scheduler/speedtest_task.go` | Main flow for latency, speed, quality, and unlock checks |
| Unlock checks | `services/unlock/*.go` | Provider registry / runtime / orchestrator / checkers |
| Tag rules | `services/tag_service.go` | Automatic tag rule execution |
| Subscription generation | `api/clients.go` | Subscription output, node filtering, rename |
| Chain proxy | `api/subscription_chain.go` / `models/subscription_chain_rule.go` | Subscription chain proxy rules and condition based node selection |
| Host management | `models/host.go` | Host mappings, batch writes, cache management |
| DNS resolution | `services/mihomo/dns_resolver.go` | Custom DNS and proxy based resolution |
| Data migration | `models/db_migrate.go` | Database migration scripts |

---

## 🔌 Protocol Extension Guide

The protocol system has been refactored into a **self registration + capability interface** model. The goal is:

> When adding a protocol, a developer should only need to add one protocol file under `node/protocol/`, implement the protocol, export capabilities, and register it.

### Protocol extension entry point

Use these as references:

- `node/protocol/protocol_demo.go`: standard sample protocol
- Real protocol files such as:
  - `node/protocol/vmess.go`
  - `node/protocol/ss.go`
  - `node/protocol/http.go`

### Current protocol architecture

Core capabilities live in `node/protocol/protocol_meta.go`:

- `Protocol`: core protocol specification
- `ProxyCapable`: supports conversion to Clash Proxy structs
- `SurgeCapable`: supports Surge line export
- `SupportsClient(...)`: declares subscription output compatibility for Clash / mihomo / v2ray / Surge and other clients
- `MustRegisterProtocol(...)`: protocol registration entry point

After adding a protocol, these flows are connected automatically without adding extra switches:

- Protocol recognition, alias / scheme
- Node raw parsing
- Node raw field updates
- Node identity extraction, name / host / port / address
- Deduplication field reads
- Node link renaming
- `LinkToProxy` dispatch
- `EncodeSurge` dispatch
- `EncodeProxyLink` dispatch
- v2ray raw output compatibility filtering, through client support declared in the protocol file
- Protocol UI metadata output

### Recommended steps for adding a protocol

1. Add a protocol file under `node/protocol/`, for example:

   ```text
   node/protocol/myprotocol.go
   ```

2. Define the protocol struct.

   Struct fields are the default source for UI field metadata, so names should be stable and clear.

3. Implement link `Decode` / `Encode`.

   At minimum:

   - `DecodeXxxURL(string) (Xxx, error)`
   - `EncodeXxxURL(Xxx) string`

4. If you need to convert back from Clash Proxy to a link, add `ConvertProxyToXxx(proxy Proxy) Xxx`.

5. If the protocol supports Clash export, implement `buildXxxProxy(link Urls, config OutputConfig)` in the same file.

6. If the protocol supports Surge export, implement `buildXxxSurgeLine(link string, config OutputConfig)` in the same file.

7. Self register in `init()` in the same file.

8. Declare client compatibility in the protocol file. Defaults are:

   - `newProtocolSpec(...)` supports `ClientV2ray` by default.
   - `newProxyProtocolSpec(...)` supports `ClientClash`, `ClientMihomo`, and `ClientV2ray` by default.
   - `newProxySurgeProtocolSpec(...)` supports `ClientClash`, `ClientMihomo`, `ClientV2ray`, and `ClientSurge` by default.
   - If the protocol is suitable for only some clients, call `WithClientSupport(...)` on `base` to override defaults. For example, Mieru declares only `ClientClash` and `ClientMihomo`, so v2ray / Surge output skips it.

   Available client constants currently include `ClientClash`, `ClientMihomo`, `ClientV2ray`, and `ClientSurge`. Before adding a new client renderer, don't add protocol special cases only in `api/clients.go`; first let protocol registration files declare support relationships.

### Standard registration template

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
        // Optional: manual field schema. If omitted, reflection generates it from the struct.
    )

    // Optional: override client compatibility. If omitted, constructor defaults are used.
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

If the protocol supports Clash but not Surge, use:

```go
MustRegisterProtocol(newProxyProtocolSpec(...))
```

If the protocol is only a demo protocol and needs only parsing plus UI metadata, registering only `newProtocolSpec(...)` is also fine.

If a protocol has extra share link prefixes that are not suitable for full Decode / Import, but still need to participate in client compatibility checks, use `WithClientSupportAliases(...)` to add aliases for compatibility checks only. This does not register that prefix as a full parser entry. For example, Mieru uses `mierus://` only to decide that v2ray should not output it. It does not claim full field by field parsing support for the official `mierus://` share link.

### VLESS XHTTP mapping conventions

This repository handles `vless + xhttp` with these rules:

- Top level URL fields:
  - `type=xhttp` maps to Clash / mihomo `network: xhttp`
  - `encryption` maps to top level Clash / mihomo `encryption`
  - `path` maps to `xhttp-opts.path`
  - `host` maps to `xhttp-opts.host`
  - `mode` maps to `xhttp-opts.mode`
  - `extra` is decoded as JSON first, then mapped to `xhttp-opts`
- Supported fields inside `extra`:
  - `headers` maps to `xhttp-opts.headers`
  - `noGRPCHeader` maps to `xhttp-opts.no-grpc-header`
  - `xPaddingBytes` maps to `xhttp-opts.x-padding-bytes`
  - `downloadSettings` maps to `xhttp-opts.download-settings`
- Common supported subfields inside `downloadSettings` include:
  - `path`, `host`, `headers`, `server`, `port`, `tls`, `alpn`
  - `skipCertVerify` maps to `skip-cert-verify`
  - `clientFingerprint` maps to `client-fingerprint`
  - `privateKey` maps to `private-key`
  - `realityOpts` maps to `reality-opts`
  - `echOpts` maps to `ech-opts`

Two ECH meanings must be kept separate:

- Top level VLESS URL `ech=...` corresponds to Xray/VLESS `echConfigList` semantics. Not every form can round trip losslessly with mihomo `ech-opts`.
- When `ech` is a fixed **base64 ECHConfig**, it maps to top level `ech-opts.enable: true` + `ech-opts.config`.
- When `ech` is Xray DNS / URI style, such as `domain+https://...`, it is mapped on a best effort basis to what mihomo can express. `enable: true` is preserved, and `query-server-name` is written when recognizable. The resolver URI itself is not preserved.
- When a node comes from the **Clash/mihomo YAML import flow**, including Clash YAML airport subscriptions and manual Clash YAML import, and only top level `ech-opts.query-server-name` can be restored, the system rebuilds it before writing `Node.Link` as `ech=<query-server-name>+https://dns.alidns.com/dns-query` using local compatibility rules.
- `extra.downloadSettings.echOpts` is used only for nested `xhttp` download settings and maps to mihomo `xhttp-opts.download-settings.ech-opts`.
- Top level `ech` and `extra.downloadSettings.echOpts` each map to their own `ech-opts` level. They are not merged or overwritten with each other.

Implementation notes:

- `xhttp` is allowed only on VLESS. Don't reuse it for other protocols.
- Don't silently downgrade `xhttp` to `http`, `h2`, or `grpc`.
- When users enable “skip certificate verification” in subscription settings, `OutputConfig.Cert` force overrides output configuration. For `xhttp`, this applies to both top level `skip-cert-verify` and `download-settings.skip-cert-verify`.

### Field metadata

`newProtocolSpec(...)` can take optional `FieldMeta` entries at the end to drive frontend field display:

- `Name`: field name
- `Label`: display label
- `Type`: `string` / `int` / `bool`
- `Group`: group, such as `basic` / `auth` / `transport` / `tls` / `advanced`
- `Description`: field description
- `Placeholder`: placeholder text
- `Options`: enum options
- `Advanced`: whether this is an advanced field
- `Secret`: whether this is sensitive
- `Multiline`: whether multiline display is recommended

If `FieldMeta` is omitted, the system falls back to struct reflection metadata, which enables minimal integration.

### When do you need changes outside the protocol file?

The ideal target is: **only add and register the protocol file.**

A small amount of “outside protocol” work still remains, but it is not core protocol integration:

- Add unit tests for the protocol
- Update README / docs support matrix if public docs need it
- Add better field metadata for frontend interaction, still in the protocol file when possible

Normally you should no longer change:

- Protocol dispatch in `node/protocol/clash.go`
- Protocol dispatch in `node/protocol/surge.go`
- Link generation switches in `node/sub.go`
- Protocol detection in `api/node.go`
- Name extraction switches in `api/node_raw.go`

If adding a protocol still requires changes there, the abstraction has regressed. Fix the abstraction before adding more cases.

### Purpose of ProtocolDemo

`node/protocol/protocol_demo.go` is not a production protocol. It is a protocol extension template.

It shows:

- How to define a protocol struct
- How to implement Decode / Encode
- How to add `LinkIdentity`
- How to declare field metadata
- How to implement Clash / Surge export capabilities
- How to complete registration in one file

When adding a real protocol, it is recommended to copy the `ProtocolDemo` structure and adapt it, instead of building everything from scratch.

---

## ⏰ Scheduled Task Development Guide

SublinkPro uses a modular scheduled task system based on `robfig/cron`.

### Directory structure

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

### Basic steps for adding a task

1. Define the task ID in `job_ids.go`.
2. Add a task file under `services/scheduler/`.
3. Wire it into the loading logic in `manager.go`.
4. If the frontend needs task progress, connect it to `TaskManager`.

### Task with progress reporting

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

## 🌍 Unlock Check Extension Guide

Unlock checks reuse the node check / speed test strategy flow. They don't start a separate task system.

### Key files

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

### Design principles

- One independent Checker per Provider
- Unified registry / orchestrator
- Shared runtime, including proxy HTTP client, timeout, and landing country
- Unified result structure: `models.UnlockProviderResult`

### Add a Provider

1. Add `services/unlock/checker_<provider>.go`.
2. Implement:

```go
type UnlockChecker interface {
    Key() string
    Aliases() []string
    Check(runtime UnlockRuntime) models.UnlockProviderResult
}
```

3. Register it with `RegisterUnlockChecker(...)` in `init()`.
4. Declare Provider metadata in the checker, including display name, category, rename variables, and related fields.
5. If new status semantics are added, add status metadata in `services/unlock/meta.go`.
6. Update `docs/features/unlock-check.md` only when the docs need to list the current built in Providers.

> [!IMPORTANT]
> The current frontend node filters, tag rules, chain proxy conditions, and unlock options in subscription editing all consume backend metadata dynamically.
> Normally, adding one checker **does not** require adding Provider or status enums to the frontend or manually syncing option lists across multiple pages.

### Rename builder variables

Provider specific forms are recommended:

- `$Unlock(gemini)`
- `$Unlock(openai)`
- `$Unlock(netflix)`

These variables are delivered dynamically through backend metadata.

### Multi condition unlock filtering

Node lists and subscription filters currently support multiple rules.

- Inside one rule: AND
- Between multiple rules: OR / AND, user selectable
- No rules: unlock filtering is disabled

### Unlock conditions in Tag / Chain rules

Current automatic Tag rules and Chain rules support:

- `unlock_provider`
- `unlock_status`
- `unlock_keyword`
- `unlock_result`

Prefer `unlock_provider` and `unlock_status` for exact matches. `unlock_keyword` is better for fuzzy search.

Schemas, operators, and enum values for these fields are all delivered by the backend:

- `unlock_provider` reads the list of registered checker Providers dynamically
- `unlock_status` reads backend status metadata dynamically
- `unlock_keyword` / `unlock_result` are treated as text fields

### Parallel unlock checks

Multiple Provider checks for a single node are run with **controlled parallelism** in `services/unlock/orchestrator.go`.

- Inside each node: multiple Providers run in parallel
- A small concurrency limit is used
- Result order stays stable

---

## 🕐 Cron Expression Format

This project uses a 5 field Cron format, without seconds:

| Field | Range | Description |
|:---|:---|:---|
| Minute | 0-59 | Minute of the hour |
| Hour | 0-23 | Hour of the day |
| Day | 1-31 | Day of the month |
| Month | 1-12 | Month of the year |
| Weekday | 0-6 | Day of the week, 0=Sunday |

Common examples:

| Expression | Description |
|:---|:---|
| `*/5 * * * *` | Every 5 minutes |
| `0 */2 * * *` | Every 2 hours |
| `30 8 * * *` | Every day at 08:30 |
| `0 0 * * 0` | Every Sunday at 00:00 |
| `0 2 1 * *` | Every month on day 1 at 02:00 |

---

## 💡 Development Advice

1. Tasks should be idempotent when possible.
2. Long running tasks should support cancellation through `ctx.Done()`.
3. Update docs when configuration semantics change.
4. Frontend commands and production build flow should follow `webs/package.json`, CI, and Dockerfile first.
5. Don't document commands that don't exist in the repository.
