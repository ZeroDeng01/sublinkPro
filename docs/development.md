# 开发指南

欢迎参与 SublinkPro 的开发！以下是项目结构和开发相关说明。

---

## 📁 项目结构

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
│       │   ├── 📂 hosts/     # Host 映射管理
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

---

## 🔧 技术栈

| 层级 | 技术 |
|:---|:---|
| **后端框架** | Go + Gin |
| **ORM** | GORM |
| **数据库** | SQLite |
| **前端框架** | React 18 + Vite |
| **UI 组件库** | Material UI (MUI) |
| **状态管理** | React Context |
| **构建工具** | Vite |

---

## 💻 本地开发

### 1. 克隆项目

```bash
git clone https://github.com/ZeroDeng01/sublinkPro.git
cd sublinkPro
```

### 2. 后端开发

```bash
# 安装 Go 依赖
go mod download

# 运行后端（开发模式）
go run main.go
```

### 3. 前端开发

```bash
# 进入前端目录
cd webs

# 安装依赖
yarn install

# 启动开发服务器
yarn run start
```

### 4. 构建生产版本

```bash
# 构建前端
cd webs && yarn run build

# 构建后端（嵌入前端资源）
go build -o sublinkpro main.go
```

---

## 📝 开发规范

- **代码风格**：后端遵循 Go 官方规范，前端使用 ESLint + Prettier
- **提交规范**：使用语义化提交信息（feat/fix/docs/refactor 等）
- **分支管理**：`main` 为稳定分支，`dev` 为开发分支
- **API 设计**：RESTful 风格，统一响应格式

---

## 🔍 关键模块说明

| 模块 | 文件 | 说明 |
|:---|:---|:---|
| 节点测速 | `services/scheduler/speedtest_task.go` | 包含延迟测试、速度测试的核心逻辑 |
| 标签规则 | `services/tag_service.go` | 自动标签规则的执行与匹配 |
| 订阅生成 | `api/clients.go` | 订阅链接的生成与节点筛选 |
| 协议解析 | `node/protocol/*.go` | 各种代理协议的解析实现 |
| Host 管理 | `models/host.go` | Host 映射 CRUD、批量操作、缓存管理 |
| DNS 解析 | `services/mihomo/resolver.go` | 自定义 DNS 服务器与代理解析 |
| 数据迁移 | `models/db_migrate.go` | 数据库版本升级迁移脚本 |
| 定时任务 | `services/scheduler/*.go` | 定时任务调度器与任务实现 |

---

## ⏰ 定时任务开发指南

SublinkPro 使用模块化的定时任务调度系统，基于 [robfig/cron](https://github.com/robfig/cron) 库实现。

### 📁 目录结构

```
services/scheduler/
├── manager.go              # 核心调度管理器（SchedulerManager 单例）
├── job_ids.go              # 系统任务ID常量定义
├── subscription_task.go    # 订阅更新任务
├── speedtest_task.go       # 节点测速任务
├── host_cleanup_task.go    # Host过期清理任务
├── reporter.go             # TaskManagerReporter（进度报告）
├── utils.go                # 工具函数
└── bridge.go               # 依赖注入桥接
```

### 🔧 添加新的定时任务

#### 步骤 1：定义任务ID

在 `services/scheduler/job_ids.go` 中添加新的任务ID常量：

```go
const (
    JobIDSpeedTest   = -100  // 节点测速定时任务ID
    JobIDHostCleanup = -101  // Host过期清理任务ID
    JobIDYourTask    = -102  // 你的新任务ID（使用负数区间，避免与用户订阅ID冲突）
)
```

> [!NOTE]
> **任务ID规则**：
> - 系统任务使用负数ID（-100 到 -199 预留区间）
> - 用户订阅任务使用正整数ID（数据库自增主键）

#### 步骤 2：创建任务文件

在 `services/scheduler/` 目录下创建新的任务文件，例如 `your_task.go`：

```go
package scheduler

import (
    "sublink/models"
    "sublink/utils"
)

// StartYourTask 启动你的定时任务
func (sm *SchedulerManager) StartYourTask(cronExpr string) error {
    sm.mutex.Lock()
    defer sm.mutex.Unlock()

    // 清理Cron表达式
    cleanCronExpr := cleanCronExpression(cronExpr)

    // 如果任务已存在，先删除
    if entryID, exists := sm.jobs[JobIDYourTask]; exists {
        sm.cron.Remove(entryID)
        delete(sm.jobs, JobIDYourTask)
    }

    // 添加新任务
    entryID, err := sm.cron.AddFunc(cleanCronExpr, func() {
        ExecuteYourTask()
    })

    if err != nil {
        utils.Error("添加你的任务失败 - Cron: %s, Error: %v", cleanCronExpr, err)
        return err
    }

    sm.jobs[JobIDYourTask] = entryID
    utils.Info("成功添加你的任务 - Cron: %s", cleanCronExpr)
    return nil
}

// StopYourTask 停止你的定时任务
func (sm *SchedulerManager) StopYourTask() {
    sm.mutex.Lock()
    defer sm.mutex.Unlock()

    if entryID, exists := sm.jobs[JobIDYourTask]; exists {
        sm.cron.Remove(entryID)
        delete(sm.jobs, JobIDYourTask)
        utils.Info("成功停止你的任务")
    }
}

// ExecuteYourTask 执行你的任务业务逻辑
func ExecuteYourTask() {
    utils.Info("开始执行你的任务...")
    
    // TODO: 在这里实现你的业务逻辑
    
    utils.Info("你的任务执行完成")
}
```

#### 步骤 3：在启动时加载任务

修改 `services/scheduler/manager.go` 中的 `LoadFromDatabase` 方法，添加任务加载逻辑：

```go
func (sm *SchedulerManager) LoadFromDatabase() error {
    // ... 现有代码 ...
    
    // 启动你的定时任务
    yourTaskEnabled, _ := models.GetSetting("your_task_enabled")
    if yourTaskEnabled == "true" {
        yourTaskCron, _ := models.GetSetting("your_task_cron")
        if err := sm.StartYourTask(yourTaskCron); err != nil {
            utils.Error("创建你的定时任务失败: %v", err)
        }
    }
    
    return nil
}
```

### 📊 带进度报告的任务

如果你的任务需要报告进度（类似测速任务），可以使用 `TaskManager`：

```go
func ExecuteYourTaskWithProgress() {
    // 获取任务管理器
    tm := getTaskManager()
    
    // 创建任务（会在前端任务面板显示）
    task, ctx, err := tm.CreateTask(
        models.TaskTypeYourType,  // 需要在 models/task.go 中定义
        "你的任务名称",
        models.TaskTriggerScheduled,  // 或 TaskTriggerManual
        100,  // 总任务数
    )
    if err != nil {
        utils.Error("创建任务失败: %v", err)
        return
    }
    
    taskID := task.ID
    
    // 执行任务并更新进度
    for i := 1; i <= 100; i++ {
        // 检查是否被用户取消
        select {
        case <-ctx.Done():
            utils.Info("任务被取消")
            return
        default:
        }
        
        // 执行单个子任务...
        
        // 更新进度
        tm.UpdateProgress(taskID, i, "当前处理项", map[string]interface{}{
            "status": "success",
        })
    }
    
    // 任务完成
    tm.CompleteTask(taskID, "任务完成", map[string]interface{}{
        "total": 100,
    })
}
```

---

## 🕐 Cron 表达式格式

本项目使用 5 字段 Cron 格式（不含秒）：

| 字段 | 取值范围 | 说明 |
|:---|:---|:---|
| 分钟 | 0-59 | 每小时的第几分钟 |
| 小时 | 0-23 | 每天的第几小时 |
| 日 | 1-31 | 每月的第几天 |
| 月 | 1-12 | 每年的第几月 |
| 周 | 0-6 | 每周的第几天（0=周日） |

**常用示例**：

| 表达式 | 说明 |
|:---|:---|
| `*/5 * * * *` | 每 5 分钟 |
| `0 */2 * * *` | 每 2 小时 |
| `30 8 * * *` | 每天 8:30 |
| `0 0 * * 0` | 每周日 00:00 |
| `0 2 1 * *` | 每月 1 日 02:00 |

---

## 💡 开发建议

1. **任务幂等性**：确保任务可以安全地重复执行
2. **错误处理**：妥善处理异常，避免影响其他定时任务
3. **日志记录**：使用 `utils.Info/Debug/Error` 记录关键信息
4. **取消支持**：长时间任务应支持用户取消（检查 `ctx.Done()`）
5. **资源释放**：任务结束时确保释放所有资源
