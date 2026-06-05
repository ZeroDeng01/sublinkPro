English | [简体中文](cloudflare-tunnel.zh-CN.md)

# Cloudflare Tunnel Remote Access

Cloudflare Tunnel lets SublinkPro serve Web access through the Cloudflare network without exposing a public IP or opening inbound ports. SublinkPro includes a Cloudflare Tunnel management page that saves the Tunnel token, starts and stops the `cloudflared` process, and shows connection status.

> [!NOTE]
> This document is for publishing the SublinkPro Web admin UI to a public domain. Cloudflare's official docs also cover private networks, SSH, RDP, and more Zero Trust use cases. See [Cloudflare Tunnel official docs](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/) for details.

---

## What Problem Does It Solve

Traditional reverse proxy setups usually need a public IP, open inbound `80/443` ports, Nginx/Caddy configuration, and certificate maintenance. Cloudflare Tunnel works differently:

```text
Browser
  │
  ▼
Cloudflare edge network
  │  HTTPS / Zero Trust / DNS
  ▼
cloudflared actively creates an outbound connection to Cloudflare
  │
  ▼
SublinkPro http://localhost:8000
```

`cloudflared` runs on the same machine or container as SublinkPro. It only creates outbound connections to Cloudflare. External traffic reaches Cloudflare first, then returns to SublinkPro through the Tunnel.

![Cloudflare Tunnel request path](https://developers.cloudflare.com/_astro/handshake.eh3a-Ml1_26dKUX.webp)

Suitable scenarios:

| Scenario | Why it fits |
|:---|:---|
| Home broadband / NAT / no public IP | No public IP or port mapping required |
| Docker deployment | SublinkPro image includes `cloudflared`; enter token on the page and start |
| No desire to maintain Nginx / Caddy | Public entry, HTTPS, and DNS are handled by Cloudflare |
| Smaller exposed surface | The server only needs outbound access to Cloudflare |

---

## Prerequisites

1. You have a Cloudflare account.
2. Your domain has been added to Cloudflare and DNS is hosted by Cloudflare.
3. SublinkPro can be accessed locally, for example `http://localhost:8000`.
4. The server can reach Cloudflare Tunnel connection ports outbound. Cloudflare recommends checking `7844` connectivity in restricted firewall environments. See [Connectivity pre-checks](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/troubleshoot-tunnels/connectivity-prechecks/).

Deployment differences:

| Deployment | cloudflared source |
|:---|:---|
| Official Docker image | Image includes `cloudflared` |
| Binary / systemd / manual deployment | Install `cloudflared` according to Cloudflare official docs and make sure it is in `PATH` |

---

## Install cloudflared

If you use the official SublinkPro Docker image, skip this section. The image already includes `cloudflared`. If you run SublinkPro with a binary, systemd, BaoTa, a panel, or another method, install `cloudflared` on the same machine and confirm the SublinkPro process can find it in `PATH`.

After installation, run:

```bash
cloudflared version
```

If the version prints successfully, return to SublinkPro **User Center -> Cloudflare Tunnel** and start it.

### macOS

Homebrew is recommended:

```bash
brew install cloudflared
cloudflared version
```

To upgrade later:

```bash
brew upgrade cloudflared
```

### Windows

Winget is recommended first:

```powershell
winget install --id Cloudflare.cloudflared
cloudflared version
```

You can also download the Windows executable from Cloudflare Downloads or GitHub Releases:

- [Cloudflare Downloads](https://developers.cloudflare.com/tunnel/downloads/)
- [cloudflared Releases](https://github.com/cloudflare/cloudflared/releases)

If you download a single `cloudflared.exe`, add its directory to system `PATH`, then reopen the terminal and verify:

```powershell
cloudflared version
```

### Debian / Ubuntu / Linux Mint and other Debian based distributions

Cloudflare recommends using the Cloudflare Package Repository for Debian, Ubuntu, Linux Mint, and other Debian based distributions:

```bash
sudo mkdir -p --mode=0755 /usr/share/keyrings
curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg | sudo tee /usr/share/keyrings/cloudflare-main.gpg >/dev/null
echo 'deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared any main' | sudo tee /etc/apt/sources.list.d/cloudflared.list
sudo apt-get update
sudo apt-get install cloudflared
cloudflared version
```

If your system does not support the `any` source, choose the distribution codename:

| System | APT source codename |
|:---|:---|
| Debian 12 | `bookworm` |
| Ubuntu 20.04 | `focal` |
| Ubuntu 22.04 | `jammy` |
| Ubuntu 24.04 | `noble` |

Example for Ubuntu 24.04:

```bash
echo 'deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared noble main' | sudo tee /etc/apt/sources.list.d/cloudflared.list
sudo apt-get update
sudo apt-get install cloudflared
cloudflared version
```

### RHEL / CentOS / Rocky Linux / AlmaLinux / Fedora / Amazon Linux

RPM based distributions can use the Cloudflare Package Repository or directly download the RPM package. General quick method:

```bash
curl -L --output cloudflared.rpm "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-x86_64.rpm"
sudo rpm -i cloudflared.rpm
cloudflared version
```

For ARM64 servers, use:

```bash
curl -L --output cloudflared.rpm "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-aarch64.rpm"
sudo rpm -i cloudflared.rpm
cloudflared version
```

If an older version is already installed, use:

```bash
sudo rpm -U cloudflared.rpm
```

### Arch Linux / Manjaro

Cloudflare mainly provides binary, deb, rpm, pkg, and similar packages. Arch users can usually download the Linux binary directly:

```bash
sudo curl -L "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64" -o /usr/local/bin/cloudflared
sudo chmod +x /usr/local/bin/cloudflared
cloudflared version
```

For ARM64 devices:

```bash
sudo curl -L "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64" -o /usr/local/bin/cloudflared
sudo chmod +x /usr/local/bin/cloudflared
cloudflared version
```

### Alpine Linux

Alpine users can also use the official Linux binary directly:

```bash
sudo mkdir -p /usr/local/bin
sudo curl -L "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64" -o /usr/local/bin/cloudflared
sudo chmod +x /usr/local/bin/cloudflared
cloudflared version
```

For ARM64 devices, replace `cloudflared-linux-amd64` with `cloudflared-linux-arm64`.

### Direct binary download, general Linux

If you are unsure whether your distribution package manager is supported, download directly by CPU architecture:

| Architecture | Download URL |
|:---|:---|
| x86_64 / amd64 | `https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64` |
| ARM64 / aarch64 | `https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64` |
| ARM | `https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm` |
| 32 bit x86 | `https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-386` |

Example:

```bash
sudo curl -L "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64" -o /usr/local/bin/cloudflared
sudo chmod +x /usr/local/bin/cloudflared
cloudflared version
```

> [!TIP]
> SublinkPro does not require you to run `cloudflared service install`. You only need to install the `cloudflared` command, then paste the final token segment from Cloudflare's install command into the SublinkPro page. SublinkPro hosts the `cloudflared` process start and stop lifecycle.

Official download page: [Cloudflare Downloads](https://developers.cloudflare.com/tunnel/downloads/).

---

## Step 1: Create a Tunnel in Cloudflare Zero Trust

1. Open [Cloudflare Dashboard](https://dash.cloudflare.com/).
2. Open **Zero Trust**.
3. In the left menu, open **Networks** -> **Connectors** -> **Cloudflare Tunnels**. Some newer interfaces may show **Networks** -> **Tunnels**.
4. Click **Create a tunnel**.
5. Choose **Cloudflared** as Connector type, then click **Next**.
6. Enter a Tunnel name, such as `sublinkpro-home` or `sublinkpro-prod`.
7. Click **Save tunnel**.

Official steps: [Create a tunnel, dashboard](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/get-started/create-remote-tunnel/).

---

## Step 2: Copy the Tunnel token

After saving the Tunnel, Cloudflare opens the connector install page and asks you to choose a runtime environment. The page shows an install command like:

```bash
sudo cloudflared service install eyJhIjoi...省略...In0=
```

Or on Windows:

```powershell
cloudflared.exe service install eyJhIjoi...省略...In0=
```

You can copy the whole command or only the final token. SublinkPro splits by whitespace and uses the final segment as the token, so all of these are accepted:

```text
eyJhIjoi...In0=
```

```text
sudo cloudflared service install eyJhIjoi...In0=
```

```text
brew install cloudflared && sudo cloudflared service install eyJhIjoi...In0=
```

> [!IMPORTANT]
> A Tunnel token is a credential for connecting to that Tunnel. Don't post it in issues, chat groups, logs, or screenshots. SublinkPro encrypts the submitted token and status APIs only return a masked token.

---

## Step 3: Configure Public Hostname in Cloudflare

When creating the Tunnel, Cloudflare continues to guide you through publishing an application. Choose **Published applications** and add a public entry:

| Field | Recommended value |
|:---|:---|
| Subdomain | `sub`, `sublink`, or your preferred subdomain |
| Domain | Choose the domain hosted on Cloudflare, such as `example.com` |
| Path | Usually leave empty |
| Type | `HTTP` |
| URL | `localhost:8000` |

If SublinkPro is not using the default port, replace `8000` with the actual port.

```text
Public hostname: https://sub.example.com
Service type:    HTTP
Service URL:     localhost:8000
```

Save after configuration. Cloudflare automatically creates the DNS record for this hostname and forwards traffic to the Tunnel.

![Cloudflare Tunnel connector online example](https://developers.cloudflare.com/_astro/connector.BnVS4T_M_ZxLFu6.webp)

---

## Step 4: Start the Tunnel in SublinkPro

1. Sign in to the SublinkPro admin UI.
2. Open **User Center** from the avatar menu in the upper right.
3. Open the **Cloudflare Tunnel** tab.
4. Paste the token or Cloudflare install command into **Cloudflare Tunnel Token**.
5. Click **Save Configuration**.
6. Click **Start Cloudflared**.

After successful start, the page shows:

- `cloudflared installed`
- `Running`
- `Saved token: eyJh********...`
- `Registered tunnel connection` and similar logs in `Runtime logs`

If you want SublinkPro to connect the Tunnel automatically after restart, enable **Auto connect Tunnel on service startup** and save the configuration. This switch uses only the token saved on the page. It does not support injecting the token through environment variables.

---

## Step 5: Verify Access

1. Return to the Cloudflare Zero Trust Tunnel page and confirm the connector is online.
2. Open the public domain you configured, for example:

```text
https://sub.example.com
```

3. Confirm that the SublinkPro login page is visible.
4. After signing in, check that subscription, node, task, and other pages load normally.

> [!TIP]
> If you access SublinkPro only through Cloudflare Tunnel, Docker deployments can avoid mapping host ports. Before first configuration, you still need internal network access or a temporary port to reach the admin UI.

---

## FAQ

### Page shows “cloudflared not detected”

SublinkPro cannot find the `cloudflared` command in the current runtime environment.

- Official Docker image: confirm that you are using a new image that includes this feature, and pull the image again.
- Non Docker: install `cloudflared` according to Cloudflare official docs, and make sure `cloudflared version` can run in the SublinkPro process `PATH`.

### It keeps failing to connect to Cloudflare after start

Check:

1. Whether the token belongs to the current Tunnel.
2. Whether the server can access Cloudflare outbound.
3. Whether the firewall allows `cloudflared` to connect to Cloudflare, especially port `7844` in restricted networks.
4. Whether the Cloudflare Zero Trust connector page shows an online instance.

### Domain access returns 502 / 1033

This usually means Cloudflare can reach the Tunnel, but the Tunnel cannot reach local SublinkPro.

Check the Service URL in Public Hostname:

- When built in `cloudflared` and SublinkPro run in the same Docker container, `localhost:8000` is usually correct.
- For non Docker deployments, use the actual SublinkPro listening address and port.
- If `SUBLINK_PORT` was changed, update this value too.

### Access stops after stopping the Tunnel

This is expected. Cloudflare Tunnel is the external access entry. After `cloudflared` stops, Cloudflare cannot forward requests to SublinkPro.

### Is Nginx / Caddy still needed?

Usually no. Cloudflare Tunnel already handles the public entry and HTTPS termination. If you need complex path routing, aggregation of multiple internal services, or custom auth on the local machine, you can still use Nginx / Caddy locally and point the Tunnel to the reverse proxy address.

---

## Security Advice

- Don't write Tunnel tokens into public docs, screenshots, or issues.
- If you suspect token leakage, delete the old connector or regenerate the install command in Cloudflare Zero Trust, then clear the old token in SublinkPro and save the new one.
- Enable a strong password and MFA for the SublinkPro admin UI.
- If the admin UI is exposed publicly, consider Cloudflare Access, IP rules, or other access control policies to restrict who can reach it.

---

## Official References

- [Cloudflare Tunnel overview](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/)
- [Create a tunnel, dashboard](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/get-started/create-remote-tunnel/)
- [Cloudflare Tunnel public hostname](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/routing-to-tunnel/)
- [Connectivity pre-checks](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/troubleshoot-tunnels/connectivity-prechecks/)
