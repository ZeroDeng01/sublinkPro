# SublinkPro AI 技能

[简体中文](README.zh-CN.md) | [English](README.md)

一个面向 [SublinkPro](https://github.com/ZeroDeng01/sublinkPro) 的 AI 操作接口——通过自然语言管理代理节点、订阅与分享。

## 这是什么？

这是一个技能（Skill）。安装后，你无需打开 Web 管理界面，直接用自然语言对 AI 助手说话即可操作你的 SublinkPro 实例。比如"添加这个节点"或"创建一个订阅"，AI 助手会自动调用 REST API 帮你完成。

**它默认是引导式的。** 你不需要知道 API、参数名称，也不需要清楚实例里已有哪些数据。只要说出你想做什么——哪怕只是模糊地说"我想加个节点"——AI 助手都会一步步带你完成：它会先查出你已有的分组/订阅/节点，让你从列表里选；只问必要的信息；在做任何改动前先确认；最后用大白话告诉你结果。不知道能做什么？直接问一句 **"我能做些什么？"** 即可。

它采用可移植、与厂商无关的 `SKILL.md` + 助手脚本格式，任何支持技能的 AI agent 或客户端都能使用（例如 Claude Code，但不限于此）。底层只是调用 SublinkPro 的 REST API，因此凡是能运行该助手脚本的工具都能用它。

**它还能帮你部署 SublinkPro 本身。** 如果你还没有运行中的实例，可以让 AI 助手引导你用 docker、docker-compose 或一键安装脚本进行部署，支持本地或远程主机，全程对话式完成；部署后再引导你登录、改密码、创建 API Key，无缝衔接到日常操作。

它是除 Web 界面之外，使用本系统的**另一种方式**。

## 安装

### 第 0 步 —— 获取 skill

获取这个 `skill-sublinkpro/` 目录有两种方式，按你的情况任选其一：

- **从运行中的 SublinkPro 实例下载（已有实例时最方便）。** 登录 Web 界面，点击顶部工具栏的 **AI 技能包下载** 图标（在捐助图标旁边），即可下载 `skill-sublinkpro.zip`，解压即得到该目录。这份内置副本始终与你的服务端版本一致。
- **从 GitHub 下载（适合还没有 SublinkPro 的纯新手）。** 从仓库下载/克隆 `skill-sublinkpro/` 目录。拿到后你可以**用这个 skill 本身来部署 SublinkPro**（见上文"它还能帮你部署 SublinkPro 本身"）——从零开始时特别有用。

一句话：**还没有实例 → 走 GitHub**（再让 skill 帮你部署）；**已经有实例 → 直接从 Web 界面下载**。

### 第 1 步 —— 安装

1. **复制到你的 AI agent 的技能目录。** 具体路径取决于你使用的工具。以 Claude Code 为例：
   ```bash
   cp -r skill-sublinkpro ~/.claude/skills/sublinkpro
   # 若已克隆仓库，也可用软链接
   ln -s "$(pwd)/skill-sublinkpro" ~/.claude/skills/sublinkpro
   ```
   其他 agent 请放到对应工具加载技能的目录下。

2. **设置环境变量：**
   ```bash
   export SUBLINK_BASE_URL=http://localhost:8000   # 你的 SublinkPro 服务地址
   export SUBLINK_API_KEY=prefix_xxx_yyy           # 你的 API Key
   ```

3. **获取 API Key：**
   - 打开 SublinkPro Web 界面
   - 进入 **设置 → 访问密钥（Access Keys）**
   - 点击 **创建**
   - 复制密钥（仅显示一次）

## 使用示例

你不需要知道有哪些接口、要填哪些参数、系统里已经有哪些数据。直接用大白话说你想做什么，AI 助手会一步步引导你完成。如果你不确定能做什么，直接问 **"我能做些什么？"** 就会得到一份能力清单。

开场白示例：

- **"我想添加一个节点"** → 助手会一步步带你完成
- "添加这个节点：`vless://...`"
- **"帮我新建一个订阅"** → 助手会问你要哪些节点、起什么名字
- "列出我所有的订阅"
- "创建一个分享链接"
- "导入一个机场"
- "用 AI 帮我改改 'ACL4SSR' 模板，屏蔽广告"
- "显示仪表盘统计"

无论是添加/修改节点、新建或修改订阅、导入或修改机场、创建分享、编辑模板，助手都会：先列出系统里现有的分组/订阅/节点供你选择，逐步收集必填信息，并在真正写入前用大白话跟你确认。

## 包含内容

- **`SKILL.md`** — 技能定义（[Agent Skills](https://code.claude.com/docs/en/skills) 开放格式，任何兼容的 AI agent 均可使用）
- **`reference/api.md`** — 完整 API 参考（每个接口都标注了 curl 调用方式）
- **`reference/deploy.md`** — 部署命令目录（docker / compose / 脚本）
- **`scripts/sublink.py`** — 可选的 Python 助手脚本（自动处理鉴权、内容类型、错误校验）；仅在已安装 Python 3 时使用

## 环境要求

- **curl**（必需）——几乎所有 macOS、Windows 10+、Linux 系统自带，是默认调用方式
- 一个运行中的 SublinkPro 实例
- 一个 API Key（在 Web 界面创建）
- Python 3（可选，仅用标准库无需 pip 依赖）——装了才能用便利脚本 `scripts/sublink.py`，不装则用 curl

## 许可证

与 SublinkPro 主项目一致（详见主仓库）。
