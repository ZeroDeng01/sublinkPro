# 解锁检测

SublinkPro 现在支持在节点检测流程中附带执行 **流媒体 / AI 服务可用区检测**。

这个功能不是单独起一套任务系统，而是直接挂在现有的 **节点检测 / 测速策略** 上执行：

- 选择节点范围
- 执行延迟 / 速度检测
- 可选检测落地国家、IP 质量
- 可选检测解锁情况
- 将结果直接写回节点信息，并在前端节点列表、详情面板、任务中心中展示

---

## 当前支持的检查项

首批内置 Provider：

- Netflix
- Disney+
- YouTube Premium
- OpenAI
- Gemini
- Claude

> [!NOTE]
> 当前版本优先追求 **可维护性、运行效率和稳定性**，因此首版主要使用低成本 HTTP 探测与区域可用性判断，不依赖外部脚本、浏览器自动化或复杂登录流程。

---

## 使用方式

进入：

`节点检测 -> 新建 / 编辑策略`

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
  - `YouTube Premium + 直连`
  - 在 OR 模式下：满足其中任意一条即可通过筛选
  - 在 AND 模式下：需要同时满足全部规则

如果用户没有新增任何解锁规则，则表示 **不启用解锁筛选**。

---

## 命名与展示建议

解锁信息可以参与节点重命名，但推荐优先使用 **紧凑摘要**，而不是把所有平台结果完整展开到名称里。

推荐：

- `$Unlock(provider)`：按具体 Provider 输出紧凑结果，例如 `$Unlock(openai)` → `直连-US`
- `$Unlock`：主解锁摘要，例如 `Netflix-解锁-US-+2`

不建议把多个平台的详细结果全部拼进节点名称，否则会导致名称过长、难读、难搜索。

这适合以下场景：

- 想找“Gemini 解锁”的节点
- 想找“Gemini 解锁 或 YouTube Premium 直连”的节点
- 想找“Claude 直连且带 US 关键词”的节点

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
> 对部分 AI 或媒体平台，首版只提供“该地区是否可进入 / 可访问”的可维护判定，而不是模拟完整登录后真实业务调用。

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

> **新增一个 Provider，应尽量只需要：新增 checker 文件 + 注册。**

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

### 步骤 4：前端展示

如果这是一个全新的 Provider，通常还需要同步更新：

- 前端对 `/api/v1/node-check/meta` 的展示消费

主要是补充：

- 友好展示名称
- 如有必要的筛选 / 选择器专属文案

### 步骤 5：文档同步

新增 Provider 后，应同步更新：

- 本文档
- `README.md` 中的功能说明（如有必要）
- `docs/development.md` 中的开发说明（如涉及扩展方式变化）

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
