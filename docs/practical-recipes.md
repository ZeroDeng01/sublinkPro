English | [简体中文](practical-recipes.zh-CN.md)

# Practical Recipes

Common development patterns and step-by-step guides for frequent tasks in SublinkPro.

---

## 📋 Quick Reference

| Task | Files to Check | Documentation |
|---|---|---|
| Add backend feature | `api/`, `services/`, `models/`, `routers/` | This file |
| Add/change scheduled task | `services/scheduler/` | This file |
| Change mihomo behavior | `services/mihomo/` | This file |
| Change deployment | `Dockerfile`, `docker-compose.yml`, docs | `build-and-deployment.md` |
| Add frontend feature | `webs/src/views/`, `webs/src/api/` | `development.md` |
| Change theme | `webs/src/themes/`, components | `frontend-theme-guidelines.md` |

---

## 🔧 Recipe 1: Adding a Backend Feature

### Typical Flow

A complete backend feature usually touches these layers:

1. **Models** (`models/`) - If persistence is needed
2. **Services** (`services/`) - Business logic
3. **API** (`api/`) - HTTP handlers
4. **Routers** (`routers/`) - Route registration
5. **Frontend** (`webs/src/api/` + `webs/src/views/`) - If UI is needed

### Step-by-Step Example: Add "Tags" Feature

#### Step 1: Define the Model

**File**: `models/tag.go`

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

**Add migration** in `models/db_migrate.go`:

```go
func RunMigrations(db *gorm.DB) error {
    return db.AutoMigrate(
        // ... existing models
        &Tag{},
    )
}
```

#### Step 2: Create the Service

**File**: `services/tag_service.go`

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

// ... more methods
```

#### Step 3: Create the API Handler

**File**: `api/tag.go`

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
            "msg":  "Failed to list tags",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "code": 200,
        "data": tags,
    })
}

// ... more handlers
```

#### Step 4: Register Routes

**File**: `routers/router.go`

```go
func SetupRouter(/* ... dependencies */) *gin.Engine {
    // ... existing setup
    
    tagAPI := api.NewTagAPI(tagService)
    
    apiV1 := r.Group("/api/v1")
    {
        // Protected routes
        apiV1.Use(middlewares.AuthToken())
        {
            apiV1.GET("/tags", tagAPI.ListTags)
            apiV1.POST("/tags", tagAPI.CreateTag)
            // ... more routes
        }
    }
    
    return r
}
```

#### Step 5: Add Frontend Support (Optional)

**File**: `webs/src/api/tags.js`

```javascript
import request from './request';

export const getTags = () => {
  return request.get('/api/v1/tags');
};

export const createTag = (data) => {
  return request.post('/api/v1/tags', data);
};
```

**File**: `webs/src/views/tags/index.jsx`

```jsx
import { getTags } from '@/api/tags';

export default function TagsView() {
  // ... implementation
}
```

#### Step 6: Cross-Layer Sync

When adding a backend feature with frontend:

1. ✅ Add API endpoint documentation to `skill-sublinkpro/reference/api.md`
2. ✅ Update i18n if user-facing text added
3. ✅ Add tests for business logic
4. ✅ Document in `docs/features/` if significant

### Validation

```bash
# Backend
gofmt -w models/tag.go services/tag_service.go api/tag.go
golangci-lint run
go test ./...

# Frontend (if added)
cd webs
yarn run lint
yarn run build
```

---

## ⏰ Recipe 2: Adding or Changing a Scheduled Task

### Understanding the Scheduler

**Key files**:
- `services/scheduler/manager.go` - Scheduler management
- `services/scheduler/job_ids.go` - Job ID constants
- `services/scheduler/speedtest_task.go` - Example task

### Task Types

1. **System tasks**: Negative IDs (e.g., `-1`, `-2`)
2. **User tasks**: Positive IDs from database

### Step-by-Step Example: Add Cleanup Task

#### Step 1: Define Job ID

**File**: `services/scheduler/job_ids.go`

```go
const (
    JobIDSpeedTest  = -1
    JobIDSubscriptionUpdate = -2
    JobIDCleanup    = -3  // New system task
)
```

#### Step 2: Implement Task Logic

**File**: `services/scheduler/cleanup_task.go`

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
    log.Println("Starting cleanup task...")
    
    // Cleanup logic here
    cutoff := time.Now().AddDate(0, 0, -30) // 30 days ago
    
    result := t.db.Where("created_at < ?", cutoff).
        Delete(&models.LogEntry{})
    
    if result.Error != nil {
        return result.Error
    }
    
    log.Printf("Cleanup completed: %d rows deleted", result.RowsAffected)
    return nil
}
```

#### Step 3: Register Task in Scheduler

**File**: `services/scheduler/manager.go`

```go
func (m *Manager) RegisterSystemTasks() {
    // ... existing tasks
    
    // Register cleanup task
    cleanupTask := NewCleanupTask(m.db)
    m.AddSystemTask(
        JobIDCleanup,
        "0 2 * * *",  // Run at 2 AM daily
        func() {
            if err := cleanupTask.Run(); err != nil {
                log.Printf("Cleanup task failed: %v", err)
            }
        },
    )
}
```

#### Step 4: Test the Task

```go
// In cleanup_task_test.go
func TestCleanupTask(t *testing.T) {
    db := setupTestDB(t)
    task := NewCleanupTask(db)
    
    // Add test data
    // ...
    
    err := task.Run()
    assert.NoError(t, err)
    
    // Verify cleanup
    // ...
}
```

### Cron Format Reference

```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
* * * * *
```

**Examples**:
- `0 2 * * *` - Every day at 2 AM
- `*/15 * * * *` - Every 15 minutes
- `0 */6 * * *` - Every 6 hours
- `0 0 * * 0` - Every Sunday at midnight

### Validation

```bash
# Test the task
go test ./services/scheduler/ -run TestCleanupTask

# Test cron parsing
go test ./services/scheduler/ -v
```

---

## 🌐 Recipe 3: Changing mihomo-Related Behavior

### Understanding mihomo Integration

mihomo is the core network/proxy engine. Changes here affect:
- Speed testing
- Proxy adapter creation
- DNS resolution
- Host mapping
- Proxied downloads

**Key files** to check together:
- `services/mihomo/mihomo.go` - Core wrapper
- `services/mihomo/dns_resolver.go` - DNS logic
- `services/mihomo/host_resolver.go` - Host mapping
- `services/scheduler/speedtest_task.go` - Speed test integration
- `utils/proxy_client.go` - Proxied HTTP client
- `node/sub.go`, `node/usage.go` - Proxied downloads

### Example: Modify Speed Test Behavior

#### If Changing Test Logic

**File**: `services/mihomo/mihomo.go`

```go
func (m *MihomoService) TestNodeDelay(
    node *models.Node,
    timeout time.Duration,
) (uint16, error) {
    // Your modified logic here
}
```

**Then check**: `services/scheduler/speedtest_task.go`

Make sure the scheduler task still works with your changes.

#### If Changing DNS Behavior

**File**: `services/mihomo/dns_resolver.go`

```go
func (r *DNSResolver) ResolveWithProxy(
    domain string,
    proxyURL string,
) ([]string, error) {
    // Your modified logic here
}
```

**Then check**:
- `api/host.go` - Host management endpoints
- `models/host.go` - Host model
- `services/mihomo/host_resolver.go` - Host sync

#### If Changing Proxy Client Behavior

**File**: `utils/proxy_client.go`

```go
func NewProxyClient(proxyURL string) (*http.Client, error) {
    // Your modified logic here
}
```

**Then check**:
- `node/sub.go` - Subscription downloads
- `node/usage.go` - Usage queries
- Any service using proxied HTTP requests

### Cross-Check Pattern

For mihomo changes, always verify:

1. ✅ Core logic (`services/mihomo/`)
2. ✅ Scheduled tasks (`services/scheduler/speedtest_task.go`)
3. ✅ HTTP clients (`utils/proxy_client.go`)
4. ✅ Related API endpoints (`api/host.go`, `api/node.go`)
5. ✅ Related models (`models/host.go`, `models/node.go`)
6. ✅ Feature documentation (`docs/features/speedtest.md`, etc.)

### Validation

```bash
# Test mihomo integration
go test ./services/mihomo/ -v

# Test scheduler integration
go test ./services/scheduler/ -v -run TestSpeedTest

# Test full flow
go run main.go
# Manually trigger speed test through UI
```

---

## 🚀 Recipe 4: Changing Deployment or Build Behavior

### When to Use This Recipe

- Modifying `Dockerfile`
- Changing frontend asset embedding
- Updating `docker-compose.yml`
- Changing production build process
- Modifying base path behavior

### Files to Check Together

1. `Dockerfile` - Docker build
2. `embed_prod.go` - Asset embedding
3. `.github/workflows/build-release.yml` - CI build
4. `main.go` - Startup and serving
5. `webs/vite.config.js` - Frontend build config
6. `docs/build-and-deployment.md` - Build documentation

### Example: Change Static Asset Path

If you need to serve assets from a different path:

#### Step 1: Update Embedding

**File**: `embed_prod.go`

```go
//go:build prod

//go:embed static
var staticFS embed.FS
```

If changing directory, update the path here.

#### Step 2: Update Serving Logic

**File**: `main.go`

```go
// Serve static files
r.StaticFS("/static", http.FS(staticFS))
```

#### Step 3: Update Frontend Build

**File**: `webs/vite.config.js`

```javascript
export default defineConfig({
  base: '/static/', // Match backend path
  build: {
    outDir: 'dist',
  },
});
```

#### Step 4: Update Docker Build

**File**: `Dockerfile`

```dockerfile
# Copy frontend build to correct path
COPY --from=frontend-builder /app/webs/dist ./static
```

#### Step 5: Update CI

**File**: `.github/workflows/build-release.yml`

```yaml
- name: Copy frontend assets
  run: cp -R webs/dist ./static
```

### Validation Checklist

After build/deployment changes:

- [ ] Test local development mode
- [ ] Test production build
- [ ] Test Docker build
- [ ] Verify assets load correctly
- [ ] Test with different base paths
- [ ] Check CI workflow runs
- [ ] Update documentation

```bash
# Local production build
cd webs && yarn run build && cd ..
rm -rf static && cp -R webs/dist static
CGO_ENABLED=0 go build -tags=prod -o sublinkPro
./sublinkPro

# Docker build
docker build -t sublinkpro:test .
docker run -p 8000:8000 sublinkpro:test
```

---

## 🎨 Recipe 5: Adding a Frontend Feature with Theme Support

### When Theme Matters

When adding UI components that use:
- Colors, backgrounds, borders
- Dialogs, drawers, popovers
- Cards, panels, overlays
- Status colors, chips, badges

### Step-by-Step Example: Add Notification Panel

#### Step 1: Create Component with Theme

**File**: `webs/src/components/NotificationPanel.jsx`

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

#### Step 2: Add i18n Support

**File**: `webs/src/i18n/locales/en-US.json`

```json
{
  "notifications": {
    "title": "Notifications",
    "empty": "No notifications",
    "markAllRead": "Mark all as read"
  }
}
```

**File**: `webs/src/i18n/locales/zh-CN.json`

```json
{
  "notifications": {
    "title": "通知",
    "empty": "暂无通知",
    "markAllRead": "全部标记为已读"
  }
}
```

#### Step 3: Use in Component

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

#### Step 4: Validate Theme

Use the theme-check skill:

```bash
# Check .agents/skills/theme-check/SKILL.md
```

- [ ] Tested in light mode
- [ ] Tested in dark mode
- [ ] Tested on desktop
- [ ] Tested on mobile
- [ ] Checked hover states
- [ ] Used semantic colors only

### Validation

```bash
cd webs
yarn run lint
yarn run build

# Manual testing
yarn run start
# Switch between light/dark mode
# Test on different screen sizes
```

---

## 📚 Related Documentation

- **Backend development**: `development.md`
- **Build process**: `build-and-deployment.md`
- **Theme guidelines**: `frontend-theme-guidelines.md`
- **i18n guidelines**: `internationalization.md`
- **Configuration**: `configuration.md`
- **Security**: `security-guidelines.md`
- **Skills**: `.agents/skills/`

---

## 💡 Tips

### Finding the Right Files

1. **Use the high-value entry points** in `AGENTS.md` section 9
2. **Follow existing patterns** - find similar features and copy their structure
3. **Check related files** - if changing A, check B and C (see recipes above)
4. **Use skills** - run relevant `.agents/skills/` checklists

### Common Mistakes to Avoid

- ❌ Only changing one layer when multiple are affected
- ❌ Forgetting to update documentation
- ❌ Not testing in both light and dark modes
- ❌ Hardcoding strings instead of using i18n
- ❌ Not following existing code patterns
- ❌ Skipping validation commands

### Development Workflow

1. **Read relevant docs first**
2. **Find existing patterns**
3. **Make minimal changes**
4. **Test locally**
5. **Run validation commands**
6. **Use skills checklists**
7. **Update docs if needed**
8. **Create PR**

---

## 🎯 Summary

Most development tasks follow these patterns:

| Pattern | Files | Key Points |
|---|---|---|
| Backend feature | models → services → api → routers | Add from bottom up |
| Scheduled task | scheduler/job_ids → task impl → manager | Follow cron format |
| mihomo change | mihomo/ → scheduler → utils → api | Check all integration points |
| Build change | Dockerfile → CI → embed → docs | Verify all build modes |
| Frontend feature | api → views → components → i18n | Theme + i18n required |

For detailed steps, see the recipes above and related documentation.
