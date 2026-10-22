[English](unlock-check.md) | 简体中文

# 解锁检测

SublinkPro 现在支持在节点检测流程中附带执行 **流媒体 / AI 服务可用区检测**。

这个功能不是单独起一套任务系统，而是直接挂在现有的 **节点检测 / 测速策略** 上执行：

- 选择节点范围
- 执行延迟 / 速度检测
- 可选检测落地国家、IP 质量
- 可选检测解锁情况
- 将结果直接写回节点信息，并在前端节点列表、详情面板、任务中心中展示

---

## 当前内置检查项

首批内置 Provider：

- Netflix
- Disney+
- YouTube Premium
- OpenAI
- Gemini
- Claude
- 巴哈姆特动画疯

> [!NOTE]
> 当前内置 Provider 会尽量采用与主流解锁检测脚本一致的服务级探针：例如 OpenAI 会分别检查 Web / iOS 入口，Disney+ 会走设备、令牌与地区 GraphQL 探针，YouTube Premium、Gemini、Claude、Netflix 也会读取对应页面或最终跳转中的可用性标记。
>
> 上面的列表表示**当前内置** Provider，而不是前端或规则系统中的固定枚举。新增 checker 后，相关 Provider 选择器和 unlock 条件会通过后端元数据自动更新。

---

## 使用方式

进入：

`检测中心 -> 检测方案 -> 新建方案 / 配置`

在策略中可开启：

- **解锁检测**
- 选择需要检测的 Provider 列表

如果不手动选择 Provider，则系统会按后端注册的默认 Provider 集合执行。

执行后，结果会显示在：

- 节点列表
- 节点卡片
- 节点详情面板
- 任务进度面板
- 任务中心历史记录

---

## 解锁筛选规则

节点列表和订阅过滤现在都支持 **多条解锁筛选规则**。

每条规则包含：

- `Provider`
- `状态`
- `关键词`

匹配语义如下：

- **同一条规则内部**：按 **AND** 生效
- **多条规则之间**：可按 **OR** 或 **AND** 生效

也就是说：

- 一条规则：`Gemini + 解锁 + US`
  - 表示同一条解锁结果必须同时满足这三个条件
- 多条规则：
  - `Gemini + 解锁`
  - `YouTube Premium + 解锁`
  - 在 OR 模式下：满足其中任意一条即可通过筛选
  - 在 AND 模式下：需要同时满足全部规则

如果用户没有新增任何解锁规则，则表示 **不启用解锁筛选**。

---

## 命名与展示建议

解锁信息可以参与节点重命名，但推荐优先使用 **紧凑摘要**，而不是把所有平台结果完整展开到名称里。

推荐：

- `$Unlock(provider)`：按具体 Provider 输出紧凑结果，例如 `$Unlock(openai)` → `解锁-US`
- `$Unlock`：主解锁摘要，例如 `Netflix-解锁-US-+2`

不建议把多个平台的详细结果全部拼进节点名称，否则会导致名称过长、难读、难搜索。

这适合以下场景：

- 想找“Gemini 解锁”的节点
- 想找“Gemini 解锁 或 YouTube Premium 解锁”的节点
- 想找“Claude 解锁且带 US 关键词”的节点

---

## 结果语义

单个 Provider 的结果采用统一结构，核心字段包括：

- `provider`：Provider 标识
- `status`：检测结果状态
- `region`：检测得到的地区（如果适用）
- `reason`：失败 / 受限原因
- `detail`：额外说明

当前常见状态：

| 状态 | 含义 |
|:---|:---|
| `available` | 明确可用 |
| `partial` | 部分，例如仅 Originals |
| `reachable` | 直连，可访问入口但不代表完整能力 |
| `restricted` | 受限，当前地区或出口被限制 |
| `unsupported` | 不支持，当前地区不在官方支持范围 |
| `error` | 异常，本轮检测失败 |
| `unknown` | 无法可靠判断 |

> [!IMPORTANT]
> `reachable` 与 `available` 不完全等价。
> 
> Provider 返回 `available` 时表示检测到了该服务对应的解锁标记或 API 可用性结果；`reachable` 仅表示入口可达，不应视为完整解锁。
> Gemini 检测会进一步读取页面内的功能可用性标记；只有出现明确可用标记时才判定为 `available`，未找到该标记时按 `restricted` 处理，避免把地域受限但仍返回 HTTP 成功的页面误判为“解锁”。
> OpenAI 的 `partial` 表示上游检测语义中的“Only Available with Web Browser”或“Only Available with Mobile APP”；Disney+ 的 `partial` 表示该地区处于 `Available Soon` 状态；Netflix 的 `partial` 表示 Originals Only。

---

## 架构设计

该功能采用 **注册表 + 独立 Checker 模块 + 总调度器** 的结构。

### 1. 调度层

节点检测仍由现有链路负责：

- `api/node_check.go`
- `models/node_check_profile.go`
- `services/scheduler/speedtest_config.go`
- `services/scheduler/speedtest_task.go`

这些文件负责：

- 保存策略配置
- 将配置转为运行时参数
- 决定何时执行解锁检测
- 将结果写回节点与任务结果

### 2. 解锁子系统

解锁检测核心位于：

- `services/unlock/registry.go`
- `services/unlock/runtime.go`
- `services/unlock/orchestrator.go`
- `services/unlock/checker_*.go`

其中：

- **registry**：注册并解析 Checker
- **runtime**：提供共享 HTTP client、timeout、落地国家等运行时上下文
- **orchestrator**：按策略选择 Provider，统一调度执行并生成结果汇总
- **checker 文件**：每个 Provider 单独维护自己的探测逻辑

### 3. 数据层

节点侧以统一结构保存结果：

- `models/unlock.go`
- `models/node.go`

当前节点会保存：

- `unlockSummary`
- `unlockCheckAt`

这样后续新增 Provider 时，不需要给 `Node` 再增加一列字段。

---

## 如何新增一个解锁检查项

这是本功能最核心的可维护性目标：

> **新增一个 Provider，应尽量只需要：新增 checker 文件 + 注册 + provider 元数据声明。**

### 步骤 1：新增 checker 文件

在 `services/unlock/` 下新增一个类似文件：

`unlock_checker_example.go`

实现统一接口：

```go
type UnlockChecker interface {
    Key() string
    Aliases() []string
    Check(runtime UnlockRuntime) models.UnlockProviderResult
}
```

同时建议在 checker 中实现对应的元数据方法，让后端自动下发展示信息和 rename 变量：

```go
type UnlockCheckerMeta interface {
    Meta() models.UnlockProviderMeta
}

type UnlockCheckerRenameMeta interface {
    RenameVariableMeta() models.UnlockRenameVariableMeta
}
```

### 步骤 2：实现 `Check`

在 `Check(runtime UnlockRuntime)` 中：

- 使用共享运行时中的代理 HTTP client
- 执行该 Provider 自己的低成本探测
- 返回统一的 `models.UnlockProviderResult`

不要：

- 直接操作数据库
- 直接依赖 scheduler / task manager
- 在 checker 内关心别的 Provider 的逻辑

### 步骤 3：注册 checker

在新文件中通过 `init()` 调用注册：

```go
func init() {
    RegisterUnlockChecker(exampleUnlockChecker{})
}
```

### 步骤 4：前端展示与规则系统

正常情况下，不需要再去前端手动补 Provider / 状态枚举。

当前这些位置都会通过后端元数据动态消费 unlock 信息：

- 节点检测策略里的 Provider 选择器
- 节点页面解锁筛选
- 订阅编辑中的 unlock 过滤规则
- 标签规则里的 unlock 条件
- 链式代理里的 unlock 条件
- 节点重命名中的 `$Unlock(provider)` 变量列表

也就是说，**新增 checker 后，只要后端元数据完整，下游 UI 会自动更新。**

只有在以下场景才需要额外前端改动：

- 你新增了全新的 unlock 条件字段，而不是新增 Provider
- 你希望为某个 Provider 做额外的专属视觉呈现，而不仅仅是普通枚举选项

### 步骤 5：文档同步

新增 Provider 后，应按需同步更新：

- 本文档中的“当前内置检查项”（如果你希望文档反映最新内置 Provider）
- `README.md` 中的功能说明（通常不需要逐个 Provider 更新）
- `docs/development.md` 中的开发说明（如果扩展方式本身变化）

> [!TIP]
> 如果只是新增一个普通 checker，而没有改变扩展机制，通常不需要再更新多个前端文档来同步枚举，因为运行时选择器会自动读取后端元数据。

---

## 维护约束建议

为了让这个子系统长期可维护，建议遵守以下规则：

1. **每个 Provider 只维护自己的逻辑**
2. **共享逻辑放在 runtime / orchestrator，不要在 checker 之间复制**
3. **不要把 Provider dispatch 再改回集中式 switch**
4. **结果结构尽量稳定，避免前后端反复改字段**
5. **优先低成本探测，避免浏览器自动化和高资源脚本依赖**

---

## 适用边界

这个功能适合：

- 批量筛选节点的地区能力
- 判断流媒体 / AI 服务是否大概率可用
- 给订阅筛选、节点运营、标签规则提供补充依据
- 在订阅过滤里按解锁结果筛选节点
- 在节点命名规则中插入解锁结果摘要

这个功能暂时不追求：

- 完全模拟真实登录后业务态
- 对所有平台做到 100% 精准
- 通过复杂反爬绕过手段来提高命中率

首版目标是：

**可维护、可扩展、可批量运行、可稳定展示。**

---

## 检测中心：方案、任务、记录与结果

**检测中心**将四类职责分开管理：

- **检测方案**定义连通性/延迟、可选下载测速、出口 IP/国家、IP 质量及元数据驱动的 AI/流媒体模块。请求 URL、超时、并发、采样和流量统计仍在详细设置中。延迟为基础检测，下载测速使用现有 Mihomo 完整测速模式。节点列表的单个/选中节点检测继续使用方案选择器，并明确传递节点 ID。
- **自动任务**复用方案，独立选择全部节点或机场/原分组/标签交集。五段 Cron 使用服务端/容器时区，生成器预览使用浏览器时区。机场更新事件仅在该机场成功更新后检测其符合条件的节点。
- **运行记录**复用现有 `speed_test` 任务记录、进度及停止操作。名称不是稳定的方案 ID；记录是执行摘要，不是逐节点历史快照。
- **节点结果**分页展示各项目的**最新状态**、原机场、Provider 状态/原因/详情和检测时间，支持重测单个节点。此接口不返回原始代理链接或凭据。

定时筛选任务可指定机场并启用**先更新指定机场**。等待全部机场更新和自动标签完成后，再解析最新范围并检测；失败/取消即停止。同一流程的前置更新抑制更新后检测钩子，防止递归或重复检测。

每个独立任务防止自身重叠：重复手动启动报错，定时重叠跳过，检测期间到达的机场事件合并为一批补跑，补跑重新读取启用状态和范围。防重是**每任务隔离**，不是全局节点锁；其他任务及机场原有更新后检测仍独立，避免给同一批节点配置重复路径。空范围不回退到全库或方案旧范围。

已有方案 ID、机场引用、API 和 Telegram 定时开关保持兼容。旧策略调度映射为负 ID 的**兼容任务**，不复制或重复注册。在这里编辑 Cron/分组/标签；更换方案或前置更新机场时新建独立任务。删除兼容任务只清除调度，不删除方案。被暂停任务或机场引用的方案也不可删除。数据库恢复保留独立任务及方案原 ID，旧备份恢复会清除目标端残留任务。

**升级前请备份数据库**：首次启动会自动迁移新增的自动任务表。检测结果供现有订阅筛选读取，不移动节点，也不另启一次检测。网络探测不证明账号登录或付费模型可调用，第三方 IP 风险仅供参考，并非平台官方评分。

OpenClash 等客户端独立运行；请按节点数量、并发和超时预留缓冲，在客户端定时刷新订阅前完成检测。订阅请求不会等待未完成的任务。
