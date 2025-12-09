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

## 🎯 项目特色

<table>
<tr>
<td width="50%">

### 🔐 高安全性与自由度
- 支持访问订阅记录
- 简易配置管理
- Token 授权功能

</td>
<td width="50%">

### 🔔 Webhooks 通知
- 支持 PushDeer、Bark
- 钉钉、方糖等
- 订阅更新/测速完成通知

</td>
</tr>
<tr>
<td>

### 📡 多协议支持
| 客户端 | 支持协议 |
|:---|:---|
| **v2ray** | base64 通用格式 |
| **clash** | ss, ssr, trojan, vmess, vless, hy, hy2, tuic, AnyTLS, Socks5 |
| **surge** | ss, trojan, vmess, hy2, tuic |

</td>
<td>

### 🚀 高级功能
- 🔥 自动检测落地IP所属国家
- 🔥 按照国家过滤节点
- 🔥 节点快速重命名
- 📜 JavaScript 脚本订阅操作
- 🛡️ IP 黑/白名单功能
- ⚡ 节点测速功能

</td>
</tr>
</table>

---

## 🔧 快速安装

### 🐳 Docker 运行（推荐）

<details open>
<summary><b>稳定版</b></summary>

```bash
docker run --name sublinke -p 8000:8000 \
  -v $PWD/db:/app/db \
  -v $PWD/template:/app/template \
  -v $PWD/logs:/app/logs \
  -d zerodeng/sublink-pro
```

</details>

<details>
<summary><b>开发版（功能尝鲜）</b></summary>

```bash
docker run --name sublinke -p 8000:8000 \
  -v $PWD/db:/app/db \
  -v $PWD/template:/app/template \
  -v $PWD/logs:/app/logs \
  -d zerodeng/sublink-pro:dev
```

</details>

### 📦 Docker Compose 运行

```yaml
services:
  sublinkx:
    # image: zerodeng/sublink-pro:dev # 开发版（功能尝鲜使用）
    image: zerodeng/sublink-pro # 稳定版
    container_name: sublinkx
    ports:
      - "8000:8000"
    volumes:
      - "./db:/app/db"
      - "./template:/app/template"
      - "./logs:/app/logs"
    restart: unless-stopped
```

### 📝 一键安装脚本

```bash
wget https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/install.sh && sh install.sh
```

> [!TIP]
> 推荐优先使用 **Docker 部署** 以获得最佳兼容性，或可选择 **Debian / Ubuntu** 等发行版。

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

## 📊 项目统计

<div align="center">
  <img src="https://repobeez.abhijithganesh.com/api/insert/ZeroDeng01/sublinkPro" alt="Repobeez" height="0" width="0" style="display: none"/>
  
  ![Star History Chart](https://api.star-history.com/svg?repos=ZeroDeng01/sublinkPro&type=Date)
</div>

---

## 🤝 贡献与支持

如果这个项目对您有帮助，欢迎：

- ⭐ **Star** 这个项目
- 🐛 提交 [Issue](https://github.com/ZeroDeng01/sublinkPro/issues) 反馈问题
- 🔧 提交 Pull Request 贡献代码

---

<div align="center">
  <sub>Made with ❤️ by <a href="https://github.com/ZeroDeng01">ZeroDeng01</a></sub>
</div>
