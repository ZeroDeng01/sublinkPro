# 安装部署指南

本文档介绍 SublinkPro 的完整安装、更新和卸载方法。

---

## 📦 Docker Compose 运行（推荐）

> [!TIP]
> **推荐使用 Docker Compose 部署**，便于管理配置、升级和维护。

创建 `docker-compose.yml` 文件：

```yaml
services:
  sublinkpro:
    # image: zerodeng/sublink-pro:dev # 开发版（功能尝鲜使用）
    image: zerodeng/sublink-pro # 稳定版
    container_name: sublinkpro
    ports:
      - "8000:8000"
    volumes:
      - "./db:/app/db"
      - "./template:/app/template"
      - "./logs:/app/logs"
    restart: unless-stopped
```

启动服务：

```bash
docker-compose up -d
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

```bash
wget https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/install.sh && sh install.sh
```

> [!NOTE]
> 安装脚本支持以下功能：
> - **全新安装**：首次安装时自动完成所有配置
> - **更新程序**：检测到已安装时，可选择更新（保留所有数据）
> - **重新安装**：可选择是否保留现有数据
> - **恢复安装**：检测到旧数据时，可选择恢复安装

---

## 🗑️ 一键卸载脚本

```bash
wget https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/uninstall.sh && sh uninstall.sh
```

> [!NOTE]
> 卸载脚本会询问是否保留数据目录（db、logs、template），选择保留可用于后续重新安装时恢复数据。

---

## 🔄 项目更新

### 📝 一键脚本更新

如果您使用一键脚本安装，可以再次运行安装脚本进行更新：

```bash
wget https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/install.sh && sh install.sh
```

脚本会自动检测已安装的版本，并提供以下选项：
- **更新程序**：保留所有数据，仅更新程序文件
- **重新安装**：可选择是否保留数据

### 📦 Docker Compose 手动更新

```bash
# 进入 docker-compose.yml 所在目录
cd /path/to/your/sublinkpro

# 拉取最新镜像
docker-compose pull

# 重新创建并启动容器
docker-compose up -d

# （可选）清理旧镜像
docker image prune -f
```

### 🐳 Docker 手动更新

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

## 🤖 Watchtower 自动更新

Watchtower 是一个可以自动更新 Docker 容器的工具，非常适合希望保持项目始终最新的用户。

### 方式一：独立运行 Watchtower

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
> - `--cleanup`：更新后自动清理旧镜像
> - `--interval 86400`：每 24 小时检查一次更新（单位：秒）
> - 最后的 `sublinkpro` 是要监控更新的容器名称，不指定则监控所有容器

### 方式二：集成到 Docker Compose

在您的 `docker-compose.yml` 中添加 Watchtower 服务：

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
    command: sublinkpro  # 只监控 sublinkpro 容器
```

> [!TIP]
> **Watchtower 高级配置**：
> - 可以设置 `WATCHTOWER_NOTIFICATIONS` 环境变量来配置更新通知（支持邮件、Slack、Gotify 等）
> - 更多配置请参考 [Watchtower 官方文档](https://containrrr.dev/watchtower/)



---

### ☁️ Zeabur 部署

https://zeabur.com/projects

**部署步骤：**

1. **新建项目与 Service**
   - 点击 "创建项目" > "Docker 容器镜像"
   - 输入镜像名称：`zerodeng/sublink-pro:latest`  (推荐稳定版 latest，开发版 dev 用于测试新功能)
   - 配置端口：`8000` (HTTP)
   - **配置卷（重要）**：
     * 点击卷
     * 点击 "添加卷" 添加新卷
     * 卷名称 > 容器路径
      ```env
      sublink-db = /app/db
      sublink-template = /app/template
      sublink-logs = /app/logs
      ```

2. **配置环境变量**

   环境变量中添加：

   ```env
   # 基础配置
   SUBLINK_PORT=8000
   SUBLINK_LOG_LEVEL=error
   SUBLINK_EXPIRE_DAYS=14

   # 登录安全
   SUBLINK_ADMIN_PASSWORD=123456 #默认管理员密码，仅首次启动有效
   SUBLINK_LOGIN_FAIL_COUNT=5
   SUBLINK_LOGIN_FAIL_WINDOW=1
   SUBLINK_LOGIN_BAN_DURATION=10

   # 安全密钥 !需填写! 随机32位以上字符串
   SUBLINK_JWT_SECRET=
   SUBLINK_API_ENCRYPTION_KEY=


   # 验证码(1为关闭)
   SUBLINK_CAPTCHA_MODE=2
   ```

3. **部署完成**
   - Zeabur 会自动拉取镜像并启动服务
   - 等待服务就绪后，需要手动设置访问域名（见下一步）

4. **设置访问域名（必须）**

   - 在服务页面，点击 "Networking" 或 "网络" 标签
   - 点击 "Generate Domain" 生成 Zeabur 提供的免费域名（如 `xxx.zeabur.app`）
   - 或者绑定自定义域名：
     * 点击 "Add Domain" 添加你的域名
     * 按照提示配置 DNS CNAME 记录指向 Zeabur 提供的目标地址
   - 设置完域名后即可通过域名访问,使用默认账号 `admin` / `123456` 登录



