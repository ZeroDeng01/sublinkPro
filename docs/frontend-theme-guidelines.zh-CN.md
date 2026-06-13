# 前端主题适配规范

## 概述

凡是前端 UI 改动涉及颜色、表面、阴影、边框、图标、弹窗、面板、预览卡、悬浮层、提示块、标签、富文本或状态色,都必须同时检查 light / dark 两种模式,而不是只在当前模式下"看起来正常"。

## 主题事实来源

- 优先使用 `theme.vars?.palette` 与 `theme.palette` 中现有的 token。
- 表面层优先使用:`background.default`、`background.paper`、`divider`、`text.primary`、`text.secondary`。
- 状态或强调色优先使用现有 palette 语义:`primary` / `secondary` / `info` / `success` / `warning` / `error`。
- 透明度和层级优先通过 `alpha()` 或仓库已有 helper(如 `withAlpha()`)表达,不要硬写一组新的浅深色体系。

## 实施规则

- **外层卡片、弹窗主体、页面主 panel**:优先使用 `background.paper` 语义;暗色模式下允许使用轻量渐变、弱边框和轻微内高光,但不要做成大块发白表面。
- **内层嵌套 panel、底栏、预览 footer、提示块**:优先用 `background.default` 或基于语义色的低透明度染色表面,不要"白底嵌白底"。
- **主值 / 标题 / 可点击重点信息**:优先用 `text.primary` 及字重体现层级;暗色模式可以提高 alpha,但不要默认全改成纯白。
- **次级说明 / 辅助信息 / 百分比 / 提示文案**:优先用 `text.secondary` 或基于 `text.primary` 的较低 alpha,而不是继续叠一个新的灰色常量。
- **图标按钮、header action、通知入口、theme toggle 等高频交互元素**:必须同时检查静态、hover、active/selected、disabled 四种状态在 light / dark 下的可见性。
- **富文本、markdown、代码块、列表和 footer overlay**:不要只修容器背景,必须同步检查正文、列表项、code、link、marker 和底栏文字对比度。

## 明确禁区

- 不要在组件里直接硬编码 `#fff`、`white`、`#000`、`black` 作为大面积表面颜色,除非确实是品牌/图标语义且无法用 token 表达。
- 不要通过"把 dark mode 所有字都调成纯白"来解决可读性问题。
- 不要在一个页面里为同类 surface 写多套互相无关的 `isDark ? ... : ...` magic number。
- 不要只修一个组件实例而放任同类 preview / dialog / footer / chip 在其他文件继续使用旧语义。

## 推荐工作方式

1. 先在相关文件附近寻找现有 helper、复用组件或同类 `sx` 模式,再决定是否新增小型共享 helper。
2. 如果同类问题出现在 2 个以上文件中,优先抽一个小而明确的 helper,而不是复制粘贴一组颜色表达式。
3. 交付说明中必须写明:检查了哪些 light / dark 场景、改了哪些 surface / text / icon 层级、哪些相关层已检查但无需修改。

## 主题覆盖范围

### 影响范围

主题适配不是"修当前看到的那个组件"就结束;只要改动涉及某个页面、模块、共享 helper、theme override 或公共组件的主题语义,就必须把该功能链路中的同类页面、弹窗、抽屉、预览卡、嵌套 panel、底栏、overlay 与移动端入口一并检查。

### 共享基础设施

如果你改的是公共主题基础设施,例如 `webs/src/themes/*`、`webs/src/utils/colorUtils.js`、共享 `sx` helper、通用 dialog/card/panel 样式,那么默认影响范围应视为"所有复用该模式的页面",不能只验证当前需求触达的单一路径。

### 变体覆盖

如果一个模块同时存在桌面版与移动版、dialog 与 fullscreen mobile、drawer 与页面内 panel、卡片列表与详情面板两套入口,主题适配时必须成套覆盖,不能只修桌面或只修主入口。

### 高风险组件

凡是带有 `useMediaQuery`、`fullScreen`、`Drawer`、`Popover`、`Menu`、`Tooltip`、`Collapse`、`Tabs`、`Stepper`、`ReactMarkdown`、`code` / `pre`、自定义 footer overlay、sticky action bar 的组件,都应默认视为主题适配高风险区域。

### 完整结构检查

对话框和弹层必须按完整结构检查:`Dialog` / `Drawer` / `Popover` / `Menu` / `Tooltip` 的 trigger、paper、title、content、actions、backdrop、scroll container、内部嵌套 panel、关闭按钮、底部操作区与 fullscreen mobile 形态都要核对,不允许只改容器背景。

### 参考样板

当某个模块已经在仓库中存在成熟的 dark/light 适配样板时,优先复用其 surface 分层、边框透明度、状态色和移动端处理,而不是重新发明一套局部规则。当前优先参考簇包括:

- 节点预览:`NodePreviewDialog.jsx`、`NodePreviewCard.jsx`、`NodePreviewDetailsPanel.jsx`
- 链式代理:`ChainProxyDialog.jsx`、`ChainPreviewDialog.jsx`、`MobileChainBuilder.jsx`、`ConditionBuilder.jsx`
- 通用 UI:`layout/MainLayout/Header/NotificationSection/index.jsx`、`views/dashboard/Default/index.jsx`
- 机场管理:`views/airports/index.jsx`、`AirportMobileList.jsx`、`AirportFormDialog.jsx`、`AirportBatchEditDialog.jsx`、`components/CronExpressionGenerator.jsx`

## 浅色迁移到深色时的高风险遗漏点

- 只改外层页面容器,没有同步处理内层 header strip、footer strip、nested panel、统计块、详情块、空状态、skeleton 或 preview footer。
- 只改文本颜色,没有同步处理边框、阴影、divider、hover、active、selected、disabled、focus-visible 与滚动容器边界。
- 只修主页面,没有同步修同模块的 dialog、drawer、popover、menu、tooltip、fullscreen mobile dialog、底部菜单或二级详情入口。
- 只修桌面端,没有同步检查移动端卡片、响应式折叠布局、safe-area、sticky footer、横向滚动、点击热区与 icon button 可见性。
- 只修容器,没有同步修 `ReactMarkdown`、`code`、`pre`、列表 marker、链接、表格、chip、badge、tag、状态提示块与通知浮层。
- 状态色仍沿用浅色主题的 tint,导致 success / warning / error / info 在暗色模式下背景不够区分、文字发灰、边框发脏。
- 图标、次级文字、百分比、caption、placeholder、helper text 与 disabled 文案没有跟随正文层级一起调整,导致"主标题能看,辅文全部发灰"。
- DataGrid、Autocomplete、Select、ToggleButtonGroup、Tabs、列表 hover 行、搜索框、过滤条和表单控件只修了默认态,没有连同 popup、选中态、聚焦态一起检查。

## 深色主题迁移修改原则

- 优先复用仓库既有模式:`const palette = theme.vars?.palette || theme.palette;`,再抽局部的 `dialogSurface`、`mutedPanelSurface`、`nestedPanelSurface`、`panelBorder` 等语义变量,避免在 JSX 中散落大量 `isDark ? ... : ...`。
- 深色模式优先做"语义表面分层",不是简单把浅色背景替换成黑色。外层通常使用 `background.paper`,内层面板、页脚、提示块与弱化区域优先使用 `background.default` 或基于语义色的低透明度染色。
- 边框、分隔线和层级优先通过 `divider`、`alpha()`、`withAlpha()` 与轻量 inset highlight 表达,不要在每个组件里手工发明一套新的灰阶。
- 主文本、强调文本、次级文本、占位文案、禁用文案、图标与状态色必须保持层级关系稳定;不要用"全部调亮"掩盖层级混乱。
- 状态色优先保持"语义文字 + 低透明度背景 + 轻量边框"的组合,而不是在暗色模式下直接铺满高饱和纯色块。
- 当同类问题在多个文件重复出现时,优先抽小型共享 helper 或复用现有 override;不要把相同的 dark-mode magic number 复制到多个模块。

## 主题迁移执行手册

当你在本仓库做主题适配时,默认按下面顺序执行,不要跳步:

1. **先界定影响范围**:当前改动触及的是页面、模块、共享组件、theme override,还是全局 theme helper;如果是后两者,默认把所有复用该模式的页面都列入检查范围。

2. **先找仓库内现成样板,再写代码**:优先参考 `NodePreview*`、`Chain*`、`MobileChainBuilder.jsx`、`ConditionBuilder.jsx`、`NotificationSection`、`dashboard/Default`,以及机场管理链路中的 `airports/index.jsx`、`AirportMobileList.jsx`、`AirportFormDialog.jsx`、`AirportBatchEditDialog.jsx`、`CronExpressionGenerator.jsx`,先复用它们的 surface 分层和状态表达,再决定是否抽新的 helper。

3. **先修语义分层,再修局部颜色**:优先处理外层 surface、内层 panel、footer/header strip、边框、divider、文本层级、状态色和 icon,再处理局部渐变、阴影或装饰。

4. **改主路径时必须同步检查分支路径**:桌面端改完后,继续检查移动端、fullscreen dialog、drawer、popover、menu、tooltip、collapse、empty state、skeleton、footer overlay、markdown/code/pre、列表 hover 与表单 focus。

5. **交付前必须写清覆盖面**:说明本次适配覆盖了哪些页面/模块/弹层/移动端入口,复用了哪些仓库现有样板,哪些同类实例已检查但无需修改。

如果你没有时间把一个模块的桌面端、移动端、dialog/drawer、overlay、details panel 一起检查完,那就不要宣称该模块"已完成主题适配";应明确标记为部分覆盖。

## 主题改动完成条件

- 至少核对 light / dark 两种模式。
- 至少核对相关 hover / active / disabled 或展开态。
- 不允许出现"容器变暗了,但内层 panel、footer、markdown、percent、icon 仍然发白或发灰不可读"的交付状态。
- 不允许把主题适配交付成"仅当前页面修好、同模块弹窗/移动端/浮层仍然沿用旧语义"的半完成状态。
- 只要本次改动触及共享主题基础设施、公共组件或模块级视觉语义,交付说明中必须明确列出已覆盖的页面、弹层、移动端入口和已检查但无需修改的相关实例。
