[English](practical-recipes.md) | 简体中文

# 实用开发模式

SublinkPro 中常见开发模式和频繁任务的分步指南。

---

## 📋 快速参考

| 任务 | 检查文件 | 文档 |
|---|---|---|
| 添加后端功能 | `api/`, `services/`, `models/`, `routers/` | 本文件 |
| 添加/修改定时任务 | `services/scheduler/` | 本文件 |
| 修改 mihomo 行为 | `services/mihomo/` | 本文件 |
| 修改部署 | `Dockerfile`, `docker-compose.yml`, docs | `build-and-deployment.md` |
| 添加前端功能 | `webs/src/views/`, `webs/src/api/` | `development.md` |
| 修改主题 | `webs/src/themes/`, components | `frontend-theme-guidelines.md` |

---

## 🔧 模式 1：添加后端功能

### 典型流程

完整的后端功能通常涉及这些层：

1. **模型** (`models/`) - 如果需要持久化
2. **服务** (`services/`) - 业务逻辑
3. **API** (`api/`) - HTTP 处理器
4. **路由** (`routers/`) - 路由注册
5. **前端** (`webs/src/api/` + `webs/src/views/`) - 如果需要 UI

### 分步示例：添加"标签"功能

#### 步骤 1：定义模型

**文件**：`models/tag.go`

```go
package models

import "gorm.io/gorm"

type Tag struct {
    ID          uint   `gorm:"primaryKey"`
    Name        string `gorm:"uniqueIndex;not null"`
    Color       string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func (Tag) TableName() string {
    return "tags"
}
```

**在 `models/db_migrate.go` 中添加迁移**：

```go
func RunMigrations(db *gorm.DB) error {
    return db.AutoMigrate(
        // ... 现有模型
        &Tag{},
    )
}
```

#### 步骤 2：创建服务

**文件**：`services/tag_service.go`

```go
package services

import (
    "sublinkPro/models"
    "gorm.io/gorm"
)

type TagService struct {
    db *gorm.DB
}

func NewTagService(db *gorm.DB) *TagService {
    return &TagService{db: db}
}

func (s *TagService) ListTags() ([]models.Tag, error) {
    var tags []models.Tag
    err := s.db.Find(&tags).Error
    return tags, err
}

func (s *TagService) CreateTag(tag *models.Tag) error {
    return s.db.Create(tag).Error
}

// ... 更多方法
```

#### 步骤 3：创建 API 处理器

**文件**：`api/tag.go`

```go
package api

import (
    "net/http"
    "sublinkPro/services"
    "github.com/gin-gonic/gin"
)

type TagAPI struct {
    tagService *services.TagService
}

func NewTagAPI(tagService *services.TagService) *TagAPI {
    return &TagAPI{tagService: tagService}
}

func (a *TagAPI) ListTags(c *gin.Context) {
    tags, err := a.tagService.ListTags()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "code": 500,
            "msg":  "获取标签列表失败",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "code": 200,
        "data": tags,
    })
}

// ... 更多处理器
```

#### 步骤 4：注册路由

**文件**：`routers/router.go`

```go
func SetupRouter(/* ... 依赖 */) *gin.Engine {
    // ... 现有设置
    
    tagAPI := api.NewTagAPI(tagService)
    
    apiV1 := r.Group("/api/v1")
    {
        // 受保护的路由
        apiV1.Use(middlewares.AuthToken())
        {
            apiV1.GET("/tags", tagAPI.ListTags)
            apiV1.POST("/tags", tagAPI.CreateTag)
            // ... 更多路由
        }
    }
    
    return r
}
```

#### 步骤 5：添加前端支持（可选）

**文件**：`webs/src/api/tags.js`

```javascript
import request from './request';

export const getTags = () => {
  return request.get('/api/v1/tags');
};

export const createTag = (data) => {
  return request.post('/api/v1/tags', data);
};
```

**文件**：`webs/src/views/tags/index.jsx`

```jsx
import { getTags } from '@/api/tags';

export default function TagsView() {
  // ... 实现
}
```

#### 步骤 6：跨层同步

当添加带前端的后端功能时：

1. ✅ 将 API 端点文档添加到 `skill-sublinkpro/reference/api.md`
2. ✅ 如果添加了用户可见文本，更新 i18n
3. ✅ 为业务逻辑添加测试
4. ✅ 如果重要，在 `docs/features/` 中记录

### 验证

```bash
# 后端
gofmt -w models/tag.go services/tag_service.go api/tag.go
golangci-lint run
go test ./...

# 前端（如果添加）
cd webs
yarn run lint
yarn run build
```

---

## ⏰ 模式 2：添加或修改定时任务

### 理解调度器

**关键文件**：
- `services/scheduler/manager.go` - 调度器管理
- `services/scheduler/job_ids.go` - 作业 ID 常量
- `services/scheduler/speedtest_task.go` - 示例任务

### 任务类型

1. **系统任务**：负数 ID（例如 `-1`、`-2`）
2. **用户任务**：来自数据库的正数 ID

### 分步示例：添加清理任务

#### 步骤 1：定义作业 ID

**文件**：`services/scheduler/job_ids.go`

```go
const (
    JobIDSpeedTest  = -1
    JobIDSubscriptionUpdate = -2
    JobIDCleanup    = -3  // 新系统任务
)
```

#### 步骤 2：实现任务逻辑

**文件**：`services/scheduler/cleanup_task.go`

```go
package scheduler

import (
    "log"
    "time"
    "gorm.io/gorm"
)

type CleanupTask struct {
    db *gorm.DB
}

func NewCleanupTask(db *gorm.DB) *CleanupTask {
    return &CleanupTask{db: db}
}

func (t *CleanupTask) Run() error {
    log.Println("开始清理任务...")
    
    // 清理逻辑
    cutoff := time.Now().AddDate(0, 0, -30) // 30天前
    
    result := t.db.Where("created_at < ?", cutoff).
        Delete(&models.LogEntry{})
    
    if result.Error != nil {
        return result.Error
    }
    
    log.Printf("清理完成：删除了 %d 行", result.RowsAffected)
    return nil
}
```

#### 步骤 3：在调度器中注册任务

**文件**：`services/scheduler/manager.go`

```go
func (m *Manager) RegisterSystemTasks() {
    // ... 现有任务
    
    // 注册清理任务
    cleanupTask := NewCleanupTask(m.db)
    m.AddSystemTask(
        JobIDCleanup,
        "0 2 * * *",  // 每天凌晨2点运行
        func() {
            if err := cleanupTask.Run(); err != nil {
                log.Printf("清理任务失败: %v", err)
            }
        },
    )
}
```

#### 步骤 4：测试任务

```go
// 在 cleanup_task_test.go 中
func TestCleanupTask(t *testing.T) {
    db := setupTestDB(t)
    task := NewCleanupTask(db)
    
    // 添加测试数据
    // ...
    
    err := task.Run()
    assert.NoError(t, err)
    
    // 验证清理
    // ...
}
```

### Cron 格式参考

```
┌───────────── 分钟 (0 - 59)
│ ┌───────────── 小时 (0 - 23)
│ │ ┌───────────── 月中的某天 (1 - 31)
│ │ │ ┌───────────── 月份 (1 - 12)
│ │ │ │ ┌───────────── 星期几 (0 - 6)（周日到周六）
│ │ │ │ │
* * * * *
```

**示例**：
- `0 2 * * *` - 每天凌晨2点
- `*/15 * * * *` - 每15分钟
- `0 */6 * * *` - 每6小时
- `0 0 * * 0` - 每周日午夜

### 验证

```bash
# 测试任务
go test ./services/scheduler/ -run TestCleanupTask

# 测试 cron 解析
go test ./services/scheduler/ -v
```

---

## 🌐 模式 3：修改 mihomo 相关行为

### 理解 mihomo 集成

mihomo 是核心网络/代理引擎。这里的更改会影响：
- 速度测试
- 代理适配器创建
- DNS 解析
- Host 映射
- 代理下载

**需要一起检查的关键文件**：
- `services/mihomo/mihomo.go` - 核心封装
- `services/mihomo/dns_resolver.go` - DNS 逻辑
- `services/mihomo/host_resolver.go` - Host 映射
- `services/scheduler/speedtest_task.go` - 速度测试集成
- `utils/proxy_client.go` - 代理 HTTP 客户端
- `node/sub.go`, `node/usage.go` - 代理下载

### 示例：修改速度测试行为

#### 如果更改测试逻辑

**文件**：`services/mihomo/mihomo.go`

```go
func (m *MihomoService) TestNodeDelay(
    node *models.Node,
    timeout time.Duration,
) (uint16, error) {
    // 你修改的逻辑在这里
}
```

**然后检查**：`services/scheduler/speedtest_task.go`

确保调度器任务仍然能够使用你的更改。

#### 如果更改 DNS 行为

**文件**：`services/mihomo/dns_resolver.go`

```go
func (r *DNSResolver) ResolveWithProxy(
    domain string,
    proxyURL string,
) ([]string, error) {
    // 你修改的逻辑在这里
}
```

**然后检查**：
- `api/host.go` - Host 管理端点
- `models/host.go` - Host 模型
- `services/mihomo/host_resolver.go` - Host 同步

#### 如果更改代理客户端行为

**文件**：`utils/proxy_client.go`

```go
func NewProxyClient(proxyURL string) (*http.Client, error) {
    // 你修改的逻辑在这里
}
```

**然后检查**：
- `node/sub.go` - 订阅下载
- `node/usage.go` - 使用量查询
- 任何使用代理 HTTP 请求的服务

### 交叉检查模式

对于 mihomo 更改，始终验证：

1. ✅ 核心逻辑 (`services/mihomo/`)
2. ✅ 定时任务 (`services/scheduler/speedtest_task.go`)
3. ✅ HTTP 客户端 (`utils/proxy_client.go`)
4. ✅ 相关 API 端点 (`api/host.go`, `api/node.go`)
5. ✅ 相关模型 (`models/host.go`, `models/node.go`)
6. ✅ 功能文档 (`docs/features/speedtest.md` 等)

### 验证

```bash
# 测试 mihomo 集成
go test ./services/mihomo/ -v

# 测试调度器集成
go test ./services/scheduler/ -v -run TestSpeedTest

# 测试完整流程
go run main.go
# 通过 UI 手动触发速度测试
```

---

## 🚀 模式 4：修改部署或构建行为

### 何时使用此模式

- 修改 `Dockerfile`
- 更改前端资源嵌入
- 更新 `docker-compose.yml`
- 更改生产构建流程
- 修改 base path 行为

### 需要一起检查的文件

1. `Dockerfile` - Docker 构建
2. `embed_prod.go` - 资源嵌入
3. `.github/workflows/build-release.yml` - CI 构建
4. `main.go` - 启动和服务
5. `webs/vite.config.js` - 前端构建配置
6. `docs/build-and-deployment.md` - 构建文档

### 示例：更改静态资源路径

如果需要从不同路径提供资源：

#### 步骤 1：更新嵌入

**文件**：`embed_prod.go`

```go
//go:build prod

//go:embed static
var staticFS embed.FS
```

如果更改目录，在这里更新路径。

#### 步骤 2：更新服务逻辑

**文件**：`main.go`

```go
// 提供静态文件
r.StaticFS("/static", http.FS(staticFS))
```

#### 步骤 3：更新前端构建

**文件**：`webs/vite.config.js`

```javascript
export default defineConfig({
  base: '/static/', // 匹配后端路径
  build: {
    outDir: 'dist',
  },
});
```

#### 步骤 4：更新 Docker 构建

**文件**：`Dockerfile`

```dockerfile
# 将前端构建复制到正确路径
COPY --from=frontend-builder /app/webs/dist ./static
```

#### 步骤 5：更新 CI

**文件**：`.github/workflows/build-release.yml`

```yaml
- name: 复制前端资源
  run: cp -R webs/dist ./static
```

### 验证检查清单

构建/部署更改后：

- [ ] 测试本地开发模式
- [ ] 测试生产构建
- [ ] 测试 Docker 构建
- [ ] 验证资源正确加载
- [ ] 使用不同 base path 测试
- [ ] 检查 CI 工作流运行
- [ ] 更新文档

```bash
# 本地生产构建
cd webs && yarn run build && cd ..
rm -rf static && cp -R webs/dist static
CGO_ENABLED=0 go build -tags=prod -o sublinkPro
./sublinkPro

# Docker 构建
docker build -t sublinkpro:test .
docker run -p 8000:8000 sublinkpro:test
```

---

## 🎨 模式 5：添加带主题支持的前端功能

### 何时主题很重要

添加使用以下内容的 UI 组件时：
- 颜色、背景、边框
- 对话框、抽屉、弹出框
- 卡片、面板、覆盖层
- 状态颜色、标签、徽章

### 分步示例：添加通知面板

#### 步骤 1：创建带主题的组件

**文件**：`webs/src/components/NotificationPanel.jsx`

```jsx
import { useTheme } from '@mui/material/styles';
import { Box, Typography } from '@mui/material';

export default function NotificationPanel({ notifications }) {
  const theme = useTheme();
  const palette = theme.vars?.palette || theme.palette;
  
  return (
    <Box
      sx={{
        backgroundColor: palette.background.paper,
        borderRadius: 1,
        border: `1px solid ${palette.divider}`,
        p: 2,
      }}
    >
      {notifications.map((notif) => (
        <Box
          key={notif.id}
          sx={{
            p: 1.5,
            mb: 1,
            backgroundColor: palette.background.default,
            borderRadius: 0.5,
            '&:hover': {
              backgroundColor: theme.palette.action.hover,
            },
          }}
        >
          <Typography color="text.primary">
            {notif.title}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {notif.message}
          </Typography>
        </Box>
      ))}
    </Box>
  );
}
```

#### 步骤 2：添加 i18n 支持

**文件**：`webs/src/i18n/locales/en-US.json`

```json
{
  "notifications": {
    "title": "Notifications",
    "empty": "No notifications",
    "markAllRead": "Mark all as read"
  }
}
```

**文件**：`webs/src/i18n/locales/zh-CN.json`

```json
{
  "notifications": {
    "title": "通知",
    "empty": "暂无通知",
    "markAllRead": "全部标记为已读"
  }
}
```

#### 步骤 3：在组件中使用

```jsx
import { useTranslation } from 'react-i18next';

export default function NotificationPanel({ notifications }) {
  const { t } = useTranslation();
  
  return (
    <Box>
      <Typography variant="h6">
        {t('notifications.title')}
      </Typography>
      {/* ... */}
    </Box>
  );
}
```

#### 步骤 4：验证主题

使用 theme-check skill：

```bash
# 检查 .agents/skills/theme-check/SKILL.md
```

- [ ] 在浅色模式下测试
- [ ] 在深色模式下测试
- [ ] 在桌面端测试
- [ ] 在移动端测试
- [ ] 检查悬停状态
- [ ] 仅使用语义颜色

### 验证

```bash
cd webs
yarn run lint
yarn run build

# 手动测试
yarn run start
# 在浅色/深色模式之间切换
# 在不同屏幕尺寸上测试
```

---

## 📚 相关文档

- **后端开发**：`development.md`
- **构建流程**：`build-and-deployment.md`
- **主题指南**：`frontend-theme-guidelines.md`
- **i18n 指南**：`internationalization.md`
- **配置**：`configuration.md`
- **安全**：`security-guidelines.md`
- **Skills**：`.agents/skills/`

---

## 💡 提示

### 找到正确的文件

1. **使用高价值入口点**，在 `AGENTS.md` 第 9 节
2. **遵循现有模式** - 查找类似功能并复制其结构
3. **检查相关文件** - 如果更改 A，检查 B 和 C（见上面的模式）
4. **使用 skills** - 运行相关的 `.agents/skills/` 检查清单

### 要避免的常见错误

- ❌ 当多个层受影响时只更改一层
- ❌ 忘记更新文档
- ❌ 不在浅色和深色模式下都测试
- ❌ 硬编码字符串而不是使用 i18n
- ❌ 不遵循现有代码模式
- ❌ 跳过验证命令

### 开发工作流程

1. **首先阅读相关文档**
2. **查找现有模式**
3. **进行最小更改**
4. **本地测试**
5. **运行验证命令**
6. **使用 skills 检查清单**
7. **如需要更新文档**
8. **创建 PR**

---

## 🎯 总结

大多数开发任务遵循这些模式：

| 模式 | 文件 | 关键点 |
|---|---|---|
| 后端功能 | models → services → api → routers | 从底层向上添加 |
| 定时任务 | scheduler/job_ids → task impl → manager | 遵循 cron 格式 |
| mihomo 更改 | mihomo/ → scheduler → utils → api | 检查所有集成点 |
| 构建更改 | Dockerfile → CI → embed → docs | 验证所有构建模式 |
| 前端功能 | api → views → components → i18n | 需要主题 + i18n |

有关详细步骤，请参见上面的模式和相关文档。
