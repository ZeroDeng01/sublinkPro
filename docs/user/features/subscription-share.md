English | [简体中文](subscription-share.zh-CN.md)

[Guide index](../README.md)

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

Use **Copy** next to a client URL or in its QR code dialog. If the browser does not allow automatic copying, select the displayed URL and copy it manually. This can depend on browser support when accessing the Web UI over a NAS's HTTP address.

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

## Subscription Output Formats

- Built-in outputs include `clash`, `mihomo`, `surge`, `v2ray`, and `uri`.
- Select **URI** in the share details or batch export. `/c/?token=<shareToken>&client=uri` returns the stored links of all nodes selected by the subscription's filtering, ordering and renaming rules, one link per line in UTF-8. It includes links for protocols that V2Ray skips, such as Snell, OpenVPN and Sudoku. Output and node filter scripts receive `uri` as the client type. Each importing app determines which links it supports, including any project-specific link formats.
- When **Application Settings -> Sub-Store** is enabled and targets are selected, share links can also request the selected expanded targets: `loon`, `egern`, `stash`, `surfboard`, `shadowrocket`, `quanx`, `sing-box`, and `json`.
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

## Node Naming Variables

`NodeNameRule` controls the node name rendered in subscription output. If it is empty, SublinkPro keeps the node's effective name. Variables are replaced when the subscription is generated, so values such as speed, delay, country, tags, and unlock status reflect the node data currently stored in the system.

Common variables:

| Variable | Meaning |
|:---|:---|
| `$Name` | Effective node name, based on the node's selected name mode |
| `$LinkName` | Original upstream node name |
| `$LinkCountry` | Node country code, such as `HK` or `US`; empty country values render as `未知` |
| `$LinkCountryName` | Country name resolved from `Settings -> Country Rules` by `$LinkCountry`; falls back to the country code when no country-rule name exists |
| `$Flag` | Flag emoji generated from the country code |
| `$Group` | Node group; empty group values render as `未分组` |
| `$Source` | Node source; manual nodes render as `手动` |
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

Country-name logic depends on the node's stored country code. For example, if a node has `$LinkCountry = HK` and the `HK` country rule name is `香港`, `$LinkCountryName` becomes `香港`; if the country code has no matching country rule, `$LinkCountryName` renders as `HK`. To change displayed country names in naming rules, update the country name in `Settings -> Country Rules`.

Example rule:

```text
[$Flag] $LinkCountryName - $LinkName $DuplicateIndex
```

For a Hong Kong node named `Premium 01`, the output could become `[🇭🇰] 香港 - Premium 01`. If another output node resolves to the same name, `$DuplicateIndex` can add a suffix value for later duplicates.

## Mieru Output Notes

- Clash/mihomo subscriptions render Mieru as YAML with `server`, `port` or `port-range`, `transport`, credentials, and optional `multiplexing`, `handshake-mode`, `traffic-pattern`, and `dialer-proxy`.
- Adding nodes accepts both [official Mieru formats](https://github.com/enfein/mieru/blob/v3.39.0/docs/client-install.md#sharing-client-settings). `mieru://` carries a Base64 Protobuf client configuration. `mierus://` uses query parameters such as `port=2090-2099&profile=default&protocol=TCP`. Each link represents one SublinkPro node. Clash/mihomo export requires one server, one transport and a single port or one contiguous port range. Multiple bindings can form one range when they use the same transport and have no gaps. SublinkPro skips incompatible nodes. Raw edits preserve the input format and the other bindings.
- The startup migration converts valid historical SublinkPro links to `mierus://` while preserving the node IDs. Migration logs identify records whose links need correction.
- `/c?client=v2ray` and Surge currently do not support Mieru. SublinkPro skips Mieru nodes, does not write `mieru://` links into v2ray base64, and does not generate Surge config for them.

## Sudoku Support

- To add a node, paste an official `sudoku://` short link or Clash/mihomo YAML containing a `type: sudoku` proxy. The raw node editor can change the endpoint, key, encoding and HTTPMask settings.
- Clash/mihomo subscriptions export `type: sudoku` with the key, AEAD method, padding (including zero), downlink mode, table settings, multiplexing and HTTPMask options. UDP and chain proxy settings are applied during export. The built-in Mihomo adapter handles node testing. [Mihomo Sudoku options](https://wiki.metacubex.one/en/config/proxies/sudoku/)
- Use a Mihomo-based client that supports Sudoku. The native v2ray and Surge subscription formats skip Sudoku nodes.
- When SublinkPro stores a YAML import as `sudoku://`, it adds optional fields for Mihomo settings such as padding and `dialer-proxy`. Standalone Sudoku clients read the official short-link fields and apply their own defaults for settings outside that format. [Official short-link format](https://github.com/SUDOKU-ASCII/sudoku/blob/main/doc/README.md#short-link-format)

## Snell Output Notes

- Snell supports Clash/mihomo and Surge output. `/c?client=clash` outputs mihomo YAML fields including `type: snell`, `server`, `port`, `psk`, plus optional `version`, `udp`, `obfs-opts` (`mode` / `host`), the shared connection-layer options `tfo`, `mptcp`, `interface-name`, `routing-mark`, `ip-version`, and chain proxy `dialer-proxy`. `/c?client=surge` outputs `snell, server, port, psk=...` and appends `version`, `obfs`, `obfs-host`, `udp-relay`, and `tfo` when present (the other mihomo-only connection-layer options have no Surge equivalent).
- Snell has no official general share-link schema; mihomo describes Snell with Clash YAML fields. SublinkPro internally uses `snell://server:port?psk=...&version=...&obfs=...&obfs-host=...#name` as the raw edit and Clash/mihomo, Surge import write back format. `version` defaults to mihomo's Snell v1 and accepts 1/2/3.
- `/c?client=v2ray` currently does not support Snell. SublinkPro skips Snell nodes and does not write `snell://` links into v2ray base64.

## OpenVPN Output Notes

- `/c?client=clash` writes OpenVPN nodes as `type: openvpn` and preserves the imported Mihomo fields, including PEM/key blocks, credentials, transport and cipher options, keepalive, IP stack, DNS, shared connection-layer options, and chain proxy `dialer-proxy`.
- SublinkPro stores OpenVPN nodes as `openvpn://server:port?...#name` for raw editing and Clash/mihomo round trips. The URI output includes this project-specific link format. Direct `.ovpn` import is not supported.
- `/c?client=v2ray` and Surge skip OpenVPN nodes. Sensitive key material is URL-encoded, not encrypted, in the stored internal link.

## VLESS XHTTP Output Notes

- When a subscription node is VLESS with `xhttp` transport, `/c?client=clash` outputs `network: xhttp` and `xhttp-opts`.
- If the VLESS URL carries `encryption`, `/c?client=clash` preserves it as the top level mihomo `encryption` field.
- `/c?client=v2ray` continues to output the VLESS URL and preserves `type=xhttp`, `path`, `host`, `mode`, and `extra`.
- When top level VLESS `ech` is Xray DNS / URI style, `/c?client=clash` outputs top level `ech-opts` within what mihomo can express. Recognizable query domains map to `query-server-name`.
- In reverse, when a node comes from Clash/mihomo YAML import and only `ech-opts.query-server-name` can be restored, the system fills it before saving the node link as `ech=<query-server-name>+https://dns.alidns.com/dns-query` using local compatibility rules.
- To avoid generating configurations that look valid but are semantically distorted, the system does not silently convert `xhttp` into `http`, `h2`, or `grpc`.

## Share link device limits

New subscriptions default to unlimited devices (limit 0). Administrators can opt in per subscription or share with a limit from 0 to 10000; 0 means unlimited. New shares (including the default link and batch-created shares) inherit the subscription default, then maintain independent limits. Changing the subscription default does not change existing shares. Existing subscriptions and shares remain unlimited after migration until an administrator opts in.

Create a separate share for each recipient and set its expiry independently. The first successful download binds a device; subsequent updates with the same ID reuse its slot. A positive limit requires a valid, single `X-HWID` request header. IP addresses, user agents, models and URL parameters are not device identities. The limit works with any client that sends this header, including compatible versions of Karing, Happ, v2RayTun and Clash Mi; it does not check a client brand or require a Karing User-Agent. Missing or invalid IDs, revoked devices and full limits return HTTP 403 without nodes; database validation failure returns 503.

### Client import

Enable the client's device identification / `X-HWID` option before importing or refreshing a limited share. For example, in Karing open Add configuration → Add configuration link, paste the link, enable **X-HWID**, then save; existing profiles need this option in their edit screen too. Other clients may use different names or enable it by default, so check the actual request headers for the installed version. Both Clash YAML and Base64 outputs use the same device policy. Have the intended recipient import first: forwarding an unbound link may let someone else claim a slot.

The share list shows bound devices / limit. Manage devices lets administrators edit notes, revoke or restore installations. Revoking frees a slot but permanently denies that ID until restored; restoration requires capacity. Revoke excess devices before lowering a limit. Ordinary Token rotation preserves bindings. Reset all devices revokes all registered IDs and rotates the link; deliver the new link to the customer's new device.

Batch updates default to keeping the original value. Choose the setting to change explicitly; changing a device limit does not change Token, expiry or enabled state. Each share is committed independently. If a lower limit is rejected for one share, other shares can still be updated. The API returns per-item results and success/failure counts while keeping its existing HTTP status and message. After partial failure, the page refreshes the list, displays each failed share and reason, and keeps only failed shares selected for retry.

Database migration and WebDAV restoration include both bound and revoked device records. An overwrite restore replaces the target's previous device records together with the shares, so reused share IDs cannot inherit unrelated devices. Backups made before device limits existed restore without bindings.

### Reverse proxies and HTTP 403 troubleshooting

Every proxy or custom subscription sidecar on `/c/` must preserve `User-Agent`, `X-HWID`, `X-Device-OS`, and `X-Device-Model`. A header allowlist that drops `X-HWID` causes `hwid_required` even when the client enabled it. Forward one original HWID value; never invent an ID or substitute an IP address. Both repeated HWID headers and comma-separated values produced by proxy merging are rejected as `hwid_invalid`. Do not cache subscription responses across devices.

Verify the complete public subscription URL, not only `/api/v1/version` or the backend port. `HEAD` with an available device identity should succeed without binding; the same request without `X-HWID` must return 403 when the device limit is enabled. If the limit is full, an unbound identity is also rejected.

The response JSON distinguishes `hwid_required`, `hwid_invalid`, `device_limit_exceeded`, and `device_revoked`. Some clients show a generic HTTP error instead. For example, iOS Karing 1.2.22.2502 may show `http statusCode: 403` and suggest changing the User-Agent; that fixed dialog does not identify the server's reason. Check the bound count / limit, response code and X-HWID option before changing settings.

### Capability and verification boundaries

This controls subscription downloads, not tamper-proof hardware authentication. An unbound link forwarded before its intended recipient imports it can be claimed by someone else. Revoking a device affects only subsequent downloads; previously downloaded nodes remain usable. Exported nodes, forged HWIDs and network sharing cannot be prevented by this limit alone. Setting the limit to 0 disables HWID and revocation checks and uses the ordinary download path without a device registration transaction. Token and share availability are still checked before sending a response. `HEAD` does not allocate slots, and failed generation does not bind devices.

Earlier PR verification used Windows Karing 1.2.23+2606 for real first-import and repeated manual-update requests: enabling the option sent `X-HWID`, with a stable digest across those requests. The UA began with `Karing/1.2.23.2606 platform/windows`. Those recorded checks do not establish automatic refresh, other clients or platforms, or behavior after app restart, upgrade, network change or backup restoration. Simulated requests are not physical-client validation. Reinstallation or backup restoration that changes the ID requires an administrator-assisted device replacement.

The earlier end-to-end Windows Karing check successfully downloaded the subscription and registered the device. After administrator revocation, Karing displayed `download profile failed: http statusCode: 403`. The test device was then restored.

#### Verification recorded on 2026-11-20

Karing successful real-client download counts are fixed to `2026-11-20T21:06:58+08:00`. Its subsequent revocation check is recorded through `2026-11-20T21:11:43+08:00`; rejected requests are not added to successful counts. The later v2rayN revocation checkpoint is 21:53:18; after restoring the test binding and correcting the fixture, both client update timestamps advanced to 21:58:46 and the server recorded a successful access at 21:58:47. All times are UTC+08:00 on 2026-11-20.

| Environment | Check | Observed result |
|:---|:---|:---|
| Simulated HTTP requests with Karing 1.2.23.2606 and unversioned Happ user agents | First download, repeated download and revoked-device refresh for both Clash YAML and Base64 output | Initial and repeated requests returned 200 and kept one binding; after revocation, the same ID returned 403 with `device_revoked` |
| Windows, isolated MySQL 8.4.13 instance | 10 rounds of 24 concurrent admissions competing for one available slot | Every round admitted one device, returned `device_limit_exceeded` for the other 23, persisted one device and had zero database errors |
| Windows, isolated PostgreSQL 17.12 instance | 10 rounds of 24 concurrent admissions competing for one available slot | Every round admitted one device, returned `device_limit_exceeded` for the other 23, persisted one device and had zero database errors |
| Windows Karing 1.2.26.2931, operated by the user | Initial, repeated manual and automatic Clash YAML downloads | Three real GETs sent `X-HWID` and returned 200, reusing one binding before revocation. The user identified the second as manual; the third arrived during observation with instructions to avoid manual refresh |
| Windows Karing 1.2.26.2931, operated by the user | Initial, repeated and automatic Base64 downloads | Six real GETs sent `X-HWID` and returned 200, reusing one binding before revocation. The user explicitly confirmed the third request was automatic; later requests also arrived during observation with instructions to avoid manual refresh |
| Windows Karing 1.2.26.2931, operated by the user | Refresh after revoking both devices | One real YAML GET and three real Base64 GETs sent `X-HWID` and returned 403. Both bindings were revoked, both active counts were zero, and successful access counters and timestamps did not increase |
| Windows v2rayN 7.25.4, operated by the user | Base64 initial and automatic import | The user reported success; one profile was persisted with a one-minute update interval and an updated client database timestamp. A manually fixed `X-HWID` was used |
| Windows v2rayN 7.25.4, operated by the user | Clash YAML import after correcting the fixture | The user confirmed successful import; the client persisted one custom configuration with core type 13. Both output groups had one-minute update intervals and timestamps of 21:58:46, alongside automatic-update log entries at 21:57:46 and 21:58:46 |
| Windows v2rayN 7.25.4, operated by the user | Refresh after revoking the shared device binding | Real client requests returned HTTP 403; the saved checkpoint had zero active devices and an unchanged successful access count of eight. The user supplied a YAML-group error; the server log did not distinguish output formats |

The initial agent-operated attempt using a portable Karing 1.2.26.2934 copy and isolated v2rayN did not complete a real-client download. The user later operated their existing Karing 1.2.26.2931 instance, producing the real GET evidence above. The YAML access counter is four, including one earlier diagnostic request, so it represents three real-client downloads.

Simulated HTTP tests verify the server contract. Real-client evidence now confirms initial, repeated and automatic downloads, plus rejection after revocation, for both Karing outputs. The automatic YAML request arrived about five minutes six seconds after the user-confirmed manual refresh, during subsequent observation with instructions to avoid manual refresh and the user's confirmation that the five-minute setting was enabled. Its source was not separately identified by the user. These observations do not guarantee an exact five-minute schedule. The real revocation logs captured HTTP 403, but the response JSON error code was not inspected; `device_revoked` in the response body was verified only by simulated HTTP tests.

The user later tested their existing v2rayN while disconnecting their own network; the agent inspected logs and databases. Both output URLs shared a token and manually configured `X-HWID`. Seven aggregate real GETs returned 200, with one active binding; the access counter of eight included one earlier diagnostic request. The log omitted the output query parameter, so these requests are not divided by format. This does not validate a native v2rayN hardware identifier.

The saved Base64 profile and the user's automatic-update report support initial and automatic Base64 import. The original YAML test fixture failed client detection because it used `port: 7890`; after switching the test service to `mixed-port: 17891`, the user confirmed import and the client persisted the custom configuration. Subsequent timer log entries and both groups' updated timestamps support automatic refresh; the shared server binding remained one device. Before the fixture correction, revocation produced real client HTTP 403 responses with zero active bindings and no increase in successful accesses. Those requests were not counted by output format, so separate Base64 and YAML revocation results are not claimed. The binding was restored for the corrected-fixture test.

Clash, Clash Verge, Clash Mi, Happ and v2RayTun were not installed or tested as real clients in this run. Testing Clash YAML output in Karing or v2rayN does not establish compatibility with those applications. Other platforms, native v2rayN hardware identification, and restart, upgrade, network-change and client-backup scenarios remain unverified.

## VLESS REALITY compatibility

Clash / mihomo YAML imports preserve `reality-opts.support-x25519mlkem768` through stored VLESS URLs and rebuilt proxies used for subscription output, latency checks, and speed tests. The URL parameter is `support-x25519mlkem768=true` or `false` (`1` / `0` are also accepted on input). Explicit `false` stays distinct from an unspecified value; empty or invalid URL values are treated as unspecified. Plain TLS nodes do not acquire REALITY options from this parameter. In the node field editor, choose `true`, `false`, or the empty selection to restore the upstream default.

VLESS and Trojan share links preserve REALITY `spx` (`spiderX`, crawler initial path) and `pqv` (`mldsa65Verify`, additional ML-DSA-65 certificate verification public key) through parsing, field editing and URL export. The editor exposes `Query.SpiderX` and `Query.Mldsa65Verify`; clearing a field removes its URL parameter. Values are URL-encoded, so crawler paths can include query parameters.

Clash YAML imports map `reality-opts.spider-x` and `reality-opts.mldsa65-verify` to the stored URL, and `/c?client=clash` or `mihomo` maps them back to YAML without losing the values. Nonempty fields are emitted only for REALITY nodes; empty fields are omitted. Trojan REALITY imports also preserve `security=reality`, `pbk` and `sid` through this round trip.

Use `/c?client=v2ray` for Xray clients such as v2rayN and v2rayNG; no separate `xray` client type is needed. Preserving configuration does not add execution support to a core. Current upstream mihomo has no corresponding verifier; the existing internal probe behavior is unchanged. Use a compatible core to execute and verify exported configurations. Surge skips VLESS REALITY nodes with nonempty `pqv` and skips Trojan REALITY nodes; crawler-only `spx` is omitted from Surge output. ML-DSA-65 verification is separate from the `support-x25519mlkem768` key-exchange option.

## VLESS in Surge

`/c?client=surge` exports VLESS for Surge iOS 6.0.0+ / Mac 6.12.0+. Supported transports are TCP/RAW, WebSocket, HTTPUpgrade, gRPC (TLS or REALITY), and XHTTP/SplitHTTP. TLS, REALITY keys, Vision, VLESS Encryption, SNI, ALPN, certificate pinning, transport headers and supported XHTTP extra/XMUX parameters are mapped to native Surge options. Host replacement and the subscription certificate override also apply. Surge uses XUDP automatically.

Nodes with unsupported transports, ECH, TCP header obfuscation, incompatible Vision/REALITY settings or XHTTP `downloadSettings` are skipped, including their proxy group entries. The uTLS `fp` and REALITY `support-x25519mlkem768` options have no Surge equivalent and are omitted; Surge handles its own ClientHello and ML-KEM fallback.

See the [official Surge VLESS reference](https://manual.nssurge.com/policies/vless.html).

## Order airports within a group

On the **Subscriptions** page, open **Group sort**, choose a group, and drag airports into the desired order. Save the sort to apply it. Nodes from airports higher in the list appear earlier in subscription output for that group.
