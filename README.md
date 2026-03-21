<div align="center">
  <img src="webs/src/assets/images/logo.svg" width="280px" />
  
  **✨ 强大的代理订阅管理与转换工具 ✨**

  <p>
    <img src="https://img.shields.io/github/go-mod/go-version/ZeroDeng01/sublinkPro?style=flat-square&logo=go&logoColor=white" alt="Go Version"/>
    <img src="https://img.shields.io/github/package-json/dependency-version/ZeroDeng01/sublinkPro/react?filename=webs%2Fpackage.json&style=flat-square&logo=react&logoColor=white&color=61DAFB" alt="React Version"/>
    <img src="https://img.shields.io/github/package-json/dependency-version/ZeroDeng01/sublinkPro/@mui/material?filename=webs%2Fpackage.json&style=flat-square&logo=mui&logoColor=white&label=MUI&color=007FFF" alt="MUI Version"/>
    <img src="https://img.shields.io/github/package-json/dependency-version/ZeroDeng01/sublinkPro/vite?filename=webs%2Fpackage.json&style=flat-square&logo=vite&logoColor=white&color=646CFF" alt="Vite Version"/>
  </p>
  <p>
    <img src="https://img.shields.io/github/v/release/ZeroDeng01/sublinkPro?style=flat-square&logo=github&label=Latest" alt="Latest Release"/>
    <img src="https://img.shields.io/github/release-date/ZeroDeng01/sublinkPro?style=flat-square&logo=github&label=Release%20Date" alt="Release Date"/>
  </p>
  <p>
    <img src="https://img.shields.io/docker/v/zerodeng/sublink-pro/latest?style=flat-square&logo=docker&logoColor=white&label=Docker%20Stable" alt="Docker Stable Version"/>
    <img src="https://img.shields.io/docker/pulls/zerodeng/sublink-pro?style=flat-square&logo=docker&logoColor=white&label=Docker%20Pulls" alt="Docker Pulls"/>
    <img src="https://img.shields.io/docker/image-size/zerodeng/sublink-pro/latest?style=flat-square&logo=docker&logoColor=white&label=Image%20Size" alt="Docker Image Size"/>
  </p>
  <p>
    <img src="https://img.shields.io/github/stars/ZeroDeng01/sublinkPro?style=flat-square&logo=github&label=Stars" alt="GitHub Stars"/>
    <img src="https://img.shields.io/github/forks/ZeroDeng01/sublinkPro?style=flat-square&logo=github&label=Forks" alt="GitHub Forks"/>
    <img src="https://img.shields.io/github/issues/ZeroDeng01/sublinkPro?style=flat-square&logo=github&label=Issues" alt="GitHub Issues"/>
    <img src="https://img.shields.io/github/license/ZeroDeng01/sublinkPro?style=flat-square&label=License" alt="License"/>
  </p>
  <p>
    <a href="https://github.com/ZeroDeng01/sublinkPro/issues">
      <img src="https://img.shields.io/badge/问题反馈-Issues-blue?style=flat-square&logo=github" alt="Issues"/>
    </a>
    <a href="https://github.com/ZeroDeng01/sublinkPro/releases">
      <img src="https://img.shields.io/badge/版本下载-Releases-green?style=flat-square&logo=github" alt="Releases"/>
    </a>
  </p>
</div>

---

## 📖 项目简介

`SublinkPro` 是基于优秀的开源项目 [sublinkX](https://github.com/gooaclok819/sublinkX) / [sublinkE](https://github.com/eun1e/sublinkE) 进行二次开发，在原项目基础上做了部分定制优化。感谢原作者的付出与贡献。

- 🎨 **前端框架**：基于 [Berry Free React Material UI Admin Template](https://github.com/codedthemes/berry-free-react-admin-template)
- ⚡ **后端技术**：Go + Gin + Gorm
- 🔐 **默认账号**：`admin` / `123456`（请安装后务必修改）
- 💻 **演示系统**: [https://sublink-pro-demo.zeabur.app](https://sublink-pro-demo.zeabur.app/) 用户名：admin 密码：123456

> [!WARNING]
> ⚠️ 本项目和原项目数据库不兼容，请不要混用。
>
> ⚠️ 请不要使用本项目以及任何本项目的衍生项目进行违反您以及您所服务用户的所在地法律法规的活动。本项目仅供个人开发和学习交流使用。

---

## ✨ 功能亮点

| 功能 | 说明 | 详情 |
|:---|:---|:---:|
| 🏷️ **智能标签系统** | 自动规则打标签、零代码筛选、支持 IP 质量条件 | [📖](docs/features/tags.md) |
| ⚡ **专业测速系统** | 双阶段测试、智能延迟测量、支持 IP 质量检测 | [📖](docs/features/speedtest.md) |
| 🔗 **链式代理** | Dialer-Proxy 原生支持、可视化配置、支持按 IP 质量选节点 | [📖](docs/features/chain-proxy.md) |
| ✈️ **机场管理** | 多格式导入、定时更新、流量监控、一键全量拉取 | [📖](docs/features/airport.md) |
| 🗂️ **分组排序** | 分组内机场优先级拖拽排序，控制订阅输出中的节点顺序 | - |
| 📋 **订阅分享** | 多链接管理、过期策略、访问统计 | [📖](docs/features/subscription-share.md) |
| 🌐 **Host 管理** | 域名映射、DNS 配置、CDN 优选 | [📖](docs/features/host.md) |
| 🤖 **Telegram Bot** | 远程测速、订阅管理、系统监控 | [📖](docs/features/telegram-bot.md) |
| 📜 **脚本系统** | 节点过滤、内容后处理、多脚本链式执行 | [📖](docs/script_support.md) |
| 🔔 **Webhooks** | 支持 PushDeer、Bark、钉钉、方糖等多平台通知 | - |
| 🔐 **安全特性** | Token 授权、API Key、IP 黑/白名单、访问日志 | - |

---

## 🚀 快速开始

### Docker Compose（推荐）

创建 `docker-compose.yml`：

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

启动服务：

```bash
docker-compose up -d
```

访问 `http://localhost:8000`，使用默认账号 `admin` / `123456` 登录。

默认使用 SQLite；如需切换到 MySQL 或 PostgreSQL，可通过 `SUBLINK_DSN`、配置文件 `dsn:` 或命令行 `--dsn` 指定数据库连接，示例见 [⚙️ 配置说明](docs/configuration.md)。

> [!TIP]
> 更多安装方式（Docker、一键脚本、更新升级等）请参阅 [📦 安装部署指南](docs/installation.md)

### 从 SQLite 迁移到 MySQL / PostgreSQL

如果您早期使用的是 SQLite，现在希望迁移到 MySQL 或 PostgreSQL，建议按以下流程操作：

1. 在旧的 SQLite 实例中登录后台，点击右上角头像菜单中的 **系统备份**，导出 `backup.zip`
2. 在新实例中配置好 MySQL 或 PostgreSQL 的 `DSN`，并确保目标库是一个全新的空库
3. 启动新实例后，进入 `设置 -> 数据迁移`
4. 上传旧实例导出的 `backup.zip`
5. 根据需要选择是否迁移 `AccessKey`、订阅访问日志，然后开始迁移
6. 迁移完成后，**请手动重启项目实例**，再重新登录检查数据

> [!IMPORTANT]
> 推荐使用 `backup.zip` 迁移。直接上传 `.db` 只会迁移数据库记录，不会恢复模板目录。

> [!NOTE]
> 如果迁移了 `AccessKey`，请确保新旧实例使用相同的 `API 加密密钥`；否则旧 API Key 可能无法继续使用。

> [!TIP]
> 如果迁移完成后提示“有 N 条警告”，可以到 `任务中心` 打开对应的“数据库迁移”任务查看详细警告内容。

---

## 📖 文档导航

### 🔧 安装与配置

| 文档 | 说明 |
|:---|:---|
| [📦 安装部署](docs/installation.md) | Docker、一键脚本、更新升级、Watchtower 自动更新 |
| [⚙️ 配置说明](docs/configuration.md) | 环境变量、命令行参数、验证码配置 |

### ✨ 功能详解

| 文档 | 说明 |
|:---|:---|
| [🏷️ 智能标签系统](docs/features/tags.md) | 自动规则打标签、零代码筛选、IP 质量规则 |
| [⚡ 测速系统](docs/features/speedtest.md) | 测速原理、IP 质量检测、参数配置 |
| [🔗 链式代理](docs/features/chain-proxy.md) | Dialer-Proxy、条件选节点、配置流程 |
| [✈️ 机场管理](docs/features/airport.md) | 订阅导入、定时更新、流量监控 |
| [📋 订阅分享](docs/features/subscription-share.md) | 多链接管理、过期策略、访问统计 |
| [🌐 Host 管理](docs/features/host.md) | 域名映射、DNS 配置、测速持久化 |
| [🤖 Telegram 机器人](docs/features/telegram-bot.md) | 命令列表、配置指南 |
| [📜 脚本功能](docs/script_support.md) | 节点过滤、内容后处理、函数参考 |
| [🔐 双重验证（MFA）](docs/features/mfa.md) | TOTP 设置、恢复码、应急重置流程 |

### 👨‍💻 开发者

| 文档 | 说明 |
|:---|:---|
| [🛠️ 开发指南](docs/development.md) | 项目结构、本地开发、定时任务开发 |

---

## 📡 多协议支持

| 客户端 | 支持协议 |
|:---|:---|
| **v2ray** | base64 通用格式 |
| **clash** | ss, ssr, trojan, vmess, vless, hy, hy2, tuic, AnyTLS, Socks5, HTTP, HTTPS |
| **surge** | ss, trojan, vmess, hy2, tuic |

---

## 🖼️ 项目预览

<details open>
<summary><b>点击展开/收起预览图</b></summary>

| | |
|:---:|:---:|
| ![预览1](docs/images/1.jpg) | ![预览2](docs/images/2.jpg) |
| ![预览3](docs/images/3.jpg) | ![预览4](docs/images/4.jpg) |
| ![预览5](docs/images/5.jpg) | ![预览6](docs/images/6.jpg) |
| ![预览7](docs/images/7.jpg) | ![预览8](docs/images/8.jpg) |
| ![预览9](docs/images/9.jpg) | ![预览10](docs/images/10.jpg) |
| ![预览11](docs/images/11.jpg) | ![预览12](docs/images/12.jpg) |

</details>

---

## 📊 项目统计

<div align="center">

[//]: # (  <img src="https://repobeez.abhijithganesh.com/api/insert/ZeroDeng01/sublinkPro" alt="Repobeez" height="0" width="0" style="display: none"/>)
  
  ![Star History Chart](https://api.star-history.com/svg?repos=ZeroDeng01/sublinkPro&type=Date)
</div>

---

## 🤝 贡献与支持

如果这个项目对您有帮助，欢迎：

- ⭐ **Star** 这个项目表示支持
- 🐛 提交 [Issue](https://github.com/ZeroDeng01/sublinkPro/issues) 反馈问题或建议
- 🔧 提交 Pull Request 贡献代码
- 📖 完善文档和使用教程

### 🙏 致谢

感谢以下项目的开源贡献：

- [sublinkX](https://github.com/gooaclok819/sublinkX) / [sublinkE](https://github.com/eun1e/sublinkE) - 原始项目
- [Berry Free React Admin Template](https://github.com/codedthemes/berry-free-react-admin-template) - 前端模板
- [Mihomo](https://github.com/MetaCubeX/mihomo) - 代理核心

---

<div align="center">
  <sub>Made with ❤️ by <a href="https://github.com/ZeroDeng01">ZeroDeng01</a></sub>
</div>
