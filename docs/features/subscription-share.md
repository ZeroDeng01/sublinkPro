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

## 订阅更新间隔

- 在订阅管理的「订阅设置」->「基础设置」中，可为每个订阅配置「更新间隔（小时）」。
- 该值按小时保存，最大为 `8760` 小时；设置为 `0` 或不填写时使用默认更新间隔：Clash 为 `24` 小时，Surge 为 `86400` 秒。
- 当客户端通过订阅链接获取 Clash 配置时，响应头会带上 `profile-update-interval`，单位为小时。
- 当客户端获取 Surge 配置时，`#!MANAGED-CONFIG` 中的 `interval` 会按设置自动换算为秒。

## Mieru 输出说明

- Mieru 当前仅支持 Clash/mihomo 输出；`/c?client=clash` 会按 mihomo YAML 字段输出 `type: mieru`、`server`、`port` 或 `port-range`、`transport`、`username`、`password`，并保留可选的 `multiplexing`、`traffic-pattern` 与链式代理 `dialer-proxy`。
- Mieru 官方存在 `mieru://` / `mierus://` 分享链接，但官方文档未定义适合逐字段编辑的通用 URL schema。SublinkPro 内部使用 `mieru://username:password@server:port?...#name` 作为原始编辑和 Clash/mihomo 导入回写格式；需要端口范围时使用 `portRange=2090-2099`，不写 `port`。
- `/c?client=v2ray` 与 Surge 当前不支持 Mieru；SublinkPro 会跳过 Mieru 节点，不会把 `mieru://` 链接写入 v2ray base64，也不会生成 Surge 配置。

## VLESS XHTTP 输出说明

- 当订阅中的节点为 VLESS 且传输层为 `xhttp` 时，`/c?client=clash` 会输出 `network: xhttp` 与 `xhttp-opts`。
- 如果 VLESS URL 带有 `encryption`，`/c?client=clash` 会保留为 mihomo 顶层 `encryption` 字段。
- `/c?client=v2ray` 会继续输出 VLESS URL，并保留 `type=xhttp`、`path`、`host`、`mode` 与 `extra`。
- 当顶层 VLESS `ech` 为 Xray 的 DNS / URI 风格时，`/c?client=clash` 会按 mihomo 可表达的范围输出顶层 `ech-opts`，其中可识别的查询域名会映射到 `query-server-name`。
- 反过来，当节点来源于 Clash/mihomo YAML 导入且只有 `ech-opts.query-server-name` 可恢复时，系统会在保存节点链接前按本地兼容规则补成 `ech=<query-server-name>+https://dns.alidns.com/dns-query`。
- 为避免生成表面可用但实际失真的配置，系统不会把 `xhttp` 静默转换成 `http`、`h2` 或 `grpc`。
