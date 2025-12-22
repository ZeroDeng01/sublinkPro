# 配置说明

本文档详细介绍 SublinkPro 的配置方式和各项参数。

---

## 配置优先级

SublinkPro 支持多种配置方式，优先级从高到低为：

1. **命令行参数** - 适用于临时覆盖，如 `--port 9000`
2. **环境变量** - 推荐用于 Docker 部署
3. **配置文件** - `db/config.yaml`
4. **数据库存储** - 敏感配置自动存储
5. **默认值** - 程序内置默认配置

---

## 环境变量列表

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `SUBLINK_PORT` | 服务端口 | 8000 |
| `SUBLINK_DB_PATH` | 数据库目录 | ./db |
| `SUBLINK_LOG_PATH` | 日志目录 | ./logs |
| `SUBLINK_JWT_SECRET` | JWT签名密钥 | (自动生成) |
| `SUBLINK_API_ENCRYPTION_KEY` | API加密密钥 | (自动生成) |
| `SUBLINK_EXPIRE_DAYS` | Token过期天数 | 14 |
| `SUBLINK_LOGIN_FAIL_COUNT` | 登录失败次数限制 | 5 |
| `SUBLINK_LOGIN_FAIL_WINDOW` | 登录失败窗口(分钟) | 1 |
| `SUBLINK_LOGIN_BAN_DURATION` | 登录封禁时间(分钟) | 10 |
| `SUBLINK_GEOIP_PATH` | GeoIP数据库路径 | ./db/GeoLite2-City.mmdb |
| `SUBLINK_CAPTCHA_MODE` | 验证码模式 (1=关闭, 2=传统, 3=Turnstile) | 2 |
| `SUBLINK_TURNSTILE_SITE_KEY` | Cloudflare Turnstile Site Key | - |
| `SUBLINK_TURNSTILE_SECRET_KEY` | Cloudflare Turnstile Secret Key | - |
| `SUBLINK_TURNSTILE_PROXY_LINK` | Turnstile 验证代理链接（mihomo 格式） | - |
| `SUBLINK_ADMIN_PASSWORD` | 初始管理员密码 | 123456 |
| `SUBLINK_ADMIN_PASSWORD_REST` | 重置管理员密码 | 输入新管理员密码 |

---

## 命令行参数

```bash
# 查看帮助
./sublinkpro help

# 指定端口启动
./sublinkpro run --port 9000

# 指定数据库目录
./sublinkpro run --db /data/db

# 重置管理员密码
./sublinkpro setting -username admin -password newpass
```

---

## 敏感配置说明

> [!TIP]
> **JWT Secret** 和 **API 加密密钥** 是敏感配置，系统会按以下方式处理：
> 1. 优先从环境变量读取
> 2. 如未设置环境变量，从数据库读取
> 3. 如数据库也没有，自动生成随机密钥并存储到数据库
> 
> **特别说明**：如果您通过环境变量设置了这些值，系统会自动同步到数据库。这样即使后续忘记设置环境变量，系统也能从数据库恢复，方便迁移部署。

> [!WARNING]
> 如果您需要**多实例部署**或**集群部署**，请务必通过环境变量设置相同的 `SUBLINK_JWT_SECRET` 和 `SUBLINK_API_ENCRYPTION_KEY`，以确保各实例间的登录状态和 API Key 一致。

---

## 验证码配置

SublinkPro 支持三种验证码模式，通过 `SUBLINK_CAPTCHA_MODE` 环境变量配置：

| 模式 | 说明 |
|:---:|:---|
| **1** | 关闭验证码（不推荐，仅限内网环境） |
| **2** | 传统图形验证码（默认） |
| **3** | Cloudflare Turnstile（推荐，更安全） |

### Cloudflare Turnstile 配置

如需使用 Turnstile，请：

1. 访问 [Cloudflare Turnstile 控制台](https://dash.cloudflare.com/?to=/:account/turnstile) 创建站点
2. 获取 **Site Key** 和 **Secret Key**
3. 配置环境变量：

```yaml
environment:
  - SUBLINK_CAPTCHA_MODE=3
  - SUBLINK_TURNSTILE_SITE_KEY=your-site-key
  - SUBLINK_TURNSTILE_SECRET_KEY=your-secret-key
```

> [!NOTE]
> **降级机制**：如果配置了 Turnstile 模式但未提供完整的密钥配置，系统会自动降级为传统图形验证码。

### Turnstile 代理配置

如果您的服务器无法直接访问 Cloudflare API，可能会遇到 `context deadline exceeded` 超时错误。此时可以配置代理：

```yaml
environment:
  - SUBLINK_TURNSTILE_PROXY_LINK=vless://your-proxy-link...
```

> [!TIP]
> **代理链接格式**：使用 mihomo 支持的代理链接格式（如 `vless://`、`vmess://`、`ss://` 等）。与 Telegram 代理配置类似。

### Turnstile 验证模式

Cloudflare Turnstile 支持三种验证模式，在 Cloudflare 控制台创建 Site Key 时选择：

| 模式 | 说明 |
|:---:|:---|
| **Managed** | Cloudflare 自动决策是否需要交互，大多数用户无感通过 |
| **Non-Interactive** | 显示加载指示器，但无需用户交互 |
| **Invisible** | 完全不可见，后台静默完成验证 |

前端 widget 会自动根据 Site Key 对应的模式进行渲染，无需额外配置。

---

## Docker 部署示例（带环境变量）

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
      - SUBLINK_EXPIRE_DAYS=14
      - SUBLINK_LOGIN_FAIL_COUNT=5
      # GeoIP 数据库路径（可选，默认为 ./db/GeoLite2-City.mmdb）
      # - SUBLINK_GEOIP_PATH=/app/db/GeoLite2-City.mmdb
      # 敏感配置（可选，不设置则自动生成）
      # - SUBLINK_JWT_SECRET=your-secret-key
      # - SUBLINK_API_ENCRYPTION_KEY=your-encryption-key
    restart: unless-stopped
```

> [!NOTE]
> 完整的 Docker Compose 模板请参考项目根目录的 `docker-compose.example.yml` 文件。
