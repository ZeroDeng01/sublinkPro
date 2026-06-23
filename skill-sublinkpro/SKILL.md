---
name: sublinkpro
description: Manage SublinkPro proxy subscriptions via its REST API — add/query nodes, build and share subscriptions, manage airports, templates (incl. AI template editing), tags, and view dashboard stats. Use when the user wants to operate their SublinkPro instance.
---

# SublinkPro AI Skill

Control your SublinkPro instance through natural language. This skill lets an AI assistant call the SublinkPro API directly, so you can manage nodes, subscriptions, shares, airports, and templates without opening the web UI.

## Installation

1. **Copy this directory** into your AI agent's skills directory. The exact path depends on your tool — for Claude Code it is `~/.claude/skills/sublinkpro`:
   ```bash
   cp -r skill-sublinkpro ~/.claude/skills/sublinkpro
   # or symlink if you cloned the repo:
   ln -s /path/to/sublinkE/skill-sublinkpro ~/.claude/skills/sublinkpro
   ```
   For other agents, place the directory wherever that tool loads skills from.

2. **Set environment variables**:
   ```bash
   export SUBLINK_BASE_URL=http://localhost:8000  # your SublinkPro server address
   export SUBLINK_API_KEY=prefix_xxx_yyy          # your API key
   ```

3. **Get an API key**:
   - Open SublinkPro web UI
   - Go to **Settings → Access Keys**
   - Click **Create** (or **Add**)
   - Copy the key (shown only once) and set `SUBLINK_API_KEY`

## How to Use

You don't need to know the API, the parameter names, or what data already exists. Just say what you want in plain language and the assistant guides you the rest of the way. If you're not sure what's possible, simply ask **"what can I do?"** and you'll get a menu.

Example openers:

- **"I want to add a node"** → the assistant walks you through it step by step
- **"Add this node: `vless://...`"**
- **"Help me build a new subscription"**
- **"Show me all my subscriptions"**
- **"Create a share link"**
- **"Import an airport"**
- **"Edit the 'ACL4SSR' template to block ads"**
- **"Show dashboard stats"**
- **"How does the smart tag system work?"** → the assistant fetches and explains the relevant documentation
- **"I don't have SublinkPro yet — help me set it up"** → the assistant guides you through deploying it (see Deploy below)

Under the hood the assistant calls the SublinkPro REST API with curl (or the optional Python helper). Authentication and API quirks are handled for you.

---

## Interaction Style: Guided by Default

**This is the most important section. The agent must treat every request as a guided, conversational flow — never a one-shot command that fails on a missing field.** Users generally do not know the API, the parameter names, or what data already exists. Lead them.

Core rules for the agent:

1. **Never guess or invent required values.** If a required field is missing, either discover it (rule 2) or ask the user — do not fabricate names, IDs, links, or URLs.
2. **Discover before you ask the user to type.** Before asking "which group / which subscription / which node?", call the matching GET endpoint and present the existing options as a short numbered list so the user can pick a number instead of typing. Examples:
   - which group → `GET /api/v1/nodes/groups`
   - which subscription → `GET /api/v1/subcription/get`
   - which nodes → `GET /api/v1/nodes/selector` (compact id+name list)
   - which airport → `GET /api/v1/airports`
   - which template → `GET /api/v1/template/get`
   - which country / protocol / source filters → `GET /api/v1/nodes/countries` / `/protocols` / `/sources`
3. **Collect required fields one step at a time**, in plain language. Don't dump a 15-field form on the user. Ask for what's required first; then offer optional refinements as a single "want to set any filters/options? (or say 'no' to skip)" step.
4. **Explain choices in the user's terms.** When a field has fixed options (e.g. share `expire_type`: never / N days / specific date), describe them in words and let the user choose, then map to the correct value yourself.
5. **Always confirm before any write (POST/PUT/DELETE).** Show a plain-language summary of exactly what will happen ("I'll create a subscription named 'US Servers' with these 3 nodes, no filters") and wait for a yes.
6. **Confirm extra carefully for destructive or bulk actions** — delete, batch-delete, batch-update-*, share token refresh, airport delete. Show the count and the names/IDs that will be affected, and make clear it can't be easily undone.
7. **Report results in plain language, not raw JSON.** On success, say what was created/changed and surface the useful bits (e.g. the share URL). On error, read the `msg`, explain it simply, and propose the fix.
8. **Offer the logical next step.** After adding nodes → offer to put them in a subscription. After creating a subscription → offer to create a share link. After importing an airport → offer to pull its nodes now.
9. **Respect mixed content types and the envelope** (see Technical Reference). The user never sees `--form` vs `--json`; the agent handles it.

## Capability Menu

When the user asks "what can you do", "help", or seems unsure where to start, present this menu and ask which one they'd like:

- **Nodes** — add a node (paste a link / WireGuard / Clash YAML), list or search nodes, move nodes between groups, set country/source, delete nodes
- **Subscriptions** — build a new subscription, edit one, preview which nodes it will output, sort nodes, copy a subscription, delete one
- **Sharing** — create a share link for a subscription, set expiration, list/refresh/disable share links, view access logs, fetch the actual subscription output
- **Airports** — import an airport (URL + schedule), pull nodes now, list airports, edit/batch-edit, refresh usage, delete
- **Templates** — list templates, edit a template by hand, or use **AI template editing** (describe a change in words → generate → review → apply)
- **Chain proxy** — view/add chain-proxy rules on a subscription
- **Tags** — list tags, create rules, apply tagging
- **Tasks** — view running/finished tasks (speed tests etc.), stop a task
- **Dashboard** — node/subscription counts, fastest node, country/protocol stats, system stats
- **Documentation & help** — explain features, configuration, installation, or troubleshooting by fetching the project's official docs (smart tags, speed tests, chain proxy, MFA, etc.)
- **Deploy / install SublinkPro** — stand up a new instance via docker, docker-compose, or the install script, locally or on a remote host (see the Deploy workflow below)
- **Manage an instance** — check status / view logs, update to the latest version, or uninstall

---

## Documentation & Help

When the user asks how a feature works, how to configure something, what a feature does, or how to troubleshoot — fetch the project's official documentation from GitHub and answer from it. **Do not invent or paraphrase from memory.** The documentation is always current and authoritative.

**How to use the docs:**

1. Match the user's question to a topic in `reference/docs.md` (the documentation map).
2. Fetch the relevant doc's raw URL from GitHub using your web-fetch capability.
3. Read it and answer in the user's language — summarize and quote the doc directly.
4. Cite the human-readable GitHub page URL so the user can read the full doc themselves.
5. If you can't fetch the doc (offline / no web access), tell the user and give them the GitHub link to read it themselves — never fabricate the content.

**Topics covered:** installation & configuration, smart tags, speed tests, unlock checks, chain proxy, AI template editing, airport management, subscription sharing, host management, Cloudflare Tunnel, Telegram bot, MFA, script support, development guide, protocol extensions, internationalization.

**When docs don't solve the problem:**

If the documentation can't answer the user's question, or they've tried the documented solution and it still doesn't work, guide them to escalate:

- **Submit a GitHub issue:** [https://github.com/ZeroDeng01/sublinkPro/issues](https://github.com/ZeroDeng01/sublinkPro/issues) — for bug reports, feature requests, or questions the docs don't cover.
- **Contact the author via Telegram:** [https://t.me/SublinkPro_ChatMeBot](https://t.me/SublinkPro_ChatMeBot) — for real-time help or urgent issues.

Phrase it like: "The docs cover X but your specific case (Y) isn't documented. I'd recommend opening a GitHub issue at [link] or reaching out to the author on Telegram at [link] — they can help diagnose this."

---

## Technical Reference (for the AI agent)

### How to Call the API

Call the REST API with **curl** — it ships on virtually every macOS, Windows 10+, and Linux system, so it's the most portable choice and the default. (A Python convenience wrapper, `scripts/sublink.py`, also exists for hosts that have Python 3 — see the end of this section — but it is optional.)

First ensure the env vars are set (check before calling; if missing, guide the user — see "Error Handling Playbook"):

```bash
export SUBLINK_BASE_URL=http://localhost:8000
export SUBLINK_API_KEY=prefix_xxx_yyy
```

**Each endpoint in `reference/api.md` is tagged with a content type. Map it to curl like this:**

| Tag in api.md | curl form |
|---|---|
| **GET** + query | `curl -s -H "X-API-Key: $SUBLINK_API_KEY" "$SUBLINK_BASE_URL<path>?k=v&k2=v2"` |
| **form** (POST) | `curl -s -H "X-API-Key: $SUBLINK_API_KEY" --data-urlencode 'k=v' --data-urlencode 'k2=v2' "$SUBLINK_BASE_URL<path>"` |
| **JSON** (POST/PUT) | `curl -s -H "X-API-Key: $SUBLINK_API_KEY" -H "Content-Type: application/json" -d '{"k":"v"}' "$SUBLINK_BASE_URL<path>"` |
| **DELETE** | `curl -s -X DELETE -H "X-API-Key: $SUBLINK_API_KEY" "$SUBLINK_BASE_URL<path>?id=123"` |
| **raw** (`/c/`, SSE) | same as GET, but read the body as-is (no `{code,msg,data}` wrapper) |

Which endpoints are form vs JSON:
- **form** (`c.PostForm` in backend): auth login, node add/update, subscription add/update, template add/update/delete.
- **JSON** (`c.ShouldBindJSON`): access keys, **node batch operations**, **subscription preview**, subscription sort/batch-sort, chain rules, airports, tags (except `/tags/delete` which is a query param), scripts, hosts, geoip, group-sort, node-check, settings, AI template endpoints.
- **query**: all GET list/filter params, plus several deletes and a few POSTs (`copy`, `refresh`, `trigger`).

**Judge success by the body's `code`, not the HTTP status** — a failure often returns HTTP 200 with `code:500`. Parse the JSON, check `.code` (200 = ok; 400/403/404/500 = failure, read `.msg` and translate it). The #1 cause of `参数错误` failures is a form-vs-JSON mismatch or wrong field casing — re-check the endpoint's tag and exact field names in `reference/api.md` before retrying.

**Examples:**
```bash
# Get version (public, no key needed)
curl -s "$SUBLINK_BASE_URL/api/v1/version"

# Add a node (form)
curl -s -H "X-API-Key: $SUBLINK_API_KEY" \
  --data-urlencode 'link=vless://...' --data-urlencode 'group=US' \
  "$SUBLINK_BASE_URL/api/v1/nodes/add"

# Create an airport (JSON)
curl -s -H "X-API-Key: $SUBLINK_API_KEY" -H "Content-Type: application/json" \
  -d '{"name":"MyAirport","url":"https://...","cronExpr":"0 */6 * * *","enabled":true}' \
  "$SUBLINK_BASE_URL/api/v1/airports"

# List nodes (GET + query)
curl -s -H "X-API-Key: $SUBLINK_API_KEY" \
  "$SUBLINK_BASE_URL/api/v1/nodes/get?page=1&pageSize=50"

# Fetch subscription output (raw — uses a share token, not the API key)
curl -s "$SUBLINK_BASE_URL/c/?token=<shareToken>"
```

**Optional helper script.** If Python 3 is present, `scripts/sublink.py` does the auth header, form-vs-JSON routing, and `code` check for you, exiting non-zero on failure. The `--form`/`--json`/`--query`/`--raw` flags used throughout this doc map directly to the curl table above:

```bash
python scripts/sublink.py GET  /api/v1/nodes/get --query page=1 --query pageSize=50
python scripts/sublink.py POST /api/v1/nodes/add --form link='vless://...' --form group='US'
python scripts/sublink.py POST /api/v1/airports --json name='MyAirport' --json enabled=true
python scripts/sublink.py GET  /c/ --query token='<shareToken>' --raw
```

Use whichever is available on the host — curl first, the script if you prefer and Python is installed. Throughout the workflows below, an endpoint marked `(--form)` / `(--json)` just means "use that content type" — apply it with either tool.

### Key API Patterns

**Base URL:** All routes are under `/api/v1/...` (except `/c/` for subscription consumption)

**Response envelope:** Every JSON response has `{code, msg, data}` where `code: 200` = success, `code: 500` = error

**Subscription group spelling:** `/api/v1/subcription` (missing second 's')

**Demo mode:** Write endpoints with `DemoModeRestrict` are blocked in demo mode

**Share tokens vs API keys:** Subscription consumption (`/c/?token=...`) uses share tokens (from `POST /api/v1/shares/add`), NOT the API key

### Full Endpoint Catalog

See `reference/api.md` for the complete list of endpoints with parameters and response shapes.

### Field Discovery Map (turn blanks into multiple-choice)

**This is how the skill replaces complex web-UI widgets.** In the web UI, fields like "Clash template", "nodes", "group", "country filter" are dropdowns, multi-selects, and visual builders. In this skill, the agent reproduces that experience by calling a discovery endpoint, showing the results as a **numbered list**, and letting the user pick a number. The agent **never** asks the user to free-type a value that can be discovered, and **never** invents one.

Whenever you need one of these values, call its source first and present choices:

| Field (what the user picks) | Discovery call | What to show / how it maps |
|---|---|---|
| **Clash/Surge template** (subscription `config`) | `GET /api/v1/template/get` | List each `file` with its `category` (e.g. `clash.yaml` — clash). The chosen `file` string IS the `config` value. Empty = no template (raw node list). |
| **Nodes** (`nodeIds`) | `GET /api/v1/nodes/selector` (paginated; supports `--query group=` / `country=` / `protocol=` filters) | Show `items[].ID` + `Name` (+ `Group`/`LinkCountry` for context). Selected IDs → comma-joined `nodeIds`. For "all matching", use `GET /api/v1/nodes/ids`. |
| **Group(s)** (`groups`, or node `group`) | `GET /api/v1/nodes/groups` | Numbered list of group names. Selected → comma-joined. |
| **Country filter** (`CountryWhitelist`/`CountryBlacklist`) | `GET /api/v1/nodes/countries` | List country codes present; user picks; comma-join. |
| **Protocol filter** (`ProtocolWhitelist`/`ProtocolBlacklist`) | `GET /api/v1/nodes/protocols` | List protocols in use; user picks; comma-join. |
| **Source filter** (node `source`) | `GET /api/v1/nodes/sources` | List sources; user picks. |
| **Tag filter** (`TagWhitelist`/`TagBlacklist`) | `GET /api/v1/tags/list` | List tag names; user picks; comma-join. |
| **Scripts** (`scripts`) | `GET /api/v1/script/list` | List scripts by id+name; selected IDs → comma-joined. |
| **Subscription** (to edit/share) | `GET /api/v1/subcription/get` | List by id+name. |
| **Airport** (to pull/edit) | `GET /api/v1/airports` | List by id+name. |

If a discovery call returns an empty list (e.g. no groups yet), say so and offer the alternative (e.g. "you have no groups — want to pick individual nodes instead, or create the group while adding?").

### Field Tiers (don't dump everything at once)

For multi-field resources (subscriptions especially), split fields into three tiers and walk them in order. Never present all fields at once.

1. **Required** — ask first, can't proceed without. (Subscription: `name` + at least one of `nodeIds`/`groups`.)
2. **Common optional** — offer as a short, skippable batch with sensible defaults. (Subscription: template `config`, country/protocol filter, `DelayTime`, `MinSpeed`, `UpdateInterval`.)
3. **Advanced** — only mention they exist; set only if the user explicitly asks. These include the **structured-JSON builder fields** (`DeduplicationRule`, `NodeNamePreprocess`, `UnlockRules`) which mirror visual builders in the UI and are easy to malform. Do not hand-craft their JSON from a vague request; if the user wants them, ask precise questions or point them to the web UI for that specific widget, and confirm the exact JSON before sending.

### Guided Workflows

These describe how to *lead the user* through each task. Discovery calls (GET) need no confirmation — run them freely to populate choices. Writes (POST/PUT/DELETE) always need a confirmation step. Use the Field Discovery Map above for every "which X?" question.

**Add a node**
1. Ask for the node link (proxy URI, WireGuard config, or Clash YAML). If the user already pasted one, skip.
2. Ask which group → discover via `GET /api/v1/nodes/groups`, numbered list (plus "new group" / "no group").
3. Optionally ask for a custom name (else the link's own name is used).
4. Confirm, then `POST /api/v1/nodes/add` (`--form`). Report the result; offer to add it to a subscription.

**Build a subscription** (the flagship multi-field flow — follow the tiers)
1. **Required — name:** ask for it. (Will be rejected if it duplicates an existing one; if so, ask for another.)
2. **Required — node selection:** ask "pick by group, or pick individual nodes?"
   - By group → `GET /api/v1/nodes/groups`, numbered list, multi-pick → `groups` (comma-joined).
   - By node → `GET /api/v1/nodes/selector` (offer to narrow with a group/country/protocol filter first since lists can be long), numbered list → `nodeIds` (comma-joined). At least one of the two is required.
3. **Common optional — template (`config`):** "Which output template? I'll list your Clash/Surge templates." → `GET /api/v1/template/get`, list each `file`+`category`, the picked `file` becomes `config`. Offer "none" (raw nodes).
4. **Common optional — filters (one skippable step):** offer country / protocol filter (discover via the map), plus `DelayTime` (max latency ms), `MinSpeed` (min MB/s), `UpdateInterval` (hours). Describe each plainly; default all to off. Say "or just say 'skip' to use none."
5. **Advanced:** mention dedup / node-name preprocessing / unlock rules exist; only collect if explicitly requested (see Field Tiers caution).
6. **Confirm** the whole picture in plain words ("Subscription 'US-Servers': 12 nodes from group 'US', clash.yaml template, protocol filter vless+trojan, max delay 300ms, no other filters"), then `POST /api/v1/subcription/add` (`--form`).
7. Offer **preview** or to create a **share link** next. Preview-before-create is a good default to suggest when filters are involved. **Note:** `POST /api/v1/subcription/preview` is **JSON with capitalized field names** (`NodeIDs` as an int array, `Groups`, `DelayTime`, `CountryWhitelist`, ...) — it does NOT mirror the `add` form fields. See `reference/api.md` "Preview subscription nodes" for the exact shape.

**Edit a subscription**
1. Pick it via `GET /api/v1/subcription/get`. Show current values so the user edits from a known state, not a blank form.
2. Ask which aspect to change (name / nodes / template / filters), discover options for that aspect via the map, confirm, then `POST /api/v1/subcription/update` (`--form`). **Key detail:** the subscription is located by `oldname` (its current name); `name` is the new name (use the same value if not renaming). You must resend the full field set you want to keep — omitted fields are not preserved. So fetch current values first and re-send them plus your edits.

**Create a share link**
1. Pick the subscription via `GET /api/v1/subcription/get` (numbered list).
2. Ask about expiration in words: never / after N days / specific date → map to `expire_type` 0/1/2 (+ `expire_days` or `expire_at`).
3. Confirm, then `POST /api/v1/shares/add` (`--json`).
4. Report the share token and build the consumable URL: `GET /c/?token=<shareToken>` (use `--raw` if fetching the actual output).

**Import an airport**
1. Ask for the airport URL.
2. Ask for an update schedule in words (e.g. "every 6 hours") → cron expr; offer a sensible default.
3. Ask name + optional group/usage tracking.
4. Confirm, then `POST /api/v1/airports` (`--json`).
5. Offer to pull nodes immediately: `POST /api/v1/airports/{id}/pull`, then show new nodes via `GET /api/v1/nodes/get --query groups=<airportGroup>`.

**AI template editing**
1. Pick the template via `GET /api/v1/template/get`; capture its `text` and `category`.
2. Ask what change they want in plain words.
3. If the user asks for all/every/全部/所有 matching template entries, explain that Template AI can use exact `match: "all"` for `replace`/`delete` operations. Otherwise the default operation contract requires one exact unique match and may fail with `PATCH_AMBIGUOUS_MATCH` when the same text appears more than once.
4. `POST /api/v1/template/ai/edit-sessions/stream` (`--json` with `userPrompt`, `filename`, `category`, `currentText`, plus current template metadata such as `ruleSource`, `useProxy`, `proxyLink`, and `enableIncludeAll` when available). Read the SSE stream until `template.edit.preview.ready` or `template.edit.completed`; model deltas are progress only. Note: this requires the server's AI assistant to be configured in the web UI; if it returns an upstream/credential error, explain that and stop.
5. Summarize the server-materialized preview from the returned `candidateText`, list operations and warnings, and confirm whether the user wants to accept it. Explain that warnings are review metadata, while validation errors block accept.
6. `POST /api/v1/template/ai/edit-sessions/{sessionId}/accept` (`--json` with optional `currentText` set to the editor's current text). Send `currentText` whenever you have the current editor text, especially after a prior accepted preview has not been saved yet. It proves the editor base matches the session base so consecutive unsaved accepts can work. It is not the candidate, is never persisted, and must not be treated as the accepted output. Don't include warning-related request fields. The response returns `candidateText` for the editor but does not persist the template, so save it with the normal template update flow after the user confirms. If `currentText` is absent or mismatched, the server still protects stale bases through the saved template file check and may return `AI_EDIT_STALE_BASE`.
7. If the user rejects the preview, call `POST /api/v1/template/ai/edit-sessions/{sessionId}/discard` when a `sessionId` exists.

**Edit / delete (any resource)**
- Always fetch and show the current state first (GET), confirm the exact target by id+name, summarize the change, then confirm before the write. For deletes and any `batch-*` operation, show the count and affected names/IDs and warn it isn't easily reversible.

**Deploy or install SublinkPro** (the skill can stand up a new instance, not just operate one)

Full command reference: `reference/deploy.md`. Read it before running anything — never invent deploy commands. Lead the user through these steps:

1. **Confirm intent & location.** Are we setting up a *new* SublinkPro instance? Ask **local or remote host**. If remote, ask which mode:
   - *Mode 1 — agent runs it:* the user lets you run commands on the remote box over their existing SSH access (`ssh user@host '...'`). Only do this after they confirm the host and grant permission for each remote command.
   - *Mode 2 — user runs it:* you generate the exact commands / compose file and the user runs them on the host themselves. Default to this if unsure.
2. **Detect the environment.** Check what's available before recommending: `docker --version`, `docker compose version` (or `docker-compose version`). For the install script, note it needs root and is interactive.
3. **Pick a method** — recommend **docker-compose** for most users (easy to update and re-run). Alternatives: plain `docker run`, or the one-line install script (binary + systemd/OpenRC; root; interactive menu — hand it over, don't pipe answers).
4. **Offer configuration as one short, skippable step.** Most users want defaults — lead with "I can use sensible defaults, or set a few options; want to customize anything?". If they want to customize, walk the common ones in plain language (host port, admin password, database, captcha, hidden base path) and map them to the right env vars yourself. The full authoritative variable list and the `config.yaml` alternative are in `reference/deploy.md` ("Configuration: environment variables & config file") — consult it, never invent a variable. Only add the env vars the user actually chose; everything else stays default. For multi-instance/migration setups, remember `SUBLINK_JWT_SECRET` + `SUBLINK_API_ENCRYPTION_KEY` must be set and identical across instances.
5. **Confirm** the exact command(s) or compose file in plain language before running or handing over. For remote Mode 1, show the full `ssh`/`scp` command you're about to run and get a yes.
6. **Run or hand over** per the chosen mode.
7. **Health-check:** `GET /api/v1/version` against the new address (public, no key needed) to confirm it's up. (`curl <base>/api/v1/version` works too, e.g. for remote.)
8. **First-run hand-off (be honest — this part is manual).** You cannot auto-create an API key (login has a captcha by default), so guide the user: open `http://<host>:8000` → sign in `admin` / `123456` → **change the password immediately** → Settings → Access Keys → Create → copy the key → set `export SUBLINK_BASE_URL=...` and `export SUBLINK_API_KEY=...`. Then offer to continue into operating the instance.

**Manage an existing deployment** (status / logs / update / uninstall) — see `reference/deploy.md` "Lifecycle".
- *Status/logs:* `GET /api/v1/version` for up/down; `docker ps`, `docker logs sublinkpro`, or `docker-compose ps` / `docker-compose logs -f` for container detail.
- *Update:* compose → `docker-compose pull && docker-compose up -d`; plain docker → stop/rm/pull/run with the same flags; script install → re-run the install script.
- *Uninstall:* `docker-compose down` or `docker stop/rm`, or `uninstall.sh`. **Removing the `./db ./template ./logs` data directories is destructive** — show exactly what would be deleted and get explicit confirmation before suggesting it.

---

## Error Handling Playbook

**Whenever an API call fails (curl returns an error envelope, or the optional `scripts/sublink.py` exits non-zero), never dump raw output at the user. Translate it into plain language + the concrete next step, and offer to fix it.** Before any call, also make sure `SUBLINK_BASE_URL` and `SUBLINK_API_KEY` are set — if not, guide the user through setting them (or deploying) rather than letting the call fail. Map the failure to one of these classes:

| Condition | What it means | What to do for the user |
|---|---|---|
| `$SUBLINK_BASE_URL` is empty/unset | No instance address configured | Ask for their instance address and set it; if they don't have one running, offer to **deploy** (see workflow above). |
| `$SUBLINK_API_KEY` is empty/unset | No API key configured | Walk them through creating one: web UI → sign in (`admin`/`123456` if fresh) → Settings → Access Keys → Create → copy → set the env var. |
| curl connection error / no response (or the script prints "Could not reach...") | Server unreachable (wrong host/port, not running, firewall) | Run the checklist: is it running? right host/port? remote port exposed? Confirm with `GET /api/v1/version`. Offer to deploy or check container status. |
| `code:403` (or HTTP 403) | Auth rejected | Key missing/wrong/expired, or base URL points at a different instance than the key. Re-create the key in the web UI. (Exception: `/settings/ai-assistant*` 403s by design with an API key — explain that's expected and not fixable via the key.) |
| `code:500` with a `msg` | Server rejected the request | Read the `msg` (it's usually specific, e.g. "订阅名称不能重复"), explain it simply, and propose the correction (pick another name, fill the missing field, etc.). |
| `code:400`/`code:404` | Bad/missing parameter or wrong path | Re-check required fields and the resource id (and the form-vs-JSON tag in `reference/api.md`); re-run discovery (GET) to get valid values, then retry. |
| body isn't `{code,msg,data}` JSON | Hit a non-envelope endpoint (`/c/` output, SSE) | That's expected for those endpoints — read the body as-is; don't try to parse an envelope. |

After explaining, always offer the fix as the next action ("want me to set that now?" / "shall I create the key step by step?" / "want me to deploy an instance?") rather than leaving the user at a dead end.
