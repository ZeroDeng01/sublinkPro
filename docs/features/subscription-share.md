English | [简体中文](subscription-share.zh-CN.md)

# Subscription Share Management

The new subscription share management feature replaces the old single Token mode with safer and more flexible share link management.

---

## Core Features

| Feature | Description |
|:---|:---|
| **Multiple link management** | Each subscription can create multiple independent share links for different users or scenarios |
| **Secure Token** | Random secure Tokens are generated, with optional custom Tokens for easier memory |
| **Expiration policies** | Supports never expire, expire after a number of days, and expire at a specific time |
| **Independent statistics** | Each share link records its own access count and IP logs |
| **Enable/disable** | Enable or disable a single share link at any time without deleting it |
| **Token refresh** | Refresh Token with one click. Old links become invalid immediately |
| **QR code generation** | Generate a QR code for each share link for easy mobile import |

---

## ⏰ Expiration Policies

| Policy | Description |
|:---|:---|
| **Never expire** | Link remains valid until manually disabled or deleted |
| **Expire by days** | Link expires a specified number of days after creation, such as 7 or 30 days |
| **Expire at time** | Link expires at a specific date and time |

---

## 📋 Use Cases

```text
Case 1: Per user management
├── Create a share link for friend A, never expires
├── Create a share link for friend B, expires after 30 days
└── Each link has independent statistics

Case 2: Secure sharing
├── Create a temporary share link, 24 hours or specific expiry time
├── Disable it immediately after use
└── If the link leaks, refresh Token to invalidate the old link

Case 3: Access tracking
├── Different share links map to different sources
├── Use access logs to understand link usage
└── IP geolocation helps show user distribution
```

---

## Upgrade Notes

> [!TIP]
> **Default share**: After upgrade, the system automatically creates a “default” share link for each subscription, keeping old links available for a smooth upgrade.

> [!IMPORTANT]
> **Legacy token rotation**: Early versions derived default share tokens from `md5(subscription name)`, which made links predictable. On upgrade, migration `0044_rotate_predictable_share_tokens` automatically replaces such tokens with random ones. Links that used the old tokens stop working, so re-copy the affected share links from Share Management.

> [!NOTE]
> **Client compatibility**: Share links can detect client type automatically. Clash, Surge, V2ray, and other client formats can also be specified manually.

## Expanded Client Formats

- Native outputs remain available through `/c?client=clash`, `/c?client=mihomo`, `/c?client=surge`, and `/c?client=v2ray`. `client=clashmeta` and `client=clash-meta` are treated as aliases of `mihomo`, so they use the native template renderer instead of the Sub-Store sidecar.
- When **Application Settings -> Sub-Store** is enabled and targets are selected, share links can also request the selected expanded targets: `loon`, `egern`, `stash`, `surfboard`, `shadowrocket`, `quanx`, `sing-box`, `uri`, and `json`.
- Expanded outputs use SublinkPro's mihomo/Clash YAML as the bridge format and call the external Sub-Store sidecar. The sidecar parser converts proxy nodes only; strategy groups, rules, DNS sections, and other full Clash configuration sections are not preserved in those expanded outputs.
- If an expanded client is requested while the sidecar is not configured or the target is not selected in Application Settings, the request returns a clear error instead of silently falling back to V2ray.

## Subscription Update Interval

- In `Subscription Management -> Subscription Settings -> Basic Settings`, each subscription can configure “Update interval, hours”.
- The value is stored in hours, with a maximum of `8760` hours. If set to `0` or left empty, the default update interval is used: Clash uses `24` hours, Surge uses `86400` seconds.
- When a client fetches Clash config through a subscription link, the response header includes `profile-update-interval`, in hours.
- When a client fetches Surge config, `interval` in `#!MANAGED-CONFIG` is converted to seconds automatically according to the setting.

## Node Selection Sources

- Manual node selection stores the selected node IDs. The output keeps those specific nodes until the subscription is edited.
- Dynamic group selection stores group names and resolves the current nodes in those groups whenever the subscription is generated.
- Dynamic airport selection stores airport IDs and resolves the current nodes imported from those airports whenever the subscription is generated.
- Mixed mode can combine manual nodes, dynamic groups, and dynamic airports. Output order follows the configured sort order, and duplicate effective node names keep the first occurrence.

## Template Rendering Notes

- Clash/mihomo templates are edited on the template text itself. Comments, blank lines, indentation width, key order, anchors, and flow style stay byte-for-byte identical to the template; only the rewritten lists below change.
- Only two places are rewritten: the top-level `proxies` list and the `proxies` list of each template proxy group. Template groups keep their order and definitions, `include-all` / `use` / `filter` groups stay untouched, and `__ALL_PROXIES__` expands in place so references before and after it keep their position.
- Emoji and other characters outside the Basic Multilingual Plane are written literally, exactly as they appear in the template.
- Every Clash-family link (`client=clash`, `client=mihomo`, `client=clashmeta`, `client=clash-meta`) goes through this native path, so the same template comments, key order, and emoji survive whichever alias the client requests. To get the Sub-Store sidecar output instead (for example its `ClashMeta` target), add `substore=<target>` to the link, e.g. `/c/?token=...&client=clashmeta&substore=ClashMeta`; unsupported values are ignored and the link falls back to the native renderer.
- Templates the text merge cannot handle safely (flow-style `proxy-groups` or root mapping, a `proxies` value written as an alias, a missing top-level `proxies` key, CR-only line endings) fall back to rewriting the YAML node tree. Comments and key order still survive in that case, but indentation is re-emitted with two spaces and blank lines are dropped. Surge templates are merged line by line, so their comments and section order are preserved as well.

## Node Naming Variables

`NodeNameRule` controls the node name rendered in subscription output. If it is empty, SublinkPro keeps the node's effective name. Variables are replaced when the subscription is generated, so values such as speed, delay, country, tags, and unlock status reflect the node data currently stored in the system.

Fallback words such as `Unknown`, `Ungrouped`, `Manual`, `Residential`, and the unlock labels follow the **node name language** described below.

Common variables:

| Variable | Meaning |
|:---|:---|
| `$Name` | Effective node name, based on the node's selected name mode |
| `$LinkName` | Original upstream node name |
| `$LinkCountry` | Node country code, such as `HK` or `US`; empty country values render as `Unknown` |
| `$LinkCountryName` | Localized country name resolved from `Settings -> Country Rules` by `$LinkCountry`; falls back to the country code when no country-rule name exists |
| `$Flag` | Flag emoji generated from the country code |
| `$Group` | Node group; empty group values render as `Ungrouped` |
| `$Source` | Node source; manual nodes render as `Manual` |
| `$Protocol` | Protocol type |
| `$Index` | Output sequence number |
| `$DuplicateIndex` | Duplicate-name index; empty for the first occurrence, then `1`, `2`, `3`, and so on |
| `$Tags` | All node tags, joined with `|` |
| `$TagGroup(name)` | The node tag from the named tag group, when present |
| `$Speed`, `$SpeedIcon` | Download speed text and speed status icon |
| `$Delay`, `$DelayIcon` | Latency text and latency status icon |
| `$IpType`, `$Residential` | IP quality labels such as native/broadcast and residential/datacenter |
| `$FraudScore`, `$FraudScoreIcon` | Fraud score and fraud score icon |
| `$Unlock` | Unlock summary |
| `$Unlock(provider)` | Unlock result for a specific provider key, such as `$Unlock(netflix)` |
| `$UnlockStatus`, `$UnlockLabel`, `$UnlockRegion` | Detailed unlock status fields when available |

Country-name logic depends on the node's stored country code. For example, if a node has `$LinkCountry = HK` and the `HK` country rule name is `香港`, `$LinkCountryName` renders that name in the node name language: `Hong Kong` for `en-US` and `香港` for `zh-CN`. A country name you edited yourself in `Settings -> Country Rules` is always kept as written. If the country code has no matching country rule, `$LinkCountryName` falls back to the code itself, such as `HK`.

### Node Name Language

Node names are rendered on the server and shown inside client apps, so they do not follow the language of the browser tab alone. The header language switcher stores the choice as the node name language (API: `GET/POST /api/v1/settings/node-naming`). Supported values are `en-US` (default) and `zh-CN`, and they control the fallback words used by the rename variables, the IP quality labels, the quality-check placeholders, and the unlock labels.

Example rule:

```text
[$Flag] $LinkCountryName - $LinkName $DuplicateIndex
```

For a Hong Kong node named `Premium 01`, the output could become `[🇭🇰] Hong Kong - Premium 01`. If another output node resolves to the same name, `$DuplicateIndex` can add a suffix value for later duplicates.

## Mieru Output Notes

- Mieru currently supports Clash/mihomo output only. `/c?client=clash` outputs mihomo YAML fields including `type: mieru`, `server`, `port` or `port-range`, `transport`, `username`, `password`, and optional `multiplexing`, `traffic-pattern`, and chain proxy `dialer-proxy`.
- Official Mieru has `mieru://` and `mierus://` share links, but official docs do not define a general URL schema suitable for field by field editing. SublinkPro internally uses `mieru://username:password@server:port?...#name` as the raw edit and Clash/mihomo import write back format. When a port range is needed, use `portRange=2090-2099` and do not write `port`.
- `/c?client=v2ray` and Surge currently do not support Mieru. SublinkPro skips Mieru nodes, does not write `mieru://` links into v2ray base64, and does not generate Surge config for them.

## Snell Output Notes

- Snell supports Clash/mihomo and Surge output. `/c?client=clash` outputs mihomo YAML fields including `type: snell`, `server`, `port`, `psk`, plus optional `version`, `udp`, `obfs-opts` (`mode` / `host`), the shared connection-layer options `tfo`, `mptcp`, `interface-name`, `routing-mark`, `ip-version`, and chain proxy `dialer-proxy`. `/c?client=surge` outputs `snell, server, port, psk=...` and appends `version`, `obfs`, `obfs-host`, `udp-relay`, and `tfo` when present (the other mihomo-only connection-layer options have no Surge equivalent).
- Snell has no official general share-link schema; mihomo describes Snell with Clash YAML fields. SublinkPro internally uses `snell://server:port?psk=...&version=...&obfs=...&obfs-host=...#name` as the raw edit and Clash/mihomo, Surge import write back format. `version` defaults to mihomo's Snell v1 and accepts 1/2/3.
- `/c?client=v2ray` currently does not support Snell. SublinkPro skips Snell nodes and does not write `snell://` links into v2ray base64.

## OpenVPN Output Notes

- OpenVPN supports Clash/mihomo output only. `/c?client=clash` writes `type: openvpn` and preserves the imported Mihomo fields, including PEM/key blocks, credentials, transport and cipher options, keepalive, IP stack, DNS, shared connection-layer options, and chain proxy `dialer-proxy`.
- SublinkPro uses `openvpn://server:port?...#name` only as an internal raw-edit and Clash/mihomo import/export round-trip format. It is not an official OpenVPN share link and direct `.ovpn` import is not supported.
- `/c?client=v2ray` and Surge skip OpenVPN nodes. Sensitive key material is URL-encoded, not encrypted, in the stored internal link.

## VLESS XHTTP Output Notes

- When a subscription node is VLESS with `xhttp` transport, `/c?client=clash` outputs `network: xhttp` and `xhttp-opts`.
- If the VLESS URL carries `encryption`, `/c?client=clash` preserves it as the top level mihomo `encryption` field.
- `/c?client=v2ray` continues to output the VLESS URL and preserves `type=xhttp`, `path`, `host`, `mode`, and `extra`.
- When top level VLESS `ech` is Xray DNS / URI style, `/c?client=clash` outputs top level `ech-opts` within what mihomo can express. Recognizable query domains map to `query-server-name`.
- In reverse, when a node comes from Clash/mihomo YAML import and only `ech-opts.query-server-name` can be restored, the system fills it before saving the node link as `ech=<query-server-name>+https://dns.alidns.com/dns-query` using local compatibility rules.
- To avoid generating configurations that look valid but are semantically distorted, the system does not silently convert `xhttp` into `http`, `h2`, or `grpc`.

## VLESS REALITY compatibility

Clash / mihomo YAML imports preserve `reality-opts.support-x25519mlkem768` through stored VLESS URLs and rebuilt proxies used for subscription output, latency checks, and speed tests. The URL parameter is `support-x25519mlkem768=true` or `false` (`1` / `0` are also accepted on input). Explicit `false` stays distinct from an unspecified value; empty or invalid URL values are treated as unspecified. Plain TLS nodes do not acquire REALITY options from this parameter. In the node field editor, choose `true`, `false`, or the empty selection to restore the upstream default.
