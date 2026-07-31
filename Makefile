.PHONY: run build test lint fmt clean lint-fe fmt-fe

APP = server

run: ## 启动服务
	go run ./cmd/server

build: ## 编译
	go build -ldflags="-s -w" -o bin/$(APP) ./cmd/server

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
	cd web && npx eslint src/ && npx prettier --check src/

fmt-fe: ## 前端代码格式化
	cd web && npx prettier --write src/

clean: ## 清理
	rm -f kingfisher.db bin/$(APP)

wire: ## 生成依赖注入
	cd internal/wire && wire
