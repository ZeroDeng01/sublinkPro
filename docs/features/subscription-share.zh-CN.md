[English](subscription-share.md) | 简体中文

# 订阅分享管理

全新的订阅分享管理功能，取代了原有的单一 Token 模式，提供更安全、更灵活的分享链接管理能力。

---

## 核心特点

| 特点 | 说明 |
|:---|:---|
| **多链接管理** | 每个订阅可创建多个独立的分享链接，方便分发给不同用户或场景 |
| **安全 Token** | 采用随机生成的安全 Token，也支持自定义 Token 便于记忆 |
| **过期策略** | 支持永不过期、按天数过期、指定时间过期三种策略 |
| **独立统计** | 每个分享链接独立记录访问次数和 IP 日志 |
| **启用/禁用** | 可随时启用或禁用单个分享链接，无需删除 |
| **Token 刷新** | 一键刷新 Token，旧链接立即失效，安全便捷 |
| **二维码生成** | 支持为每个分享链接生成二维码，方便移动端扫码导入 |

---

## ⏰ 过期策略

| 策略 | 说明 |
|:---|:---|
| **永不过期** | 链接长期有效，除非手动禁用或删除 |
| **按天数过期** | 从创建时起指定天数后自动失效，如 7 天、30 天 |
| **指定时间过期** | 设置具体的过期日期和时间，到期后自动失效 |

---

## 📋 使用场景

```
场景一：分用户管理
├── 为朋友 A 创建分享链接（永不过期）
├── 为朋友 B 创建分享链接（30天后过期）
└── 各自链接独立统计，互不影响

场景二：安全分享
├── 创建临时分享链接（24小时或指定时间过期）
├── 使用完毕后可立即禁用
└── 若链接泄露，可刷新Token使旧链接失效

场景三：访问追踪
├── 不同分享链接对应不同来源
├── 通过访问日志了解各链接的使用情况
└── IP 地理位置自动识别，了解用户分布
```

---

## 升级说明

> [!TIP]
> **默认分享**：系统升级后会自动为每个订阅创建一个「默认」分享链接，保持原有链接可用，确保平滑升级。

> [!NOTE]
> **客户端兼容**：分享链接支持自动识别客户端类型，也可手动指定 Clash、Surge、V2ray 等客户端格式。

## 扩展客户端格式

- 原生输出仍通过 `/c?client=clash`、`/c?client=mihomo`、`/c?client=surge`、`/c?client=v2ray` 提供。
- 在 **用户中心 -> Sub-Store** 启用 sidecar 并选择目标后，分享链接还可以请求已选中的扩展目标：`loon`、`egern`、`stash`、`surfboard`、`shadowrocket`、`quanx`、`sing-box`、`uri`、`json`。
- 扩展输出使用 SublinkPro 的 mihomo/Clash YAML 作为桥接格式，并调用外部 Sub-Store sidecar。sidecar parser 只转换代理节点，不保留策略组、规则、DNS 等完整 Clash 配置段。
- 如果请求扩展客户端时未配置 sidecar，或目标未在用户中心选中，请求会返回明确错误，不会静默回退到 V2ray。

## 订阅更新间隔

- 在订阅管理的「订阅设置」->「基础设置」中，可为每个订阅配置「更新间隔（小时）」。
- 该值按小时保存，最大为 `8760` 小时；设置为 `0` 或不填写时使用默认更新间隔：Clash 为 `24` 小时，Surge 为 `86400` 秒。
- 当客户端通过订阅链接获取 Clash 配置时，响应头会带上 `profile-update-interval`，单位为小时。
- 当客户端获取 Surge 配置时，`#!MANAGED-CONFIG` 中的 `interval` 会按设置自动换算为秒。

## 节点选择来源

- 手动选择节点会保存具体节点 ID，订阅输出会保留这些指定节点，直到再次编辑订阅。
- 动态选择分组会保存分组名称，每次生成订阅时解析这些分组下的当前节点。
- 动态选择机场会保存机场 ID，每次生成订阅时解析这些机场当前导入的节点。
- 混合模式可以同时组合手动节点、动态分组和动态机场。输出顺序遵循已配置的排序，同名有效节点只保留最先出现的一份。

## 节点命名变量

`NodeNameRule` 控制订阅输出时的节点名称。留空时，SublinkPro 会保留节点的实际使用名称。变量会在生成订阅时替换，因此速度、延迟、国家、标签、解锁状态等值都来自系统当前保存的节点数据。

常用变量：

| 变量 | 含义 |
|:---|:---|
| `$Name` | 节点实际使用名称，取决于节点的名称模式 |
| `$LinkName` | 上游原始节点名称 |
| `$LinkCountry` | 节点国家代码，例如 `HK`、`US`；国家为空时显示 `未知` |
| `$LinkCountryName` | 根据 `$LinkCountry` 到「应用设置 -> 国家规则」中查找得到的国家名称；没有对应国家规则名称时回退为国家代码 |
| `$Flag` | 根据国家代码生成的国旗 Emoji |
| `$Group` | 节点分组；分组为空时显示 `未分组` |
| `$Source` | 节点来源；手动节点显示为 `手动` |
| `$Protocol` | 协议类型 |
| `$Index` | 输出序号 |
| `$DuplicateIndex` | 重名序号；第一次出现为空，后续依次为 `1`、`2`、`3` 等 |
| `$Tags` | 节点所有标签，使用 `|` 连接 |
| `$TagGroup(name)` | 节点在指定标签组中的标签，存在时才输出 |
| `$Speed`、`$SpeedIcon` | 下载速度文本和速度状态图标 |
| `$Delay`、`$DelayIcon` | 延迟文本和延迟状态图标 |
| `$IpType`、`$Residential` | IP 质量标签，例如原生/广播、住宅/机房 |
| `$FraudScore`、`$FraudScoreIcon` | 欺诈评分和欺诈评分图标 |
| `$Unlock` | 解锁摘要 |
| `$Unlock(provider)` | 指定服务商的解锁结果，例如 `$Unlock(netflix)` |
| `$UnlockStatus`、`$UnlockLabel`、`$UnlockRegion` | 可用时输出更细的解锁状态字段 |

国家名称逻辑依赖节点已经保存的国家代码。例如节点的 `$LinkCountry = HK`，并且 `HK` 国家规则的国家名称是 `香港`，那么 `$LinkCountryName` 会输出 `香港`；如果这个国家代码没有对应国家规则，则 `$LinkCountryName` 会输出 `HK`。如果希望调整命名规则中的国家名称，请修改「应用设置 -> 国家规则」里的国家名称。

示例规则：

```text
[$Flag] $LinkCountryName - $LinkName $DuplicateIndex
```

对于名为 `Premium 01` 的香港节点，输出可能是 `[🇭🇰] 香港 - Premium 01`。如果后续节点生成了同名结果，`$DuplicateIndex` 可以为后续重名节点追加序号。

## Mieru 输出说明

- Clash/mihomo 订阅将 Mieru 节点输出为 YAML，支持 `server`、`port` 或 `port-range`、`transport`、登录凭据及可选的 `multiplexing`、`handshake-mode`、`traffic-pattern` 和 `dialer-proxy`。
- 添加节点支持两种 [Mieru 官方格式](https://github.com/enfein/mieru/blob/v3.39.0/docs/client-install.zh_CN.md#分享客户端的设置)。`mieru://` 携带 Base64 编码的 Protobuf 客户端配置，`mierus://` 使用 `port=2090-2099&profile=default&protocol=TCP` 等参数。每条链接对应一个 SublinkPro 节点。Clash/mihomo 导出要求单服务器、单一传输协议，以及单端口或一个连续端口范围。同一协议下连续的端口绑定可合并为一个范围，无法满足这些条件的节点会跳过。编辑原始字段时会保留分享格式和其他绑定。
- 应用启动升级时，会将可转换的旧版 SublinkPro 链接迁移为 `mierus://`，原节点 ID 保持不变。迁移日志会标明需要修正链接的节点。
- `/c?client=v2ray` 与 Surge 当前不支持 Mieru；SublinkPro 会跳过 Mieru 节点，不会把 `mieru://` 链接写入 v2ray base64，也不会生成 Surge 配置。

## Snell 输出说明

- Snell 支持 Clash/mihomo 与 Surge 输出。`/c?client=clash` 会按 mihomo YAML 字段输出 `type: snell`、`server`、`port`、`psk`，并保留可选的 `version`、`udp` 与 `obfs-opts`（`mode` / `host`）、通用连接层选项（`tfo`、`mptcp`、`interface-name`、`routing-mark`、`ip-version`）以及链式代理 `dialer-proxy`；`/c?client=surge` 会输出 `snell, server, port, psk=...`，并按需追加 `version`、`obfs`、`obfs-host`、`tfo` 与 `udp-relay`（`mptcp`、`interface-name`、`routing-mark`、`ip-version` 为 mihomo 专属，Surge 无对应字段不输出）。
- Snell 官方没有定义通用的分享链接方案，mihomo 以 Clash YAML 字段描述 Snell。SublinkPro 内部使用 `snell://server:port?psk=...&version=...&obfs=...&obfs-host=...#name` 作为原始编辑和 Clash/mihomo、Surge 导入回写格式；`version` 默认为 mihomo 的 Snell v1，取值范围为 1/2/3。
- `/c?client=v2ray` 当前不支持 Snell；SublinkPro 会跳过 Snell 节点，不会把 `snell://` 链接写入 v2ray base64。

## OpenVPN 输出说明

- OpenVPN 仅支持 Clash/mihomo 输出；`/c?client=clash` 会输出 `type: openvpn`，并保留导入的 Mihomo 字段，包括 PEM/密钥块、账号凭据、传输与加密选项、保活、IP 栈、DNS、通用连接层选项及链式代理 `dialer-proxy`。
- SublinkPro 使用 `openvpn://server:port?...#name` 仅作为内部原始编辑与 Clash/mihomo 导入导出的往返格式；它不是 OpenVPN 官方分享链接，也不支持直接导入 `.ovpn` 文件。
- `/c?client=v2ray` 与 Surge 会跳过 OpenVPN 节点。敏感密钥材料在内部链接中仅经过 URL 编码，并未加密。

## VLESS XHTTP 输出说明

- 当订阅中的节点为 VLESS 且传输层为 `xhttp` 时，`/c?client=clash` 会输出 `network: xhttp` 与 `xhttp-opts`。
- 如果 VLESS URL 带有 `encryption`，`/c?client=clash` 会保留为 mihomo 顶层 `encryption` 字段。
- `/c?client=v2ray` 会继续输出 VLESS URL，并保留 `type=xhttp`、`path`、`host`、`mode` 与 `extra`。
- 当顶层 VLESS `ech` 为 Xray 的 DNS / URI 风格时，`/c?client=clash` 会按 mihomo 可表达的范围输出顶层 `ech-opts`，其中可识别的查询域名会映射到 `query-server-name`。
- 反过来，当节点来源于 Clash/mihomo YAML 导入且只有 `ech-opts.query-server-name` 可恢复时，系统会在保存节点链接前按本地兼容规则补成 `ech=<query-server-name>+https://dns.alidns.com/dns-query`。
- 为避免生成表面可用但实际失真的配置，系统不会把 `xhttp` 静默转换成 `http`、`h2` 或 `grpc`。

## 分享链接设备限额

新建订阅默认不限设备（上限为 0）。管理员可以按订阅或分享主动开启限额，设置 0–10000；0 表示不限设备数量。新建分享（包括默认链接、批量分享）继承订阅默认值，之后各自独立计数。修改订阅默认值不会修改已有分享。升级前已有的订阅与分享保持不限设备，需管理员主动启用。

建议为每位接收者创建独立分享，分别设置到期时间。新设备第一次成功下载时自动绑定；相同设备后续更新不重复占用名额。设备上限大于零时必须发送有效且唯一的 `X-HWID` 请求头，不支持用 IP、UA、型号或链接参数代替。限额适用于发送该请求头的任意客户端，包括兼容版本的 Karing、Happ、v2RayTun 和 Clash Mi；不检查客户端品牌，也不要求 Karing UA。超限、缺少或非法标识、设备被撤销均返回 HTTP 403，不下发节点；数据库检查失败返回 503。

### 客户端导入步骤

导入或刷新启用限额的分享前，请开启客户端的设备标识 / `X-HWID` 选项。例如，在 Karing 中打开“添加配置 → 添加配置链接”，粘贴分享链接，打开 **X-HWID** 开关，再保存；已有配置也需在编辑界面开启该项。其他客户端的选项名称或默认状态可能不同，应检查所安装版本实际发送的请求头。Clash YAML 与 Base64 输出使用相同设备策略。先交付给接收者本人导入，未绑定的链接提前转发可能被别人抢占名额。

管理员可在分享列表查看“设备 已绑定数 / 上限”，进入“管理设备”修改备注、撤销或恢复设备。撤销释放名额，但原标识持续被拒绝；恢复需要剩余名额。降低上限前应先撤销多余设备。普通刷新 Token 保留设备绑定；“全部换机重置”会撤销全部登记设备并更换链接，需将新链接交付给客户的新设备。

批量修改默认保持原值，由用户明确选择要修改的设置；修改设备限额不会修改 Token、过期时间或启用状态。每条分享独立提交。某个分享因已绑定设备过多而无法降低上限时，其他分享仍可能修改成功。API 保留既有 HTTP 状态和消息，并返回逐项结果与成功、失败数量。部分失败后页面会刷新列表，显示失败分享及原因，仅保留失败条目供重试。

数据库迁移和 WebDAV 恢复包含已绑定及已撤销的设备记录。覆盖恢复时，目标库旧设备记录与分享一同替换，避免复用分享 ID 时串入其他分享的设备。设备限额功能之前生成的旧备份可恢复，但不包含设备绑定。

### 反向代理与 HTTP 403 排查

`/c/` 链路上的每层代理或自定义订阅转发服务都必须保留 `User-Agent`、`X-HWID`、`X-Device-OS`、`X-Device-Model`。请求头白名单若丢弃 `X-HWID`，即使客户端开启了开关，也会返回 `hwid_required`。应转发客户端原有的唯一 HWID，不要生成替代标识或用 IP 代替。重复 HWID 请求头和代理合并后的逗号分隔值都会被作为 `hwid_invalid` 拒绝。不要跨设备缓存订阅响应。

验证时应使用完整的公开订阅地址，不能只检查 `/api/v1/version` 或后端端口。可用设备标识的 `HEAD` 请求应成功且不绑定设备；开启设备上限时，同一请求去掉 `X-HWID` 必须返回 403。名额已满时，未绑定的标识也会被拒绝。

响应 JSON 会区分 `hwid_required`、`hwid_invalid`、`device_limit_exceeded` 和 `device_revoked`。部分客户端只显示通用 HTTP 错误。例如，iOS Karing 1.2.22.2502 的固定弹窗可能显示 `http statusCode: 403` 并建议修改 User-Agent，不能说明服务端的实际拒绝原因。调整设置前请先核对已绑定数量、上限、响应错误码及 X-HWID 选项。

### 能力与验证边界

此功能限制订阅下载，不是不可伪造的硬件认证。未绑定的链接在本人导入前被转发时，可能被其他人抢占。撤销设备只影响后续下载，已下载的节点仍可使用。导出的节点、伪造的 HWID 和网络共享无法仅靠此功能阻止。关闭数量限制（设为 0）也停止 HWID 和撤销状态检查，沿用普通下载路径，不执行设备登记事务；发送响应前仍复核 Token 和分享状态。`HEAD` 不占用名额；订阅生成失败不绑定设备。

原 PR 的历史验证使用 Windows Karing 1.2.23+2606，完成真实首次导入与多次手动更新请求：开启开关后发送 `X-HWID`，相同设备的摘要保持一致，UA 以 `Karing/1.2.23.2606 platform/windows` 开头。该历史记录不能证明自动刷新、其他客户端或平台，以及应用重启、升级、网络切换、备份恢复后的表现。模拟请求测试不是实体客户端验证。重装或恢复备份若导致标识改变，需管理员办理换机。

历史后端联调中，Windows Karing 成功下载并登记设备；管理员撤销后，客户端显示 `更新失败:download profile failed: http statusCode: 403`，测试设备随后已恢复。

#### 2026-11-20 验证记录

Karing 实机成功下载数量固定截至 `2026-11-20T21:06:58+08:00`；之后的撤销检查记录截至 `2026-11-20T21:11:43+08:00`，被拒绝的请求不加入成功数量。后续 v2rayN 撤销检查点截至 21:53:18；恢复测试绑定并修正模板后，两组客户端更新时间均推进到 21:58:46，服务端最近成功访问为 21:58:47。时间均为 2026-11-20 的 UTC+08:00。

| 环境 | 检查项 | 实际结果 |
|:---|:---|:---|
| 使用 Karing 1.2.23.2606 和未带版本号的 Happ UA 的模拟 HTTP 请求 | Clash YAML 与 Base64 输出的首次下载、重复下载及撤销后刷新 | 首次和重复请求均返回 200，保持一条绑定；撤销后相同标识返回 403，错误码为 `device_revoked` |
| Windows，隔离的 MySQL 8.4.13 实例 | 10 轮，每轮 24 个并发登记请求争抢一个剩余名额 | 每轮成功一台，其余 23 台返回 `device_limit_exceeded`，实际落库一台，数据库错误为零 |
| Windows，隔离的 PostgreSQL 17.12 实例 | 10 轮，每轮 24 个并发登记请求争抢一个剩余名额 | 每轮成功一台，其余 23 台返回 `device_limit_exceeded`，实际落库一台，数据库错误为零 |
| 用户操作的 Windows Karing 1.2.26.2931 | Clash YAML 首次、重复手动及自动下载 | 三条真实 GET 均发送 `X-HWID`、返回 200，撤销前复用一条绑定。用户确认第二次为手动刷新；第三次在已要求不手动刷新的观察期间到达 |
| 用户操作的 Windows Karing 1.2.26.2931 | Base64 首次、重复及自动下载 | 六条真实 GET 均发送 `X-HWID`、返回 200，撤销前复用一条绑定。用户明确确认第三次请求是自动更新，之后也在已要求不手动刷新的观察期间收到请求 |
| 用户操作的 Windows Karing 1.2.26.2931 | 撤销两台设备后刷新 | 一条真实 YAML GET、三条真实 Base64 GET 均发送 `X-HWID` 并返回 403。两条绑定均已撤销，两个分享的有效设备数均为零，成功访问计数和时间均未增加 |
| 用户操作的 Windows v2rayN 7.25.4 | Base64 首次及自动导入 | 用户报告成功；客户端已保存一条配置，自动更新间隔为一分钟，数据库更新时间已推进。使用手工固定的 `X-HWID` |
| 用户操作的 Windows v2rayN 7.25.4 | 修正测试模板后的 Clash YAML 导入 | 用户确认导入成功；客户端保存了一条核心类型为 13 的自定义配置。两种输出分组的自动更新间隔均为一分钟、更新时间均为 21:58:46，另有 21:57:46 和 21:58:46 的自动更新日志 |
| 用户操作的 Windows v2rayN 7.25.4 | 撤销共用设备绑定后刷新 | 真实客户端请求返回 HTTP 403；保存的检查点中有效设备为零、成功访问计数仍为八。用户提供了 YAML 分组错误，服务日志未区分输出格式 |

早期助手操作使用 Karing 1.2.26.2934 便携副本和隔离的 v2rayN，未完成实机下载。之后用户操作现有的 Karing 1.2.26.2931，产生了上表中的真实 GET 证据。YAML 访问计数为四，包含一次较早的诊断请求，因此实际客户端下载为三次。

模拟 HTTP 测试验证服务端契约。实机证据现已确认 Karing 两种输出的首次、重复、自动下载及撤销后拒绝均通过。YAML 自动请求发生在用户确认的手动刷新约五分六秒后，来自随后已要求不手动刷新的持续观察期间，且用户确认已开启五分钟设置；用户未单独说明这次请求的来源。这些观察不保证精确五分钟调度。真实撤销日志只记录了 HTTP 403，未检查响应 JSON 错误码；响应体中的 `device_revoked` 仅由模拟 HTTP 测试验证。

之后用户自行断开网络，操作现有 v2rayN 测试，助手核对日志与数据库。两种输出 URL 使用相同 Token、手工固定的 `X-HWID`。合计七条真实 GET 返回 200，共一台有效绑定；访问计数为八，包含一次较早的诊断请求。日志未记录输出查询参数，不能将这些请求按格式分配，也不能视为已验证 v2rayN 原生硬件标识。

Base64 已保存的配置和用户自动更新成功报告支持首次及自动导入通过。原 YAML 测试模板使用 `port: 7890`，未通过客户端识别；测试服务改用 `mixed-port: 17891` 后，用户确认导入成功，客户端已保存自定义配置。后续定时更新日志及两组更新时间推进支持自动刷新，共用服务端绑定仍为一台设备。修正模板前，撤销已产生真实客户端 HTTP 403，有效绑定数为零、成功访问计数未增加；这些拒绝请求未按格式计数，因此不宣称已分别完成 Base64 与 YAML 撤销验证。为测试修正版模板，绑定随后已恢复。

本轮未安装和实测 Clash、Clash Verge、Clash Mi、Happ、v2RayTun 等客户端。Karing 或 v2rayN 成功导入 Clash YAML 不等于验证这些应用的兼容性。其他平台、v2rayN 原生硬件标识，以及重启、升级、网络切换和客户端备份恢复等场景仍未验证。

## VLESS REALITY 兼容说明

导入 Clash / mihomo YAML 时，`reality-opts.support-x25519mlkem768` 会经由内部存储的 VLESS URL 保留到订阅输出、延迟检测及下载测速使用的 Proxy。URL 参数为 `support-x25519mlkem768=true` 或 `false`（输入也接受 `1` / `0`）。显式 `false` 与未指定保持区分；空值或无效 URL 值视为未指定。普通 TLS 节点不会因该参数而新增 REALITY 配置。在节点字段编辑器中可选择 `true`、`false` 或空选项，空选项表示恢复上游默认行为。

VLESS、Trojan 分享链接中的 REALITY `spx`（`spiderX`，爬虫初始路径）和 `pqv`（`mldsa65Verify`，额外的 ML-DSA-65 证书签名验证公钥）会在解析、字段编辑和 URL 导出时保留。编辑器对应字段为 `Query.SpiderX`、`Query.Mldsa65Verify`；清空字段会移除对应 URL 参数。参数值采用 URL 编码，因此爬虫路径可以携带查询参数。

Clash YAML 导入会将 `reality-opts.spider-x`、`reality-opts.mldsa65-verify` 映射到内部 URL，`/c?client=clash` 或 `mihomo` 再将其还原到 YAML，保持参数值不丢失。非空字段仅为 REALITY 节点输出，空字段省略。Trojan REALITY 导入时还会保留 `security=reality`、`pbk` 和 `sid`，支持完整往返。

Xray 客户端（如 v2rayN、v2rayNG）使用 `/c?client=v2ray`，无需新增 `xray` 客户端类型。保留配置不代表内核新增了执行能力；当前上游 mihomo 没有对应验证能力，内部检测仍沿用现有行为；导出配置中的额外验证需由兼容内核执行。Surge 会跳过 `pqv` 非空的 VLESS REALITY 节点及 Trojan REALITY 节点；仅有爬虫参数 `spx` 时，Surge 输出会省略该参数。ML-DSA-65 签名验证与 `support-x25519mlkem768` 密钥交换选项相互独立。

## Surge 的 VLESS 支持

`/c?client=surge` 可为 Surge iOS 6.0.0+ / Mac 6.12.0+ 导出 VLESS。支持 TCP/RAW、WebSocket、HTTPUpgrade、gRPC（TLS 或 REALITY）及 XHTTP/SplitHTTP。TLS、REALITY 密钥、Vision、VLESS Encryption、SNI、ALPN、证书指纹、传输请求头和受支持的 XHTTP extra/XMUX 参数会映射为 Surge 原生配置。订阅的 Host 替换及跳过证书校验设置同样生效。Surge 自动使用 XUDP。

不支持的传输、ECH、TCP 请求头混淆、不兼容的 Vision/REALITY 组合及含 XHTTP `downloadSettings` 的节点会被跳过，也不会加入代理组。uTLS `fp` 和 REALITY `support-x25519mlkem768` 无对应的 Surge 参数，因此不输出；Surge 自行处理 ClientHello 和 ML-KEM 回退。

参见 [Surge 官方 VLESS 文档](https://manual.nssurge.com/policies/vless.html)。
