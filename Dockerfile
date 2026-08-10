# Kingfisher 统一构建镜像：单文件多阶段构建前端 dist + 后端二进制，运行镜像同时服务前后端。
# 前端由 node 构建，后端由 golang 构建，最终 alpine 运行后端（含 static_dir 提供前端 SPA）。

# ---- Stage 1: 构建前端 dist ----
FROM node:22-alpine AS frontend-builder
# corepack 启用：首次调用 pnpm 按 package.json 的 packageManager 自动安装锁定版本
RUN corepack enable
WORKDIR /app
COPY kingfisher-web/package.json kingfisher-web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY kingfisher-web/ .
RUN pnpm build

# ---- Stage 2: 构建后端二进制 ----
FROM golang:1.25-alpine AS backend-builder
# 纯 Go SQLite 驱动（glebarez/modernc），无需 cgo，可静态编译
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG CGO_ENABLED=0
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# ---- Stage 3: 运行镜像（后端进程同时服务前端 dist + API）----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata curl
ENV TZ=Asia/Shanghai
WORKDIR /app
# 后端二进制 + 前端构建产物（static_dir: "kingfisher-web/dist"）
COPY --from=backend-builder /app/server .
COPY --from=backend-builder /src/config ./config
COPY --from=frontend-builder /app/dist ./kingfisher-web/dist
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --retries=3 CMD curl -f http://localhost:8080/health || exit 1
ENTRYPOINT ["./server"]
