# SublinkPro API Reference

Complete, verified endpoint catalog for the SublinkPro REST API. Base URL: `http://<host>:8000` (default port).

Every command below was cross-checked against the backend handlers (`api/*.go`, `routers/*.go`) **and** the live frontend calls (`webs/src/api/*.js`), which are the authoritative source for paths, methods, content types, and exact field names.

All routes under `/api/v1/...` require authentication via the `X-API-Key` header. The subscription-consume path `/c/` uses a share token instead (see Shares).

---

## How to Call (curl)

The primary, dependency-free way to call the API is **curl**, which ships on virtually every macOS, Windows 10+, and Linux system. (A Python convenience wrapper also exists — see "Optional helper script" below — but it is not required.)

Set these once per session:

```bash
export SUBLINK_BASE_URL=http://localhost:8000   # your instance address
export SUBLINK_API_KEY=prefix_xxx_yyy           # your API key
```

Four call patterns cover every endpoint:

```bash
# 1. GET (query params go in the URL)
curl -s -H "X-API-Key: $SUBLINK_API_KEY" \
  "$SUBLINK_BASE_URL/api/v1/nodes/get?page=1&pageSize=50"

# 2. POST form-encoded  (endpoints marked "form" below)
curl -s -H "X-API-Key: $SUBLINK_API_KEY" \
  --data-urlencode 'link=vless://...' \
  --data-urlencode 'group=US' \
  "$SUBLINK_BASE_URL/api/v1/nodes/add"

# 3. POST/PUT JSON  (endpoints marked "JSON" below)
curl -s -H "X-API-Key: $SUBLINK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name":"MyAirport","url":"https://...","cronExpr":"0 */6 * * *","enabled":true}' \
  "$SUBLINK_BASE_URL/api/v1/airports"

# 4. DELETE (often a query param)
curl -s -X DELETE -H "X-API-Key: $SUBLINK_API_KEY" \
  "$SUBLINK_BASE_URL/api/v1/nodes/delete?id=123"
```

**Critical — judge success by the body's `code`, not the HTTP status.** A failure commonly comes back as **HTTP 200 with `code:500`** in the body. Always parse the JSON and check `.code` (200 = success; 400/403/404/500 = failure, read `.msg`). For example: `... | python3 -c 'import sys,json;d=json.load(sys.stdin);sys.exit(0 if d.get("code")==200 else 1)'`, or just read the `code` field directly from the response.

For non-envelope responses (subscription output at `/c/`, SSE streams) there is no `{code,msg,data}` wrapper — read the raw body as-is.

Before calling, make sure `$SUBLINK_BASE_URL` and `$SUBLINK_API_KEY` are set; if not, guide the user to set them (see SKILL.md "Error Handling Playbook") rather than sending a broken request.

### Optional helper script

If Python 3 is available, `scripts/sublink.py` wraps the four patterns above (auth header, form-vs-JSON routing, and the `code` check with a non-zero exit on failure):

```bash
python scripts/sublink.py GET  /api/v1/nodes/get --query page=1
python scripts/sublink.py POST /api/v1/nodes/add --form link='vless://...' --form group='US'
python scripts/sublink.py POST /api/v1/airports --json name='MyAirport' --json enabled=true
```

It is a convenience only — everything it does can be done with curl. Use whichever is available.

---

## Response Envelope

All JSON endpoints return:
```json
{
  "code": 200,        // 200=success, 500=error, 400/403/404=client error
  "msg": "...",       // human-readable message
  "data": <any>,      // actual response data
  "i18nKey": "...",   // optional i18n key
  "i18nParams": {}     // optional i18n params
}
```
`code` lives in the body and is what matters — a failure can come back as HTTP 200 with `code:500`. Check `.code`, not the HTTP status.

## Content-Type Trap (read this — it's the #1 cause of "参数错误" failures)

Endpoints are NOT consistent. Sending the wrong body shape produces `参数错误` / validation errors. The rule, verified per-endpoint below:

- **Form-encoded** (`--form`): auth login, node add/update, subscription add/update, template add/update/delete. These read `c.PostForm(...)`.
- **JSON** (`--json`): access keys, **node batch operations**, **subscription preview**, subscription sort/batch-sort, chain rules, airports, tags, scripts, hosts, geoip, group-sort, node-check, settings, AI template endpoints. These use `c.ShouldBindJSON(...)`.
- **Query string** (`--query`): all GET lists/filters, plus several deletes and a few POSTs that take their args in the query (`copy`, `refresh`, `trigger`).

When a call fails with a parameter error, the most likely cause is form-vs-JSON mismatch or a wrong field name — re-check against the exact entry below before retrying.

Field-name casing is inconsistent across modules (e.g. node filters are lowerCamel query params, but subscription form fields and the preview JSON body are UpperCamel). Use the exact casing shown.

---

## Authentication & Access Keys

### Login (JWT) — usually not needed; prefer X-API-Key
**POST** `/api/v1/auth/login` — **form-encoded**
- Fields: `username`, `password`, `captchaKey`, `captchaCode`, `rememberMe`, optional `turnstileToken`
- Returns a JWT (14-day expiry) or an MFA challenge. Captcha is on by default, which is why headless login is impractical — use an API key instead.

### Get Captcha
**GET** `/api/v1/auth/captcha`

### Create API Key
**POST** `/api/v1/accesskey/add` — **JSON** (demo-restricted)
```json
{
  "username": "admin",                    // required — an existing username
  "expiredAt": "2025-12-31T23:59:59Z",    // optional, RFC3339
  "description": "AI skill"               // optional
}
```
Response data: `{"accessKey": "prefix_xxx_yyy"}` — **shown only once**.

### List API Keys
**GET** `/api/v1/accesskey/get/{userId}` (query: `?page=1&pageSize=20`)

### Delete API Key
**DELETE** `/api/v1/accesskey/delete/{accessKeyId}` (demo-restricted)

---

## Users

Base: `/api/v1/users`

- **GET** `/users/me` — current user info
- **POST** `/users/update` — **form** `username`, `password` (updates login credentials)
- **POST** `/users/change-password` — **JSON** (change password with old-password check)
- **POST** `/users/update-profile` — **JSON** (username / nickname)
- **GET** `/users/mfa` — TOTP status
- **POST** `/users/mfa/totp/begin` — **JSON**
- **POST** `/users/mfa/totp/confirm` — **JSON**
- **POST** `/users/mfa/totp/disable` — **JSON**
- **POST** `/users/mfa/recovery-codes/regenerate` — **JSON**

---

## Nodes

Base: `/api/v1/nodes`

### Add Node
**POST** `/nodes/add` — **form**
- **Required:** `link` (proxy URI, WireGuard config text, or Clash YAML containing `proxies:`)
- Optional: `name`, `nameMode` (`link` = use upstream name, `remark` = use custom `name`; `custom` is accepted as an alias of `remark`), `dialerProxyName`, `group`

### Update Node
**POST** `/nodes/update` — **form**
- Located by `oldlink` (and/or `oldname`), NOT by id.
- Fields: `oldname`, `oldlink`, `link` (required, the new/confirmed link), `name`, `nameMode`, `dialerProxyName`, `group`, `tags`

### Delete Node
**DELETE** `/nodes/delete` (query: `?id=123`)

### Batch Delete
**DELETE** `/nodes/batch-delete` — **JSON** `{"ids": [1,2,3]}`

### Batch Update Group / Dialer-Proxy / Source / Country (all JSON)
- **POST** `/nodes/batch-update-group` — `{"ids":[...], "group":"US"}`
- **POST** `/nodes/batch-update-dialer-proxy` — `{"ids":[...], "dialerProxyName":"..."}`
- **POST** `/nodes/batch-update-source` — `{"ids":[...], "source":"..."}`
- **POST** `/nodes/batch-update-country` — `{"ids":[...], "country":"US"}` manually assigns the same country code to selected nodes. Empty `country` clears the country marker.
- **POST** `/nodes/batch-fill-country` — **JSON** `{"airportId":123, "onlyEmpty":true}` fills node countries from enabled country rules by matching each node name. `airportId` is optional; when omitted, it scans all nodes. This is rule-based filling of empty/missing country values, not landing-IP detection. Use `/nodes/batch-update-country` when you already know the exact country code to set.

### List / Get Nodes
**GET** `/nodes/get` — query filters (all lowerCamel): `search`, `group`, `source`, `protocol`, `maxDelay`, `minSpeed`, `sortBy`, `sortOrder`, `page`, `pageSize`, array filters `countries[]`, `tags[]`. Returns `{items, total, page, pageSize, totalPages}` when paginated.

### Node Selector (compact picker list)
**GET** `/nodes/selector` — same query filters as `/nodes/get`; returns `{items:[{ID, Name, Group, Source, LinkCountry, ...}], total, ...}`. Use this to present nodes as a numbered pick list.

### Node Selector by IDs
**GET** `/nodes/selector/by-ids` (query: `?ids=1,2,3`)

### Get Node IDs (for "select all matching")
**GET** `/nodes/ids` — same filters as `/nodes/get`; returns the matching id array.

### Discovery lists (for building choices)
- **GET** `/nodes/groups` — all group names
- **GET** `/nodes/sources` — all sources
- **GET** `/nodes/countries` — all country codes present
- **GET** `/nodes/protocols` — all protocols in use
- **GET** `/nodes/group-stats` — per-group counts

### Other
- **GET** `/nodes/ip-info` (query: `?ip=1.2.3.4`)
- **GET** `/nodes/ip-cache/stats`
- **DELETE** `/nodes/ip-cache`
- **GET** `/nodes/protocol-ui-meta`
- **GET** `/nodes/parse-link` (query: `?link=vless://...`)
- **GET** `/nodes/raw-info` (query: `?id=123`)
- **POST** `/nodes/update-raw` — **JSON** `{"nodeId":123, "fields":{...}}` (demo-restricted)

---

## Subscriptions

Base: `/api/v1/subcription` (note the spelling — missing the second "s").

### Add Subscription
**POST** `/subcription/add` — **form**
- **Required:** `name` AND at least one of `nodeIds` (comma-separated node IDs) or `groups` (comma-separated group names).
- Output template: `config` = a template **filename** (the `file` value from `GET /template/get`, e.g. `clash.yaml`); empty = raw node list.
- Other optional form fields (UpperCamel): `scripts` (comma-separated script IDs), `IPWhitelist`, `IPBlacklist`, `DelayTime`, `MinSpeed`, `CountryWhitelist`, `CountryBlacklist`, `NodeNameRule`, `NodeNamePreprocess`, `NodeNameWhitelist`, `NodeNameBlacklist`, `TagWhitelist`, `TagBlacklist`, `ProtocolWhitelist`, `ProtocolBlacklist`, `DeduplicationRule`, `RefreshUsageOnRequest`, `UpdateInterval`, `MaxFraudScore`, `OnlyResidential`, `OnlyNative`, `ResidentialType`, `IPType`, `QualityStatus`, `UnlockProvider`, `UnlockStatus`, `UnlockKeyword`, `UnlockRules`, `UnlockRuleMode`.
- Name must be unique (duplicate → `订阅名称不能重复`).

### Update Subscription
**POST** `/subcription/update` — **form**, located by `oldname` (current name). Same field set as add, plus `oldname`.

### Get Subscriptions
**GET** `/subcription/get` (query: `?page=1&pageSize=20`). Paginated → `{items, total, ...}`; without paging → array.

### Delete Subscription
**DELETE** `/subcription/delete` (query: `?id=123`)

### Copy Subscription
**POST** `/subcription/copy` (query: `?id=123`)

### Sort Nodes
**POST** `/subcription/sort` — **JSON** `{"ID":123, "NodeSort":[{"ID":1,"Name":"...","Sort":0,"IsGroup":false}, ...]}`

### Batch Sort
**POST** `/subcription/batch-sort` — **JSON** `{"ID":123, "sortBy":"delay", "sortOrder":"asc"}` (`sortBy`: source|name|protocol|delay|speed|country)

### Preview Nodes (which nodes survive the filters)
**POST** `/subcription/preview` — **JSON** (NOT form — UpperCamel field names):
```json
{
  "SubscriptionID": 123,            // either preview a saved sub by id...
  "NodeIDs": [1,2,3], "NodeSorts": [0,1,2],   // ...or preview an unsaved form
  "Groups": ["US"], "GroupSorts": [0],
  "Scripts": [1],
  "DelayTime": 300, "MinSpeed": 5,
  "CountryWhitelist": "", "CountryBlacklist": "",
  "TagWhitelist": "", "TagBlacklist": "",
  "ProtocolWhitelist": "", "ProtocolBlacklist": "",
  "NodeNameWhitelist": "", "NodeNameBlacklist": "",
  "MaxFraudScore": 0, "OnlyResidential": false, "OnlyNative": false,
  "ResidentialType": "", "IPType": "", "QualityStatus": "",
  "UnlockProvider": "", "UnlockStatus": "", "UnlockKeyword": "",
  "UnlockRules": "", "UnlockRuleMode": ""
}
```

### Metadata helpers
- **GET** `/subcription/protocol-meta`
- **GET** `/subcription/node-fields-meta`

### NodeNameRule variables
`NodeNameRule` is evaluated when subscription output names are rendered. Country-related variables are based on the node's stored country code:

- `$LinkCountry` — country code, such as `HK` or `US`; empty values render as `未知`.
- `$LinkCountryName` — country name looked up from Country Rules by `$LinkCountry`; falls back to the code when no matching country-rule name exists.
- `$Flag` — flag emoji generated from the country code.
- `$DuplicateIndex` — empty for the first duplicate-name occurrence, then `1`, `2`, `3`, and so on.

Other supported variables include `$Name`, `$LinkName`, `$Group`, `$Source`, `$Protocol`, `$Index`, `$Tags`, `$TagGroup(name)`, speed/delay variables, IP quality variables, and unlock variables such as `$Unlock` and `$Unlock(netflix)`.

### Chain Proxy Rules (all under a subscription id; bodies are JSON)
- **GET** `/subcription/{id}/chain-rules` — list
- **POST** `/subcription/{id}/chain-rules` — **JSON** create
- **PUT** `/subcription/{id}/chain-rules/sort` — **JSON** `{"ruleIds":[...]}` (must be defined before the `:ruleId` route)
- **PUT** `/subcription/{id}/chain-rules/{ruleId}` — **JSON** update
- **DELETE** `/subcription/{id}/chain-rules/{ruleId}`
- **PUT** `/subcription/{id}/chain-rules/{ruleId}/toggle`
- **GET** `/subcription/{id}/chain-options`
- **GET** `/subcription/{id}/chain-rules/preview`

---

## Shares

Base: `/api/v1/shares`

### Get Shares (for a subscription)
**GET** `/shares/get` (query: `?subId=123`) — `subId` is **required**.

### Add Share
**POST** `/shares/add` — **JSON**
```json
{
  "subscription_id": 123,                  // required
  "name": "Public Link",                   // optional
  "token": "",                             // optional, auto-generated if empty
  "expire_type": 0,                        // 0=never, 1=after N days, 2=specific datetime
  "expire_days": 30,                       // when expire_type=1
  "expire_at": "2025-12-31T23:59:59Z"      // when expire_type=2 (RFC3339)
}
```

### Update Share
**POST** `/shares/update` — **JSON** (same fields as add + `id` + `enabled`)

### Delete Share
**DELETE** `/shares/delete` (query: `?id=123`)

### Refresh Share Token
**POST** `/shares/refresh` (query: `?id=123`)

### Get Share Logs
**GET** `/shares/logs` (query: `?shareId=123`)

### Consume Subscription (the actual output a client fetches)
**GET** `/c/` (query: `?token=<shareToken>&client=<idx>`)
- Use `--raw` (this is NOT a JSON envelope — it's the rendered clash/v2ray/surge output).
- `token` is the **share token** (from `/shares/get` or `/shares/add`), NOT the API key.
- Client type auto-detected from User-Agent, or force with `client`.

---

## Templates

Base: `/api/v1/template`

### Get Templates
**GET** `/template/get` (query: `?page=1&pageSize=20`) — returns each template's `file`, `text`, `category` (clash|surge), `ruleSource`, etc. The `file` value is what a subscription's `config` field references.

### Add Template
**POST** `/template/add` — **form**: `filename`, `text`, `category`, `ruleSource`, `useProxy` (`true`/`false`), `proxyLink`, `enableIncludeAll` (`true`/`false`)

### Update Template
**POST** `/template/update` — **form**: `filename`, `oldname`, `text`, `category`, `ruleSource`, `useProxy`, `proxyLink`, `enableIncludeAll`

### Delete Template
**POST** `/template/delete` — **form**: `filename`

### Template Usage
**GET** `/template/usage` (query: `?filename=...`)

### ACL4SSR Presets
**GET** `/template/presets`

### Convert Rules
**POST** `/template/convert` — **JSON**

### AI Template Editing (edit sessions, all JSON; require the server's AI assistant to be configured in the web UI)

The current Template AI contract is an edit-session preview flow. The model must return structured operations, not a full replacement template. The server applies those operations to the current template, validates the candidate preview, then the client can accept the preview into the editor. Accepting does **not** save the template. Save through `/template/update` after the candidate is copied into the editor.

Short-lived session limits:

- Session TTL: `15m`
- Cleanup interval: `5m`
- Max active sessions per authenticated user: `10`
- Sessions are in memory only and are not persistent edit history

#### Start Edit Session Stream

**POST** `/api/v1/template/ai/edit-sessions/stream`, **SSE** (use `--raw`)

Request body:

```json
{
  "filename": "clash.yaml",
  "category": "clash",
  "currentText": "mixed-port: 7890\n...",
  "userPrompt": "Add a Hong Kong auto select policy group",
  "ruleSource": "",
  "useProxy": false,
  "proxyLink": "",
  "enableIncludeAll": false
}
```

SSE event names:

- `template.edit.session.created`
- `template.edit.model.delta`
- `template.edit.operations.ready`
- `template.edit.preview.validating`
- `template.edit.preview.ready`
- `template.edit.warning`
- `template.edit.error`
- `template.edit.completed`

Final preview payload fields include:

```json
{
  "sessionId": "unguessable-id",
  "status": "preview_ready",
  "baseHash": "sha256-or-revision-hash",
  "candidateHash": "sha256-or-revision-hash",
  "candidateText": "mixed-port: 7890\n...",
  "operations": [
    {
      "op": "insert",
      "anchor": "proxy-groups:\n",
      "position": "after",
      "newString": "  - name: HK Auto\n",
      "description": "Add Hong Kong auto group"
    }
  ],
  "validation": {
    "errors": [],
    "warnings": [],
    "detectedType": "clash",
    "protectedTokensFound": []
  },
  "warningFingerprint": "",
  "expiresAt": "2026-06-24T12:15:00Z"
}
```

`candidateText` here is the server-materialized preview after operations were applied and validated. It is not AI model output.

#### Get Edit Session

**GET** `/api/v1/template/ai/edit-sessions/:sessionId`

Returns the current session preview state, including `sessionId`, `status`, `filename`, `category`, `baseHash`, `candidateText` when available, `operations`, `validation`, `warningFingerprint`, `createdAt`, `expiresAt`, and `lastError`.

#### Accept Edit Session

**POST** `/api/v1/template/ai/edit-sessions/:sessionId/accept`, **JSON**

Request body:

```json
{
  "currentText": "mixed-port: 7890\n..."
}
```

`currentText` is optional, but clients should send the editor's current text when accepting a preview, especially after a previous accepted preview has not been saved yet. It is used only as editor base proof: when its revision hash matches the session `baseHash`, accept can proceed even if the template file on disk still contains an earlier saved version. It is not the candidate, is never persisted, and does not replace the server-built `candidateText` returned by accept.

If `currentText` is absent or doesn't match the session base, the server falls back to the saved template file freshness check. If the saved file changed since the session was created, accept fails with `AI_EDIT_STALE_BASE`. Validation errors still block acceptance. Validation warnings remain in the session payload as review metadata, but clients don't send warning-related request fields when accepting.

Response data:

```json
{
  "sessionId": "unguessable-id",
  "candidateText": "mixed-port: 7890\n...",
  "candidateHash": "sha256-or-revision-hash",
  "validation": {
    "errors": [],
    "warnings": []
  }
}
```

Accept verifies the owner, session expiry, `preview_ready` status, base hash freshness through either matching `currentText` or saved file content, and that validation has no errors. It returns the candidate text for the editor. It does not persist the template.

#### Discard Edit Session

**POST** `/api/v1/template/ai/edit-sessions/:sessionId/discard`

Marks the session discarded. It does not change the template or editor content.

#### Operation Schema

AI output must be an operation array. Supported v1 operations are exact-match only. `match` is optional and defaults to `unique` when omitted:

```json
[
  {
    "op": "replace",
    "oldString": "old exact text",
    "newString": "new text",
    "match": "unique",
    "description": "optional human summary"
  },
  {
    "op": "insert",
    "anchor": "exact anchor text",
    "position": "before",
    "newString": "inserted text",
    "description": "optional human summary"
  },
  {
    "op": "delete",
    "oldString": "exact text to remove",
    "match": "all",
    "description": "optional human summary"
  }
]
```

Rules:

- `replace` requires non-empty `oldString`, allows any `newString`, allows optional `match` of `unique` or `all` (empty is treated as the default unique mode), and forbids `anchor` and `position`.
- `insert` requires non-empty `anchor`, non-empty `newString`, and `position` of `before` or `after`, and forbids `oldString` and any explicit non-empty `match`. Anchors must stay exact and unique.
- `delete` requires non-empty `oldString`, allows optional `match` of `unique` or `all` (empty is treated as the default unique mode), and forbids `newString`, `anchor`, and `position`.
- In default `unique` mode, every target or anchor must match exactly once after line endings are normalized for matching.
- In explicit `all` mode, `replace` and `delete` apply to every exact occurrence of `oldString`; zero occurrences still fail with `PATCH_NO_MATCH`.
- `all` mode is intended only when the user clearly asks for all/every/全部/所有 exact occurrences. It does not enable fuzzy, regex, YAML-path, AST, or line-number matching.
- Operation application is atomic. If one operation fails, no preview is ready.

#### Error Codes

Common edit-session error codes:

- `AI_EDIT_INVALID_OPERATION`: operation type or fields are invalid
- `AI_EDIT_EMPTY_OPERATIONS`: model returned no operations
- `AI_EDIT_MALFORMED_OUTPUT`: model output was not valid contract JSON
- `AI_EDIT_LEGACY_OUTPUT`: model tried to return legacy full-template output
- `AI_EDIT_SESSION_LIMIT`: user has too many active sessions
- `AI_EDIT_SESSION_NOT_FOUND`: session id doesn't exist
- `AI_EDIT_SESSION_EXPIRED`: session expired, generate a new preview
- `AI_EDIT_STALE_BASE`: template base changed since the session was created, or the supplied `currentText` doesn't prove the editor still matches the session base while the saved file has also changed
- `AI_EDIT_VALIDATION_FAILED`: validation errors block acceptance
- `PATCH_NO_MATCH`: operation target or anchor was not found
- `PATCH_AMBIGUOUS_MATCH`: operation target or anchor matched more than once in default `unique` mode
- `PATCH_EMPTY_TARGET`: operation target was empty
- `PATCH_NOOP`: operation would not change content
- `PATCH_INVALID_INSERT_POSITION`: insert position was not `before` or `after`
- `PATCH_INVALID_MATCH_MODE`: operation used an unsupported `match` value or requested `match: "all"` for an insert at patch-application time

#### Breaking Change From Legacy Full-Template Generation

The old full-template AI endpoints are removed or return `AI_EDIT_LEGACY_REMOVED`:

- `POST /api/v1/template/ai/generate`
- `POST /api/v1/template/ai/generate-stream`
- `POST /api/v1/template/ai/validate`
- `POST /api/v1/template/ai/apply`

There is no supported fallback that accepts AI-generated full-template `candidateText` as the source of truth. Clients should use `/api/v1/template/ai/edit-sessions/*` and save accepted text through the normal template save flow.

> If `/api/v1/template/ai/edit-sessions/stream` returns an upstream or credential error, the server's AI assistant isn't configured, or its provider failed. The `/api/v1/settings/ai-assistant*` config endpoints reject API-key auth (login-session only), so this must be set up in the web UI.

---

## Airports

Base: `/api/v1/airports` (all write bodies are JSON)

- **GET** `/airports` (query: `?page=1&pageSize=20`) — list
- **GET** `/airports/{id}` — one
- **POST** `/airports` — **JSON** create. Key fields (see `dto.AirportRequest`): `name`(required), `url`(required, valid URL), `cronExpr`(required), `enabled`, `group`, `downloadWithProxy`, `proxyLink`, `userAgent`, `fetchUsageInfo`, `skipTLSVerify`, plus filter/rename/dedup fields (`nodeNameWhitelist`, `nodeNameBlacklist`, `protocolWhitelist`, `protocolBlacklist`, `nodeNamePreprocess`, `deduplicationRule`, `nodeNameUniquify`, ...), and country-fill fields (`autoFillCountry`, `backfillExistingCountry`).
- **PUT** `/airports/{id}` — **JSON** update (same schema)
- **POST** `/airports/batch-update` — **JSON** `{ids:[...], applyGroup, group, applySchedule, cronExpr}`
- **DELETE** `/airports/{id}` (query: `?deleteNodes=false`)
- **POST** `/airports/pull-all` — pull all enabled airports now
- **POST** `/airports/{id}/pull` — pull one now
- **POST** `/airports/{id}/refresh-usage`

Country-fill fields apply during airport pulls and use Country Rules against upstream node names:

- `autoFillCountry`: fills the country code for newly imported nodes whose country is empty.
- `backfillExistingCountry`: fills the country code for every existing node from the same airport whose country is still empty, including nodes that otherwise do not need a name/link/order update. Saving an airport with `backfillExistingCountry:true` also persists `autoFillCountry:true`.

Neither field overwrites an existing country code, and neither performs landing-IP detection.

---

## Country Rules

Base: `/api/v1/country-rules`

Country Rules define how SublinkPro parses country codes from node names and how it resolves country names for naming variables. Enabled rules are regex-based, sorted by `priority` descending, and matching stops at the first rule that matches the node name.

Fields:

| Field | Meaning |
|:---|:---|
| `countryCode` | Country code stored on the node, usually ISO-style codes such as `HK`, `US`, `JP` |
| `countryName` | Display name used by `$LinkCountryName` and country-related UI |
| `pattern` | Regular expression matched against the node name; use `(?i)` for case-insensitive patterns |
| `priority` | Integer priority from `0` to `1000`; higher numbers match first |
| `enabled` | Whether the rule participates in matching |

Endpoints:

- **GET** `/country-rules` — list rules. Query supports `page`, `pageSize`, and `keyword`.
- **POST** `/country-rules` — **JSON** create, for example `{"countryCode":"HK","countryName":"香港","pattern":"(?i)香港|HK|Hong Kong|🇭🇰","priority":100,"enabled":true}`.
- **PUT** `/country-rules/{id}` — **JSON** update.
- **DELETE** `/country-rules/{id}` — delete.
- **POST** `/country-rules/test` — **JSON** test one pattern against one node name, for example `{"pattern":"(?i)香港|HK","testName":"HK Premium 01"}`.
- **POST** `/country-rules/batch-test` — **JSON** test enabled rules against multiple node names.
- **POST** `/country-rules/batch` — **JSON** import or replace rule lists.
- **GET** `/country-rules/export` — export all country rules as text.
- **POST** `/country-rules/sync` — **JSON** full sync from text, `{"text":"..."}`.

Text export/sync format is one rule per line:

```text
countryCode countryName priority enabled pattern
HK 香港 100 true (?i)香港|HK|Hong Kong|🇭🇰
```

Use Country Rules together with airport `autoFillCountry` / `backfillExistingCountry`, `/nodes/batch-fill-country`, and subscription `NodeNameRule` variables such as `$LinkCountryName`.

---

## Tags

Base: `/api/v1/tags`

- **GET** `/tags/list` — all tags
- **POST** `/tags/add` — **JSON**
- **POST** `/tags/update` — **JSON**
- **DELETE** `/tags/delete` (query: `?name=<tagName>`) — note: deletes by **name**, not id
- **GET** `/tags/groups` — tag groups
- **GET** `/tags/rules` — rule list
- **POST** `/tags/rules/add` — **JSON**
- **POST** `/tags/rules/update` — **JSON**
- **DELETE** `/tags/rules/delete` (query: `?id=<ruleId>`)
- **POST** `/tags/rules/trigger` (query: `?id=<ruleId>`) — apply a rule
- **POST** `/tags/node/add` — **JSON** add a tag to a node
- **POST** `/tags/node/remove` — **JSON**
- **POST** `/tags/node/batch-add` — **JSON**
- **POST** `/tags/node/batch-set` — **JSON** (overwrite mode)
- **POST** `/tags/node/batch-remove` — **JSON**
- **GET** `/tags/node/tags` (query: `?nodeId=123`)

---

## Scripts

Base: `/api/v1/script`

- **GET** `/script/list` (query: `?page=1&pageSize=20`)
- **POST** `/script/add` — **JSON** (demo-restricted)
- **POST** `/script/update` — **JSON** (demo-restricted)
- **DELETE** `/script/delete` — **JSON** (demo-restricted)
- **GET** `/script/usage` (query: `?...`)

---

## Tasks

Base: `/api/v1/tasks`

- **GET** `/tasks` (query: `?page=...&pageSize=...`) — list (path is exactly `/api/v1/tasks`, no trailing segment)
- **GET** `/tasks/stats`
- **GET** `/tasks/running`
- **GET** `/tasks/{id}`
- **GET** `/tasks/{id}/traffic` (query: filters/paging)
- **POST** `/tasks/{id}/stop`
- **DELETE** `/tasks` — **JSON** body for clearing history

---

## Node Check (speed test / unlock check)

Base: `/api/v1/node-check`

- **GET** `/node-check/meta`
- **GET** `/node-check/profiles` — list profiles
- **GET** `/node-check/profiles/{id}`
- **POST** `/node-check/profiles` — **JSON** create (demo-restricted)
- **PUT** `/node-check/profiles/{id}` — **JSON** (demo-restricted)
- **DELETE** `/node-check/profiles/{id}` (demo-restricted)
- **POST** `/node-check/profiles/{id}/run` — run a profile (demo-restricted)
- **POST** `/node-check/run` — **JSON** `{"profileId":0, "nodeIds":[...]}` — general run entry (demo-restricted)

---

## Hosts

Base: `/api/v1/hosts`

- **GET** `/hosts/list`
- **POST** `/hosts/add` — **JSON**
- **POST** `/hosts/update` — **JSON**
- **DELETE** `/hosts/delete` (query: `?id=123`)
- **DELETE** `/hosts/batch-delete` — **JSON** `{"ids":[...]}`
- **GET** `/hosts/export` — hosts as text
- **POST** `/hosts/sync` — **JSON** `{"text":"..."}`
- **GET** `/hosts/settings`
- **POST** `/hosts/settings` — **JSON**
- **POST** `/hosts/pin` — **JSON** `{"id":123, "pinned":true}`

---

## GeoIP

Base: `/api/v1/geoip`

- **GET** `/geoip/config`
- **PUT** `/geoip/config` — **JSON**
- **GET** `/geoip/status`
- **POST** `/geoip/download`
- **POST** `/geoip/stop`

---

## Group Sort

Base: `/api/v1/group-sort`

- **GET** `/group-sort/groups`
- **GET** `/group-sort/detail` (query: `?group=<name>`)
- **POST** `/group-sort/save` — **JSON** `{"groupName":"...", "airportSorts":[...]}`

---

## Settings

Base: `/api/v1/settings` (write bodies are JSON unless noted; most writes are demo-restricted)

**Webhooks:** **GET** `/settings/webhooks` · **POST** `/settings/webhooks` · **PUT** `/settings/webhooks/{id}` · **DELETE** `/settings/webhooks/{id}` · **POST** `/settings/webhooks/{id}/test`

**Base templates:** **GET** `/settings/base-templates` · **POST** `/settings/base-templates` (`{category, content}`)

**System domain:** **GET** `/settings/system-domain` · **POST** `/settings/system-domain`

**Node dedup:** **GET** `/settings/node-dedup` · **POST** `/settings/node-dedup`

**Global node processing:** **GET** `/settings/global-node-processing` · **POST** `/settings/global-node-processing` (`{nodeNameWhitelist, nodeNameBlacklist, protocolWhitelist, protocolBlacklist, nodeNamePreprocess}`)

**Telegram:** **GET** `/settings/telegram` · **POST** `/settings/telegram` · **POST** `/settings/telegram/test` · **GET** `/settings/telegram/status` · **POST** `/settings/telegram/reconnect`

**Cloudflared:** **GET** `/settings/cloudflared` · **POST** `/settings/cloudflared` · **POST** `/settings/cloudflared/start` · **POST** `/settings/cloudflared/stop` · **DELETE** `/settings/cloudflared/token`

**Sub-Store:** **GET** `/settings/substore` · **POST** `/settings/substore` · **POST** `/settings/substore/test`

**Database migration:** **POST** `/settings/database-migration/import` — **multipart/form-data** (upload a backup.zip / .db)

**AI assistant (login-session only — reject API key with 403 by design):**
- **GET** `/settings/ai-assistant`
- **POST** `/settings/ai-assistant` (`{baseUrl, model, requestType, apiKey, temperature, maxTokens, extraHeaders}`; `requestType` is `responses` or `chat_completions`)
- **POST** `/settings/ai-assistant/models`
- **POST** `/settings/ai-assistant/test` (accepts the same provider fields and uses `requestType` to test `/responses` or streamed `/chat/completions`)

> These four configure the stored AI provider key, so they only work in a logged-in web session. An API-key call always gets 403 here — that's expected, not a bug. Configure the AI assistant once in the web UI; afterwards `/template/ai/*` works with API-key auth.

---

## Dashboard / Stats

Base: `/api/v1/total`

- **GET** `/total/sub` — subscription count
- **GET** `/total/node` — node count
- **GET** `/total/fastest-speed` — fastest node
- **GET** `/total/lowest-delay` — lowest-delay node
- **GET** `/total/country-stats`
- **GET** `/total/dashboard-country-stats`
- **GET** `/total/protocol-stats`
- **GET** `/total/system-stats`
- **GET** `/total/tag-stats`
- **GET** `/total/group-stats`
- **GET** `/total/source-stats`
- **GET** `/total/dashboard-grouped-stats`
- **GET** `/total/quality-stats`

---

## Other Endpoints

### Version (public — no API key needed; ideal health check)
**GET** `/api/v1/version`

### Backup
**GET** `/api/v1/backup/download` — download a backup (demo-restricted). (Import is via `POST /settings/database-migration/import`.)

### Server-Sent Events
**GET** `/api/se` (query: `?token=<jwt>`) — SSE stream (auth-protected).

---

## Notes

- **Demo mode:** write endpoints marked demo-restricted are blocked when the instance runs in demo mode.
- **Pagination:** most list endpoints accept `?page=&pageSize=` and then return `{items, total, page, pageSize, totalPages}`.
- **Content type is per-endpoint** — see the Content-Type Trap section above. When in doubt, match the exact entry here rather than guessing.
