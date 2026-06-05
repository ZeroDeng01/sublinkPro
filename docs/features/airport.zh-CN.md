[English](airport.md) | 简体中文

# 机场订阅管理

SublinkPro 提供了完善的机场订阅管理功能，不仅能将订阅转换为节点，还能全方位监控和管理您的机场服务。

---

## 💡 核心功能

| 功能 | 说明 |
|:---|:---|
| **📥 多格式导入** | 支持 Clash/mihomo、V2Ray 订阅格式的自动解析与导入；Mieru 仅支持 Clash/mihomo YAML |
| **⏱️ 智能定时更新** | 内置 Crontab 级调度器，支持按时间间隔或 Cron 表达式自动更新订阅，确保节点时刻在线 |
| **📊 流量用量监控** | 自动解析订阅返回的 `Subscription-Userinfo` 头，直观展示**已用上传**、**已用下载**、**总流量**及**过期时间** |
| **🚀 立即更新机制** | 支持一键「立即拉取」，配合实时回调机制，无需刷新页面即可看到最新的流量数据和节点列表 |
| **⚡ 更新后检测** | 机场可绑定节点检测策略，订阅更新成功后立即针对该机场节点执行检测 |
| **🤖 Bot 集成管理** | 通过 Telegram Bot 可随时查询各订阅的剩余流量、到期时间，并支持远程触发更新任务 |

### Clash / mihomo `proxy-providers` 兼容说明

- 机场订阅导入支持 Clash / mihomo YAML 中的 `proxy-providers`，当顶层没有 `proxies` 但 provider 中包含节点时，也会展开并导入为普通节点。
- 当前会远程拉取 `type: http` 的 provider，并复用机场订阅本身的代理下载和忽略 TLS 证书验证等行为；provider 请求始终携带 User-Agent，只有 provider URL 与根订阅 URL 同 host 时才会继承自定义 Header，跨 host provider 和跨 host 跳转不会携带自定义 Header。
- 系统优先展开 `proxy-groups[].use` 引用到的 provider；当分组启用 `include-all` / `include-all-providers`，或配置中没有任何分组引用时，会按配置声明顺序展开所有 HTTP provider。
- Provider 响应支持标准 YAML 顶层 `proxies`；如果 provider 返回的是 base64 或明文 URI 行列表，也会按普通订阅链接兼容解析。
- 当前不会实现本地 `file` provider、provider 缓存、定时健康检查、完整 mihomo override 语义或 provider 自身的 `proxy` 路由；这些运行时能力仍应由 Clash / mihomo 客户端处理。

### VLESS / XHTTP 兼容说明

- 机场订阅导入现已支持 `vless://` 链接中的 `type=xhttp`。
- 当上游订阅是 Clash / mihomo YAML 且节点为 `type: vless`、`network: xhttp` 时，系统会识别 `xhttp-opts` 并回写为 VLESS URL。
- 当前已支持的 URL 顶层字段包括：`type`、`encryption`、`path`、`host`、`mode`、`extra`、`ech`。
- `extra` 中已支持映射到 mihomo 的字段包括：`headers`、`noGRPCHeader`、`xPaddingBytes`、`downloadSettings` 及其已知子字段。
- 顶层 `ech` 会优先映射到 mihomo 顶层 `ech-opts`：当值是固定 base64 ECHConfig 时写入 `config`，当值是 Xray 的 DNS / URI 风格时会按 mihomo 可表达的范围做最佳努力映射。
- 当机场订阅本身是 Clash/mihomo YAML，且导入时只能从顶层 `ech-opts` 恢复出 `query-server-name` 时，系统会在保存节点链接前按本地兼容规则重建为 `ech=<query-server-name>+https://dns.alidns.com/dns-query`。
- `extra.downloadSettings.echOpts` 仍只映射到 mihomo `xhttp-opts.download-settings.ech-opts`，不会和顶层 `ech-opts` 混写。
- `xmux`、`sessionPlacement` 等在 Xray 侧存在但 mihomo 当前没有公开承载字段的扩展项，会被视为未支持，不会静默降级成 `http`、`h2` 或 `grpc`。

### Mieru 兼容说明

- 机场订阅导入支持 Clash/mihomo YAML 中的 `type: mieru` 节点，并保留 mihomo 官方字段：`server`、`port` 或 `port-range`、`transport`、`username`、`password`、`multiplexing`、`traffic-pattern`。
- Mieru 官方存在 `mieru://` / `mierus://` 分享链接，但未定义适合 SublinkPro 原始编辑器逐字段修改的通用 URL schema。系统保存节点时使用内部可编辑形态 `mieru://username:password@server:port?...#name`，端口范围使用 `portRange=2090-2099`，用于 Clash/mihomo YAML 导入后的回写与后续导出。
- Mieru 不会输出到 v2ray 或 Surge；这些客户端当前不在 SublinkPro 的 Mieru 支持范围内。

---

## 📱 界面展示

系统在订阅列表中清晰展示了每个机场的详细状态，包括：
- 上次更新时间
- 下次计划更新时间
- 可视化的流量进度条

让您对机场使用情况一目了然。

---

## 使用流程

### 添加机场订阅

1. 进入「机场管理」页面
2. 点击「添加机场」
3. 填写订阅链接和名称
4. 按需配置请求设置（如 User-Agent、自定义 Header、代理下载）
5. 配置更新策略（可选）；如需订阅更新后立刻检测节点，可在高级选项开启「更新后检测」并选择节点检测策略
6. 保存并拉取节点

开启「更新后检测」后，只有当前机场订阅更新成功时才会触发检测；检测范围限定为该机场本次更新后的节点，不会等待节点检测策略自身的定时时间，也不会扩大到其它机场节点。由该功能触发的节点检测任务会在任务中心标记为「机场更新」，以便和「手动」「定时」触发区分。

### 节点处理：名称唯一化

在机场编辑弹窗的「节点处理（拉取时生效）」中，可使用“节点名称唯一化”相关配置：

- **节点名称唯一化**：为当前机场导入的节点统一添加稳定前缀，用于避免不同机场之间出现重名节点。
- **机场内节点名称唯一化**：在同一机场内如果存在重名节点，会在当前节点名称后依次追加 `-1`、`-2`、`-3` 等数字编号。

说明：

- 两个开关都在**拉取订阅时生效**，修改后需要重新拉取该机场才能应用到已存在节点。
- 机场间前缀唯一化与机场内顺序编号可以同时开启；此时系统会先生成机场前缀，再对同机场内的重名节点追加 `-1`、`-2`… 数字编号。
- 机场内顺序编号也可以单独开启；单独开启时不会添加机场前缀，只会在同机场的重名节点后追加数字编号。
- 编号是**按重名组分别计算**的，不是全机场共享一套连续序号；例如 `HK` 重名组会得到 `HK-1`、`HK-2`，而 `US` 重名组会单独从 `US-1` 开始。

### 节点名称与备注

机场拉取到的节点会同时保存两个名称：

- **原始名称**：来自机场订阅链接本身，机场更新时会随上游变化刷新。
- **备注名称**：用户在节点编辑弹窗中维护的自定义名称，机场更新不会覆盖该字段；全局节点备注不能重复。

在「节点管理」编辑节点时，可以通过「实际使用名称」选择订阅输出、节点重命名规则、链式代理和节点脚本使用哪个名称：

- **使用原始名称**：按机场订阅中的原始名称操作，适合希望完全跟随上游命名的节点。
- **使用备注名称**：按用户备注作为实际节点名称操作，适合自定义命名并希望机场更新后继续保留的节点。

切回「使用原始名称」即可恢复使用机场订阅里的名称；备注名称会保留，后续可随时再次切换回来。

说明：手动编辑备注时，如果与其他节点备注重复会被拒绝；机场自动导入产生重复备注时，系统优先生成 `原始名称@机场名称`，如果仍重复则继续追加编号，例如 `香港 01@机场B-2`。没有机场来源的自动导入会使用普通编号以保持全局唯一。

### 请求设置

机场的「请求设置」支持配置拉取订阅时附带的请求参数：

- `User-Agent`：使用专用输入框设置常见客户端 UA 或手动输入。
- **自定义 Header**：可按 `Header 名称` + `Header 值` 的方式添加多条请求头，适合需要额外鉴权或来源标识的机场。
- `使用代理下载`：通过指定节点或自动选择最佳节点拉取订阅。

说明：

- 自定义 Header 会在请求机场订阅地址时一并附带。
- 如果开启了「获取用量信息」，系统在刷新机场用量时也会复用相同的自定义 Header。
- `User-Agent` 使用单独字段管理，自定义 Header 中不支持再次填写 `User-Agent`。

### 配置定时更新

支持两种定时更新方式：

| 方式 | 说明 |
|:---|:---|
| **按间隔更新** | 设置固定时间间隔，如每 6 小时更新一次 |
| **Cron 表达式** | 灵活的 Cron 表达式配置，如 `0 */6 * * *` |

### 流量监控

系统自动解析订阅响应头中的 `Subscription-Userinfo`，提取以下信息：
- `upload`：已用上传流量
- `download`：已用下载流量
- `total`：总流量额度
- `expire`：到期时间戳

---

## Telegram Bot 集成

通过 Telegram Bot 可以：
- 查询各机场剩余流量
- 查看到期时间
- 远程触发订阅更新

详见 [Telegram 机器人文档](telegram-bot.zh-CN.md)
