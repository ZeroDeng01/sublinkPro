English | [简体中文](airport.zh-CN.md)

# Airport Subscription Management

SublinkPro provides complete airport subscription management. It can convert subscriptions into nodes, and also monitor and manage airport services in detail.

---

## 💡 Core Features

| Feature | Description |
|:---|:---|
| **📥 Multi format import** | Automatically parses and imports Clash/mihomo and V2Ray subscription formats. Mieru supports Clash/mihomo YAML only |
| **⏱️ Smart scheduled updates** | Built in Crontab level scheduler, with interval or Cron expression updates, keeping nodes fresh |
| **📊 Traffic usage monitoring** | Parses the `Subscription-Userinfo` response header and shows **used upload**, **used download**, **total traffic**, and **expiration time** |
| **🚀 Update now** | One click “fetch now” plus real time callbacks, so latest traffic data and node lists appear without refreshing the page |
| **⚡ Post update check** | Airports can bind node check profiles. After a subscription update succeeds, the airport's nodes can be checked immediately |
| **🤖 Bot integration** | Telegram Bot can query remaining traffic, expiration time, and trigger update tasks remotely |

### Clash / mihomo `proxy-providers` compatibility

- Airport subscription import supports `proxy-providers` in Clash / mihomo YAML. If top level `proxies` is missing but providers contain nodes, those provider nodes are expanded and imported as normal nodes.
- The system currently fetches remote providers with `type: http`, reusing the airport subscription's proxy download and ignore TLS certificate verification behavior. Provider requests always carry User-Agent. Custom Header values are inherited only when the provider URL and root subscription URL have the same host. Cross host providers and cross host redirects do not carry custom Header values.
- The system first expands providers referenced by `proxy-groups[].use`. If groups enable `include-all` / `include-all-providers`, or if no group references any provider, all HTTP providers are expanded in declared order.
- Provider responses support standard YAML top level `proxies`. If a provider returns base64 or plaintext URI line lists, they are parsed with normal subscription link compatibility.
- Local `file` providers, provider cache, scheduled health checks, full mihomo override semantics, and a provider's own `proxy` routing are not implemented. These runtime features should still be handled by Clash / mihomo clients.

### VLESS / XHTTP compatibility

- Airport subscription import now supports `type=xhttp` in `vless://` links.
- When the upstream subscription is Clash / mihomo YAML and a node is `type: vless` with `network: xhttp`, the system reads `xhttp-opts` and writes it back as a VLESS URL.
- Supported top level URL fields include `type`, `encryption`, `path`, `host`, `mode`, `extra`, and `ech`.
- Supported fields in `extra` that map to mihomo include `headers`, `noGRPCHeader`, `xPaddingBytes`, `downloadSettings`, and known subfields.
- Top level `ech` maps first to mihomo top level `ech-opts`: fixed base64 ECHConfig values are written to `config`, while Xray DNS / URI style values are mapped on a best effort basis to what mihomo can express.
- When the airport subscription itself is Clash/mihomo YAML and import can only restore `query-server-name` from top level `ech-opts`, the system rebuilds the saved node link before storage as `ech=<query-server-name>+https://dns.alidns.com/dns-query` using local compatibility rules.
- `extra.downloadSettings.echOpts` still maps only to mihomo `xhttp-opts.download-settings.ech-opts`. It is not mixed with top level `ech-opts`.
- Extensions that exist on the Xray side but have no current public mihomo carrying field, such as `xmux` and `sessionPlacement`, are treated as unsupported. They are not silently downgraded to `http`, `h2`, or `grpc`.

### Mieru compatibility

- Airport subscription import supports `type: mieru` nodes in Clash/mihomo YAML, preserving official mihomo fields: `server`, `port` or `port-range`, `transport`, `username`, `password`, `multiplexing`, and `traffic-pattern`.
- Official Mieru has `mieru://` and `mierus://` share links, but does not define a general URL schema suitable for field by field editing in the SublinkPro raw editor. When saving nodes, the system uses an internal editable form, `mieru://username:password@server:port?...#name`, with port ranges written as `portRange=2090-2099`, for Clash/mihomo YAML import write back and later export.
- Mieru is not output to v2ray or Surge. Those clients are currently outside SublinkPro's Mieru support scope.

---

## 📱 UI Display

The subscription list clearly shows the state of each airport, including:

- Last update time
- Next scheduled update time
- Visual traffic progress bar

This gives a direct overview of airport usage.

---

## Usage Flow

### Add an airport subscription

1. Open the “Airport Management” page.
2. Click “Add Airport”.
3. Fill in the subscription link and name.
4. Configure request settings as needed, such as User-Agent, custom Header, and proxy download.
5. Configure an update policy, optional. To check nodes immediately after subscription update, enable “post update check” in advanced options and choose a node check profile.
6. Save and fetch nodes.

After “post update check” is enabled, checks are triggered only when the current airport subscription updates successfully. The check range is limited to nodes from this airport's latest update. It does not wait for the node check profile's own schedule, and it does not expand to nodes from other airports. Node check tasks triggered by this feature are marked as “airport update” in Task Center, so they can be distinguished from “manual” and “scheduled” triggers.

### Node processing: unique names

In “Node Processing, applied during fetch” inside the airport edit dialog, you can configure unique node names:

- **Unique node names**: add a stable prefix to nodes imported from the current airport to avoid name conflicts between airports.
- **Unique names inside airport**: when duplicate names exist inside the same airport, append `-1`, `-2`, `-3`, and so on to the current node name.

Notes:

- Both switches take effect **when fetching subscriptions**. After changes, refetch the airport to apply them to existing nodes.
- Airport prefix uniqueness and within airport sequence numbering can be enabled together. The system first generates the airport prefix, then appends `-1`, `-2`, and so on for duplicates inside the same airport.
- Within airport sequence numbering can be enabled alone. In that case, no airport prefix is added, and only same airport duplicates get numeric suffixes.
- Numbering is calculated **per duplicate name group**, not as one continuous sequence for the whole airport. For example, duplicated `HK` nodes become `HK-1`, `HK-2`, while duplicated `US` nodes start separately from `US-1`.

### Node names and remarks

Nodes fetched from airports store two names:

- **Original name**: comes from the airport subscription link and refreshes when the airport updates.
- **Remark name**: a custom name maintained by users in the node edit dialog. Airport updates do not overwrite it. Global node remarks must be unique.

When editing a node in “Node Management”, use “Actual use name” to choose which name is used by subscription output, node rename rules, chain proxy, and node scripts:

- **Use original name**: operate on the original name from the airport subscription. Suitable when you want to follow upstream naming exactly.
- **Use remark name**: operate on the user remark as the actual node name. Suitable for custom naming that should survive airport updates.

Switching back to “Use original name” restores use of the airport subscription name. The remark name is kept, so you can switch back later.

Note: Manual remark edits are rejected if they duplicate another node remark. When airport automatic import produces duplicate remarks, the system first generates `original name@airport name`. If that is still duplicated, it appends numbering, for example `香港 01@机场B-2`. Automatic imports without an airport source use normal numbering to keep global uniqueness.

### Request settings

Airport “Request Settings” configure request parameters used when fetching subscriptions:

- `User-Agent`: set with a dedicated input for common client UAs or manual input.
- **Custom Header**: add multiple headers with `Header name` + `Header value`, useful for airports that need extra auth or source identification.
- `Use proxy download`: fetch subscriptions through a specified node or automatically selected best node.

Notes:

- Custom Header values are sent with airport subscription requests.
- If “fetch usage information” is enabled, the system reuses the same custom Header values when refreshing airport usage.
- `User-Agent` is managed as a separate field. Custom Header does not allow another `User-Agent` entry.

### Configure scheduled updates

Two scheduled update methods are supported:

| Method | Description |
|:---|:---|
| **Interval update** | Set a fixed interval, such as every 6 hours |
| **Cron expression** | Flexible Cron expression, such as `0 */6 * * *` |

### Traffic monitoring

The system automatically parses `Subscription-Userinfo` from subscription response headers and extracts:

- `upload`: used upload traffic
- `download`: used download traffic
- `total`: total traffic quota
- `expire`: expiration timestamp

---

## Telegram Bot Integration

Telegram Bot can:

- Query remaining traffic for each airport
- View expiration time
- Trigger subscription updates remotely

See [Telegram Bot documentation](telegram-bot.md).
