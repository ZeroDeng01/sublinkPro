# SublinkPro Dockerfile
# ============================================
# 多阶段构建：前端 -> 后端 -> 运行时镜像
# ============================================

# 1. 构建前端
FROM node:22-alpine AS frontend-builder
WORKDIR /frontend
COPY webs ./webs

# Enable Corepack for Yarn
RUN corepack enable

# 使用 Yarn 安装依赖并构建
RUN cd webs && yarn install && yarn run build


# 2. 构建后端
FROM golang:1.25.5 AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# 把前端构建产物复制到 static 目录
COPY --from=frontend-builder /frontend/webs/dist ./static

RUN CGO_ENABLED=0 go build -tags=prod -ldflags="-s -w" -o sublinkPro


# 3. 运行镜像
FROM alpine:latest
WORKDIR /app

# ============================================
# 环境变量说明
# ============================================
# 基础配置
ENV GIN_MODE=release
# SUBLINK_PORT        - 服务端口 (默认: 8000)
# SUBLINK_DB_PATH     - 数据库目录 (默认: ./db)
# SUBLINK_LOG_PATH    - 日志目录 (默认: ./logs)
# SUBLINK_EXPIRE_DAYS - Token过期天数 (默认: 14)
#
# 敏感配置 (可选，不设置则自动生成并存储到数据库)
# SUBLINK_JWT_SECRET        - JWT签名密钥
# SUBLINK_API_ENCRYPTION_KEY - API加密密钥
#
# 登录安全配置
# SUBLINK_LOGIN_FAIL_COUNT    - 登录失败次数限制 (默认: 5)
# SUBLINK_LOGIN_FAIL_WINDOW   - 登录失败窗口时间(分钟) (默认: 1)
# SUBLINK_LOGIN_BAN_DURATION  - 登录封禁时间(分钟) (默认: 10)
#
# 管理员配置
# SUBLINK_ADMIN_PASSWORD      - 初始管理员密码 (仅首次启动时生效)
# SUBLINK_ADMIN_PASSWORD_REST - 重置管理员密码

# 安装 tzdata 和 ca-certificates，并设置时区
RUN apk add --no-cache tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone
RUN mkdir -p /app/db /app/logs /app/template && chmod 777 /app/db /app/logs /app/template

COPY --from=backend-builder /app/sublinkPro /app/sublinkPro
COPY --from=backend-builder /app/static /app/static


EXPOSE 8000
CMD ["/app/sublinkPro"]
