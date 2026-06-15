# SublinkPro Deployment Reference

The factual, copy-pasteable command catalog the agent uses when deploying or
managing a SublinkPro instance. **Do not invent commands or flags beyond what is
listed here.** Every command below is verified against the project's
`docs/installation.md`, `README.md`, `docker-compose.example.yml`, and
`install.sh`.

This is the *deploy / install* half of the skill. The *operate* half (managing
nodes, subscriptions, etc. via the REST API) is documented in `api.md` and the
main `SKILL.md`.

---

## Defaults & key facts

- **Image:** `zerodeng/sublink-pro` (stable) or `zerodeng/sublink-pro:dev` (dev/preview builds).
- **Port:** `8000` (web UI + API).
- **Default login:** `admin` / `123456` — tell the user to change it immediately.
- **Data directories** (persist across upgrades; removing them is destructive):
  - `./db` — database, config files, GeoIP, default SQLite DB
  - `./template` — template files
  - `./logs` — runtime logs
- **Health check:** `GET /api/v1/version` is **public** (no API key needed). Use it to confirm an instance is up.
- **The Docker image bundles `cloudflared`** (for Cloudflare Tunnel). Non-Docker installs need it installed separately.

---

## Prerequisites & detection

Before choosing a method, check what the target host has. Run these on whichever
host will run SublinkPro (local shell, or via `ssh user@host '...'` for remote):

```bash
docker --version              # is Docker installed?
docker compose version        # v2 compose plugin (preferred)
docker-compose version        # v1 standalone compose (older hosts)
uname -m                      # arch: x86_64 -> amd64, aarch64 -> arm64, armv7l -> armv7, i386/i686 -> x86
id -u                         # 0 means root (the install script requires root)
```

Method choice guidance:
- **docker-compose** — recommended default for most users (declarative, easy upgrades).
- **docker run** — fine for a quick single-container start.
- **install script** — for non-Docker hosts; it installs a native binary + system service, but is **interactive and root-only** (see Method C).

Note: compose v2 is invoked as `docker compose ...` (space); v1 as `docker-compose ...` (hyphen). Detect which exists and use that form consistently.

---

## Method A — docker run (local)

Stable:

```bash
docker run --name sublinkpro -p 8000:8000 \
  -v $PWD/db:/app/db \
  -v $PWD/template:/app/template \
  -v $PWD/logs:/app/logs \
  -d zerodeng/sublink-pro
```

Dev/preview build: replace the image with `zerodeng/sublink-pro:dev`.

---

## Method B — docker-compose (local, recommended)

**Minimal `docker-compose.yml`** (from the README quick start):

```yaml
services:
  sublinkpro:
    image: zerodeng/sublink-pro
    container_name: sublinkpro
    ports:
      - "8000:8000"
    volumes:
      - "./db:/app/db"
      - "./template:/app/template"
      - "./logs:/app/logs"
    restart: unless-stopped
```

Start:

```bash
docker-compose up -d      # or: docker compose up -d
```

Then open `http://localhost:8000` and sign in with `admin` / `123456`.

**Annotated "with options" version** — only add the env vars the user actually
wants; every one is optional and has a sensible default. Field meanings are from
`docker-compose.example.yml`:

```yaml
services:
  sublinkpro:
    image: zerodeng/sublink-pro
    container_name: sublinkpro
    ports:
      - "8000:8000"                  # change left side to remap host port, e.g. "9000:8000"
    volumes:
      - "./db:/app/db"
      - "./template:/app/template"
      - "./logs:/app/logs"
    environment:
      - SUBLINK_PORT=8000            # in-container port
      # - SUBLINK_DSN=mysql://user:pass@tcp(mysql:3306)/sublink?charset=utf8mb4&parseTime=True&loc=Local
      # - SUBLINK_DSN=postgres://user:pass@postgres:5432/sublink?sslmode=disable
      #   (leave unset to use the default SQLite DB at /app/db/sublink.db)
      - SUBLINK_LOG_LEVEL=info       # debug | info | warn | error | fatal
      # - SUBLINK_ADMIN_PASSWORD=your-admin-password   # initial admin password (first start only)
      - SUBLINK_CAPTCHA_MODE=2       # 1=off | 2=traditional (default) | 3=Cloudflare Turnstile
      # - SUBLINK_WEB_BASE_PATH=/admin   # hide the admin UI behind a path; does NOT affect /api/* or /c/*
      # - SUBLINK_JWT_SECRET=at-least-32-chars          # auto-generated if unset
      # - SUBLINK_API_ENCRYPTION_KEY=at-least-32-chars  # auto-generated if unset
    restart: unless-stopped
```

> Multi-instance / migration note: when running more than one instance against
> the same data (or migrating), `SUBLINK_JWT_SECRET` and
> `SUBLINK_API_ENCRYPTION_KEY` must be identical across instances, or existing
> tokens and API keys stop working.

**Optional Sub-Store sidecar** (from `docs/installation.md`) — add as a second
service; keep it on the internal compose network and do **not** publish its port:

```yaml
  substore:
    image: xream/sub-store
    container_name: substore
    environment:
      - SUB_STORE_BACKEND_API_PORT=3000
      - SUB_STORE_BODY_JSON_LIMIT=10mb
    restart: unless-stopped
```

Sub-Store is then enabled in-app under **User Center -> Sub-Store** (base URL
e.g. `http://substore:3000`), not via env vars.

---

## Method C — one-line install script (local, root, INTERACTIVE)

```bash
sh -c "$(wget -qO- https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/install.sh)"
```

**Important:** this script is interactive (`read -r` menus) and must run as
**root**. The agent must **not** try to pipe answers into it or run it headless.
Instead, hand the user the exact command to paste into their own terminal, and
explain the menu they'll see:

- **Fresh install** — runs automatically on first install (binary + system service, default admin/123456, port 8000).
- **Update** — detected when already installed; updates the program, keeps all data.
- **Reinstall** — lets the user choose whether to keep or wipe data.
- **Restore install** — detected when old data dirs exist; offers to restore them.

The script installs to `/usr/local/bin/sublink` and registers a `systemd`
(or OpenRC on Alpine) service named `sublink`.

---

## Configuration (env vars & config file)

**Everything below is optional.** SublinkPro runs with sensible defaults out of
the box — a fresh deploy needs *zero* configuration. Offer these only as a
short, skippable step ("want to customize anything, or use defaults?"). Set just
the ones the user actually asks for.

### How config is resolved (priority, high → low)

1. Command-line flags (`--port`, `--db`, etc.)
2. Environment variables (`SUBLINK_*`) — the usual way for docker/compose
3. Config file `db/config.yaml`
4. Database-stored settings (sensitive values)
5. Built-in defaults

For docker/compose, **environment variables are the right tool.** The config
file matters mainly for native installs or advanced setups; if the user wants
it, copy `config.example.yaml` to `db/config.yaml` (the `./db` volume) and edit.

### Environment variable reference

Verified against `config.example.yaml` and the Go source. Complete annotated
file: `config.example.yaml` in the repo root — point power users there.

| Env var | Default | Meaning |
|---|---|---|
| `SUBLINK_PORT` | `8000` | Web UI + API port (in-container). |
| `SUBLINK_DSN` | _(SQLite)_ | DB connection. Empty = SQLite at `db/sublink.db`. MySQL/Postgres examples below. |
| `SUBLINK_DB_PATH` | `./db` | Local data dir (config, SQLite, GeoIP). |
| `SUBLINK_LOG_PATH` | `./logs` | Log directory. |
| `SUBLINK_LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` / `fatal`. |
| `SUBLINK_EXPIRE_DAYS` | `14` | Login token validity (days). |
| `SUBLINK_ADMIN_PASSWORD` | `123456` | Initial admin password — **first start only**. |
| `SUBLINK_ADMIN_PASSWORD_REST` | _(unset)_ | Resets admin password on **every** start while set (then should be removed). |
| `SUBLINK_CAPTCHA_MODE` | `2` | `1`=off (intranet only) / `2`=image captcha / `3`=Cloudflare Turnstile. |
| `SUBLINK_TURNSTILE_SITE_KEY` | _(unset)_ | Turnstile site key (mode 3). |
| `SUBLINK_TURNSTILE_SECRET_KEY` | _(unset)_ | Turnstile secret key (mode 3). |
| `SUBLINK_TURNSTILE_PROXY_LINK` | _(unset)_ | Proxy (mihomo link) for Turnstile verify when the server can't reach Cloudflare. |
| `SUBLINK_WEB_BASE_PATH` | _(unset)_ | Hide admin UI behind a path (e.g. `/admin`). Does **not** affect `/api/*` or `/c/*`. |
| `SUBLINK_GEOIP_PATH` | `db/GeoLite2-City.mmdb` | GeoIP DB path; auto-downloaded if missing. |
| `SUBLINK_TRUSTED_PROXIES` | local/private CIDRs | Comma-separated trusted reverse-proxy IPs/CIDRs for real-client-IP. |
| `SUBLINK_LOGIN_FAIL_COUNT` | `5` | Failed logins before IP ban. |
| `SUBLINK_LOGIN_FAIL_WINDOW` | `1` | Failure-count window (minutes). |
| `SUBLINK_LOGIN_BAN_DURATION` | `10` | Ban duration (minutes). |
| `SUBLINK_JWT_SECRET` | _(auto-generated)_ | Login-token signing key. Set (≥32 chars) for multi-instance/migration. |
| `SUBLINK_API_ENCRYPTION_KEY` | _(auto-generated)_ | API-Key encryption key. **Must match** across instances or existing API keys break. |
| `SUBLINK_MFA_RESET_SECRET` | _(auto-generated)_ | Secret for the MFA emergency-reset flow. |

DSN examples:

```text
mysql://user:pass@tcp(mysql:3306)/sublink?charset=utf8mb4&parseTime=True&loc=Local
postgres://user:pass@postgres:5432/sublink?sslmode=disable
```

> **Sensitive keys** (`SUBLINK_JWT_SECRET`, `SUBLINK_API_ENCRYPTION_KEY`,
> `SUBLINK_MFA_RESET_SECRET`) auto-generate and persist to the DB if unset — fine
> for a single instance. For **multi-instance or migration**, set them explicitly
> and identically everywhere, or tokens/API keys stop working.

### How the agent should handle config

- **Default-first:** never force the user through every variable. Ask once if
  they want to customize; if not, deploy with defaults.
- When they do want changes, map their plain-language goal to the right
  variable(s) (e.g. "use port 9000" → `SUBLINK_PORT=9000` + remap the published
  port; "use my MySQL" → build the `SUBLINK_DSN`; "hide the panel" →
  `SUBLINK_WEB_BASE_PATH`).
- Add the chosen vars to the compose `environment:` list (or `-e` flags for
  `docker run`), show the final file/command, and confirm before applying.

---

## Remote deployment

Ask the user which mode they want **each time** (don't assume):

### Mode 1 — the agent runs it over SSH

Only after the user confirms the target host and that the agent may run remote
commands. Prefix docker/compose commands with their existing SSH access:

```bash
ssh user@host 'docker run --name sublinkpro -p 8000:8000 \
  -v $PWD/db:/app/db -v $PWD/template:/app/template -v $PWD/logs:/app/logs \
  -d zerodeng/sublink-pro'
```

For compose, write the file remotely then bring it up:

```bash
ssh user@host 'mkdir -p ~/sublinkpro'
scp docker-compose.yml user@host:~/sublinkpro/docker-compose.yml
ssh user@host 'cd ~/sublinkpro && docker-compose up -d'
```

(Or create the file remotely with a heredoc over ssh.) Then health-check
against the remote address: `GET http://host:8000/api/v1/version`.

### Mode 2 — the user runs it themselves

Generate the exact command(s) / the full `docker-compose.yml` and tell the user
to copy them to the remote host and run them there. Remind them the firewall /
security group must allow the chosen port.

---

## Lifecycle management

### Status / health / logs

```bash
# Is it up? (public endpoint, no key)
curl http://<host>:8000/api/v1/version

docker ps                        # container running?
docker logs sublinkpro           # container logs (add -f to follow)
docker logs --tail 100 sublinkpro
# compose equivalents:
docker-compose ps
docker-compose logs -f
```

For native (install-script) deployments: `systemctl status sublink`,
`journalctl -u sublink` (or `rc-service sublink status` on Alpine).

### Update / upgrade

docker-compose:

```bash
cd /path/to/your/sublinkpro
docker-compose pull
docker-compose up -d
docker image prune -f        # optional: clean old images
```

docker run:

```bash
docker stop sublinkpro && docker rm sublinkpro
docker pull zerodeng/sublink-pro
# re-run the SAME docker run command used at install (same -v mounts!)
docker run --name sublinkpro -p 8000:8000 \
  -v $PWD/db:/app/db -v $PWD/template:/app/template -v $PWD/logs:/app/logs \
  -d zerodeng/sublink-pro
docker image prune -f        # optional
```

install script: re-run the one-line install command; it detects the existing
install and offers Update (keeps data).

### Uninstall

Confirm with the user first, and **always warn that deleting the data
directories (`./db ./template ./logs`) is irreversible.** Removing the container
alone keeps the data.

```bash
# docker run install:
docker stop sublinkpro && docker rm sublinkpro
docker rmi zerodeng/sublink-pro            # optional: remove image
# rm -rf ./db ./template ./logs            # DESTRUCTIVE — only if user confirms

# docker-compose install:
cd /path/to/your/sublinkpro
docker-compose down                        # add -v only if they want named volumes gone
# rm -rf ./db ./template ./logs            # DESTRUCTIVE — only if user confirms
```

Native (install-script) install — hand the user the interactive uninstall
script (it asks whether to keep data):

```bash
sh -c "$(wget -qO- https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/uninstall.sh)"
```

---

## First-run hand-off (do this after every fresh deploy)

After the health check passes, the agent **cannot** finish setup headlessly —
creating an API key requires a logged-in web session, and login has a captcha by
default (`SUBLINK_CAPTCHA_MODE=2`). So guide the user through these final steps:

1. Open `http://<host>:8000` in a browser.
2. Sign in with `admin` / `123456`.
3. **Change the admin password immediately.**
4. Go to **Settings -> Access Keys -> Create**; copy the key (shown once).
5. Set the env vars so the operate half of the skill works:
   ```bash
   export SUBLINK_BASE_URL=http://<host>:8000
   export SUBLINK_API_KEY=prefix_xxx_yyy
   ```
6. Confirm with `curl -s "$SUBLINK_BASE_URL/api/v1/version"` (or, if Python 3 is
   present, `python scripts/sublink.py GET /api/v1/version`), then offer to continue
   into operating the instance (add nodes, build subscriptions, etc.).
