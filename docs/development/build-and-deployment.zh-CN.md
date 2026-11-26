[English](build-and-deployment.md) | 简体中文

[返回指南目录](README.zh-CN.md)

# 构建与部署指南

本指南涵盖 SublinkPro 的构建流程、生产部署和 CI/CD 工作流。

---

## 🏗️ 构建架构

SublinkPro 使用**两阶段构建流程**：

1. **前端构建**：将 React + Vite 应用编译为静态资源
2. **后端构建**：将 Go 应用与嵌入的前端资源一起编译

这种架构允许 Go 二进制文件从单个进程同时提供 API 和前端服务。

---

## 💻 本地开发构建

### 后端构建

用于本地开发或测试：

```bash
# 从仓库根目录执行
go build -o sublinkpro .
```

这会创建一个不包含嵌入前端资源的开发版二进制文件。

### 前端构建

本地前端构建：

```bash
# 从 webs/ 目录执行
cd webs
yarn run build
```

构建输出到 `webs/dist/`。

---

## 🚀 生产构建

### 前置条件

- Go 1.27.3 或更新版本
- Node.js 24
- Yarn 4（通过 corepack）

### 生产构建详细步骤

#### 1. 安装前端依赖

```bash
cd webs
corepack enable
yarn install --immutable
```

#### 2. 检查和构建前端

```bash
cd webs
yarn run lint
yarn run build
```

这会在 `webs/dist/` 中创建优化的静态资源。

#### 3. 准备静态资源

```bash
# 从仓库根目录执行
rm -rf static
mkdir -p static
cp -R webs/dist/. static/
```

`static/` 目录将被嵌入到 Go 二进制文件中。

#### 4. 构建带嵌入资源的后端

```bash
# 从仓库根目录执行
CGO_ENABLED=0 go build -tags=prod -ldflags="-s -w" -o sublinkPro
```

**构建标志说明**：
- `CGO_ENABLED=0` - 禁用 CGO 以实现静态链接
- `-tags=prod` - 使用生产构建标签（启用 `//go:embed`）
- `-ldflags="-s -w"` - 剥离调试信息以减小二进制大小
- `-o sublinkPro` - 输出二进制文件名

#### 5. 验证构建

```bash
./sublinkPro --version
```

二进制文件应该启动并同时提供 API 和前端服务。

---

## 🐳 Docker 构建

### 标准 Docker 构建

`Dockerfile` 自动实现两阶段构建：

```bash
docker build -t sublinkpro:latest .
```

### Docker 构建流程（内部）

Dockerfile 遵循以下流程：

1. **前端阶段**：
   ```dockerfile
   FROM node:24-alpine AS frontend-builder
   WORKDIR /frontend
   COPY webs ./webs
   RUN corepack enable
   RUN cd webs && yarn install && yarn run build
   ```

2. **后端阶段**：
   ```dockerfile
   FROM golang:1.27.3 as backend-builder
   WORKDIR /app
   COPY . .
   COPY --from=frontend-builder /frontend/webs/dist ./static
   RUN CGO_ENABLED=0 go build -tags=prod -ldflags="-s -w" -o sublinkPro
   ```

3. **运行时阶段**：
   ```dockerfile
   FROM alpine:latest
   COPY --from=backend-builder /app/sublinkPro /app/sublinkPro
   EXPOSE 8000
   CMD ["/app/sublinkPro"]
   ```

### Docker Compose 构建

```bash
docker compose build
docker compose up -d
```

---

## 🔄 CI/CD 流水线

### GitHub Actions 工作流

仓库使用 `.github/workflows/build-release.yml` 进行自动化构建。推送到 `dependabot/**` 的提交会跳过此工作流。以 `dev` 为目标的 Dependabot PR 会运行 PR 检查，合并到 `dev` 后触发常规构建流程。

| 渠道 | 触发 | Docker 镜像标签 | GitHub Release |
| --- | --- | --- | --- |
| 开发版 | 推送 `dev` | `dev` | 无 |
| 预发布 | 推送 `main` | `pre-YYMMDDHHMMSS` 和 `preview` | 附多平台二进制的预发布 |
| 稳定版 | 推送 `v*` 标签 | 版本标签和 `latest` | 附多平台二进制的正式 Release |

预发布标题、二进制版本、Git 标签和固定 Docker 标签统一使用 `pre-261125093045` 这样的版本号。工作流首次为提交分配标签时记录北京时间，年份使用两位数字，精确到秒。重跑复用该提交已有的标签。时间标签已被其他提交占用时，工作流尝试下一秒。

预发布公开后，工作流将 Docker 滚动标签 `:preview` 指向最近公开的预发布所对应的多架构镜像。使用时间标签可固定到某一次构建。

预发布按发布时间选择当前 Git 历史中的上一条公开 Release 作为起点，正式版和预发布都参与比较。当前历史首次发布时，日志覆盖全部提交。稳定版以此前的 `v*` 标签为起点。两类发布都会列出 Docker 镜像和各平台安装包。

开发版内嵌版本号 `dev-<7 位 SHA>`，用于版本卡展示。升级开发版时，重新拉取滚动更新的 `:dev` 镜像并重建容器。构建如果已被更新的提交替代，就跳过镜像推送步骤。

一键脚本默认选择稳定版。安装或更新到预发布时使用 `--preview`。[安装部署指南](../user/installation.zh-CN.md)

其他分支运行构建验证。预发布和稳定版会在 Docker 镜像及二进制产物就绪后发布。

#### 工作流步骤

1. **设置**：
   - 检出代码
   - 设置 Node.js 24
   - 设置 Go 1.27.3
   - 启用 corepack

2. **前端构建**：
   ```bash
   cd webs
   yarn install --immutable
   yarn run lint
   yarn run build
   ```

3. **后端验证**：
   ```bash
   gofmt -w .
   golangci-lint run
   go test ./...
   ```

4. **生产构建**：
   ```bash
   cp -R webs/dist ./static
   CGO_ENABLED=0 go build -tags=prod -ldflags="-s -w" -o sublinkPro
   ```

5. **发布**：
   - 按上表推送对应渠道的 Docker 镜像
   - 为预发布和稳定版附上二进制产物

### PR 检查

`.github/workflows/pr-checks.yml` 在目标为 `dev` 的 PR 创建、重新打开、标记为可审查，或改为以 `dev` 为目标时运行：

**后端检查**：
- `go build ./...`
- `golangci-lint run`
- `go test ./...`

**前端检查**：
- `yarn run lint`
- `yarn run build`

自动检查和 `/recheck` 都会在 PR 对话发布详细报告。复查权限和目标分支规则见 [PR 自动检查](getting-started.zh-CN.md#6-pr-自动检查)。

---

## 📦 构建产物

### 二进制产物

生产构建生成：
- `sublinkPro-linux-amd64` - Linux x64 二进制文件
- `sublinkPro-linux-arm64` - Linux ARM64 二进制文件
- `sublinkPro-linux-armv7` - Linux ARMv7 32 位二进制文件
- `sublinkPro-linux-x86` - Linux x86 32 位二进制文件
- `sublinkPro-windows-amd64.exe` - Windows x64 二进制文件
- `sublinkPro-windows-x86.exe` - Windows x86 32 位二进制文件
- `sublinkPro-darwin-amd64` - macOS x64 二进制文件
- `sublinkPro-darwin-arm64` - macOS ARM64 二进制文件

### Docker 镜像

发布的镜像：
- `zerodeng/sublink-pro:latest`
- `zerodeng/sublink-pro:v{version}`

### 发布整改项

独立二进制文件和运行镜像需要随附 [MIT 许可](../../LICENSE) 声明。将两个打包流程中的许可文件和发布产物检查作为独立发布改动处理。

---

## 🔧 构建模式

### 开发模式

**特征**：
- 前端运行在 Vite 开发服务器上（端口 3000）
- 后端单独运行（端口 8000）
- 启用热重载
- 提供 source maps
- 无资源嵌入

**启动**：
```bash
# 终端 1：后端
go run .

# 终端 2：前端
cd webs && yarn run start
```

### 生产模式

**特征**：
- 单个二进制文件同时提供 API 和前端服务
- 前端资源嵌入到二进制文件中
- 优化和压缩
- 无 source maps
- 从嵌入的文件系统提供静态文件

**构建标签**：`-tags=prod`

---

## 🎯 构建标签系统

### `prod` 标签

控制前端资源嵌入：

**不使用 `-tags=prod`**：
```go
// 开发模式：从文件系统提供服务
http.FileServer(http.Dir("./webs/dist"))
```

**使用 `-tags=prod`**：
```go
//go:embed static
var staticFS embed.FS

// 生产模式：从嵌入的文件系统提供服务
http.FS(staticFS)
```

---

## 🔍 构建故障排查

### 前端构建失败

**问题**：`yarn run build` 失败并显示错误

**常见原因**：
- 依赖过期
- Node 版本不匹配
- TypeScript/ESLint 错误

**解决方案**：
```bash
cd webs
rm -rf node_modules .yarn/cache
yarn install --immutable
yarn run lint:fix
yarn run build
```

### 后端构建失败

**问题**：Go 构建失败并显示 embed 错误

**原因**：`static/` 目录不存在或为空

**解决方案**：
```bash
# 确保首先构建前端
cd webs && yarn run build && cd ..

# 准备 static 目录
rm -rf static
mkdir -p static
cp -R webs/dist/. static/

# 然后构建后端
go build -tags=prod -o sublinkPro
```

### 二进制文件过大

**问题**：二进制文件大小异常大

**检查**：
- 使用 `-ldflags="-s -w"` 剥离调试信息了吗？
- 使用 `CGO_ENABLED=0` 进行静态链接了吗？
- 前端构建是生产优化的吗？

**优化**：
```bash
# 前端优化
cd webs
yarn run build  # Vite 已自动优化

# 后端优化
CGO_ENABLED=0 go build -tags=prod -ldflags="-s -w" -o sublinkPro

# 进一步压缩（可选）
upx --best sublinkPro
```

### 资源加载失败

**问题**：生产环境中前端资源返回 404

**原因**：
1. 后端构建前未构建前端
2. `static/` 目录未正确复制
3. 缺少 `-tags=prod` 标志
4. `SUBLINK_WEB_BASE_PATH` 配置错误

**调试**：
```bash
# 检查 static 目录是否存在且有内容
ls -la static/

# 验证二进制文件是否包含嵌入文件
go version -m sublinkPro

# 检查运行时日志中的资源路径
./sublinkPro 2>&1 | grep static
```

---

## 🌐 Base Path 处理

### SUBLINK_WEB_BASE_PATH

此环境变量控制 Web UI 路由前缀。

**示例**：
```bash
SUBLINK_WEB_BASE_PATH=/sublink ./sublinkPro
```

**效果**：
- Web UI 访问地址：`http://localhost:8000/sublink/`
- API 仍然在：`http://localhost:8000/api/*`
- 订阅访问仍然在：`http://localhost:8000/c/*`

**重要**：Base path 仅影响 Web UI 路由，不影响 API 或订阅路径。

### 构建注意事项

更改 base path 行为时：
1. 在本地开发模式下测试
2. 测试带嵌入资源的生产构建
3. 验证所有资源路径正确解析
4. 使用不同的 base path 值测试

---

## 🔄 增量构建

### 前端更改时

```bash
cd webs
yarn run build
cd ..
rm -rf static
cp -R webs/dist static
go build -tags=prod -o sublinkPro
```

### 后端更改时

```bash
# 前端资源未更改，只需重新构建后端
go build -tags=prod -o sublinkPro
```

---

## 📊 构建性能

### 典型构建时间

| 阶段 | 时间 | 备注 |
|---|---|---|
| 前端依赖安装 | ~30s | 首次运行后缓存 |
| 前端 lint | ~5s | |
| 前端构建 | ~15s | Vite 优化 |
| 后端测试 | ~20s | 完整测试套件 |
| 后端构建 | ~10s | 包含 embed |
| **总计** | **~80s** | 冷构建 |

### 构建优化技巧

1. **使用构建缓存**：
   ```bash
   # Go 构建缓存
   go env GOCACHE
   
   # Yarn 缓存
   yarn config get cacheFolder
   ```

2. **并行构建**：
   ```bash
   # 前端和后端验证并行
   (cd webs && yarn run lint) & golangci-lint run & wait
   ```

3. **跳过不必要的步骤**（仅开发）：
   ```bash
   # 跳过测试快速重建
   go build -tags=prod -o sublinkPro
   ```

---

## 🎓 相关文档

- **本地开发**：参见 `development.md`
- **配置**：参见 `configuration.md`
- **部署**：参见 `installation.md`
- **Docker compose**：参见 `skills/sublinkpro/reference/deploy.md`

---

## 📝 注意事项

- 生产环境中始终先构建前端再构建后端
- 生产构建使用 `-tags=prod`
- 构建更改后验证嵌入资源
- 发布前使用真实数据测试
- 查看 CI 工作流以了解规范构建步骤
