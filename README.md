<div align="center">
<img src="webs/src/assets/images/logo.svg" width="150px" height="150px" />
</div>

<div align="center">
  <img src="https://img.shields.io/badge/Go-1.24.3-green.svg"/>
  <img src="https://img.shields.io/badge/license-MIT-green.svg"/>



</div>

# 项目简介

`sublinkPro` 是基于优秀的开源项目  [sublinkX](https://github.com/gooaclok819/sublinkX) /[sublinkE](https://github.com/eun1e/sublinkE)  进行二次开发，在原项目基础上做了部分定制优化。感谢原作者的付出与贡献。

**⚠️本项目和原项目数据库不兼容，请不要混用。**

**⚠️请不要使用本项目以及任何本项目的衍生项目进行违反您以及您所服务用户的所在地法律法规的活动。本项目仅供个人开发和学习交流使用。**

- 前端基于 [Berry Free React Material UI Admin Template](https://github.com/codedthemes/berry-free-react-admin-template)；
- 后端采用 Go + Gin + Gorm；
- 默认账号：admin 密码：123456，请安装后务必自行修改；

# 修改内容


- [x] 修复部分页面BUG
- [x] 支持 Clash `dialer-proxy` 属性
- [x] 允许添加并使用 API KEY 访问 API
- [x] 导入、定时更新订阅链接中的节点（可以通过前置代理订阅）
- [x] 支持AnyTLS、Socks5协议
- [x] 订阅节点排序
- [x] 支持订阅的IP黑/白名单功能
- [x] 支持节点测速功能
- [x] 支持按照测速结果作为条件筛选返回的节点
- [x] 支持javascript脚本进行订阅操作
- [x] 支持webhooks通知，比如pushdeer、Bark、钉钉、方糖等在订阅更新或者测速结束的时候进行通知
- [ ] ...

# 项目特色

- 高自由度与安全性，支持访问订阅记录及简易配置管理；
- 支持多种客户端协议及格式，包括：
    - v2ray（base64 通用格式）
    - clash（支持 ss, ssr, trojan, vmess, vless, hy, hy2, tuic, AnyTLS, Socks5）
    - surge（支持 ss, trojan, vmess, hy2, tuic）
- 新增 token 授权及订阅导入功能，增强安全性和便捷性。

# 安装说明

## Docker 运行
### 稳定版
```bash
docker run --name sublinke -p 8000:8000 \
-v $PWD/db:/app/db \
-v $PWD/template:/app/template \
-v $PWD/logs:/app/logs \
-d zerodeng/sublink-pro
```
### 开发版（功能尝鲜使用）
```bash
docker run --name sublinke -p 8000:8000 \
-v $PWD/db:/app/db \
-v $PWD/template:/app/template \
-v $PWD/logs:/app/logs \
-d zerodeng/sublink-pro:dev
```
## Docker Compose 运行
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

## 一键安装
```bash
wget https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/refs/heads/main/install.sh   && sh install.sh
```

> ⚠ **注意**
> 推荐优先使用 **Docker 部署** 以获得最佳兼容性，或可选择 **Debian / Ubuntu** 等发行版。



# 项目预览

![预览1](docs/images/1.jpg)
![预览2](docs/images/2.jpg)
![预览3](docs/images/3.jpg)
![预览4](docs/images/4.jpg)
![预览5](docs/images/5.jpg)
![预览6](docs/images/6.jpg)
![预览7](docs/images/7.jpg)
![预览8](docs/images/8.jpg)
![预览9](docs/images/9.jpg)
![预览10](docs/images/10.jpg)

# 脚本功能说明

SublinkPro 支持使用 JavaScript 脚本对订阅内容进行自定义处理。脚本可以包含以下两个主要函数：

## 1. 节点过滤 (filterNode)

`filterNode` 函数在生成订阅内容之前执行，用于对节点列表进行过滤或修改。

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

## 2. 内容后处理 (subMod)

`subMod` 函数在生成最终订阅内容之后执行，用于对最终的文本内容进行修改。

**函数签名:**
```javascript
function subMod( input, clientType) {
    // input: 原始输入内容
    // clientType: 客户端类型
    // 返回值: 修改后的内容字符串
    return input; // 注意：此处示例仅为示意，实际应返回处理后的字符串
}
```

**脚本支持的函数请查看【[脚本文档](docs/script_support.md)】**

**注意:**
- 脚本中可以使用 `console.log()` 输出日志到后台。
- 多个脚本会按照排序顺序依次执行。
