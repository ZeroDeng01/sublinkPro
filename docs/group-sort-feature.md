# 分组管理功能 — 需求与实现文档

## 1. 需求背景

当前系统中，订阅展开分组节点时（`models/subcription.go` 的 `GetSub` 方法），固定使用 `ORDER BY nodes.id ASC` 排序。这导致同一分组内来自不同机场的节点无法按用户期望的机场优先级排列。

**用户期望**：能对分组内各机场的排列顺序进行管理。例如指定"免费分组"中机场C的节点排在最前面，然后是机场A，最后是机场B。

**核心目标**：新增"分组管理"功能模块，允许用户为每个节点分组配置机场排序权重，订阅输出时按此权重重排节点。

## 2. 技术方案概述

### 2.1 数据模型

新建 `GroupAirportSort` 表，存储每个分组内各机场的排序权重：

```
┌──────────────────────────────────────────────────────────┐
│ group_airport_sorts                                      │
├──────────┬────────────┬──────────────────────────────────┤
│ id (PK)  │ int        │ 自增主键                         │
│ group_name│ string(200)│ 分组名称（联合唯一索引）          │
│ airport_id│ int        │ 机场ID / 0=手动添加（联合唯一索引）│
│ sort      │ int        │ 排序权重，值越小越靠前            │
└──────────┴────────────┴──────────────────────────────────┘
联合唯一索引: idx_group_airport (group_name, airport_id)
```

### 2.2 缓存策略

使用项目现有的 `cache.MapCache[int, GroupAirportSort]` 泛型缓存，以 `ID` 为主键，添加 `groupName` 二级索引。参照 `models/tag.go` 的缓存模式。

### 2.3 数据流

```
用户拖拽排序 → 前端 POST /api/v1/group-sort/save
            → SaveGroupAirportSorts() 事务写库 + 更新缓存
            → GetSub() 调用 GetGroupAirportSortMap() 从缓存获取映射
            → sort.SliceStable 按权重重排节点
            → 订阅输出
```

### 2.4 API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/group-sort/groups` | 获取所有分组概要信息（名称、节点数、机场数、是否已配置排序） |
| GET | `/api/v1/group-sort/detail?group=分组名` | 获取分组详情（机场列表含名称、节点数、当前排序） |
| POST | `/api/v1/group-sort/save` | 保存分组内机场排序配置 |

所有接口均需 `AuthToken` 鉴权。

**detail 响应示例**：
```json
{
  "groupName": "免费分组",
  "airports": [
    { "airportId": 3, "airportName": "机场C", "nodeCount": 10, "sort": 0 },
    { "airportId": 1, "airportName": "机场A", "nodeCount": 5,  "sort": 1 },
    { "airportId": 0, "airportName": "手动添加", "nodeCount": 2, "sort": 2 }
  ]
}
```

**save 请求示例**：
```json
{
  "groupName": "免费分组",
  "airportSorts": [
    { "airportId": 3, "sort": 0 },
    { "airportId": 1, "sort": 1 },
    { "airportId": 0, "sort": 2 }
  ]
}
```

## 3. 文件清单

### 3.1 新建文件（5个，共 627 行）

| 文件 | 行数 | 用途 |
|------|------|------|
| `models/group_airport_sort.go` | 254 | 数据模型、缓存 init、CRUD、聚合查询 |
| `api/group_sort.go` | 52 | HTTP Handler（groups / detail / save） |
| `routers/group_sort.go` | 18 | 路由注册 |
| `webs/src/api/groupSort.js` | 16 | 前端 API 封装 |
| `webs/src/views/group-sort/index.jsx` | 287 | 分组管理页面（左右布局 + 拖拽排序） |

### 3.2 修改文件（5个）

| 文件 | 改动行数 | 修改内容 |
|------|---------|---------|
| `models/db_migrate.go` | +5 | 在 Airport AutoMigrate 之后添加 GroupAirportSort 表迁移 |
| `models/subcription.go` | +20 | GetSub 分组展开后，按机场排序权重 `sort.SliceStable` 重排节点 |
| `main.go` | +4 | 添加 `InitGroupAirportSortCache()` + `routers.GroupSort(r)` |
| `webs/src/menu-items/subscription.js` | +11 | 在"节点检测"和"订阅列表"之间插入"分组管理"菜单项（IconCategory） |
| `webs/src/routes/MainRoutes.jsx` | +5 | 添加 `group-sort` 路由和懒加载组件 |

## 4. 核心实现详解

### 4.1 数据模型层 — `models/group_airport_sort.go`

**缓存初始化**（参照 `tag.go:52-77`）：
```go
var groupAirportSortCache *cache.MapCache[int, GroupAirportSort]

func init() {
    groupAirportSortCache = cache.NewMapCache(func(g GroupAirportSort) int { return g.ID })
    groupAirportSortCache.AddIndex("groupName", func(g GroupAirportSort) string { return g.GroupName })
}
```

**GetGroupDetail** — 聚合逻辑：
1. 从 `nodeCache` 按 `group` 二级索引获取该分组所有节点
2. 按 `SourceID` 聚合节点数量
3. 从 `groupAirportSortCache` 按 `groupName` 索引获取已存排序
4. 从 `airportCache` 获取机场名称（SourceID=0 显示"手动添加"）
5. 按 sort 排序后重新编号返回

**SaveGroupAirportSorts** — 写入逻辑：
1. `database.WithTransaction` 内删除旧记录 + 批量插入新记录
2. 事务提交后更新缓存（删旧 → 从 DB 重载该分组记录）

**GetGroupAirportSortMap** — 供 GetSub 调用：
- 从缓存获取，返回 `map[int]int`（airportID → sortWeight）
- 无配置时返回 `nil`

### 4.2 订阅排序逻辑 — `models/subcription.go:541-558`

在现有 `ORDER BY nodes.id ASC` 数据库查询之后，增加内存重排：

```go
airportSortMap := GetGroupAirportSortMap(group.GroupName)
if len(airportSortMap) > 0 {
    sort.SliceStable(groupNodes, func(i, j int) bool {
        sortI, okI := airportSortMap[groupNodes[i].SourceID]
        sortJ, okJ := airportSortMap[groupNodes[j].SourceID]
        if !okI { sortI = 999999 }
        if !okJ { sortJ = 999999 }
        if sortI != sortJ {
            return sortI < sortJ
        }
        return groupNodes[i].ID < groupNodes[j].ID
    })
}
```

**设计要点**：
- `sort.SliceStable` 保持同一机场内节点的原有 ID 顺序
- 未配置排序的机场赋予 `999999` 权重排到最后
- `len(airportSortMap) > 0` 守卫确保无配置时零开销
- 从缓存读取映射，无额外数据库查询

### 4.3 前端页面 — `webs/src/views/group-sort/index.jsx`

**布局**：Grid 左右两栏
- 左侧（md=4）：分组列表，带搜索过滤，已配置排序的分组显示 ✓ 图标
- 右侧（md=8）：选中分组的机场拖拽排序区域

**拖拽**：使用 `@hello-pangea/dnd`（项目中已有 7 处使用此库），标准三层结构 `DragDropContext > Droppable > Draggable`。

**数据流**：
1. 页面加载 → `getGroupSortGroups()` → 渲染左侧列表
2. 点击分组 → `getGroupSortDetail(groupName)` → 渲染右侧机场列表
3. 拖拽调整顺序 → 本地 state 更新
4. 点击"保存排序" → `saveGroupAirportSort()` → 提示成功 → 刷新左侧列表状态

## 5. 向后兼容性

| 场景 | 行为 |
|------|------|
| 分组无排序配置 | `GetGroupAirportSortMap` 返回 `nil`，`len == 0` 跳过排序，节点按原 ID 升序 |
| 新增机场节点后未更新排序 | 新机场的 SourceID 不在 sortMap 中，赋 `999999` 排到最后 |
| 机场被删除 | 详情页显示"机场#N(已删除)"，排序配置仍生效 |
| 保存空排序列表 | 事务删除该分组所有排序记录，等同于重置为默认顺序 |

## 6. 初始化顺序依赖

```
InitNodeCache()          ← GetGroupDetail 依赖 nodeCache
InitAirportCache()       ← GetGroupDetail 依赖 airportCache
InitGroupAirportSortCache()  ← 在以上两者之后
```

`main.go` 中已按此顺序排列。

## 7. 走查结果

### 已发现并修复的问题

1. **前端 `Divider` 未使用导入** — `index.jsx` 导入了 `Divider` 但未使用，已删除
2. **后端 4 个死代码函数** — `GetGroupAirportSortCount`、`GetGroupsWithSortConfig`、`GetGroupAirportCount`、`GetGroupNodeCount` 定义后无调用，已删除

### 功能闭环验证

| 链路节点 | 状态 |
|---------|------|
| 数据表创建（AutoMigrate） | ✅ |
| 缓存初始化（在 Airport 之后） | ✅ |
| 路由注册（AuthToken 鉴权） | ✅ |
| API → Model 调用链 | ✅ |
| 前端 API → 后端路径匹配 | ✅ |
| 前端路由 + 菜单项 | ✅ |
| GetSub 排序逻辑注入 | ✅ |
| 向后兼容（无配置 = 无变化） | ✅ |
| Go 编译通过 | ✅ |
| 前端 vite build 通过 | ✅ |

## 8. 验证方案

1. **编译验证**：`go build ./...` 通过
2. **前端构建**：`vite build` 通过
3. **功能验证**：
   - 启动服务，确认侧边栏"订阅管理"下新增"分组管理"菜单项
   - 进入分组管理页面，确认左侧显示所有分组列表
   - 点击某分组，确认右侧显示该分组下的机场列表（含名称和节点数量）
   - 拖拽调整机场顺序后点击保存，确认保存成功
   - 访问已关联该分组的订阅输出，确认节点按配置的机场顺序排列
4. **向后兼容**：未配置排序的分组，节点顺序与修改前完全一致
