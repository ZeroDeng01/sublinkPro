English | [简体中文](installation.zh-CN.md)

[Guide index](README.md)

# Installation Guide

This document explains how to install, update, and uninstall SublinkPro.

## Choose a version

| Channel | When to use it | Docker image | Linux one-line installer |
| --- | --- | --- | --- |
| Stable | Recommended for regular use | `zerodeng/sublink-pro:latest` or a fixed version such as `:v1.3.0` | Default |
| Preview | Try changes published from `main` before the next stable release | `zerodeng/sublink-pro:preview` follows published previews. Pin an exact version with the `:pre-YYMMDDHHMMSS` tag from its release notes | `--preview` |
| Development | Test the latest published builds from `dev` | `zerodeng/sublink-pro:dev` | Use the Docker image |

The version card checks stable and Preview builds for updates in their own channels. To update a development build, pull `:dev` and recreate the container. The **Changelog** page shows the 10 most recent stable releases. [GitHub Releases](https://github.com/ZeroDeng01/sublinkPro/releases)

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
    # image: zerodeng/sublink-pro:preview # Rolling Preview
    # image: zerodeng/sublink-pro:pre-YYMMDDHHMMSS # Replace with a published Preview tag to pin it
    image: zerodeng/sublink-pro:latest # Stable version
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

Keep the Sub-Store service inside the Compose network and do not publish its port unless you protect it separately. After both containers start, sign in and open **Application Settings -> Sub-Store** to enable the sidecar, set the base URL such as `http://substore:3000`, choose allowed output targets, and test the connection. Sub-Store integration is managed from that page, not through environment variables.

To expose the service through Cloudflare Tunnel, start the instance first, then open **Application Settings -> Cloudflare Tunnel**, enter the token, and start it. When auto connect is enabled, the Tunnel connects when the service starts. See [Cloudflare Tunnel remote access](features/cloudflare-tunnel.md) for the full flow.

The official Docker image includes `cloudflared`. Non Docker deployments need `cloudflared` installed first according to Cloudflare's official documentation.

Start the service:

```bash
docker compose up -d
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
<summary><b>Preview</b></summary>

```bash
docker run --name sublinkpro -p 8000:8000 \
  -v $PWD/db:/app/db \
  -v $PWD/template:/app/template \
  -v $PWD/logs:/app/logs \
  -d zerodeng/sublink-pro:preview
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

Run the script as root on a supported Linux host. By default it installs the latest stable release.

```bash
sh -c "$(wget -qO- https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/install.sh)"
```

Use `--preview` to install the latest published Preview with a binary for your Linux architecture.

```bash
sh -c "$(wget -qO- https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/install.sh)" -- --preview
```

> [!NOTE]
> The install script supports:
> - **Fresh install**: completes all setup automatically on first install
> - **Update**: detects an existing install and updates the program while keeping data
> - **Reinstall**: lets you choose whether to keep existing data
> - **Restore install**: detects old data and lets you restore it
> - **Architectures**: Linux x64, Linux ARM64, Linux ARMv7 32-bit, and Linux x86 32-bit

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

Run the install command again to update. Use `--preview` each time you want to update to a Preview. The command without arguments selects the stable release.

The script detects the installed version and provides these options:

- **Update program**: keep all data and update program files only
- **Reinstall**: choose whether to keep data

### 📦 Manual Docker Compose update

For a pinned Preview or stable version, first set `image:` to the new tag listed in that release's notes. For the rolling `:latest`, `:preview`, and `:dev` images, run the commands with your current Compose file.

```bash
# Enter the directory containing docker-compose.yml
cd /path/to/your/sublinkpro

# Pull the latest image
docker compose pull

# Recreate and start the container
docker compose up -d

# Optional: clean old images
docker image prune -f
```

### 🐳 Manual Docker update

The example follows the rolling stable image. For another channel, use the chosen image in both the pull and run commands.

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

## Reverse proxies and subscription device checks

For installations using a reverse proxy or custom subscription sidecar, preserve client device request headers on `/c/` and check the public subscription URL after updates. See [device header forwarding and HTTP 403 troubleshooting](features/subscription-share.md#reverse-proxies-and-http-403-troubleshooting).

## Existing Watchtower deployments

The upstream Watchtower project is archived. Use the [Docker Compose update steps](#-manual-docker-compose-update) above for new deployments. If you already use Watchtower, review your update configuration before switching. [Watchtower repository](https://github.com/containrrr/watchtower)
