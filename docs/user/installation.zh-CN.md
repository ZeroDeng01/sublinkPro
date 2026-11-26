[English](installation.md) | 简体中文

[返回指南目录](README.zh-CN.md)

# 安装部署指南

本文档介绍 SublinkPro 的完整安装、更新和卸载方法。

## 选择版本

| 渠道 | 使用场景 | Docker 镜像 | Linux 一键脚本 |
| --- | --- | --- | --- |
| 稳定版 | 日常使用，推荐选择 | `zerodeng/sublink-pro:latest`，也可固定到 `:v1.3.0` 等版本 | 默认 |
| 预发布 | 提前体验 `main` 已发布的改动 | `zerodeng/sublink-pro:preview` 跟随最新预发布，也可使用发布说明中的 `:pre-YYMMDDHHMMSS` 时间标签固定版本 | `--preview` |
| 开发版 | 测试 `dev` 最新发布的构建 | `zerodeng/sublink-pro:dev` | 使用 Docker 镜像 |

稳定版和预发布的版本卡分别检查各自渠道的更新。开发版通过拉取 `:dev` 镜像并重建容器更新。**更新日志**页面显示最近 10 条正式发布。[GitHub Releases](https://github.com/ZeroDeng01/sublinkPro/releases)

---

## 📦 Docker Compose 运行（推荐）

> [!TIP]
> **推荐使用 Docker Compose 部署**，便于管理配置、升级和维护。

> [!IMPORTANT]
> `db/`、`template/`、`logs/` 属于运行时持久化目录，请在升级和迁移时保留。

创建 `docker-compose.yml` 文件：

```yaml
services:
  sublinkpro:
    # image: zerodeng/sublink-pro:dev # 开发版（功能尝鲜使用）
    # image: zerodeng/sublink-pro:preview # 滚动预发布
    # image: zerodeng/sublink-pro:pre-YYMMDDHHMMSS # 替换为已发布版本的时间标签
    image: zerodeng/sublink-pro:latest # 稳定版
    container_name: sublinkpro
    ports:
      - "8000:8000"
    volumes:
      - "./db:/app/db"
      - "./template:/app/template"
      - "./logs:/app/logs"
    restart: unless-stopped
```

如需扩展更多订阅输出格式，可选部署 Sub-Store sidecar：

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

建议把 Sub-Store 服务仅保留在 Compose 内部网络中，不要直接暴露端口。两个容器启动后，请登录后台进入 **用户中心 -> Sub-Store**，启用 sidecar、填写类似 `http://substore:3000` 的 Base URL、选择允许的输出目标并测试连接。Sub-Store 集成只通过该页面管理，不使用环境变量。

如需通过 Cloudflare Tunnel 暴露服务，可在启动后进入 **用户中心 -> Cloudflare Tunnel** 填写 token 并启动；启用自动连接后会随服务启动连接 Tunnel。完整步骤见 [Cloudflare Tunnel 远程访问](features/cloudflare-tunnel.zh-CN.md)。

官方 Docker 镜像已内置 `cloudflared`，非 Docker 部署则需要先按 Cloudflare 官方文档安装 `cloudflared`。

启动服务：

```bash
docker compose up -d
```

---

## 🐳 Docker 运行

<details>
<summary><b>稳定版</b></summary>

```bash
docker run --name sublinkpro -p 8000:8000 \
  -v $PWD/db:/app/db \
  -v $PWD/template:/app/template \
  -v $PWD/logs:/app/logs \
  -d zerodeng/sublink-pro
```

</details>

<details>
<summary><b>预发布</b></summary>

```bash
docker run --name sublinkpro -p 8000:8000 \
  -v $PWD/db:/app/db \
  -v $PWD/template:/app/template \
  -v $PWD/logs:/app/logs \
  -d zerodeng/sublink-pro:preview
```

</details>

<details>
<summary><b>开发版（功能尝鲜）</b></summary>

```bash
docker run --name sublinkpro -p 8000:8000 \
  -v $PWD/db:/app/db \
  -v $PWD/template:/app/template \
  -v $PWD/logs:/app/logs \
  -d zerodeng/sublink-pro:dev
```

</details>

---

## 📝 一键安装/更新脚本

在支持的 Linux 系统上以 root 身份运行。默认安装最新稳定版。

```bash
sh -c "$(wget -qO- https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/install.sh)"
```

加上 `--preview`，安装最新已发布且提供当前 Linux 架构安装包的预发布版本。

```bash
sh -c "$(wget -qO- https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/install.sh)" -- --preview
```

> [!NOTE]
> 安装脚本支持以下功能：
> - **全新安装**：首次安装时自动完成所有配置
> - **更新程序**：检测到已安装时，可选择更新（保留所有数据）
> - **重新安装**：可选择是否保留现有数据
> - **恢复安装**：检测到旧数据时，可选择恢复安装
> - **支持架构**：Linux x64、Linux ARM64、Linux ARMv7 32 位、Linux x86 32 位

---

## 🗑️ 一键卸载脚本

```bash
sh -c "$(wget -qO- https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/uninstall.sh)"
```

> [!NOTE]
> 卸载脚本会询问是否保留数据目录（db、logs、template），选择保留可用于后续重新安装时恢复数据。

---

## 🔄 项目更新

### 📝 一键脚本更新

再次运行安装命令即可更新。更新到预发布时，每次都要带上 `--preview`。不带参数时，脚本选择稳定版。

脚本会自动检测已安装的版本，并提供以下选项：
- **更新程序**：保留所有数据，仅更新程序文件
- **重新安装**：可选择是否保留数据

### 📦 Docker Compose 手动更新

使用固定时间标签或稳定版版本标签时，先将 `image:` 改为发布说明列出的新标签，再执行以下命令。使用 `:latest`、`:preview` 或 `:dev` 滚动标签时，直接使用现有 Compose 文件执行即可。

```bash
# 进入 docker-compose.yml 所在目录
cd /path/to/your/sublinkpro

# 拉取最新镜像
docker compose pull

# 重新创建并启动容器
docker compose up -d

# （可选）清理旧镜像
docker image prune -f
```

### 🐳 Docker 手动更新

以下示例沿用滚动稳定版镜像。其他渠道在拉取和启动时都填写所选镜像。

```bash
# 停止并删除旧容器
docker stop sublinkpro
docker rm sublinkpro

# 拉取最新镜像
docker pull zerodeng/sublink-pro

# 重新启动容器（使用与安装时相同的参数）
docker run --name sublinkpro -p 8000:8000 \
  -v $PWD/db:/app/db \
  -v $PWD/template:/app/template \
  -v $PWD/logs:/app/logs \
  -d zerodeng/sublink-pro

# （可选）清理旧镜像
docker image prune -f
```

---

## 反向代理与订阅设备校验

使用反向代理或自定义订阅转发服务时，请保留 `/c/` 上的客户端设备请求头，并在更新后检查公开订阅地址。详见[设备请求头转发与 HTTP 403 排查](features/subscription-share.zh-CN.md#反向代理与-http-403-排查)。

## 已使用 Watchtower 的部署

Watchtower 上游项目已归档。新部署请按上文的 [Docker Compose 更新步骤](#-docker-compose-手动更新)操作。已有 Watchtower 部署可先核对当前更新配置，再切换更新方式。[Watchtower 仓库](https://github.com/containrrr/watchtower)
