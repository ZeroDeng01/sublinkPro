# 1. 构建前端
FROM node:20-alpine AS frontend-builder
WORKDIR /frontend
COPY webs ./webs

# 安装 pnpm
RUN npm install -g pnpm

# 使用 pnpm 安装依赖并构建
RUN cd webs && pnpm install && pnpm run build


# 2. 构建后端
FROM golang:1.24.3 AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# 把前端构建产物复制到 static 目录
COPY --from=frontend-builder /frontend/webs/dist ./static

RUN go build -tags=prod -o sublinkE

# 3. 运行镜像
FROM debian:bookworm-slim
WORKDIR /app

# 安装 tzdata 和 ca-certificates，并设置时区
RUN apt-get update && \
    apt-get install -y --no-install-recommends tzdata ca-certificates && \
    ln -snf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    apt-get clean && rm -rf /var/lib/apt/lists/*
RUN mkdir -p /app/db /app/logs /app/template /app/plugins && chmod 777 /app/db /app/logs /app/template /app/plugins

COPY --from=backend-builder /app/sublinkE /app/sublinkE
COPY --from=backend-builder /app/static /app/static


EXPOSE 8000
CMD ["/app/sublinkE"]
