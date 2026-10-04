English | [简体中文](socks5.zh-CN.md)

# SOCKS5 Gateway

SublinkPro can expose selected stored proxy nodes through a local SOCKS5 gateway. This is separate from the existing SOCKS5 node import/export support.

## What is included

- TCP `CONNECT` only
- IPv4, IPv6, and domain targets
- Optional username/password authentication (enabled by default)
- Independent SOCKS5 accounts with separate encrypted passwords and direct routing-profile binding
- Multiple listener addresses/ports with stable IDs, per-listener default profiles, and optional authentication-policy overrides
- Best-node, random-node, round-robin, smart P2C, or specific-node selection
- Candidate node pools filtered by one or more groups, sources, protocols, and countries/regions; fields combine with AND and values inside one field combine with OR
- Optional in-memory sticky sessions keyed by client IP or authenticated SOCKS5 username, with a 60-604800 second sliding TTL
- Multiple routing profiles with independent candidate pools, selection strategies, retries, cooldown, and sticky-session settings
- Backward-compatible username routing: `username`, `username@profile`, or `username@profile.account`
- Mihomo outbound adapter pooling keyed by node ID and link hash, with an idle LRU cap
- Configurable per-request retry (up to 5 candidate nodes), dial timeout, and failed-node cooldown
- Optional fallback from a specific node; disabled by default to keep specific routing strict
- Per-listener active-connection limit (`maxConnections`, default `256`, range `1-10000`)
- Per-client-IP active-connection limit within each listener (`maxConnectionsPerClient`, default `32`, range `1..maxConnections`)
- Idle timeout (`idleTimeoutSeconds`, default `600`, range `0-86400`; `0` disables it)
- Maximum connection duration (`maxConnectionDurationSeconds`, default `0`, range `0-604800`; `0` disables it)
- Administrator-only live monitoring with active/total/success/failure counters and upload/download byte counters
- Administrator controls to disconnect one active connection or all active connections
- Optional active node health checks with a fixed HTTPS probe target, bounded concurrency, latency reporting, exponential failure cooldown, and manual probe control
- Server-paginated node load visualization with search, health filtering, sorting, active connections, success/failure totals, latency sources, and live smart scores
- Runtime-stat reset controls that preserve active connections while clearing per-node historical counters
- Health-aware routing: active-probe failures are removed from normal candidate sets, and `best` prioritizes fresh active-probe latency before applying the retry limit
- Runtime start/stop when settings are saved; no process restart is required

The existing gateway settings are exposed as the non-deletable `default` routing profile, and the legacy listen address/port become the compatible `default` listener until a listener list is saved, so upgrades require no data migration. Additional profiles are selected through the authenticated username: the original username continues to use `default`; `username@japan` selects the `japan` profile and uses client-IP affinity; `username@japan.user01` selects the same profile and uses `user01` as the affinity account. All legacy username forms use the existing gateway password. Independent accounts use their own password and are bound directly to one enabled routing profile; exact independent-account usernames are checked before the legacy syntax. Disabled or missing profiles fail authentication, and profile routing requires username/password authentication. Round-robin counters and sticky leases are isolated by profile inside each listener runtime. Connection IDs include the listener ID and gateway counters are aggregated across listeners.

Smart routing uses power-of-two choices for the first attempt, combining active-probe latency (or stored delay), active connection load, and consecutive failure penalties; remaining retry candidates are score-ordered. Sticky hits remain first and smart scoring applies to their fallbacks. Retries happen before the SOCKS5 success reply is sent. Active-probe failures are skipped while another routable candidate exists; if every candidate is unhealthy, routing fails open and keeps candidates available for recovery. A successful real connection clears the routing exclusion without erasing the most recent active-probe latency. A failed adapter is discarded, and adapters are closed when the gateway is stopped or reapplied. If every candidate is cooling down, the node whose cooldown expires first is probed so the pool cannot remain permanently unavailable.

UDP `ASSOCIATE` and `BIND` are not included yet. Sticky leases and per-node routing statistics are in-memory and reset whenever the gateway is stopped or settings are reapplied. The administrator monitor includes live connection controls plus a paginated candidate-node load table; resetting node statistics does not interrupt active connections.

## Configure

1. Sign in as an administrator.
2. Open **System Settings → SOCKS5 gateway** from the sidebar.
3. Keep **Listen address** as `127.0.0.1` for local-only access.
4. Choose a port (default `1080`) and node selection strategy.
5. Optionally restrict the candidate node pool by groups, sources, protocols, and countries/regions. Empty fields do not restrict the pool. Specific-node mode uses these filters only for fallback candidates.
6. Configure maximum attempts, per-node dial timeout, and failed-node cooldown.
7. Configure per-listener and per-client connection limits.
8. Configure idle timeout and maximum connection duration; set either value to `0` to disable that limit.
9. Optionally enable sticky sessions and choose client IP or authenticated SOCKS5 username as the affinity key. Configure a 60-604800 second TTL; each successful connection refreshes it.
10. For specific-node routing, enable fallback only if switching to another node is acceptable.
11. Keep authentication enabled and set a username/password.
12. Optionally enable active node health checks, then choose a 10-3600 second interval and 1-30 second per-node timeout.
13. Save the settings. Monitoring, health probing, and connection termination controls are administrator-only.
14. Optionally create routing profiles below the gateway settings. Use lowercase profile IDs such as `japan`, then connect with `username@japan` or `username@japan.account`.
15. Add independent accounts when clients need separate passwords; bind each account directly to an enabled routing profile.
16. Add or edit listeners to expose additional addresses/ports. Each listener can choose a default profile and inherit, require, or disable authentication. Disabling authentication is allowed only on loopback addresses.

The gateway is disabled by default. Passwords are encrypted with the instance API encryption key. Settings responses expose `hasPassword` and, when present, `maskedPassword`; plaintext `password` is never returned. Omitting `password` preserves the saved password, while `clearPassword: true` clears it.

## Use

```bash
curl --proxy socks5h://127.0.0.1:1080 \
  -U "admin:your-password" \
  https://api.ipify.org
```

If you bind to `0.0.0.0` or another non-loopback address, authentication is mandatory. Expose the port only behind a firewall or private network and use a strong password. The backend rejects unauthenticated non-loopback listeners to prevent accidental open proxies.

## Backup and restore

Database migration and WebDAV restore preserve the target instance's SOCKS5 configuration, accounts, listeners, and routing profiles. This avoids importing credentials encrypted with another instance's key or unexpectedly opening network listeners. Active connections are disconnected and gateway state is reloaded after restore; review node IDs and candidate pools against the restored nodes.

Connection limits apply separately to each listener. Connection totals aggregate all listeners; the health/routing views currently show the first enabled listener's node runtime. SOCKS5 itself does not add TLS, so non-loopback listeners should remain behind a firewall or trusted private network.

## API

All endpoints below require an authenticated administrator. The destructive `POST`/`DELETE` operations are also restricted in demo mode.

- `GET /api/v1/settings/socks5` — read public gateway settings; the plaintext password is never returned.
- `POST /api/v1/settings/socks5` — save and apply settings. JSON fields include `enabled`, `listenAddress`, `port`, `username`, optional `password`, `clearPassword`, `nodeId`, `selection` (`best`, `random`, `round_robin`, `smart`, or `specific`), `requireAuth`, `maxAttempts` (1-5), `dialTimeoutSeconds` (1-120), `failureCooldownSeconds` (0-3600), `specificFallback`, `maxConnections` (1-10000), `maxConnectionsPerClient` (1..maxConnections), `idleTimeoutSeconds` (0-86400), `maxConnectionDurationSeconds` (0-604800), `healthCheckEnabled`, `healthCheckIntervalSeconds` (10-3600), `healthCheckTimeoutSeconds` (1-30), `candidateGroups`, `candidateSources`, `candidateProtocols`, `candidateCountries`, `stickySessionEnabled`, `stickySessionMode` (`client_ip` or `username`), and `stickySessionTtlSeconds` (60-604800). Candidate pool fields are string arrays; fields are combined with AND and values within each field use OR. Optional fields may be omitted to preserve saved values for legacy clients.
- `POST /api/v1/settings/socks5/stop` — stop the listener without changing saved settings.
- `GET /api/v1/settings/socks5/status` — return `data.config`, `data.listeners`, `data.stats`, and `data.health` with listener state, aggregate connection totals, sweep state, and health details.
- `GET|POST /api/v1/settings/socks5/listeners` — list or create listeners. Listener fields are `id`, `enabled`, `listenAddress`, `port`, `defaultProfileId`, and nullable `requireAuth` (null inherits the gateway setting).
- `PUT|DELETE /api/v1/settings/socks5/listeners/:id` — replace or delete a listener. Listener IDs and enabled endpoints must be unique. A failed bind rolls the saved listener list back.
- `GET|POST /api/v1/settings/socks5/accounts` — list or create independent accounts. Account fields are `id`, `username`, `password`, `profileId`, and `enabled`; responses expose only password-presence metadata.
- `PUT|DELETE /api/v1/settings/socks5/accounts/:id` — replace or delete an account. An empty update password retains the encrypted saved password.
- `GET /api/v1/settings/socks5/profiles` — list the virtual `default` profile plus saved custom routing profiles.
- `POST /api/v1/settings/socks5/profiles` — create a custom profile. Fields include `id`, `name`, `enabled`, `selection`, `nodeId`, retry/cooldown values, candidate-pool arrays, and sticky-session settings.
- `PUT /api/v1/settings/socks5/profiles/:id` — replace an existing custom profile. The reserved `default` profile is managed through the gateway settings endpoint.
- `DELETE /api/v1/settings/socks5/profiles/:id` — delete a custom profile; existing established connections are not interrupted.
- `GET /api/v1/settings/socks5/routing` — return a server-paginated candidate-node load page. Query parameters are `profileId`, `keyword`, `status` (`healthy`, `unhealthy`, `checking`, `cooling`, or `unknown`), `sortBy`, `sortOrder`, `page`, and `pageSize` (maximum `100`). Items include node metadata, health state, latency source, active/success/failure counters, consecutive failures, smart score, cooldown, and last selection time.
- `POST /api/v1/settings/socks5/routing/reset` — clear per-node success/failure/last-selection counters without disconnecting active sessions.
- `POST /api/v1/settings/socks5/health/probe` — start an immediate health sweep. An already-running sweep is reused instead of starting a duplicate.
- `GET /api/v1/settings/socks5/connections` — return the current active connection array. Each item includes `id`, `clientAddress`, `target`, `nodeName`, `profileId`, `profileName`, optional `account`, optional `listenerId`, `phase`, `startedAt`, `lastActivity`, `uploadBytes`, and `downloadBytes`.
- `DELETE /api/v1/settings/socks5/connections/:id` — disconnect one active connection by ID.
- `DELETE /api/v1/settings/socks5/connections` — disconnect all active connections.

The status counters, per-node routing statistics, active connection list, and sticky leases belong to the current gateway server lifetime; stopping or reapplying settings resets them. The routing endpoint performs filtering, sorting, and pagination on the server so three-second UI refreshes render only the requested page.

Active health checks probe the deduplicated union of candidate nodes from every enabled routing profile. Strict specific-node profiles contribute only their configured node; profiles with specific-node fallback contribute that node plus their configured fallback pool. Other profiles contribute their configured candidate pool. Probes use a fixed Cloudflare HTTPS connectivity endpoint; administrators cannot configure an arbitrary probe URL. Sweeps use at most four workers and never overlap. Status responses include aggregate counts and at most the first 200 sorted node records to keep three-second monitoring polls bounded. Failed probes retire the failed adapter lease safely and apply exponential cooldown capped at one hour.
