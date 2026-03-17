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

| 环境变量 | 说明                              | 默认值                                 |
|----------|---------------------------------|-------------------------------------|
| `SUBLINK_PORT` | 服务端口                            | 8000                                |
| `SUBLINK_DSN` | 数据库 DSN（支持 sqlite/mysql/postgres） | 默认使用 SQLite：`sqlite://./db/sublink.db` |
| `SUBLINK_DB_PATH` | 本地数据目录 / SQLite 默认数据库目录        | ./db                                |
| `SUBLINK_LOG_PATH` | 日志目录                            | ./logs                              |
| `SUBLINK_JWT_SECRET` | JWT签名密钥                         | (自动生成)                              |
| `SUBLINK_API_ENCRYPTION_KEY` | API加密密钥                         | (自动生成)                              |
| `SUBLINK_EXPIRE_DAYS` | Token过期天数                       | 14                                  |
| `SUBLINK_LOGIN_FAIL_COUNT` | 登录失败次数限制                        | 5                                   |
| `SUBLINK_LOGIN_FAIL_WINDOW` | 登录失败窗口(分钟)                      | 1                                   |
| `SUBLINK_LOGIN_BAN_DURATION` | 登录封禁时间(分钟)                      | 10                                  |
| `SUBLINK_GEOIP_PATH` | GeoIP数据库路径                      | ./db/GeoLite2-City.mmdb             |
| `SUBLINK_CAPTCHA_MODE` | 验证码模式 (1=关闭, 2=传统, 3=Turnstile) | 2                                   |
| `SUBLINK_TURNSTILE_SITE_KEY` | Cloudflare Turnstile Site Key   | -                                   |
| `SUBLINK_TURNSTILE_SECRET_KEY` | Cloudflare Turnstile Secret Key | -                                   |
| `SUBLINK_TURNSTILE_PROXY_LINK` | Turnstile 验证代理链接（mihomo 格式）     | -                                   |
| `SUBLINK_WEB_BASE_PATH` | 前端访问基础路径（站点隐藏）                  | -                                   |
| `SUBLINK_ADMIN_PASSWORD` | 初始管理员密码                         | 123456                              |
| `SUBLINK_ADMIN_PASSWORD_REST` | 重置管理员密码                         | 输入新管理员密码                            |
| `SUBLINK_FEATURE` | 试验性功能开关                         | 目前可以设置其值为`SubNodePreview`开启订阅节点预览功能 |

---

## 命令行参数

```bash
# 查看帮助
./sublinkpro help

# 指定端口启动
./sublinkpro run --port 9000

# 指定 SQLite 数据库
./sublinkpro run --dsn "sqlite:///data/sublink.db"

# 指定 MySQL
./sublinkpro run --dsn "mysql://user:pass@tcp(127.0.0.1:3306)/sublink?charset=utf8mb4&parseTime=True&loc=Local"

# 指定 PostgreSQL
./sublinkpro run --dsn "postgres://user:pass@127.0.0.1:5432/sublink?sslmode=disable"

# 指定本地数据目录（配置文件 / GeoIP / 默认 SQLite）
./sublinkpro run --db /data

# 重置管理员密码
./sublinkpro setting -username admin -password newpass
```

---

## 数据库 DSN

SublinkPro 现在支持通过 `dsn` 统一配置数据库连接，支持以下方言：

- `sqlite://`
- `mysql://`
- `postgres://`
- `postgresql://`

如果 `dsn` 为空，系统会自动退回到 SQLite，并使用 `db_path/sublink.db` 作为数据库文件。

### SQLite 示例

```yaml
dsn: sqlite:///app/db/sublink.db
```

### MySQL 示例

```yaml
dsn: mysql://user:pass@tcp(mysql:3306)/sublink?charset=utf8mb4&parseTime=True&loc=Local
```

### PostgreSQL 示例

```yaml
dsn: postgres://user:pass@postgres:5432/sublink?sslmode=disable
```

> [!TIP]
> 使用 MySQL 或 PostgreSQL 时，`db_path` 仍然用于本地配置文件和 GeoIP 数据库存储；它不再决定实际数据库后端。

## 从 SQLite 迁移到 MySQL / PostgreSQL

如果您的旧实例一直使用 SQLite，现在希望迁移到 MySQL 或 PostgreSQL，建议使用内置的“数据库迁移”功能。

### 迁移前准备

1. 准备一个新的 MySQL 或 PostgreSQL 空库
2. 为新实例配置数据库 `DSN`
3. 确认旧实例可以正常登录后台
4. 如需保留旧 `AccessKey`，请确认新旧实例的 `SUBLINK_API_ENCRYPTION_KEY` 保持一致

### 第一步：在新实例中配置 DSN

您可以通过以下任一方式为新实例配置数据库：

- 环境变量：`SUBLINK_DSN`
- 配置文件：`db/config.yaml` 中的 `dsn:`
- 命令行参数：`./sublinkpro run --dsn "..."`

示例：

```yaml
# MySQL
dsn: mysql://user:pass@tcp(mysql:3306)/sublink?charset=utf8mb4&parseTime=True&loc=Local

# PostgreSQL
dsn: postgres://user:pass@postgres:5432/sublink?sslmode=disable
```

> [!IMPORTANT]
> 建议迁移目标使用全新的空库。不要在已有业务数据的库上直接导入。

### 第二步：在旧 SQLite 实例中导出备份

登录旧实例后台后：

1. 点击右上角头像菜单
2. 选择 **系统备份**
3. 下载生成的 `backup.zip`

推荐使用 `backup.zip` 作为迁移源文件，因为它会同时包含：

- `db` 目录中的 SQLite 数据库文件
- `template` 目录中的模板文件

> [!TIP]
> 也可以直接上传 `.db`、`.sqlite`、`.sqlite3` 文件，但这种方式只迁移数据库记录，不会恢复模板目录。

### 第三步：在新实例中执行迁移

启动新实例后，进入：

`设置 -> 数据迁移`

然后按以下步骤操作：

1. 上传旧实例导出的 `backup.zip`
2. 根据需要选择是否迁移 `AccessKey`
3. 根据需要选择是否迁移“订阅访问日志”
4. 勾选“我已确认本次导入会覆盖当前实例的业务数据”
5. 点击 **开始迁移**

迁移任务会在后台执行，您可以在以下位置查看进度与结果：

- 右下角任务进度面板
- `任务中心`

### 迁移完成后的操作

迁移完成后，请执行以下操作：

1. 查看迁移结果是否成功
2. 如果提示“有 N 条警告”，到 `任务中心` 打开对应的“数据库迁移”任务查看详细警告
3. **手动重启项目实例**
4. 重新登录后台并检查关键数据是否正常

### 迁移注意事项

- 本次导入会覆盖当前实例的业务数据
- 推荐只在新部署的 MySQL / PostgreSQL 实例首次迁移时使用
- 订阅访问日志通常较大，默认不建议迁移
- 如果旧 `AccessKey` 迁移后无法使用，请检查 `SUBLINK_API_ENCRYPTION_KEY` 是否与旧实例一致
- 如果迁移完成后登录态异常，重新登录一次即可

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

## 站点隐藏配置

通过设置 `SUBLINK_WEB_BASE_PATH` 可以隐藏管理站点入口，类似 3x-ui 的自定义路径功能。

```yaml
environment:
  - SUBLINK_WEB_BASE_PATH=/admin
```

设置后：
- 访问 `http://domain/` 返回 404
- 访问 `http://domain/admin` 才能进入管理界面
- API 接口 (`/api/*`) 和订阅获取 (`/c/*`) **不受影响**

> [!TIP]
> 路径支持带或不带前导斜杠，如 `admin` 或 `/admin` 效果相同。

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
      # 数据库 DSN（可选，未设置时默认使用 SQLite）
      # - SUBLINK_DSN=mysql://user:pass@mysql:3306/sublink?charset=utf8mb4&parseTime=True&loc=Local
      - SUBLINK_EXPIRE_DAYS=14
      - SUBLINK_LOGIN_FAIL_COUNT=5
      # 本地数据目录（可选，默认用于 config.yaml / GeoIP / SQLite）
      # - SUBLINK_DB_PATH=/app/db
      # GeoIP 数据库路径（可选，默认为 ./db/GeoLite2-City.mmdb）
      # - SUBLINK_GEOIP_PATH=/app/db/GeoLite2-City.mmdb
      # 敏感配置（可选，不设置则自动生成）
      # - SUBLINK_JWT_SECRET=your-secret-key
      # - SUBLINK_API_ENCRYPTION_KEY=your-encryption-key
    restart: unless-stopped
```

> [!NOTE]
> 完整的 Docker Compose 模板请参考项目根目录的 `docker-compose.example.yml` 文件。
