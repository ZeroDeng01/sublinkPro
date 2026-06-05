English | [简体中文](configuration.zh-CN.md)

# Configuration

This document describes SublinkPro configuration methods and parameters.

---

## Configuration Priority

SublinkPro supports several configuration methods. Priority from highest to lowest:

1. **Command line flags**, useful for temporary overrides such as `--port 9000`
2. **Environment variables**, recommended for Docker deployments
3. **Configuration file**, `db/config.yaml`
4. **Database stored settings**, used for sensitive configuration
5. **Default values**, built into the program

---

## Environment Variables

| Environment variable | Description | Default |
|----------|---------------------------------|-------------------------------------|
| `SUBLINK_PORT` | Service port | 8000 |
| `SUBLINK_DSN` | Database DSN, supports sqlite/mysql/postgres | SQLite by default: `sqlite://./db/sublink.db` |
| `SUBLINK_DB_PATH` | Local data directory and default SQLite database directory | ./db |
| `SUBLINK_LOG_PATH` | Log directory | ./logs |
| `SUBLINK_JWT_SECRET` | JWT signing secret | Generated automatically |
| `SUBLINK_API_ENCRYPTION_KEY` | API encryption key | Generated automatically |
| `SUBLINK_EXPIRE_DAYS` | Token expiration days | 14 |
| `SUBLINK_LOGIN_FAIL_COUNT` | Login failure limit | 5 |
| `SUBLINK_LOGIN_FAIL_WINDOW` | Login failure window, in minutes | 1 |
| `SUBLINK_LOGIN_BAN_DURATION` | Login ban duration, in minutes | 10 |
| `SUBLINK_GEOIP_PATH` | GeoIP database path | ./db/GeoLite2-City.mmdb |
| `SUBLINK_CAPTCHA_MODE` | CAPTCHA mode, 1=off, 2=image, 3=Turnstile | 2 |
| `SUBLINK_TURNSTILE_SITE_KEY` | Cloudflare Turnstile Site Key | - |
| `SUBLINK_TURNSTILE_SECRET_KEY` | Cloudflare Turnstile Secret Key | - |
| `SUBLINK_TURNSTILE_PROXY_LINK` | Proxy link for Turnstile verification, mihomo format | - |
| `SUBLINK_TRUSTED_PROXIES` | Trusted reverse proxy IP/CIDR list, comma separated | `127.0.0.1,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,100.64.0.0/10` |
| `SUBLINK_WEB_BASE_PATH` | Frontend base path for hiding the site entry | - |
| `SUBLINK_ADMIN_PASSWORD` | Initial admin password | 123456 |
| `SUBLINK_ADMIN_PASSWORD_REST` | Reset admin password | Enter the new admin password |
| `SUBLINK_MFA_RESET_SECRET` | Secret used to generate restricted TOTP emergency reset tokens, environment variable only | - |
| `SUBLINK_DEMO_MODE` | Enable demo mode, memory database and some sensitive operations disabled | false |
| `SUBLINK_FEATURE` | Experimental feature flags | Reserved experimental flags, comma separated. Node preview is now stable and no longer needs configuration |

---

## Cloudflare Tunnel

The **Cloudflare Tunnel** tab in User Center can host the local `cloudflared` process and connect the current SublinkPro instance to a remotely managed Tunnel in Cloudflare Zero Trust.

- The Docker image includes `cloudflared`, so you usually only need to enter the Tunnel token on the page and start it.
- Non Docker deployments need `cloudflared` installed first, with the `cloudflared` command available in `PATH`.
- The page never echoes the raw token. Status APIs only return a masked token.

At runtime this is equivalent to `cloudflared tunnel --no-autoupdate run`. The token is passed through the `TUNNEL_TOKEN` environment variable so it does not appear in process arguments.

See the full guide at [Cloudflare Tunnel remote access](features/cloudflare-tunnel.md).

---

## Command Line Flags

```bash
# Show help
./sublinkpro help

# Start with a specific port
./sublinkpro run --port 9000

# Use a specific SQLite database
./sublinkpro run --dsn "sqlite:///data/sublink.db"

# Use MySQL
./sublinkpro run --dsn "mysql://user:pass@tcp(127.0.0.1:3306)/sublink?charset=utf8mb4&parseTime=True&loc=Local"

# Use PostgreSQL
./sublinkpro run --dsn "postgres://user:pass@127.0.0.1:5432/sublink?sslmode=disable"

# Set the local data directory, used for config file, GeoIP, and default SQLite
./sublinkpro run --db /data

# Reset admin password
./sublinkpro setting -username admin -password newpass
```

---

## Database DSN

SublinkPro now supports unified database connection configuration through `dsn`, with these dialects:

- `sqlite://`
- `mysql://`
- `postgres://`
- `postgresql://`

If `dsn` is empty, the system falls back to SQLite and uses `db_path/sublink.db` as the database file.

### SQLite example

```yaml
dsn: sqlite:///app/db/sublink.db
```

### MySQL example

```yaml
dsn: mysql://user:pass@tcp(mysql:3306)/sublink?charset=utf8mb4&parseTime=True&loc=Local
```

### PostgreSQL example

```yaml
dsn: postgres://user:pass@postgres:5432/sublink?sslmode=disable
```

> [!TIP]
> When using MySQL or PostgreSQL, `db_path` is still used for local config files and GeoIP database storage. It no longer decides the actual database backend.

## Migrate from SQLite to MySQL / PostgreSQL

If an old instance has always used SQLite and you want to migrate to MySQL or PostgreSQL, use the built in “Data Migration” feature.

### Before migration

1. Prepare a new empty MySQL or PostgreSQL database.
2. Configure database `DSN` for the new instance.
3. Confirm that the old instance can sign in normally.
4. If you need to keep old `AccessKey` values, confirm that `SUBLINK_API_ENCRYPTION_KEY` is the same on both instances.

### Step 1: Configure DSN in the new instance

You can configure the database for the new instance in any of these ways:

- Environment variable: `SUBLINK_DSN`
- Config file: `dsn:` in `db/config.yaml`
- Command line flag: `./sublinkpro run --dsn "..."`

Example:

```yaml
# MySQL
dsn: mysql://user:pass@tcp(mysql:3306)/sublink?charset=utf8mb4&parseTime=True&loc=Local

# PostgreSQL
dsn: postgres://user:pass@postgres:5432/sublink?sslmode=disable
```

> [!IMPORTANT]
> A fresh empty target database is recommended. Don't import directly into a database that already has business data.

### Step 2: Export a backup from the old SQLite instance

After signing in to the old instance:

1. Click the avatar menu in the upper right.
2. Choose **System Backup**.
3. Download the generated `backup.zip`.

Using `backup.zip` is recommended because it includes:

- The SQLite database file from the `db` directory
- Template files from the `template` directory

> [!TIP]
> You can also upload a `.db`, `.sqlite`, or `.sqlite3` file directly, but that only migrates database records and won't restore the template directory.

### Step 3: Run migration in the new instance

After starting the new instance, open:

`Settings -> Data Migration`

Then:

1. Upload the `backup.zip` exported from the old instance.
2. Choose whether to migrate `AccessKey`.
3. Choose whether to migrate subscription access logs.
4. Check “I confirm that this import will overwrite business data in the current instance”.
5. Click **Start Migration**.

The migration task runs in the background. You can view progress and results in:

- The task progress panel in the lower right
- `Task Center`

### After migration

1. Check whether the migration result is successful.
2. If it reports “N warnings”, open the corresponding “Database Migration” task in `Task Center` to view details.
3. **Manually restart the project instance**.
4. Sign in again and check that important data is normal.

### Migration notes

- This import overwrites business data in the current instance.
- It is recommended only for first time migration into a newly deployed MySQL / PostgreSQL instance.
- Subscription access logs are usually large, so migrating them is not recommended by default.
- If old `AccessKey` values cannot be used after migration, check whether `SUBLINK_API_ENCRYPTION_KEY` matches the old instance.
- If login state behaves oddly after migration, sign in again.

---

## Sensitive Configuration

> [!TIP]
> **JWT Secret** and **API encryption key** are sensitive settings. The system handles them in this order:
> 1. Read from environment variables first.
> 2. If not set in environment variables, read from the database.
> 3. If missing from the database too, generate random keys automatically and store them in the database.
>
> **Special note**: If you set these values through environment variables, the system automatically syncs them to the database. That lets the system recover them from the database later even if you forget to set the environment variables, which helps migration and deployment.

> [!WARNING]
> If you need **multi instance deployment** or **cluster deployment**, set the same `SUBLINK_JWT_SECRET` and `SUBLINK_API_ENCRYPTION_KEY` through environment variables for all instances. This keeps login state and API Keys consistent across instances.

## TOTP / MFA Security Notes

SublinkPro supports TOTP based multi factor authentication. When enabled, login becomes:

1. Username + password + CAPTCHA
2. Authenticator code or one time recovery code

### Enablement and usage recommendations

- Start setup in `Settings -> Personal Settings -> Multi Factor Authentication (TOTP)`.
- After scanning the QR code, enter the current 6 digit code once to enable it.
- The system generates a set of **one time recovery codes**. Save them offline, separate from account passwords.
- If the current account has TOTP enabled, changing password, changing username or nickname, disabling TOTP, or resetting recovery codes requires the current dynamic code again.

### Recovery code policy

- Recovery codes can be used for login only **after TOTP is fully enabled**.
- Each recovery code can be used once.
- Old recovery codes become invalid immediately after recovery codes are regenerated.

### Emergency reset, break glass, policy

`SUBLINK_MFA_RESET_SECRET` is used to generate **restricted emergency TOTP reset tokens**. It is for operators helping users who lost their authenticator and cannot use recovery codes.

This setting has these constraints:

- **Environment variable only**, it is not written to config files
- It does not provide a global universal login bypass
- It can only clear TOTP for an account after username + password have been verified
- It is recommended only as a temporary operations setting, with careful rotation

### Recommended operations flow

1. Temporarily set `SUBLINK_MFA_RESET_SECRET`.
2. Generate a reset token with an expiration time for the target user.
3. Call `/api/v1/auth/mfa/reset` with:
   - `username`
   - `password`
   - `resetToken`
4. The user signs in again and binds TOTP again.

> [!WARNING]
> Don't keep `SUBLINK_MFA_RESET_SECRET` as a permanent public setting, and don't treat it as a backdoor for bypassing MFA login. It is only for **restricted TOTP reset when the account password is known**.

---

## CAPTCHA Configuration

SublinkPro supports three CAPTCHA modes through `SUBLINK_CAPTCHA_MODE`:

| Mode | Description |
|:---:|:---|
| **1** | Disable CAPTCHA, not recommended, only for internal networks |
| **2** | Traditional image CAPTCHA, default |
| **3** | Cloudflare Turnstile, recommended and more secure |

### Cloudflare Turnstile configuration

To use Turnstile:

1. Open the [Cloudflare Turnstile console](https://dash.cloudflare.com/?to=/:account/turnstile) and create a site.
2. Get the **Site Key** and **Secret Key**.
3. Configure environment variables:

```yaml
environment:
  - SUBLINK_CAPTCHA_MODE=3
  - SUBLINK_TURNSTILE_SITE_KEY=your-site-key
  - SUBLINK_TURNSTILE_SECRET_KEY=your-secret-key
```

> [!NOTE]
> **Fallback behavior**: If Turnstile mode is configured but complete keys are missing, the system automatically falls back to traditional image CAPTCHA.

### Turnstile proxy configuration

If your server cannot access the Cloudflare API directly, you may see a `context deadline exceeded` timeout. Configure a proxy in that case:

```yaml
environment:
  - SUBLINK_TURNSTILE_PROXY_LINK=vless://your-proxy-link...
```

> [!TIP]
> **Proxy link format**: Use proxy links supported by mihomo, such as `vless://`, `vmess://`, `ss://`, and others. This is similar to Telegram proxy configuration.

### Turnstile verification modes

Cloudflare Turnstile supports three verification modes. Choose one when creating the Site Key in the Cloudflare console:

| Mode | Description |
|:---:|:---|
| **Managed** | Cloudflare decides whether interaction is needed. Most users pass without noticing. |
| **Non-Interactive** | Shows a loading indicator but requires no user interaction. |
| **Invisible** | Fully invisible, verification completes silently in the background. |

The frontend widget renders automatically based on the mode associated with the Site Key. No extra configuration is needed.

---

## Site Hiding Configuration

Set `SUBLINK_WEB_BASE_PATH` to hide the admin site entry, similar to custom path features in 3x-ui.

```yaml
environment:
  - SUBLINK_WEB_BASE_PATH=/admin
```

After setting it:

- `http://domain/` returns 404
- `http://domain/admin` opens the admin UI
- API paths (`/api/*`) and subscription fetch paths (`/c/*`) are **not affected**

> [!TIP]
> Paths work with or without a leading slash. `admin` and `/admin` have the same effect.

---

## Reverse Proxy and Real IP

If you access SublinkPro through Nginx, Caddy, BaoTa, Docker reverse proxy, Cloudflare Tunnel, or another proxy, the client IP in access logs depends on whether the server trusts that proxy.

- By default, local addresses and common private network ranges are trusted, so reverse proxies on the host or container network usually expose the real source IP automatically.
- If logs keep showing proxy addresses like `127.0.0.1` or `172.x.x.x`, the proxy egress is usually not in the trusted list.
- Add proxy IPs or CIDRs through `SUBLINK_TRUSTED_PROXIES` or `trusted_proxies` in `config.yaml`.

Example:

```yaml
trusted_proxies:
  - 127.0.0.1
  - ::1
  - 10.0.0.0/8
  - 172.16.0.0/12
  - 192.168.0.0/16
  - 100.64.0.0/10
  - 203.0.113.10
  - 198.51.100.0/24
```

Docker Compose environment variable form:

```yaml
environment:
  - SUBLINK_TRUSTED_PROXIES=127.0.0.1,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,100.64.0.0/10
```

If you are sure you do not want to trust any proxy headers, explicitly disable them:

```yaml
trusted_proxies: []
```

---

## Docker Deployment Example with Environment Variables

```yaml
services:
  sublinkpro:
    image: zerodeng/sublink-pro:latest
    container_name: sublinkpro
    ports:
      - "8000:8000"
    volumes:
      - "./db:/app/db"
      - "./template:/app/template"
      - "./logs:/app/logs"
    environment:
      - SUBLINK_PORT=8000
      # Database DSN, optional. SQLite is used by default when unset.
      # - SUBLINK_DSN=mysql://user:pass@mysql:3306/sublink?charset=utf8mb4&parseTime=True&loc=Local
      - SUBLINK_EXPIRE_DAYS=14
      - SUBLINK_LOGIN_FAIL_COUNT=5
      # Local data directory, optional, default for config.yaml / GeoIP / SQLite
      # - SUBLINK_DB_PATH=/app/db
      # GeoIP database path, optional, defaults to ./db/GeoLite2-City.mmdb
      # - SUBLINK_GEOIP_PATH=/app/db/GeoLite2-City.mmdb
      # Trusted reverse proxies, optional, comma separated. Host or private network reverse proxies usually need no extra change.
      # - SUBLINK_TRUSTED_PROXIES=127.0.0.1,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,100.64.0.0/10
      # Sensitive configuration, optional, generated automatically when unset
      # - SUBLINK_JWT_SECRET=your-secret-key
      # - SUBLINK_API_ENCRYPTION_KEY=your-encryption-key
      # - SUBLINK_MFA_RESET_SECRET=your-break-glass-secret
    restart: unless-stopped
```

> [!NOTE]
> For the full Docker Compose template, see `docker-compose.example.yml` in the project root.
