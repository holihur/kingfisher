# Deploy — 部署 & CI/CD

## Dockerfile（多阶段构建）

```dockerfile
# Stage 1: Build
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /app/server ./cmd/server

# Stage 2: Run
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata curl
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /src/config ./config
COPY --from=builder /src/migrations ./migrations
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --retries=3 CMD curl -f http://localhost:8080/health || exit 1
ENTRYPOINT ["./server"]
```

## docker-compose.yaml

```yaml
version: '3.8'
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_PASSWORD}
      MYSQL_DATABASE: kingfisher
    ports: ["3306:3306"]
    volumes:
      - mysql_data:/var/lib/mysql
      - ./migrations:/docker-entrypoint-initdb.d  # 首次启动自动迁移
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      retries: 5

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
    volumes: [redis_data:/data]
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]

  app:
    build: .
    ports: ["8080:8080"]
    depends_on:
      mysql: { condition: service_healthy }
      redis: { condition: service_healthy }
    environment:
      MYSQL_HOST: mysql
      MYSQL_USER: root
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
      MYSQL_DATABASE: kingfisher
      REDIS_HOST: redis
      JWT_SECRET: ${JWT_SECRET}
    restart: unless-stopped

  # 可观测性（开发环境）
  jaeger:
    image: jaegertracing/all-in-one:1.58
    ports: ["16686:16686", "4318:4318"]

  prometheus:
    image: prom/prometheus:v2.52
    ports: ["9090:9090"]
    volumes: [./deploy/prometheus.yaml:/etc/prometheus/prometheus.yml]

  frontend:
    build:
      context: ./kingfisher-web
      dockerfile: deploy/Dockerfile.frontend
    ports: ["80:80"]
    depends_on: [app]

  grafana:
    image: grafana/grafana:11.0
    ports: ["3000:3000"]
    volumes: [grafana_data:/var/lib/grafana]

volumes:
  mysql_data:
  redis_data:
  grafana_data:
```

## Makefile

```makefile
.PHONY: help run build test lint swagger wire docker-up docker-down

APP     = server
VERSION  = $(shell git describe --tags --always 2>/dev/null || echo "dev")
COMMIT   = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)

help: ## 显示帮助
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | \
	 awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## 本地运行
	go run ./cmd/server

build: ## 编译
	go build -ldflags="$(LDFLAGS)" -o bin/$(APP) ./cmd/server

wire: ## 生成依赖注入代码
	cd internal/wire && wire

swagger: ## 生成 API 文档
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

test: ## 运行测试
	go test -v -race -count=1 ./...

test-short: ## 运行单元测试（跳过集成测试）
	go test -v -race -short ./...

lint: ## 代码检查
	golangci-lint run --timeout=5m ./...

docker-build: ## 构建镜像
	docker build -t kingfisher:$(VERSION) .

docker-up: ## 启动服务
	docker-compose up -d

docker-down: ## 停止服务
	docker-compose down

migrate-up: ## 执行迁移
	go run ./cmd/migrate up

migrate-down: ## 回滚迁移
	go run ./cmd/migrate down

cover: ## 测试覆盖率
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
```

## CI/CD (GitHub Actions)

```yaml
# .github/workflows/ci.yaml
name: CI
on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.23' }
      - uses: golangci/golangci-lint-action@v6

  test:
    runs-on: ubuntu-latest
    services:
      mysql:
        image: mysql:8.0
        env: { MYSQL_ROOT_PASSWORD: test, MYSQL_DATABASE: kingfisher }
        ports: ['3306:3306']
      redis:
        image: redis:7-alpine
        ports: ['6379:6379']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: go test -v -race -count=1 ./...

  build:
    needs: [lint, test]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/build-push-action@v5
        with: { push: false, tags: kingfisher:latest }
```

## 设计要点

- Docker 最终镜像 < 15MB（Alpine + 瘦身 binary）
- `HEALTHCHECK` 指令让 Docker 自动检测健康状态
- `docker-compose` 用 `depends_on: condition: service_healthy` 确保启动顺序
- 版本号编译时注入（`-ldflags -X main.version=...`）
- CI 并行跑 lint / test，最后 build
