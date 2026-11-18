English | [简体中文](subscription-proxy.zh-CN.md)

# Proxy Subscriptions

Subscription Management has two kinds of subscriptions:

| Tab | Type | What it returns |
|:---|:---|:---|
| **Aggregated** | `collection` (default) | Nodes assembled by SublinkPro from manual nodes, dynamic groups, and airports, filtered, renamed, and rendered through a template |
| **Proxy** | `proxy` | The response body of the upstream subscription of a selected airport, returned as-is |

Proxy subscriptions exist for the case where you do not control the node set yourself and want to expose an existing upstream subscription through SublinkPro shares: a stable share link, expiry, and access logs, without re-parsing or rewriting the upstream content. You only pick an airport: its URL and request settings are reused for the relay. IP allow/deny lists are not applied, and every share offers a single subscription link.

---

## How it works

```text
Client  --GET /c/?token=...-->  SublinkPro
                                  |  share lookup, expiry
                                  |  access log
                                  v
                              Airport (URL, client UA, headers,
                              device profile, proxy, TLS)
                                  |  HTTP(S)
                                  v
                            Upstream provider
                                  |
                                  v
Client  <--upstream body + status--  SublinkPro
```

1. The client requests a share link exactly as for an aggregated subscription (`/c/?token=...`).
2. SublinkPro resolves the share, checks expiry, and records the access.
3. SublinkPro loads the airport referenced by the subscription and requests its subscription URL with the client's `User-Agent` (the airport's effective User-Agent is used only when the client sends none) and the airport's request headers (including device profile headers and the derived HWID unless disabled), optional proxy, and TLS-skip setting. Panels such as Remnawave choose the response format from the User-Agent, so the relayed link returns the same format the airport serves to that client directly.
4. The upstream status code and a whitelist of headers are copied to the response, and the body is streamed back unchanged.

Except for the User-Agent passthrough, the relay uses exactly the same settings as the scheduled airport update, so editing the airport — its URL, headers, device profile, proxy, or TLS setting — immediately changes what the proxy subscription returns. The airport's User-Agent is still used as the fallback for clients that send no User-Agent.

Nothing is parsed, filtered, renamed, or re-rendered. The client receives byte-for-byte what the upstream returned; the upstream decides which client-specific format (Clash, Surge, sing-box, Xray JSON, base64, plain URI list) to serve based on the relayed User-Agent.

---

## Configuration

Create the subscription from the **Proxy** tab of Subscription Management.

| Field | Required | Description |
|:---|:---|:---|
| `name` | Yes | Subscription name, unique across both types |
| `AirportID` | Yes | Airport whose upstream subscription is relayed; the URL, User-Agent, headers, device profile, proxy, and TLS settings all come from the airport |

Shares, expiry policies, and access logs work as for aggregated subscriptions; configure them from the subscription's row as usual. IP allow/deny lists are not applied to proxy subscriptions, and the share link dialog shows a single subscription link (`/c/?token=...`) because client-specific variants are meaningless for a verbatim relay.

An airport that is referenced by any proxy subscription cannot be deleted: SublinkPro rejects the delete request and reports how many proxy subscriptions use it. Re-point or delete those subscriptions first.

### Relayed response headers

Only headers that carry subscription metadata are forwarded:

- `Content-Type`
- `Content-Disposition`
- `subscription-userinfo`
- `profile-update-interval`
- `profile-web-page-url`
- `profile-title`
- `encrypt-tag`

`encrypt-tag` is relayed so that Happ can decrypt encrypted subscription content when the upstream returns it.

`Content-Encoding` is intentionally not relayed: the HTTP client already decompresses the upstream body, so forwarding the original encoding header would break clients.

---

## Error handling

| Situation | Response |
|:---|:---|
| Subscription has no airport | `502` with `Subscription is not linked to an airport` |
| Referenced airport no longer exists | `502` with `The airport linked to this subscription does not exist` |
| Airport subscription URL is missing or not HTTP(S) | `502` with `Invalid airport subscription URL` |
| Upstream returns non-2xx | `502` with a generic `Failed to fetch upstream subscription` body; the status code and reason are written to the server log |
| Share expired | `410` with `Subscription expired` |
| Upstream unreachable / timeout | `502` with the same generic body |
| Body larger than 10 MiB | `502`; the read is bounded to protect server memory |

Upstream URLs, tokens, and header values are never echoed back to the client.

> [!CAUTION]
> The upstream is fetched from the server's network, so anyone who can create a proxy subscription can make the server request an internal address. Only configure URLs you trust, and protect the admin account accordingly.

---

## Limitations

- Proxy subscriptions have no nodes, groups, scripts, filters, or naming rules; the Aggregated tab UI for those is hidden on the Proxy tab.
- Request settings cannot be overridden per subscription; edit the airport instead, and every proxy subscription that references it picks up the change.
- Preview, node sorting, and chain-proxy rules do not apply to proxy subscriptions.
- `client` and `substore` query parameters on `/c/` are ignored; the upstream decides the format from the relayed `User-Agent`.
- IP allow/deny lists are not available: `IPWhitelist` and `IPBlacklist` are ignored and cleared on save, even for legacy rows.
- The upstream is fetched on every client request; there is no caching layer.

---

## API

```bash
# Create
curl -X POST "http://localhost:8000/api/v1/subcription/add" \
  -H "X-API-Key: <key>" \
  -F "type=proxy" \
  -F "name=Upstream" \
  -F "AirportID=1" \
  -F "RefreshUsageOnRequest=false"

# List only proxy subscriptions
curl "http://localhost:8000/api/v1/subcription/get?page=1&pageSize=20&type=proxy" \
  -H "X-API-Key: <key>"

# Public subscription output (share token, not the API key)
curl "http://localhost:8000/c/?token=<share-token>"
```

See `skill-sublinkpro/reference/api.md` for the complete field list.
