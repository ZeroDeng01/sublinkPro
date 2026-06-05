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

> [!NOTE]
> **Client compatibility**: Share links can detect client type automatically. Clash, Surge, V2ray, and other client formats can also be specified manually.

## Subscription Update Interval

- In `Subscription Management -> Subscription Settings -> Basic Settings`, each subscription can configure “Update interval, hours”.
- The value is stored in hours, with a maximum of `8760` hours. If set to `0` or left empty, the default update interval is used: Clash uses `24` hours, Surge uses `86400` seconds.
- When a client fetches Clash config through a subscription link, the response header includes `profile-update-interval`, in hours.
- When a client fetches Surge config, `interval` in `#!MANAGED-CONFIG` is converted to seconds automatically according to the setting.

## Mieru Output Notes

- Mieru currently supports Clash/mihomo output only. `/c?client=clash` outputs mihomo YAML fields including `type: mieru`, `server`, `port` or `port-range`, `transport`, `username`, `password`, and optional `multiplexing`, `traffic-pattern`, and chain proxy `dialer-proxy`.
- Official Mieru has `mieru://` and `mierus://` share links, but official docs do not define a general URL schema suitable for field by field editing. SublinkPro internally uses `mieru://username:password@server:port?...#name` as the raw edit and Clash/mihomo import write back format. When a port range is needed, use `portRange=2090-2099` and do not write `port`.
- `/c?client=v2ray` and Surge currently do not support Mieru. SublinkPro skips Mieru nodes, does not write `mieru://` links into v2ray base64, and does not generate Surge config for them.

## VLESS XHTTP Output Notes

- When a subscription node is VLESS with `xhttp` transport, `/c?client=clash` outputs `network: xhttp` and `xhttp-opts`.
- If the VLESS URL carries `encryption`, `/c?client=clash` preserves it as the top level mihomo `encryption` field.
- `/c?client=v2ray` continues to output the VLESS URL and preserves `type=xhttp`, `path`, `host`, `mode`, and `extra`.
- When top level VLESS `ech` is Xray DNS / URI style, `/c?client=clash` outputs top level `ech-opts` within what mihomo can express. Recognizable query domains map to `query-server-name`.
- In reverse, when a node comes from Clash/mihomo YAML import and only `ech-opts.query-server-name` can be restored, the system fills it before saving the node link as `ech=<query-server-name>+https://dns.alidns.com/dns-query` using local compatibility rules.
- To avoid generating configurations that look valid but are semantically distorted, the system does not silently convert `xhttp` into `http`, `h2`, or `grpc`.
