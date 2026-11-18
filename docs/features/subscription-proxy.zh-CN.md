[English](subscription-proxy.md) | 简体中文

# 代理订阅

订阅管理现在有两种订阅：

| 标签页 | 类型 | 返回内容 |
|:---|:---|:---|
| **聚合订阅** | `collection`（默认） | SublinkPro 从手动节点、动态分组和机场中组装节点，经过过滤、重命名并用模板渲染后输出 |
| **代理订阅** | `proxy` | 所选机场的上游订阅内容，原样返回 |

代理订阅适用于节点集合不由自己掌控的场景：希望把已有的上游订阅通过 SublinkPro 的分享链接对外提供——稳定的分享链接、过期策略和访问记录，同时不解析、不改写上游内容。只需选择机场，转发时复用该机场的地址与请求配置。不启用 IP 黑白名单，每个分享只提供一个订阅链接。

---

## 工作原理

```text
客户端  --GET /c/?token=...-->  SublinkPro
                                  |  分享校验、过期检查
                                  |  记录访问
                                  v
                              机场（地址、客户端 UA、请求头、
                              设备信息、代理、TLS）
                                  |  HTTP(S)
                                  v
                              上游服务商
                                  |
                                  v
客户端  <--上游响应体 + 状态码--  SublinkPro
```

1. 客户端使用与聚合订阅相同的分享链接请求（`/c/?token=...`）。
2. SublinkPro 校验分享、检查过期时间，并记录访问。
3. SublinkPro 加载订阅关联的机场，使用该机场的订阅地址、请求客户端的 User-Agent（仅在客户端未发送 UA 时回落到机场生效的 User-Agent）与机场请求头（包括设备信息请求头和自动派生的 HWID，除非已禁用）、可选代理和 TLS 跳过来请求上游。Remnawave 等面板按 User-Agent 选择返回格式，因此转发链接会返回该客户端直连机场时得到的相同格式。
4. 将上游状态码和一组成员白名单响应头复制到响应中，响应体原样转发。

除 User-Agent 透传外，转发与机场定时更新使用完全相同的配置，因此修改机场的地址、请求头、设备信息、代理或 TLS 设置后，代理订阅的输出会立即随之变化；机场的 User-Agent 仅作为客户端未发送 UA 时的回退值。

不解析、不过滤、不重命名、不重新渲染。客户端收到的内容与上游返回的字节完全一致；由上游根据透传的 User-Agent 决定返回哪种客户端特定格式（Clash、Surge、sing-box、Xray JSON、base64、纯链接列表等）。

---

## 配置项

在订阅管理页面的 **代理订阅** 标签页中创建。

| 字段 | 必填 | 说明 |
|:---|:---|:---|
| `name` | 是 | 订阅名称，两种类型之间也要求唯一 |
| `AirportID` | 是 | 需要原样转发的机场；订阅地址、User-Agent、请求头、设备信息、代理和 TLS 设置均取自该机场 |

分享链接、过期策略和访问记录与聚合订阅一致，仍在该订阅行的操作中配置。代理订阅不启用 IP 黑白名单；由于原样转发时 client 参数没有意义，分享链接弹窗只提供一个订阅链接（`/c/?token=...`）。

被任意代理订阅引用的机场无法删除：SublinkPro 会拒绝删除请求，并告知被多少个代理订阅使用。请先修改或删除这些代理订阅。

### 转发的响应头

仅转发携带订阅元信息的响应头：

- `Content-Type`
- `Content-Disposition`
- `subscription-userinfo`
- `profile-update-interval`
- `profile-web-page-url`
- `profile-title`
- `encrypt-tag`

转发 `encrypt-tag` 是为了让 Happ 能解密上游返回的加密订阅内容。

`Content-Encoding` 有意不转发：HTTP 客户端已自动解压上游响应体，再转发原始编码头会导致客户端解析失败。

---

## 错误处理

| 场景 | 响应 |
|:---|:---|
| 订阅未关联机场 | `502`，响应体为 `Subscription is not linked to an airport` |
| 关联的机场已不存在 | `502`，响应体为 `The airport linked to this subscription does not exist` |
| 机场订阅地址为空或不是 HTTP(S) | `502`，响应体为 `Invalid airport subscription URL` |
| 上游返回非 2xx | `502`，响应体为通用的 `Failed to fetch upstream subscription`；具体状态码与原因写入服务端日志 |
| 分享已过期 | `410`，响应体为 `Subscription expired` |
| 上游不可达或超时 | `502`，响应体同样为通用提示 |
| 响应体超过 10 MiB | `502`；读取有上限，避免占用过多服务端内存 |

上游地址、Token 和请求头内容不会回显给客户端。

> [!CAUTION]
> 上游由服务端网络发起请求，因此能创建代理订阅的人可以让服务器访问内网地址。请仅配置可信地址，并妥善保护管理员账号。

---

## 限制

- 代理订阅没有节点、分组、脚本、过滤或命名规则；这些聚合订阅的界面在代理订阅标签页中隐藏。
- 请求配置不能按订阅单独覆盖；请直接修改机场，所有引用该机场的代理订阅会自动生效。
- 节点预览、节点排序和链式代理规则对代理订阅不适用。
- 请求 `/c/` 时的 `client` 和 `substore` 参数会被忽略，格式由上游根据透传的 `User-Agent` 决定。
- 不支持 IP 黑白名单：`IPWhitelist` 和 `IPBlacklist` 会被忽略，并在保存时清空（历史数据同样不生效）。
- 每次客户端请求都会实时拉取上游，没有缓存层。

---

## API

```bash
# 创建
curl -X POST "http://localhost:8000/api/v1/subcription/add" \
  -H "X-API-Key: <key>" \
  -F "type=proxy" \
  -F "name=Upstream" \
  -F "AirportID=1" \
  -F "RefreshUsageOnRequest=false"

# 仅列出代理订阅
curl "http://localhost:8000/api/v1/subcription/get?page=1&pageSize=20&type=proxy" \
  -H "X-API-Key: <key>"

# 公开订阅输出（使用分享 Token，而非 API Key）
curl "http://localhost:8000/c/?token=<share-token>"
```

完整字段列表见 `skill-sublinkpro/reference/api.md`。
