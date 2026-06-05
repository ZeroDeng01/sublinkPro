[English](cloudflare-tunnel.md) | 简体中文

# Cloudflare Tunnel 远程访问

Cloudflare Tunnel 可以让 SublinkPro 在不暴露公网 IP、不开放入站端口的情况下，通过 Cloudflare 网络对外提供 Web 访问。SublinkPro 内置 Cloudflare Tunnel 管理页，负责保存 Tunnel token、启动/停止 `cloudflared` 进程，并展示连接状态。

> [!NOTE]
> 本文面向“把 SublinkPro Web 管理界面发布到公网域名”的场景。Cloudflare 官方文档同时覆盖私有网络、SSH、RDP 等更多 Zero Trust 用法，详见 [Cloudflare Tunnel 官方说明](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/)。

---

## 它解决什么问题

传统反向代理通常需要：公网 IP、开放 `80/443` 入站端口、Nginx/Caddy 配置和证书维护。Cloudflare Tunnel 的模式不同：

```text
浏览器
  │
  ▼
Cloudflare 边缘网络
  │  HTTPS / Zero Trust / DNS
  ▼
cloudflared 主动向 Cloudflare 建立出站连接
  │
  ▼
SublinkPro http://localhost:8000
```

`cloudflared` 运行在 SublinkPro 所在机器或容器中，只主动向 Cloudflare 建立出站连接。外部访问先到 Cloudflare，再通过这条 Tunnel 回到 SublinkPro。

![Cloudflare Tunnel 请求链路](https://developers.cloudflare.com/_astro/handshake.eh3a-Ml1_26dKUX.webp)

适合这些场景：

| 场景 | 为什么适合 |
|:---|:---|
| 家宽 / NAT / 无公网 IP | 不需要公网 IP，也不需要端口映射 |
| Docker 部署 | SublinkPro 镜像已内置 `cloudflared`，页面填 token 即可启动 |
| 不想维护 Nginx / Caddy | 公网入口、HTTPS 和 DNS 由 Cloudflare 处理 |
| 希望收敛暴露面 | 服务器只需要能主动连出到 Cloudflare |

---

## 前置条件

1. 已有 Cloudflare 账号。
2. 你的域名已经添加到 Cloudflare，并由 Cloudflare 托管 DNS。
3. SublinkPro 服务可以正常本地访问，例如 `http://localhost:8000`。
4. 服务器可以主动访问 Cloudflare Tunnel 连接端口。Cloudflare 官方建议在受限防火墙环境中先检查 `7844` 端口连通性，参考 [Connectivity pre-checks](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/troubleshoot-tunnels/connectivity-prechecks/)。

部署差异：

| 部署方式 | cloudflared 来源 |
|:---|:---|
| Docker 官方镜像 | 镜像已内置 `cloudflared` |
| 二进制 / systemd / 手动部署 | 需要先按 Cloudflare 官方安装文档安装 `cloudflared`，并确保命令在 `PATH` 中 |

---

## 安装 cloudflared

如果你使用 SublinkPro Docker 官方镜像，可以跳过本节；镜像内已经包含 `cloudflared`。如果你使用二进制、systemd、宝塔、面板或其他方式运行 SublinkPro，需要先在同一台机器上安装 `cloudflared`，并确认 SublinkPro 进程能在 `PATH` 中找到它。

安装完成后先运行：

```bash
cloudflared version
```

能正常输出版本号，再回到 SublinkPro 的 **用户中心 -> Cloudflare Tunnel** 页面启动。

### macOS

推荐使用 Homebrew：

```bash
brew install cloudflared
cloudflared version
```

如果后续需要升级：

```bash
brew upgrade cloudflared
```

### Windows

推荐优先使用 Winget：

```powershell
winget install --id Cloudflare.cloudflared
cloudflared version
```

也可以到 Cloudflare 官方下载页或 GitHub Releases 下载 Windows 可执行文件：

- [Cloudflare Downloads](https://developers.cloudflare.com/tunnel/downloads/)
- [cloudflared Releases](https://github.com/cloudflare/cloudflared/releases)

如果你下载的是单个 `cloudflared.exe`，请把它所在目录加入系统 `PATH`，然后重新打开终端验证：

```powershell
cloudflared version
```

### Debian / Ubuntu / Linux Mint 等 Debian 系发行版

Cloudflare 官方推荐使用 Cloudflare Package Repository。适用于 Debian、Ubuntu、Linux Mint 等 Debian 系发行版：

```bash
sudo mkdir -p --mode=0755 /usr/share/keyrings
curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg | sudo tee /usr/share/keyrings/cloudflare-main.gpg >/dev/null
echo 'deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared any main' | sudo tee /etc/apt/sources.list.d/cloudflared.list
sudo apt-get update
sudo apt-get install cloudflared
cloudflared version
```

如果你的系统不支持 `any` 源，也可以按发行版代号选择：

| 系统 | APT 源代号 |
|:---|:---|
| Debian 12 | `bookworm` |
| Ubuntu 20.04 | `focal` |
| Ubuntu 22.04 | `jammy` |
| Ubuntu 24.04 | `noble` |

例如 Ubuntu 24.04：

```bash
echo 'deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared noble main' | sudo tee /etc/apt/sources.list.d/cloudflared.list
sudo apt-get update
sudo apt-get install cloudflared
cloudflared version
```

### RHEL / CentOS / Rocky Linux / AlmaLinux / Fedora / Amazon Linux

RPM 系发行版建议使用 Cloudflare Package Repository 或直接下载 RPM 包。通用快速方式：

```bash
curl -L --output cloudflared.rpm "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-x86_64.rpm"
sudo rpm -i cloudflared.rpm
cloudflared version
```

如果是 ARM64 服务器，把下载地址改为：

```bash
curl -L --output cloudflared.rpm "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-aarch64.rpm"
sudo rpm -i cloudflared.rpm
cloudflared version
```

如果系统已经安装过旧版本，可以用：

```bash
sudo rpm -U cloudflared.rpm
```

### Arch Linux / Manjaro

Cloudflare 官方主要提供二进制、deb、rpm、pkg 等发布包。Arch 系用户通常可以直接下载 Linux 二进制：

```bash
sudo curl -L "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64" -o /usr/local/bin/cloudflared
sudo chmod +x /usr/local/bin/cloudflared
cloudflared version
```

ARM64 设备使用：

```bash
sudo curl -L "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64" -o /usr/local/bin/cloudflared
sudo chmod +x /usr/local/bin/cloudflared
cloudflared version
```

### Alpine Linux

Alpine 用户也可以直接使用官方发布的 Linux 二进制：

```bash
sudo mkdir -p /usr/local/bin
sudo curl -L "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64" -o /usr/local/bin/cloudflared
sudo chmod +x /usr/local/bin/cloudflared
cloudflared version
```

ARM64 设备同样把 `cloudflared-linux-amd64` 改成 `cloudflared-linux-arm64`。

### 直接下载二进制（通用 Linux）

不确定发行版包管理器是否适配时，可以按 CPU 架构直接下载：

| 架构 | 下载地址 |
|:---|:---|
| x86_64 / amd64 | `https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64` |
| ARM64 / aarch64 | `https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64` |
| ARM | `https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm` |
| 32 位 x86 | `https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-386` |

示例：

```bash
sudo curl -L "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64" -o /usr/local/bin/cloudflared
sudo chmod +x /usr/local/bin/cloudflared
cloudflared version
```

> [!TIP]
> SublinkPro 不需要你运行 `cloudflared service install`。你只需要安装 `cloudflared` 命令，然后把 Cloudflare 页面里安装命令最后一段 token 粘贴到 SublinkPro 页面。SublinkPro 会托管 `cloudflared` 进程的启动和停止。

官方下载页：[Cloudflare Downloads](https://developers.cloudflare.com/tunnel/downloads/)。

---

## 第一步：在 Cloudflare Zero Trust 创建 Tunnel

1. 打开 [Cloudflare Dashboard](https://dash.cloudflare.com/)。
2. 进入 **Zero Trust**。
3. 在左侧菜单进入 **Networks** -> **Connectors** -> **Cloudflare Tunnels**。部分新版界面也可能显示为 **Networks** -> **Tunnels**。
4. 点击 **Create a tunnel**。
5. Connector type 选择 **Cloudflared**，点击 **Next**。
6. 填写 Tunnel 名称，例如：`sublinkpro-home`、`sublinkpro-prod`。
7. 点击 **Save tunnel**。

官方步骤参考：[Create a tunnel (dashboard)](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/get-started/create-remote-tunnel/)。

---

## 第二步：复制 Tunnel token

保存 Tunnel 后，Cloudflare 会进入安装 connector 的页面，并让你选择运行环境。页面中会出现类似下面的安装命令：

```bash
sudo cloudflared service install eyJhIjoi...省略...In0=
```

或 Windows 命令：

```powershell
cloudflared.exe service install eyJhIjoi...省略...In0=
```

你可以复制整条命令，也可以只复制最后一段 token。SublinkPro 会按空白分割并取最后一段作为有效 token，所以粘贴以下内容都可以：

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
> Tunnel token 等价于连接该 Tunnel 的凭据。不要发到 issue、聊天群、日志或截图里。SublinkPro 会加密保存页面提交的 token，状态接口只返回遮罩后的 token。

---

## 第三步：在 Cloudflare 配置 Public Hostname

创建 Tunnel 时，Cloudflare 会继续引导你发布应用。选择 **Published applications**，添加一个公开访问入口：

| 字段 | 推荐值 |
|:---|:---|
| Subdomain | `sub`、`sublink` 或你希望的子域名 |
| Domain | 选择已托管在 Cloudflare 的域名，例如 `example.com` |
| Path | 通常留空 |
| Type | `HTTP` |
| URL | `localhost:8000` |

如果你的 SublinkPro 不是默认端口，请把 `8000` 改成实际端口。

```text
Public hostname: https://sub.example.com
Service type:    HTTP
Service URL:     localhost:8000
```

配置完成后保存。此时 Cloudflare 会自动为这个 hostname 创建对应 DNS 记录，并把访问流量转发到 Tunnel。

![Cloudflare Tunnel connector 在线示意](https://developers.cloudflare.com/_astro/connector.BnVS4T_M_ZxLFu6.webp)

---

## 第四步：在 SublinkPro 启动 Tunnel

1. 登录 SublinkPro 管理界面。
2. 进入右上角头像菜单中的 **个人中心**。
3. 打开 **Cloudflare Tunnel** 标签页。
4. 在 **Cloudflare Tunnel Token** 输入框中粘贴 token 或 Cloudflare 安装命令。
5. 点击 **保存配置**。
6. 点击 **启动 Cloudflared**。

启动成功后页面会显示：

- `已安装 cloudflared`
- `运行中`
- `已保存 token：eyJh********...`
- `运行日志` 中出现 `Registered tunnel connection` 等连接日志

如果希望 SublinkPro 重启后自动连接 Tunnel，打开 **随服务启动自动连接 Tunnel**，然后保存配置。这个开关只使用页面中保存的 token，不支持通过环境变量注入 token。

---

## 第五步：验证访问

1. 回到 Cloudflare Zero Trust 的 Tunnel 页面，确认 connector 处于在线状态。
2. 在浏览器访问你配置的公开域名，例如：

```text
https://sub.example.com
```

3. 确认可以看到 SublinkPro 登录页。
4. 登录后检查订阅、节点、任务等页面是否正常加载。

> [!TIP]
> 如果你只通过 Cloudflare Tunnel 访问 SublinkPro，Docker 部署可以不映射宿主机端口。但首次配置前仍需要能从内网或临时端口访问管理界面。

---

## 常见问题

### 页面提示“未检测到 cloudflared”

说明 SublinkPro 当前运行环境找不到 `cloudflared` 命令。

- Docker 官方镜像：确认使用的是包含本功能的新镜像，并重新拉取镜像。
- 非 Docker：按 Cloudflare 官方安装文档安装 `cloudflared`，并确保 `cloudflared version` 可以在 SublinkPro 进程的 `PATH` 中执行。

### 启动后一直连不上 Cloudflare

检查：

1. token 是否来自当前 Tunnel。
2. 服务器是否能主动访问 Cloudflare。
3. 防火墙是否允许 `cloudflared` 连接 Cloudflare，尤其是受限网络中的 `7844` 端口。
4. Cloudflare Zero Trust 的 connector 页面是否出现在线实例。

### 访问域名出现 502 / 1033

通常表示 Cloudflare 能到 Tunnel，但 Tunnel 到本地 SublinkPro 不通。

检查 Public Hostname 的 Service URL：

- Docker 内置 `cloudflared` 和 SublinkPro 在同一容器中时，通常填写 `localhost:8000`。
- 非 Docker 部署时，填写 SublinkPro 实际监听地址和端口。
- 如果修改了 `SUBLINK_PORT`，这里也要同步修改。

### 停止 Tunnel 后访问中断

这是正常行为。Cloudflare Tunnel 是外部访问入口，停止 `cloudflared` 后 Cloudflare 无法再把请求转发到 SublinkPro。

### 是否还需要 Nginx / Caddy

一般不需要。Cloudflare Tunnel 已经承担公网入口和 HTTPS 终止。如果你有复杂路径转发、内网多服务聚合或自定义鉴权需求，也可以在本机继续使用 Nginx / Caddy，再让 Tunnel 指向反向代理地址。

---

## 安全建议

- 不要把 Tunnel token 写进公开文档、截图或 issue。
- 如果怀疑 token 泄露，到 Cloudflare Zero Trust 删除旧 connector / 重新生成安装命令，并在 SublinkPro 中清除旧 token 后重新保存。
- 推荐为 SublinkPro 管理界面启用强密码和 MFA。
- 如果管理界面暴露到公网，建议结合 Cloudflare Access、IP 规则或其他访问控制策略限制访问人群。

---

## 官方参考

- [Cloudflare Tunnel 概览](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/)
- [Create a tunnel (dashboard)](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/get-started/create-remote-tunnel/)
- [Cloudflare Tunnel public hostname](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/routing-to-tunnel/)
- [Connectivity pre-checks](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/troubleshoot-tunnels/connectivity-prechecks/)
