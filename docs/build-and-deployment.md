English | [简体中文](build-and-deployment.zh-CN.md)

# Build and Deployment Guide

This guide covers the build process, production deployment, and CI/CD workflows for SublinkPro.

---

## 🏗️ Build Architecture

SublinkPro uses a **two-stage build process**:

1. **Frontend build**: Compile React + Vite application to static assets
2. **Backend build**: Compile Go application with embedded frontend assets

This architecture allows the Go binary to serve both API and frontend from a single process.

---

## 💻 Local Development Build

### Backend Build

For local development or testing:

```bash
# From repository root
go build -o sublinkpro main.go
```

This creates a development binary without embedded frontend assets.

### Frontend Build

For local frontend build:

```bash
# From webs/ directory
cd webs
yarn run build
```

Build output goes to `webs/dist/`.

---

## 🚀 Production Build

### Prerequisites

- Go 1.26.3 or newer
- Node.js 22
- Yarn 4 (via corepack)

### Step-by-Step Production Build

#### 1. Install Frontend Dependencies

```bash
cd webs
corepack enable
yarn install --immutable
```

#### 2. Lint and Build Frontend

```bash
cd webs
yarn run lint
yarn run build
```

This creates optimized static assets in `webs/dist/`.

#### 3. Prepare Static Assets

```bash
# From repository root
rm -rf static
mkdir -p static
cp -R webs/dist/. static/
```

The `static/` directory will be embedded into the Go binary.

#### 4. Build Backend with Embedded Assets

```bash
# From repository root
CGO_ENABLED=0 go build -tags=prod -ldflags="-s -w" -o sublinkPro
```

**Build flags explained**:
- `CGO_ENABLED=0` - Disable CGO for static linking
- `-tags=prod` - Use production build tag (enables `//go:embed`)
- `-ldflags="-s -w"` - Strip debug information to reduce binary size
- `-o sublinkPro` - Output binary name

#### 5. Verify Build

```bash
./sublinkPro --version
```

The binary should start and serve both API and frontend.

---

## 🐳 Docker Build

### Standard Docker Build

The `Dockerfile` implements the two-stage build automatically:

```bash
docker build -t sublinkpro:latest .
```

### Docker Build Process (Internal)

The Dockerfile follows this flow:

1. **Frontend stage**:
   ```dockerfile
   FROM node:22 as frontend-builder
   WORKDIR /app/webs
   RUN corepack enable
   COPY webs/package.json webs/yarn.lock ./
   RUN yarn install --immutable
   COPY webs/ ./
   RUN yarn run build
   ```

2. **Backend stage**:
   ```dockerfile
   FROM golang:1.26.3 as backend-builder
   WORKDIR /app
   COPY . .
   COPY --from=frontend-builder /app/webs/dist ./static
   RUN CGO_ENABLED=0 go build -tags=prod -ldflags="-s -w" -o sublinkPro
   ```

3. **Runtime stage**:
   ```dockerfile
   FROM alpine:latest
   COPY --from=backend-builder /app/sublinkPro /app/sublinkPro
   EXPOSE 8000
   CMD ["/app/sublinkPro"]
   ```

### Docker Compose Build

```bash
docker-compose build
docker-compose up -d
```

---

## 🔄 CI/CD Pipeline

### GitHub Actions Workflow

The repository uses `.github/workflows/build-release.yml` for automated builds.

#### Workflow Steps

1. **Setup**:
   - Checkout code
   - Setup Node.js 22
   - Setup Go 1.26.3
   - Enable corepack

2. **Frontend Build**:
   ```bash
   cd webs
   yarn install --immutable
   yarn run lint
   yarn run build
   ```

3. **Backend Validation**:
   ```bash
   gofmt -w .
   golangci-lint run
   go test ./...
   ```

4. **Production Build**:
   ```bash
   cp -R webs/dist ./static
   CGO_ENABLED=0 go build -tags=prod -ldflags="-s -w" -o sublinkPro
   ```

5. **Release** (on tag):
   - Create GitHub release
   - Upload binary artifacts
   - Push Docker images

### PR Checks

`.github/workflows/pr-checks.yml` runs on PR open/reopen/ready-for-review:

**Backend checks**:
- `golangci-lint run`
- `go test ./...`

**Frontend checks**:
- `yarn run lint`
- `yarn run build`

To manually re-trigger checks, comment `/recheck` on the PR.

---

## 📦 Build Artifacts

### Binary Artifacts

Production builds generate:
- `sublinkPro` - Linux amd64 binary
- `sublinkPro-arm64` - Linux arm64 binary
- `sublinkPro-darwin` - macOS binary (optional)

### Docker Images

Released images:
- `zerodeng/sublink-pro:latest`
- `zerodeng/sublink-pro:v{version}`

---

## 🔧 Build Modes

### Development Mode

**Characteristics**:
- Frontend runs on Vite dev server (port 3000)
- Backend runs separately (port 8000)
- Hot reload enabled
- Source maps available
- No asset embedding

**Start**:
```bash
# Terminal 1: Backend
go run main.go

# Terminal 2: Frontend
cd webs && yarn run start
```

### Production Mode

**Characteristics**:
- Single binary serves both API and frontend
- Frontend assets embedded in binary
- Optimized and minified
- No source maps
- Static file serving from embedded FS

**Build tag**: `-tags=prod`

---

## 🎯 Build Tag System

### `prod` Tag

Controls frontend asset embedding:

**Without `-tags=prod`**:
```go
// Development: serves from filesystem
http.FileServer(http.Dir("./webs/dist"))
```

**With `-tags=prod`**:
```go
//go:embed static
var staticFS embed.FS

// Production: serves from embedded FS
http.FS(staticFS)
```

---

## 🔍 Build Troubleshooting

### Frontend Build Fails

**Issue**: `yarn run build` fails with errors

**Common causes**:
- Outdated dependencies
- Node version mismatch
- TypeScript/ESLint errors

**Solution**:
```bash
cd webs
rm -rf node_modules .yarn/cache
yarn install --immutable
yarn run lint:fix
yarn run build
```

### Backend Build Fails

**Issue**: Go build fails with embed errors

**Cause**: `static/` directory doesn't exist or is empty

**Solution**:
```bash
# Ensure frontend is built first
cd webs && yarn run build && cd ..

# Prepare static directory
rm -rf static
mkdir -p static
cp -R webs/dist/. static/

# Then build backend
go build -tags=prod -o sublinkPro
```

### Binary Too Large

**Issue**: Binary size is unexpectedly large

**Check**:
- Using `-ldflags="-s -w"` to strip debug info?
- Using `CGO_ENABLED=0` for static linking?
- Frontend build is production-optimized?

**Optimize**:
```bash
# Frontend optimization
cd webs
yarn run build  # Already optimized by Vite

# Backend optimization
CGO_ENABLED=0 go build -tags=prod -ldflags="-s -w" -o sublinkPro

# Further compress (optional)
upx --best sublinkPro
```

### Assets Not Loading

**Issue**: Frontend assets return 404 in production

**Causes**:
1. Frontend not built before backend build
2. `static/` directory not copied correctly
3. Missing `-tags=prod` flag
4. `SUBLINK_WEB_BASE_PATH` misconfigured

**Debug**:
```bash
# Check if static directory exists and has content
ls -la static/

# Verify binary has embedded files
go version -m sublinkPro | grep embed

# Check runtime logs for asset paths
./sublinkPro 2>&1 | grep static
```

---

## 🌐 Base Path Handling

### SUBLINK_WEB_BASE_PATH

This environment variable controls the Web UI routing prefix.

**Example**:
```bash
SUBLINK_WEB_BASE_PATH=/sublink ./sublinkPro
```

**Effect**:
- Web UI accessible at: `http://localhost:8000/sublink/`
- API still at: `http://localhost:8000/api/*`
- Subscription access still at: `http://localhost:8000/c/*`

**Important**: Base path only affects Web UI routing, not API or subscription paths.

### Build Considerations

When changing base path behavior:
1. Test in local development mode
2. Test production build with embedded assets
3. Verify all asset paths resolve correctly
4. Test with different base path values

---

## 🔄 Incremental Build

### When Frontend Changes

```bash
cd webs
yarn run build
cd ..
rm -rf static
cp -R webs/dist static
go build -tags=prod -o sublinkPro
```

### When Backend Changes

```bash
# Frontend assets unchanged, just rebuild backend
go build -tags=prod -o sublinkPro
```

---

## 📊 Build Performance

### Typical Build Times

| Stage | Time | Notes |
|---|---|---|
| Frontend deps install | ~30s | Cached after first run |
| Frontend lint | ~5s | |
| Frontend build | ~15s | Vite optimization |
| Backend test | ~20s | Full test suite |
| Backend build | ~10s | With embed |
| **Total** | **~80s** | Cold build |

### Build Optimization Tips

1. **Use build cache**:
   ```bash
   # Go build cache
   go env GOCACHE
   
   # Yarn cache
   yarn cache dir
   ```

2. **Parallel builds**:
   ```bash
   # Frontend and backend validation in parallel
   (cd webs && yarn run lint) & golangci-lint run & wait
   ```

3. **Skip unnecessary steps** (development only):
   ```bash
   # Skip tests for quick rebuild
   go build -tags=prod -o sublinkPro
   ```

---

## 🎓 Related Documentation

- **Local development**: See `development.md`
- **Configuration**: See `configuration.md`
- **Deployment**: See `installation.md`
- **Docker compose**: See `skill-sublinkpro/reference/deploy.md`

---

## 📝 Notes

- Always build frontend before backend in production
- Use `-tags=prod` for production builds
- Verify embedded assets after build changes
- Test with real data before releasing
- Check CI workflow for canonical build steps
