# Test — 测试策略

## 分层测试金字塔

```
         ╱ E2E ╲         少量：关键链路（登录→查列表→CRUD）
       ╱ Integration ╲    中等：DB + Redis + HTTP 集成
     ╱   Unit Tests    ╲  大量：Service + Handler 单测
    ─────────────────────
```

| 层级 | 工具 | 跑在 | 速度 | 占比 |
|------|------|------|------|------|
| Unit | `testing` + testify | 内存 | <1s | 60% |
| Integration | testcontainers + dockertest | 真实 DB/Redis | 10-30s | 30% |
| E2E | go test + httptest | 真实依赖 | 30-60s | 10% |

## Unit Test

### Service 单测（mock Repository + Cache）

```go
func TestUserService_Login_Success(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepo{
        FindByUsernameFunc: func(ctx context.Context, username string) (*domain.User, error) {
            hashed, _ := bcrypt.GenerateFromPassword([]byte("Abcd1234"), bcrypt.MinCost)
            return &domain.User{ID: 1, Username: "admin", Password: string(hashed), Status: 1}, nil
        },
    }
    mockCache := &MockCache{}
    jwtMgr := jwt.NewJWTManager(jwt.JWTConfig{Secret: "test", AccessTTL: time.Hour})
    svc := user.NewUserService(mockRepo, mockCache, jwtMgr)

    // Act
    access, refresh, user, err := svc.Login(context.Background(), "admin", "Abcd1234")

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, access)
    assert.NotEmpty(t, refresh)
    assert.Equal(t, "admin", user.Username)
    assert.Empty(t, user.Password)  // 密码不返回
}

func TestUserService_Login_WrongPassword(t *testing.T) { /* ... */ }
func TestUserService_Login_UserNotFound(t *testing.T) { /* ... */ }
func TestUserService_Login_Disabled(t *testing.T) { /* ... */ }
```

### Handler 单测（用 gin.CreateTestContext）

```go
func TestUserHandler_Login(t *testing.T) {
    mockSvc := &MockUserService{
        LoginFunc: func(ctx context.Context, username, password string) (string, string, *domain.User, error) {
            return "access", "refresh", &domain.User{ID: 1, Username: "admin"}, nil
        },
    }
    handler := user.NewUserHandler(mockSvc)

    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("POST", "/api/v1/auth/login",
        strings.NewReader(`{"username":"admin","password":"Abcd1234"}`))
    c.Request.Header.Set("Content-Type", "application/json")

    handler.Login(c)

    assert.Equal(t, 200, w.Code)
    var resp response.Response
    json.Unmarshal(w.Body.Bytes(), &resp)
    assert.Equal(t, 0, resp.Code)
}
```

## Integration Test

### 使用 testcontainers

```go
func TestUserRepo_CRUD(t *testing.T) {
    ctx := context.Background()

    // 启动 MySQL 容器
    mysqlContainer, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image: "mysql:8.0",
            Env:   map[string]string{"MYSQL_ROOT_PASSWORD": "test", "MYSQL_DATABASE": "kingfisher"},
        },
        Started: true,
    })
    defer mysqlContainer.Terminate(ctx)

    // 连接 + 迁移
    port, _ := mysqlContainer.MappedPort(ctx, "3306")
    db := setupTestDB(port.Int())
    migrate(db)

    // 测试
    repo := adapter.NewUserRepo(db)
    err := repo.Create(ctx, &domain.User{Username: "test", Password: "..."})
    assert.NoError(t, err)

    user, err := repo.FindByUsername(ctx, "test")
    assert.NoError(t, err)
    assert.Equal(t, "test", user.Username)
}
```

### 使用 httptest + 真实 DB（更轻量）

```go
func TestUserAPI_LoginFlow(t *testing.T) {
    router := setupTestRouter(t)   // 完整的 Gin Engine + 所有 module
    // POST /api/v1/auth/register
    // POST /api/v1/auth/login  → 拿到 token
    // GET  /api/v1/users/:id   → 带 token 请求
}
```

## Mock 工厂

```go
// test/testutil/mock.go
func NewMockUserRepo() *MockUserRepo { ... }
func NewMockCache() *MockCache { ... }
func NewMockJWTManager() *MockJWTManager { ... }
```

## 测试文件位置

```
extends/user/
├── app/service_test.go           # Service 单测
├── transport/handler_test.go     # Handler 单测
└── adapter/mysql/repo_test.go    # Adapter 集成测试
test/
├── integration/                  # 跨模块集成测试
│   ├── user_integration_test.go
│   └── rbac_integration_test.go
├── e2e/                          # 端到端
│   └── api_test.go
└── testutil/                     # Mock + Fixture 工具
    ├── mock_cache.go
    ├── mock_repo.go
    └── fixture.go                # 造数据
```

## Makefile

```makefile
test-unit:
	go test -v -race ./internal/... -short

test-integration:
	go test -v -race ./test/integration/...

test-e2e:
	go test -v ./test/e2e/...

test: test-unit test-integration

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
```

## 设计要点

- 单测覆盖：Service 核心逻辑 100%、Handler 参数校验 80%
- 集成测试：每个 Repository 至少 CRUD + 边界条件
- Mock 接口定义在 port 层，Mock 实现在 test/testutil
- 测试用 `-race` 检测并发问题
- CI 中单元测试 < 30s，集成测试 < 3min
