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
├── static/                  # 生产构建时前端产物放置目录
├── main.go                  # 应用入口
├── Dockerfile               # Docker 构建
└── README.md
```

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

建议使用 **Go 1.26.1** 或更高版本，与仓库、Docker 和 CI 保持一致。

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

> [!NOTE]
> 当前仓库**没有权威的前端 `test` 或 `typecheck` 脚本**。不要在文档或自动化里发明不存在的校验流程。

### 5. 普通后端构建

```bash
go build -o sublinkpro main.go
```

这适合开发环境或快速本地编译，不代表生产嵌入构建。

### 6. 生产构建（实际流程）

生产构建是两阶段：

```bash
# 1) 构建前端
cd webs && yarn run build

# 2) 准备生产静态资源
cd ..
rm -rf static && mkdir -p static
cp -R webs/dist/. static/

# 3) 构建生产后端（嵌入前端资源）
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
