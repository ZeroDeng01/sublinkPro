English | [简体中文](installation.zh-CN.md)

# Installation Guide

This document explains how to install, update, and uninstall SublinkPro.

---

## 📦 Run with Docker Compose, recommended

> [!TIP]
> **Docker Compose is recommended** because it makes configuration, upgrades, and maintenance easier.

> [!IMPORTANT]
> `db/`, `template/`, and `logs/` are runtime persistence directories. Keep them during upgrades and migrations.

Create `docker-compose.yml`:

```yaml
services:
  sublinkpro:
    # image: zerodeng/sublink-pro:dev # Development version, for trying new features
    image: zerodeng/sublink-pro # Stable version
    container_name: sublinkpro
    ports:
      - "8000:8000"
    volumes:
      - "./db:/app/db"
      - "./template:/app/template"
      - "./logs:/app/logs"
    restart: unless-stopped
```

Optional Sub-Store sidecar for expanded subscription output formats:

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

  substore:
    image: xream/sub-store
    container_name: substore
    environment:
      - SUB_STORE_BACKEND_API_PORT=3000
      - SUB_STORE_BODY_JSON_LIMIT=10mb
    restart: unless-stopped
```

Keep the Sub-Store service inside the Compose network and do not publish its port unless you protect it separately. After both containers start, sign in and open **User Center -> Sub-Store** to enable the sidecar, set the base URL such as `http://substore:3000`, choose allowed output targets, and test the connection. Sub-Store integration is managed from that page, not through environment variables.

To expose the service through Cloudflare Tunnel, start the instance first, then open **User Center -> Cloudflare Tunnel**, enter the token, and start it. When auto connect is enabled, the Tunnel connects when the service starts. See [Cloudflare Tunnel remote access](features/cloudflare-tunnel.md) for the full flow.

The official Docker image includes `cloudflared`. Non Docker deployments need `cloudflared` installed first according to Cloudflare's official documentation.

Start the service:

```bash
docker-compose up -d
```

---

## 🐳 Run with Docker

<details>
<summary><b>Stable version</b></summary>

```bash
docker run --name sublinkpro -p 8000:8000 \
  -v $PWD/db:/app/db \
  -v $PWD/template:/app/template \
  -v $PWD/logs:/app/logs \
  -d zerodeng/sublink-pro
```

</details>

<details>
<summary><b>Development version, for trying new features</b></summary>

```bash
docker run --name sublinkpro -p 8000:8000 \
  -v $PWD/db:/app/db \
  -v $PWD/template:/app/template \
  -v $PWD/logs:/app/logs \
  -d zerodeng/sublink-pro:dev
```

</details>

---

## 📝 One Line Install or Update Script

```bash
sh -c "$(wget -qO- https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/install.sh)"
```

> [!NOTE]
> The install script supports:
> - **Fresh install**: completes all setup automatically on first install
> - **Update**: detects an existing install and updates the program while keeping data
> - **Reinstall**: lets you choose whether to keep existing data
> - **Restore install**: detects old data and lets you restore it

---

## 🗑️ One Line Uninstall Script

```bash
sh -c "$(wget -qO- https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/uninstall.sh)"
```

> [!NOTE]
> The uninstall script asks whether to keep the data directories, including db, logs, and template. Keeping them allows later reinstalls to restore data.

---

## 🔄 Project Updates

### 📝 Update with the one line script

If you installed with the one line script, run the install script again to update:

```bash
sh -c "$(wget -qO- https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/install.sh)"
```

The script detects the installed version and provides these options:

- **Update program**: keep all data and update program files only
- **Reinstall**: choose whether to keep data

### 📦 Manual Docker Compose update

```bash
# Enter the directory containing docker-compose.yml
cd /path/to/your/sublinkpro

# Pull the latest image
docker-compose pull

# Recreate and start the container
docker-compose up -d

# Optional: clean old images
docker image prune -f
```

### 🐳 Manual Docker update

```bash
# Stop and remove the old container
docker stop sublinkpro
docker rm sublinkpro

# Pull the latest image
docker pull zerodeng/sublink-pro

# Start the container again with the same parameters used during installation
docker run --name sublinkpro -p 8000:8000 \
  -v $PWD/db:/app/db \
  -v $PWD/template:/app/template \
  -v $PWD/logs:/app/logs \
  -d zerodeng/sublink-pro

# Optional: clean old images
docker image prune -f
```

---

## 🤖 Automatic Updates with Watchtower

Watchtower automatically updates Docker containers. It is useful if you want the project to stay current.

### Option 1: Run Watchtower separately

```bash
docker run -d \
  --name watchtower \
  -v /var/run/docker.sock:/var/run/docker.sock \
  containrrr/watchtower \
  --cleanup \
  --interval 86400 \
  sublinkpro
```

> [!NOTE]
> - `--cleanup`: remove old images after updates
> - `--interval 86400`: check for updates every 24 hours, in seconds
> - The final `sublinkpro` is the container name to monitor. If omitted, all containers are monitored.

### Option 2: Add Watchtower to Docker Compose

Add the Watchtower service to your `docker-compose.yml`:

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

  watchtower:
    image: containrrr/watchtower
    container_name: watchtower
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - TZ=Asia/Shanghai
      - WATCHTOWER_CLEANUP=true
      - WATCHTOWER_POLL_INTERVAL=86400
    restart: unless-stopped
    command: sublinkpro  # Only monitor the sublinkpro container
```

> [!TIP]
> **Advanced Watchtower configuration**:
> - Set `WATCHTOWER_NOTIFICATIONS` to configure update notifications, including email, Slack, Gotify, and others
> - See the [official Watchtower documentation](https://containrrr.dev/watchtower/) for more settings
