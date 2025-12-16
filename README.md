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

## ✨ 新增功能

| 状态 | 功能描述 |
|:---:|:---|
| ✅ | 修复部分页面BUG |
| ✅ | 支持 Clash `dialer-proxy` 属性 |
| ✅ | 允许添加并使用 API KEY 访问 API |
| ✅ | 导入、定时更新订阅链接中的节点（可通过前置代理订阅） |
| ✅ | 支持 AnyTLS、Socks5 协议 |
| ✅ | 订阅节点排序 |
| ✅ | 全新 UI，交互和操作便捷性大大提升，移动端友好 |
| 🔄 | 更多功能持续开发中... |

---

## 🎯 系统核心亮点

<table>
<tr>
<td width="50%">

### 🏷️ 智能标签系统（核心优势）
- **自动规则打标**：根据延迟、速度、国家、协议等自动分类
- **零代码筛选**：订阅使用标签白/黑名单，无需脚本编程
- **标签互斥组**：同组标签自动替换，避免标签冗余
- **动态更新**：测速后自动更新标签，始终反映最新状态

</td>
<td width="50%">

### ⚡ 专业测速系统
- **双阶段测试**：延迟测试与速度测试分离，更精准
- **智能延迟测量**：支持 UnifiedDelay 模式，可选择包含或排除握手时间
- **独立并发控制**：延迟和速度测试可分别设置并发数
- **状态自动标记**：根据测速结果自动标记节点状态

</td>
</tr>
<tr>
<td>

### 🔐 高安全性与自由度
- 支持访问订阅记录查询
- Token 授权与 API Key 访问
- IP 黑/白名单功能
- 简易安全配置管理

</td>
<td>

### 🔔 Webhooks 通知
- 支持 PushDeer、Bark、钉钉、方糖等多平台
- 订阅更新完成通知
- 测速任务完成通知

</td>
</tr>
<tr>
<td>

### 📜 JavaScript 脚本系统
- **节点过滤脚本**：订阅生成前自定义过滤逻辑
- **内容后处理**：对最终订阅内容进行二次修改
- **多脚本链式执行**：按顺序依次执行多个脚本
- **完整日志支持**：`console.log()` 输出调试信息

</td>
<td>

### ✏️ 智能节点重命名 & 模板生成
- **规则化重命名**：支持变量替换，如 `$Country`、`$Tags`、`$Protocol`
- **批量重命名**：一键按规则重命名所有节点
- **ACL4SSR 规则同步**：自动从 ACL4SSR 生成 Clash/Surge 模板
- **模板自定义**：灵活配置代理组和规则组

</td>
</tr>
</table>

---

## 🏷️ 自动标签系统详解

SublinkPro 的自动标签系统是本项目最强大的功能之一，让节点管理从「手动编辑」升级为「规则驱动」。

### 💡 核心优势

| 优势 | 说明 |
|:---|:---|
| **🚫 零代码筛选** | 通过标签规则给节点打标签后，订阅可直接使用标签进行复杂条件筛选，**无需编写任何代码或脚本** |
| **🔄 动态自动化** | 标签随节点状态自动更新，例如测速后延迟变高的节点会自动从「低延迟」标签中移除 |
| **📊 多维度分类** | 支持按延迟、速度、国家、协议、来源、节点名等多种条件组合打标签 |
| **🎛️ 灵活运算符** | 支持等于、包含、大于、小于、正则匹配等丰富的条件运算符 |
| **♻️ 互斥组管理** | 同一互斥组内的标签自动替换，确保节点分类唯一且清晰 |

### 📋 使用场景示例

```
场景一：按延迟分级
├── 规则：延迟 < 100ms → 标签「⚡极速」
├── 规则：延迟 100-300ms → 标签「✅正常」
└── 规则：延迟 > 300ms → 标签「🐌较慢」

场景二：按地区分类
├── 规则：国家 = 香港 → 标签「🇭🇰香港」
├── 规则：国家 = 日本 → 标签「🇯🇵日本」
└── 规则：国家 = 美国 → 标签「🇺🇸美国」

场景三：订阅筛选（无需编码！）
├── 订阅A：标签白名单 = 「⚡极速」「🇭🇰香港」→ 只返回香港极速节点
└── 订阅B：标签黑名单 = 「🐌较慢」→ 排除所有慢速节点
```

> [!TIP]
> **标签互斥组使用场景**：创建「优秀」、「良好」、「差」三个标签并设为同组「速度评级」，测速时节点只会保留最新的评级，避免标签堆积。

---

## ⚡ 测速系统详解

SublinkPro 提供专业级的节点测速功能，采用科学的测试方法确保结果准确可靠。

### 🔬 技术特点

| 特点 | 说明 |
|:---|:---|
| **双阶段分离测试** | 延迟测试与下载速度测试分开进行，各自使用最适合的测试URL和方法 |
| **智能延迟测量** | 支持 UnifiedDelay 模式，可选择包含或排除握手时间（系统自动发送两次请求取第二次结果） |
| **独立并发设置** | 延迟测试和速度测试可分别配置不同的并发数，平衡效率与精度 |
| **状态自动标记** | 测速完成后自动根据结果更新节点的延迟状态和速度状态 |
| **实时进度展示** | 测速过程中实时显示进度和状态，支持任务面板查看 |
| **流量统计** | 每个节点测速完成后记录消耗流量，任务完成时汇总显示 |

### 📊 测速状态分类

| 颜色 | 延迟阈值 | 速度阈值 |
|:---|:---|:---|
| 🟢 绿色 | < 200ms | >= 5 MB/s |
| 🟡 黄色 | 200-500ms | 1-5 MB/s |
| 🔴 红色 | >= 500ms 或超时/失败 | < 1 MB/s 或失败 |
| ⚪ 灰色 | 未测试 | 未测试 |

### 🔧 测速原理与流量计算

> [!IMPORTANT]
> **核心原理**：测速采用「限时下载」方式，即在设定的超时时间内尽可能多地下载数据，然后根据实际下载的字节数和耗时计算速度。

**速度计算公式**：
```
速度 (MB/s) = 实际下载字节数 ÷ 1024 ÷ 1024 ÷ 实际耗时(秒)
```

**常见疑问**：

| 现象 | 原因 |
|:---|:---|
| 设置 5MB 测速文件，但实际下载不到 5MB | 节点速度慢，在超时时间内未下载完 |
| 有的节点下载 500KB，有的下载 5MB | 速度快的节点在超时前下载完成，慢的节点下载不完 |
| 流量消耗显示几百KB | 节点速度约为 100KB/s × 超时时间(秒) |

**示例**（假设超时时间为 5 秒）：

| 节点速度 | 实际下载量 | 计算结果 |
|:---|:---|:---|
| 10 MB/s | 5 MB（提前完成） | 速度 ≈ 10 MB/s |
| 1 MB/s | 5 MB | 速度 = 1 MB/s |
| 0.2 MB/s | 1 MB | 速度 = 0.2 MB/s |
| 0.1 MB/s | 500 KB | 速度 = 0.1 MB/s |

### ⚙️ 参数配置建议

| 参数 | 说明 | 建议值 |
|:---|:---|:---|
| **测速超时时间** | 每个节点测速的最大时长 | 5-10秒（平衡速度与准确性） |
| **测速文件 URL** | 用于下载测速的文件地址 | 建议 ≥ 超时时间 × 预期最大速度 |
| **延迟测试 URL** | 用于延迟测试的地址 | 建议使用 HTTP 204 或小文件，如 Cloudflare 的 generate_204 |
| **延迟测试并发** | 同时进行延迟测试的节点数 | 10-50（可适当调高） |
| **速度测试并发** | 同时进行速度测试的节点数 | 1-3（避免带宽竞争） |

> [!TIP]
> **测速文件大小建议**：如果超时时间为 5 秒，预期最快节点为 20 MB/s，则测速文件应 ≥ 100MB。推荐使用 Cloudflare 的测速 URL：`https://speed.cloudflare.com/__down?bytes=100000000`（100MB）【⚠️注意：文件越大超时时间越大可能消耗流量越多】

> [!WARNING]
> **流量消耗提示**：每次测速会消耗实际带宽流量。100个节点 × 5MB 测速文件 = 最多消耗 500MB 流量。慢速节点消耗较少，但快速节点会下载完整文件。

---

## 📡 多协议支持

| 客户端 | 支持协议 |
|:---|:---|
| **v2ray** | base64 通用格式 |
| **clash** | ss, ssr, trojan, vmess, vless, hy, hy2, tuic, AnyTLS, Socks5 |
| **surge** | ss, trojan, vmess, hy2, tuic |

---

## 📦 功能模块总览

### 🖥️ 系统功能

| 模块 | 功能说明 |
|:---|:---|
| **📊 仪表盘** | 系统概览、节点统计、协议分布、地区分布、标签分布等可视化图表 |
| **🌐 节点管理** | 节点增删改查、批量操作、条件筛选、标签管理、测速操作 |
| **📋 订阅管理** | 订阅链接生成、模板配置、节点筛选规则、访问日志、定时更新 |
| **🏷️ 标签管理** | 标签创建编辑、自动规则配置、互斥组设置、批量打标签 |
| **📝 模板管理** | Clash/Surge 配置模板、规则组管理、ACL4SSR 规则同步 |
| **📜 脚本管理** | JavaScript 脚本编辑、节点过滤脚本、内容后处理脚本 |
| **🔑 API Key** | API 密钥管理、权限控制、访问统计 |
| **⚙️ 系统设置** | 通知配置、测速配置、安全设置、用户管理、数据备份 |

### 🛡️ 安全特性

- ✅ Token 授权访问控制
- ✅ API Key 独立权限管理
- ✅ IP 黑/白名单过滤
- ✅ 订阅访问日志记录
- ✅ 操作审计追踪

### 🔔 通知推送

- ✅ PushDeer 推送
- ✅ Bark 推送
- ✅ 钉钉机器人
- ✅ 方糖 (Server酱)
- ✅ 订阅更新通知
- ✅ 测速完成通知

### 🌍 GeoIP 数据库

- ✅ IP 地理位置查询
- ✅ 节点落地国家/地区检测
- ✅ 登录位置信息显示
- ✅ 支持代理下载数据库
- ✅ 可通过界面配置下载地址

> [!NOTE]
> GeoIP 数据库不再内置在程序中，系统首次启动时会引导用户下载。数据库必须是 MaxMind 的 mmdb 格式且包含 city 数据 (GeoLite2-City)。

---

## 🤖 Telegram 机器人

SublinkPro 内置了强大的 Telegram 机器人，让您可以通过 Telegram 随时随地管理您的系统。

### ✨ 核心功能

| 功能 | 命令 | 说明 |
|:---|:---|:---|
| **📊 仪表盘统计** | `/stats` | 查看系统概览、订阅数、节点在线情况、最快/最慢节点等 |
| **🖥️ 系统监控** | `/monitor` | 实时查看服务器 CPU、内存使用率、运行时间等 |
| **⚡ 远程测速** | `/speedtest` | 远程触发定时测速任务，或仅对未测速节点进行测试 |
| **📋 订阅管理** | `/subscriptions` | 查看所有订阅，**一键生成临时访问链接**（支持自定义域名） |
| **🌐 节点管理** | `/nodes` | 查看节点总数、在线数及地区分布详情 |
| **🏷️ 标签规则** | `/tags` | 手动触发自动标签规则，立即更新节点标签 |
| **📝 任务管理** | `/tasks` | 查看当前正在运行的后台任务，支持远程取消任务 |

### 🚀 配置指南

1. **创建机器人**：
   - 在 Telegram 中搜索 [@BotFather](https://t.me/BotFather)
   - 发送 `/newbot` 创建新机器人，获取 **Bot Token**

2. **系统配置**：
   - 进入 SublinkPro 「系统设置」 -> 「Telegram 设置」
   - 填入 **Bot Token**
   - (可选) 配置 **代理设置**，如果您的服务器无法直接访问 Telegram API
   - (可选) 配置 **远程访问域名**，用于机器人生成可访问的订阅链接

3. **绑定管理员**：
   - 在配置页面启用机器人并保存
   - 向您的机器人发送 `/start` 命令
   - 系统会自动识别并绑定您的 Chat ID

> [!TIP]
> **订阅链接生成**：机器人生成的订阅链接会自动使用您配置的「远程访问域名」。为了安全起见，生成的链接带有临时的这个 Token，建议按需生成。


## 🔧 快速安装

### 📦 Docker Compose 运行（推荐）

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

### 🐳 Docker 运行

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

### 📝 一键安装/更新脚本

```bash
wget https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/install.sh && sh install.sh
```

> [!NOTE]
> 安装脚本支持以下功能：
> - **全新安装**：首次安装时自动完成所有配置
> - **更新程序**：检测到已安装时，可选择更新（保留所有数据）
> - **重新安装**：可选择是否保留现有数据
> - **恢复安装**：检测到旧数据时，可选择恢复安装

### 🗑️ 一键卸载脚本

```bash
wget https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/uninstall.sh && sh uninstall.sh
```

> [!NOTE]
> 卸载脚本会询问是否保留数据目录（db、logs、template），选择保留可用于后续重新安装时恢复数据。

> [!TIP]
> 推荐优先使用 **Docker Compose 部署** 以获得最佳兼容性和便捷的更新体验。

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

如果您使用 Docker Compose 部署，可以通过以下命令手动更新：

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

如果您使用 `docker run` 命令部署，可以通过以下步骤更新：

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

### 🤖 Watchtower 自动更新（推荐）

Watchtower 是一个可以自动更新 Docker 容器的工具，非常适合希望保持项目始终最新的用户。

#### 方式一：独立运行 Watchtower

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

#### 方式二：集成到 Docker Compose

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

启动服务：

```bash
docker-compose up -d
```

> [!TIP]
> **Watchtower 高级配置**：
> - 可以设置 `WATCHTOWER_NOTIFICATIONS` 环境变量来配置更新通知（支持邮件、Slack、Gotify 等）
> - 更多配置请参考 [Watchtower 官方文档](https://containrrr.dev/watchtower/)

---

## ⚙️ 配置说明

### 配置优先级

SublinkPro 支持多种配置方式，优先级从高到低为：

1. **命令行参数** - 适用于临时覆盖，如 `--port 9000`
2. **环境变量** - 推荐用于 Docker 部署
3. **配置文件** - `db/config.yaml`
4. **数据库存储** - 敏感配置自动存储
5. **默认值** - 程序内置默认配置

### 环境变量列表

| 环境变量 | 说明 | 默认值      |
|----------|------|----------|
| `SUBLINK_PORT` | 服务端口 | 8000     |
| `SUBLINK_DB_PATH` | 数据库目录 | ./db     |
| `SUBLINK_LOG_PATH` | 日志目录 | ./logs   |
| `SUBLINK_JWT_SECRET` | JWT签名密钥 | (自动生成)   |
| `SUBLINK_API_ENCRYPTION_KEY` | API加密密钥 | (自动生成)   |
| `SUBLINK_EXPIRE_DAYS` | Token过期天数 | 14       |
| `SUBLINK_LOGIN_FAIL_COUNT` | 登录失败次数限制 | 5        |
| `SUBLINK_LOGIN_FAIL_WINDOW` | 登录失败窗口(分钟) | 1        |
| `SUBLINK_LOGIN_BAN_DURATION` | 登录封禁时间(分钟) | 10       |
| `SUBLINK_GEOIP_PATH` | GeoIP数据库路径 | ./db/GeoLite2-City.mmdb |
| `SUBLINK_ADMIN_PASSWORD` | 初始管理员密码 | 123456   |
| `SUBLINK_ADMIN_PASSWORD_REST` | 重置管理员密码 | 输入新管理员密码 |

### 命令行参数

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

### 敏感配置说明

> [!TIP]
> **JWT Secret** 和 **API 加密密钥** 是敏感配置，系统会按以下方式处理：
> 1. 优先从环境变量读取
> 2. 如未设置环境变量，从数据库读取
> 3. 如数据库也没有，自动生成随机密钥并存储到数据库
> 
> **特别说明**：如果您通过环境变量设置了这些值，系统会自动同步到数据库。这样即使后续忘记设置环境变量，系统也能从数据库恢复，方便迁移部署。

> [!WARNING]
> 如果您需要**多实例部署**或**集群部署**，请务必通过环境变量设置相同的 `SUBLINK_JWT_SECRET` 和 `SUBLINK_API_ENCRYPTION_KEY`，以确保各实例间的登录状态和 API Key 一致。

### Docker 部署示例（带环境变量）

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

</details>

---

## 📜 脚本功能说明

SublinkPro 支持使用 JavaScript 脚本对订阅内容进行自定义处理。

### 1️⃣ 节点过滤 `filterNode`

在生成订阅内容之前执行，用于对节点列表进行过滤或修改。

<details>
<summary><b>查看函数签名与示例</b></summary>

**函数签名:**
```javascript
function filterNode(nodes, clientType) {
    // nodes: 节点对象数组
    // clientType: 客户端类型 (v2ray, clash, surge)
    // 返回值: 修改后的节点对象数组
    return nodes;
}
```

**示例:**
```javascript
function filterNode(nodes, clientType) {
    // 过滤掉名称包含 "测试" 的节点
    var newNodes = [];
    for (var i = 0; i < nodes.length; i++) {
        if (nodes[i].Name.indexOf("测试") === -1) {
            newNodes.push(nodes[i]);
        }
    }
    return newNodes;
}
```

</details>

### 2️⃣ 内容后处理 `subMod`

在生成最终订阅内容之后执行，用于对最终的文本内容进行修改。

<details>
<summary><b>查看函数签名与示例</b></summary>

**函数签名:**
```javascript
function subMod(input, clientType) {
    // input: 原始输入内容
    // clientType: 客户端类型
    // 返回值: 修改后的内容字符串
    return input;
}
```

</details>

> [!NOTE]
> - 脚本中可以使用 `console.log()` 输出日志到后台
> - 多个脚本会按照排序顺序依次执行
> - 脚本支持的函数请查看 [📚 脚本文档](docs/script_support.md)

---

## 🛠️ 开发指南

欢迎参与 SublinkPro 的开发！以下是项目结构和开发相关说明。

### 📁 项目结构

```
sublinkPro/
├── 📂 api/                    # API 接口层
│   ├── node.go               # 节点相关 API
│   ├── sub.go                # 订阅相关 API
│   ├── tag.go                # 标签相关 API
│   ├── template.go           # 模板相关 API
│   ├── setting.go            # 设置相关 API
│   └── ...
├── 📂 models/                 # 数据模型层
│   ├── node.go               # 节点模型
│   ├── subcription.go        # 订阅模型
│   ├── tag.go                # 标签模型
│   ├── template.go           # 模板模型
│   ├── db_migrate.go         # 数据库迁移
│   └── ...
├── 📂 services/               # 业务服务层
│   ├── scheduler.go          # 定时任务调度器
│   ├── tag_service.go        # 标签服务
│   ├── 📂 geoip/             # GeoIP 服务
│   ├── 📂 mihomo/            # Mihomo 核心服务
│   └── 📂 sse/               # Server-Sent Events
├── 📂 routers/                # 路由定义
│   ├── node.go               # 节点路由
│   ├── tag.go                # 标签路由
│   └── ...
├── 📂 node/                   # 节点协议解析
│   ├── sub.go                # 订阅链接解析
│   └── 📂 protocol/          # 各协议解析器
├── 📂 utils/                  # 工具函数
│   ├── speedtest.go          # 测速工具
│   ├── node_renamer.go       # 节点重命名工具
│   ├── script_executor.go    # 脚本执行器
│   └── ...
├── 📂 middlewares/            # 中间件
├── 📂 constants/              # 常量定义
├── 📂 database/               # 数据库连接
├── 📂 cache/                  # 缓存管理
├── 📂 dto/                    # 数据传输对象
├── 📂 webs/                   # 前端代码 (React)
│   └── 📂 src/
│       ├── 📂 api/           # API 调用
│       ├── 📂 views/         # 页面视图
│       │   ├── 📂 dashboard/ # 仪表盘
│       │   ├── 📂 nodes/     # 节点管理
│       │   ├── 📂 subscriptions/ # 订阅管理
│       │   ├── 📂 tags/      # 标签管理
│       │   ├── 📂 templates/ # 模板管理
│       │   └── 📂 settings/  # 系统设置
│       ├── 📂 components/    # 公共组件
│       ├── 📂 contexts/      # React Context
│       ├── 📂 hooks/         # 自定义 Hooks
│       ├── 📂 themes/        # 主题配置
│       └── 📂 layout/        # 布局组件
├── 📂 template/               # 订阅模板文件
├── 📂 docs/                   # 文档
├── main.go                   # 程序入口
├── go.mod                    # Go 依赖管理
├── Dockerfile                # Docker 构建文件
└── README.md                 # 项目说明
```

### 🔧 技术栈

| 层级 | 技术 |
|:---|:---|
| **后端框架** | Go + Gin |
| **ORM** | GORM |
| **数据库** | SQLite |
| **前端框架** | React 18 + Vite |
| **UI 组件库** | Material UI (MUI) |
| **状态管理** | React Context |
| **构建工具** | Vite |

### 💻 本地开发

#### 1. 克隆项目
```bash
git clone https://github.com/ZeroDeng01/sublinkPro.git
cd sublinkPro
```

#### 2. 后端开发
```bash
# 安装 Go 依赖
go mod download

# 运行后端（开发模式）
go run main.go
```

#### 3. 前端开发
```bash
# 进入前端目录
cd webs

# 安装依赖
yarn install

# 启动开发服务器
yarn run run start
```

#### 4. 构建生产版本
```bash
# 构建前端
cd webs && yarn run build

# 构建后端（嵌入前端资源）
go build -o sublinkpro main.go
```

### 📝 开发规范

- **代码风格**：后端遵循 Go 官方规范，前端使用 ESLint + Prettier
- **提交规范**：使用语义化提交信息（feat/fix/docs/refactor 等）
- **分支管理**：`main` 为稳定分支，`dev` 为开发分支
- **API 设计**：RESTful 风格，统一响应格式

### 🔍 关键模块说明

| 模块 | 文件 | 说明 |
|:---|:---|:---|
| 节点测速 | `services/scheduler.go` | 包含延迟测试、速度测试的核心逻辑 |
| 标签规则 | `services/tag_service.go` | 自动标签规则的执行与匹配 |
| 订阅生成 | `api/clients.go` | 订阅链接的生成与节点筛选 |
| 协议解析 | `node/protocol/*.go` | 各种代理协议的解析实现 |
| 数据迁移 | `models/db_migrate.go` | 数据库版本升级迁移脚本 |

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

