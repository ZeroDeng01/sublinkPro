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

- Mieru 当前仅支持 Clash/mihomo 输出；`/c?client=clash` 会按 mihomo YAML 字段输出 `type: mieru`、`server`、`port` 或 `port-range`、`transport`、`username`、`password`，并保留可选的 `multiplexing`、`traffic-pattern` 与链式代理 `dialer-proxy`。
- Mieru 官方存在 `mieru://` / `mierus://` 分享链接，但官方文档未定义适合逐字段编辑的通用 URL schema。SublinkPro 内部使用 `mieru://username:password@server:port?...#name` 作为原始编辑和 Clash/mihomo 导入回写格式；需要端口范围时使用 `portRange=2090-2099`，不写 `port`。
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

## Karing 设备数量限制

新建订阅默认开启“仅允许 Karing”，默认设备上限为 1。可在后台设置 0–10000；0 表示不限设备数量。新建分享（包括默认链接、批量分享）继承订阅默认值，之后各自独立计数。修改订阅默认值不会修改已有分享。升级前已有的订阅与分享保持不限设备、不限制客户端，需管理员主动启用。

建议每位客户创建一个独立分享，分别设置年费到期时间。新设备第一次成功下载时自动绑定；相同设备后续更新不重复占用名额。设备上限大于零时必须发送有效的 `X-HWID` 请求头，不支持用 IP、UA、型号或链接参数代替。开启“仅允许 Karing”时还会校验 Karing UA。超限、缺少标识或设备被撤销均返回 HTTP 403，不下发节点；数据库检查失败返回 503。

### 客户导入步骤

在 Karing 中打开“添加配置 → 添加配置链接”，粘贴分享链接，打开 **X-HWID** 开关，再保存。已有配置也需在编辑界面开启该项。请保留客户端发送的 Karing UA。先交付给客户本人导入，首次绑定前转发链接可能被别人抢占。

管理员可在分享列表查看“设备 已绑定数 / 上限”，进入“管理设备”修改备注、撤销或恢复设备。撤销释放名额，但原标识持续被拒绝；恢复需要剩余名额。降低上限前应先撤销多余设备。普通刷新 Token 保留设备绑定；“全部换机重置”会撤销全部登记设备并更换链接，需将新链接交付给客户的新设备。

批量修改会逐个处理分享。某个分享因已绑定设备过多而无法降低上限时，其他分享仍可能修改成功。页面会在部分失败后刷新列表，请核对失败提示后再重试。

### 反向代理与 HTTP 403 排查

`/c/` 链路上的每层代理或自定义订阅转发服务都必须保留 `User-Agent`、`X-HWID`、`X-Device-OS`、`X-Device-Model`。请求头白名单若丢弃 `X-HWID`，即使客户端开启了开关，也会返回 `hwid_required`。应转发客户端原有的唯一 HWID；不要生成替代标识、用 IP 代替，或把重复的 HWID 请求头压成一个有效值。不要跨设备缓存订阅响应。

验证时应使用完整的公开订阅地址，不能只检查 `/api/v1/version` 或后端端口。有效 Karing UA 和可用设备标识的 `HEAD` 请求应成功且不绑定设备；开启设备上限时，同一请求去掉 `X-HWID` 必须返回 403。名额已满时，未绑定的标识也会被拒绝。

响应 JSON 会区分 `hwid_required`、`device_limit_exceeded` 和 `device_revoked`。Karing 可能只显示通用的 `http statusCode: 403` 并建议修改 User-Agent，这段文案不能说明服务端的实际拒绝原因。调整设置前请先核对已绑定数量、上限及响应错误码。仅修改服务器文案无法替换 iOS Karing 1.2.22.2502 的固定弹窗；请保留默认 Karing UA 并开启 X-HWID。

### 能力与验证边界

此功能限制订阅下载，不是不可伪造的硬件认证。导出的节点、伪造的 HWID、网络共享及已下载配置无法仅靠此功能阻止。关闭数量限制（设为 0）也停止 HWID 和撤销状态检查。`HEAD` 不占用名额；订阅生成失败不绑定设备。

Windows Karing 1.2.23+2606 已通过真实首次导入与多次手动更新请求验证：开启开关后发送 `X-HWID`，相同设备的摘要保持一致。其 UA 以 `Karing/1.2.23.2606 platform/windows` 开头。自动更新、应用重启、升级、网络切换、备份恢复，以及 Android、iOS/iPadOS、macOS、Linux、tvOS 等平台的实机兼容性尚未验证；不要将模拟请求测试视为全平台验证。重装或恢复备份若导致标识改变，需管理员办理换机。

真实后端联调也已验证：Windows Karing 成功下载并登记设备；管理员撤销后，客户端显示 `更新失败:download profile failed: http statusCode: 403`。测试设备随后已恢复。
