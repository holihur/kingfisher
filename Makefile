.PHONY: run dev build test lint fmt clean lint-fe fmt-fe wire swagger

APP = server

run: ## 启动服务
	go run ./cmd/server

dev: ## 同时启动前后端（Ctrl+C 一并退出）
	@trap 'kill 0' INT TERM EXIT; \
	(go run ./cmd/server) & \
	(cd kingfisher-web && npm run dev) & \
	wait

build: ## 编译
	go build -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev) -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo unknown) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/$(APP) ./cmd/server

test: ## 运行测试
	go test -v -race -count=1 ./...

test-short: ## 单元测试
	go test -v -race -short ./...

cover: ## 测试覆盖率
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint: ## Go 代码检查
	golangci-lint run ./...

fmt: ## Go 代码格式化
	goimports -w -local kingfisher .

lint-fe: ## 前端代码检查
	cd kingfisher-web && npx eslint src/ && npx prettier --check src/

fmt-fe: ## 前端代码格式化
	cd kingfisher-web && npx prettier --write src/

test-e2e: ## 运行 E2E 测试（需要 Redis 在 localhost:6379）
	@echo "Starting E2E tests..."
	cd kingfisher-web && npm run test:e2e

test-e2e-ui: ## Playwright UI 调试模式
	cd kingfisher-web && npm run test:e2e:ui

test-e2e-report: ## 查看 E2E 报告
	cd kingfisher-web && npx playwright show-report ../test/e2e/playwright-report

clean: ## 清理
	rm -f kingfisher.db bin/$(APP)

wire: ## 生成依赖注入
	cd internal/wire && wire

swagger: ## 生成 API 文档
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
