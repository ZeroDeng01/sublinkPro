[English](internationalization.md) | 简体中文

# 国际化实施契约

本文档定义 SublinkPro 国际化 Wave 1 的实施契约，供后续贡献者在现有 React/Vite/MUI 前端和 Go/Gin 后端中接入 i18n 时使用。

---

## 范围

- 首批支持语言为 `zh-CN` 和 `en-US`。
- 当前 UI 大量文本以 `zh-CN` 为源语言。`en-US` 必须完整覆盖后，首轮国际化才算可用。
- 新增前端可见文本时，必须在同一次变更中补齐首批两种语言。
- 后续新增语言时，必须沿用本文定义的命名空间布局、语言检测规则、MUI 映射和格式化 helper。

SublinkPro 不使用 URL path 或 subdomain 做语言路由。`SUBLINK_WEB_BASE_PATH` 已经是 Web UI 的路由关注点，语言选择不能再引入新的路径前缀或基于域名的路由层。

---

## 前端契约

前端 i18n 应放在 `webs/src/` 下。翻译资源要跟随前端应用维护，不要放到后端目录或运行时数据目录中。

按功能或稳定 UI 边界划分 namespace。建议首批布局如下：

```text
webs/src/i18n/
├── index.js
├── locales.js
└── locales/
    ├── en-US.json
    └── zh-CN.json
```

Wave 1 翻译资源采用每个 locale 一个 JSON 文件。如果后续翻译文件变得过大，应在明确的后续变更中按 namespace 拆分，并同步更新本文档契约。

Namespace 规则：

- `common` 用于共享操作、状态、校验、空状态和通用标签。
- 功能 namespace 与页面或模块名保持一致，例如 `airports`、`subscriptions`、`dashboard`、`settings` 和 `tasks`。
- 除非组件会跨多个无关功能复用，不要按单个组件拆 namespace。
- 文案变化时尽量保持 key 稳定。key 表达语义，不表达当前措辞。

Key 命名规则：

- 使用 lower camel case 的点路径，例如 `actions.save`、`fields.username` 和 `errors.loginFailed`。
- 优先使用语义化 key，不要把英文文案本身当作 key。
- 插值使用具名变量，例如 `taskCount: "{{count}} tasks"`。
- 数量变化使用复数能力，不要手动拼接字符串。
- 避免在 JSX 中拼接多个翻译片段。短语和句子应作为完整文本翻译。

---

## 语言状态

语言检测顺序如下：

1. local storage 中保存的用户偏好。
2. 浏览器语言，且该语言能匹配已支持 locale 或已支持基础语言。
3. 默认回退到 `zh-CN`。

用户显式选择的语言应写入 local storage。该选择需要在同一浏览器中跨刷新、退出登录和重新登录保留。Wave 1 不把语言写入后端运行时目录或服务端 session 状态。

匹配浏览器语言时，只做保守别名归一化：

- `zh`、`zh-CN`、`zh-Hans` 映射为 `zh-CN`。
- `en`、`en-US`、`en-GB` 和其他英语变体映射为 `en-US`。

不支持的语言在补齐翻译资源和 MUI locale 映射前，统一回退到 `zh-CN`。

---

## MUI Locale 与格式化

MUI locale 配置必须由当前应用语言决定：

| App locale | MUI locale |
|:---|:---|
| `zh-CN` | `zhCN` |
| `en-US` | `enUS` |

该映射应集中在一个前端 helper 中，例如 `webs/src/i18n/muiLocale.js`，方便后续新增语言时只改一个入口。

日期、时间、相对时间、数字、百分比、字节和流量值都必须通过接收当前 locale 的共享 helper 格式化。功能组件中不要硬编码中文或英文日期顺序。

格式化规则：

- 尽量使用 `Intl.DateTimeFormat` 和 `Intl.NumberFormat`。
- 协议值、节点名称、域名、文件路径、API key 和模板语法保持原样。
- 可见 UI 中的单位文本应可翻译。
- 日志和调试标识保留机器可读值，只在需要时翻译周边标签。

---

## 后端契约

后端需要保持现有 API 客户端兼容，同时支持前端本地化展示。

已经暴露 `msg` 或 `code` 的 API 响应必须保持字段稳定，除非该接口契约被有意修改并同步写入文档。新增或改造响应时，可以增加 `i18nKey` 和可选插值数据，让前端渲染本地化文本。

面向用户的错误或任务结果推荐响应结构：

```json
{
  "code": 400,
  "msg": "Invalid request",
  "i18nKey": "errors.invalidRequest",
  "i18nParams": {
    "field": "name"
  }
}
```

兼容规则：

- `msg` 保留为旧客户端和诊断场景的回退文本。
- `code` 保留为程序化状态或业务码。
- `i18nKey` 初期可选，随着覆盖率提升，可逐步成为前端展示的优先来源。
- `i18nParams` 只能包含普通 JSON 值。
- 不要翻译协议输出、订阅内容、模板变量、机器日志、ID、token 或外部客户端会消费的值。

当后端本身就是产品展示面时，后端仍可拥有人工可读文本，例如 CLI help、启动错误、日志、webhook payload、Telegram Bot 回复、生成的任务消息和导出文件。若这些文本会显示在 Web UI 中，并且 UI 需要本地化，优先使用 `i18nKey` 加 params。

---

## 新增语言

新增未来语言时：

1. 将 locale code 加入支持列表。
2. 在 `webs/src/i18n/locales/<locale>.json` 增加完整资源文件，并保持与 `zh-CN`、`en-US` 相同的 key 结构。
3. 按需补充浏览器语言归一化规则。
4. 补充 MUI locale 映射。
5. 确认该语言下日期、时间、数字、百分比和流量格式正确。
6. 如果行为或贡献者指南发生变化，同步更新英文和简体中文文档。

除非首批 namespace 中所有用户可见字符串已有翻译，或已经写明分阶段计划，否则不要新增该语言。

---

## 文档规则

i18n 相关文档必须成对维护英文和简体中文版本。如果修改 `docs/internationalization.md`，必须在同一次变更中同步修改 `docs/internationalization.zh-CN.md`，并保持语言切换链接有效。

后续 README 或功能文档如果涉及国际化，也遵循同一规则：英文 canonical 文档和对应 `*.zh-CN.md` 文件必须保持一致。
