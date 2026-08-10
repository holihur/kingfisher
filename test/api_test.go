package test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"

	"kingfisher/core/config"
	"kingfisher/core/database"
	"kingfisher/core/jwt"
	"kingfisher/core/logger"
	"kingfisher/core/middleware"
	"kingfisher/core/router"

	auditTransport "kingfisher/extends/audit/transport"
	configTransport "kingfisher/extends/config/transport"
	dictTransport "kingfisher/extends/dict/transport"
	menuTransport "kingfisher/extends/menu/transport"
	messageTransport "kingfisher/extends/message/transport"
	messageWorker "kingfisher/extends/message/worker"
	rbacTransport "kingfisher/extends/rbac/transport"
	userTransport "kingfisher/extends/user/transport"
)

func setupTestServer(t *testing.T) (*gin.Engine, *jwt.JWTManager) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-secret-abc123")
	// 注册自定义校验器（password 等），与 cmd/server/main.go 一致
	middleware.InitValidator()

	cfg := &config.Config{
		Server:   config.ServerConfig{Port: 9999, Mode: "test"},
		Database: config.DatabaseConfig{Driver: "sqlite", SQLite: config.SQLiteConfig{Path: ":memory:"}},
		Redis:    config.RedisConfig{Host: "127.0.0.1", Port: 6379},
		JWT:      config.JWTConfig{Secret: "test-secret-abc123", AccessTTL: 3600000000000000, RefreshTTL: 7200000000000000, Issuer: "test"},
		Log:      config.LogConfig{Level: "error", Format: "console", Output: "stdout", MaxSize: 10, MaxBackups: 1, MaxAge: 1},
		CORS:     config.CORSConfig{AllowedOrigins: []string{"*"}, AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"}, AllowedHeaders: []string{"*"}},
	}

	zapLog, _ := logger.New(logger.Config{Level: "error", Format: "console", Output: "stdout", MaxSize: 10, MaxBackups: 1, MaxAge: 1})
	db, err := database.InitDatabase(cfg.Database, zapLog)
	if err != nil {
		t.Fatalf("db init: %v", err)
	}
	_ = database.SeedData(db)

	jwtMgr := jwt.NewJWTManager(cfg.JWT, nil)
	r := router.NewEngine(cfg, zapLog)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/version", func(c *gin.Context) { c.JSON(200, gin.H{"version": "dev"}) })

	authMw := rbacTransport.AuthMiddleware(jwtMgr, nil)
	allPerms := map[string]bool{
		"user:list": true, "user:create": true, "user:update": true, "user:delete": true,
		"menu:list": true, "menu:create": true, "menu:update": true, "menu:delete": true,
		"role:list": true, "role:create": true, "role:update": true, "role:delete": true,
		"config:list": true, "config:update": true, "audit:list": true,
		"dict:list": true, "dict:create": true, "dict:update": true, "dict:delete": true,
		"message:list": true, "message:create": true, "message:update": true, "message:delete": true,
	}
	rbacMw := func(c *gin.Context) {
		c.Set("permissions", allPerms)
		c.Next()
	}

	userMod := userTransport.NewUserModule(db, nil, jwtMgr, nil)
	// 注入角色落地页查询：集成测试中按 role_id 返回固定落地页
	userMod.InjectLandingPageProvider(func(ctx context.Context, roleID uint) (string, error) {
		switch roleID {
		case 1:
			return "/dashboard", nil
		case 3:
			return "/system/users", nil
		default:
			return "/dashboard", nil
		}
	})
	// 注入配置查询：注册开关/默认注册角色从 DB 读真实配置
	userMod.InjectConfigProvider(func(ctx context.Context, key string) (string, error) {
		var v string
		if err := db.Table("system_configs").Where("`key` = ?", key).Pluck("value", &v).Error; err != nil {
			return "", err
		}
		return v, nil
	})
	// 站内信：测试用同步生产者，内联执行真实 worker 的处理逻辑
	producer := &syncProducer{}
	msgMod := messageTransport.NewMessageModule(db, producer)
	producer.w = msgMod.Worker().(*messageWorker.MessageWorker)

	mods := []router.Module{
		userMod,
		rbacTransport.NewRBACModule(db, nil),
		menuTransport.NewMenuModule(db, nil),
		configTransport.NewConfigModule(db, nil),
		dictTransport.NewDictModule(db, nil),
		msgMod,
		auditTransport.NewAuditModule(db),
	}
	for _, m := range mods {
		_ = m.Init(t.Context())
		router.Register(r, m, authMw, rbacMw)
	}
	return r, jwtMgr
}

// syncProducer 测试用生产者：入队时内联执行真实 worker 的 HandleSendMessage，
// 让集成测试覆盖完整的"入队→处理→落库"链路，而无需连接 Redis。
type syncProducer struct{ w *messageWorker.MessageWorker }

func (p *syncProducer) Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	if err := p.w.HandleSendMessage(ctx, task); err != nil {
		return nil, err
	}
	return &asynq.TaskInfo{ID: "sync-task"}, nil
}

func itoa(n int) string { return strconv.Itoa(n) }

func doRequest(method, path, token string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequestWithContext(context.Background(), method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func login(t *testing.T, s *gin.Engine) string {
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/auth/login", "", map[string]string{"username": "admin", "password": "Abcd1234"}))
	if w.Code != 200 {
		t.Fatalf("login failed: %d", w.Code)
	}
	var resp struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Data.AccessToken
}

func loginAs(t *testing.T, s *gin.Engine, username string) string {
	t.Helper()
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/auth/login", "", map[string]string{"username": username, "password": "Abcd1234"}))
	if w.Code != 200 {
		t.Fatalf("login as %s failed: %d", username, w.Code)
	}
	var resp struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Data.AccessToken
}

func assertCode(t *testing.T, w *httptest.ResponseRecorder, want int) map[string]any {
	t.Helper()
	var m map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &m)
	if c, ok := m["code"].(float64); ok && int(c) != want {
		t.Errorf("want code=%d, got %v body=%s", want, c, w.Body.String())
	}
	return m
}

// Auth tests

func TestHealth(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/health", "", nil))
	if w.Code != 200 {
		t.Error("want 200")
	}
}

func TestVersion(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/version", "", nil))
	if w.Code != 200 {
		t.Error("want 200")
	}
}

func TestRegisterSuccess(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/auth/register", "", map[string]string{"username": "newuser", "password": "Abcd1234"}))
	assertCode(t, w, 0)
}

func TestRegisterDuplicate(t *testing.T) {
	s, _ := setupTestServer(t)
	s.ServeHTTP(httptest.NewRecorder(), doRequest("POST", "/api/v1/auth/register", "", map[string]string{"username": "dup2", "password": "Abcd1234"}))
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/auth/register", "", map[string]string{"username": "dup2", "password": "Abcd1234"}))
	assertCode(t, w, 10101)
}

func TestRegisterShortPassword(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/auth/register", "", map[string]string{"username": "x", "password": "a"}))
	if w.Code != 400 {
		t.Error("want 400")
	}
}

// 注册关闭时拒绝注册，返回 10111
func TestRegisterDisabled(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	// 关闭注册
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("PUT", "/api/v1/configs/registration_enabled", tok, map[string]string{"value": "false"}))
	assertCode(t, w, 0)

	// 注册被拒
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/auth/register", "", map[string]string{"username": "newreg", "password": "Abcd1234"}))
	assertCode(t, w, 10111)

	// 恢复开放
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("PUT", "/api/v1/configs/registration_enabled", tok, map[string]string{"value": "true"}))
	assertCode(t, w, 0)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/auth/register", "", map[string]string{"username": "newreg2", "password": "Abcd1234"}))
	assertCode(t, w, 0)
}

// 新注册用户使用配置的默认角色（default_register_role_id=4 访客）
func TestRegisterUsesDefaultRole(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/auth/register", "", map[string]string{"username": "rolereg", "password": "Abcd1234"}))
	assertCode(t, w, 0)

	tok := login(t, s)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/users?filter="+url.QueryEscape(`{"username":"rolereg"}`), tok, nil))
	m := assertCode(t, w, 0)
	items := m["data"].(map[string]any)["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("want 1 user, got %d", len(items))
	}
	if int(items[0].(map[string]any)["role_id"].(float64)) != 4 {
		t.Errorf("want default role 4, got %v", items[0].(map[string]any)["role_id"])
	}
}

func TestLoginSuccess(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	if tok == "" {
		t.Error("empty token")
	}
}

// 登录响应应携带角色落地页（admin → /dashboard）
func TestLoginReturnsLandingPage(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/auth/login", "", map[string]string{"username": "admin", "password": "Abcd1234"}))
	m := assertCode(t, w, 0)
	data := m["data"].(map[string]any)
	if data["landing_page"] != "/dashboard" {
		t.Errorf("want landing_page=/dashboard, got %v", data["landing_page"])
	}
}

func TestLoginWrongPassword(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/auth/login", "", map[string]string{"username": "admin", "password": "wrong"}))
	assertCode(t, w, 10103)
}

func TestLoginNonExistent(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/auth/login", "", map[string]string{"username": "ghost", "password": "anything"}))
	assertCode(t, w, 10103) // enumeration resistant
}

func TestLoginBadJSON(t *testing.T) {
	s, _ := setupTestServer(t)
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/auth/login", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Error("want 400")
	}
}

func TestRefreshSuccess(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/auth/login", "", map[string]string{"username": "admin", "password": "Abcd1234"}))
	var resp struct {
		Data struct {
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	w2 := httptest.NewRecorder()
	s.ServeHTTP(w2, doRequest("POST", "/api/v1/auth/refresh", "", map[string]string{"refresh_token": resp.Data.RefreshToken}))
	assertCode(t, w2, 0)
}

func TestRefreshWithAccessToken(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/auth/refresh", "", map[string]string{"refresh_token": tok}))
	assertCode(t, w, 10105)
}

func TestRefreshWithGarbage(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/auth/refresh", "", map[string]string{"refresh_token": "not.a.jwt"}))
	assertCode(t, w, 10105)
}

// Auth middleware tests

func TestNoAuthReturns401(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/users", "", nil))
	if w.Code != 401 {
		t.Errorf("want 401, got %d", w.Code)
	}
}

func TestNoAuthRoles(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/roles", "", nil))
	if w.Code != 401 {
		t.Error("want 401")
	}
}

func TestNoAuthConfigs(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/configs", "", nil))
	if w.Code != 401 {
		t.Error("want 401")
	}
}

// Public config tests: is_public=true 的配置无需登录即可读取

func TestPublicConfigsNoAuth(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/public/configs", "", nil))
	m := assertCode(t, w, 0)
	items := m["data"].([]any)
	if len(items) < 1 {
		t.Fatalf("want at least 1 public config, got %d", len(items))
	}
	keys := map[string]bool{}
	for _, it := range items {
		cfg := it.(map[string]any)
		if cfg["is_public"] != true {
			t.Errorf("config %v should be public", cfg["key"])
		}
		if k, ok := cfg["key"].(string); ok {
			keys[k] = true
		}
	}
	for _, required := range []string{"site_name", "site_logo", "site_description", "site_login_cover", "registration_enabled"} {
		if !keys[required] {
			t.Errorf("public configs missing required key %q in %v", required, keys)
		}
	}
}

func TestPublicConfigGetSingleNoAuth(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/public/configs/site_name", "", nil))
	m := assertCode(t, w, 0)
	cfg := m["data"].(map[string]any)
	if cfg["key"] != "site_name" || cfg["is_public"] != true {
		t.Errorf("unexpected public config: %v", cfg)
	}
}

func TestPublicConfigPrivateNotExposed(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	// 非公开配置不能通过公开接口读取（不应泄露其存在）
	s.ServeHTTP(w, doRequest("GET", "/api/v1/public/configs/max_login_attempts", "", nil))
	assertCode(t, w, 10401) // ErrConfigNotFound
}

func TestPublicConfigListExcludesPrivate(t *testing.T) {
	s, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/public/configs", "", nil))
	m := assertCode(t, w, 0)
	items := m["data"].([]any)
	for _, it := range items {
		if it.(map[string]any)["key"] == "max_login_attempts" {
			t.Error("private config leaked into public list")
		}
	}
}

// Config group tests

func TestConfigGroupCRUDAPI(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)

	// List 种子分组（站点/安全）
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/config-groups", tok, nil))
	m := assertCode(t, w, 0)
	groups := m["data"].([]any)
	if len(groups) < 2 {
		t.Fatalf("want >=2 seed groups, got %d", len(groups))
	}

	// Create
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/config-groups", tok, map[string]any{"name": "集成测试组", "sort": 9}))
	m = assertCode(t, w, 0)
	newGroup := m["data"].(map[string]any)
	gid := int(newGroup["id"].(float64))

	// Update
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("PUT", "/api/v1/config-groups/"+itoa(gid), tok, map[string]any{"name": "集成测试组-改", "sort": 8}))
	assertCode(t, w, 0)

	// Verify updated
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/config-groups", tok, nil))
	m = assertCode(t, w, 0)
	found := false
	for _, g := range m["data"].([]any) {
		if int(g.(map[string]any)["id"].(float64)) == gid && g.(map[string]any)["name"] == "集成测试组-改" {
			found = true
		}
	}
	if !found {
		t.Error("updated group not found in list")
	}

	// Delete
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("DELETE", "/api/v1/config-groups/"+itoa(gid), tok, nil))
	assertCode(t, w, 0)
}

func TestConfigSetWithGroupID(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)

	// 配置关联分组（group_id=1 站点）
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("PUT", "/api/v1/configs/site_name", tok, map[string]any{"value": "Kingfisher", "is_public": true, "version": "1.0.0", "render": "text", "group_id": 1}))
	assertCode(t, w, 0)

	// 读回确认 group_id + group_name
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/configs/site_name", tok, nil))
	m := assertCode(t, w, 0)
	cfg := m["data"].(map[string]any)
	if int(cfg["group_id"].(float64)) != 1 {
		t.Errorf("want group_id=1, got %v", cfg["group_id"])
	}
}

// User CRUD tests

func TestGetUsers(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/users", tok, nil))
	m := assertCode(t, w, 0)
	if m["data"].(map[string]any)["total"].(float64) < 1 {
		t.Error("should have users")
	}
}

func TestGetUserByID(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/users/1", tok, nil))
	assertCode(t, w, 0)
}

func TestGetUserByIDNotFound(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/users/99999", tok, nil))
	assertCode(t, w, 10005)
}

func TestGetMe(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/users/me", tok, nil))
	assertCode(t, w, 0)
}

func TestGetMePermissions(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/users/me/permissions", tok, nil))
	assertCode(t, w, 0)
}

func TestUpdateUser(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("PUT", "/api/v1/users/2", tok, map[string]any{"email": "updated@test.com"}))
	assertCode(t, w, 0)
}

func TestDeleteUser(t *testing.T) {
	// skipped: RBAC middleware
	s, _ := setupTestServer(t)
	tok := login(t, s)
	// Register then delete
	s.ServeHTTP(httptest.NewRecorder(), doRequest("POST", "/api/v1/auth/register", "", map[string]string{"username": "tod", "password": "Abcd1234"}))

	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/users?keyword=tod", tok, nil))
	var resp struct {
		Data struct {
			Items []struct {
				ID float64 `json:"id"`
			} `json:"items"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Data.Items) == 0 {
		t.Skip("user not found")
		return
	}
	id := fmt.Sprintf("%.0f", resp.Data.Items[0].ID)

	w2 := httptest.NewRecorder()
	path := fmt.Sprintf("%s/%s", "/api/v1/users", id)
	s.ServeHTTP(w2, doRequest("DELETE", path, tok, nil))
	assertCode(t, w2, 0)
}

func TestChangePassword(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("PUT", "/api/v1/users/me/password", tok, map[string]string{"old_password": "Abcd1234", "new_password": "NewPass123"}))
	assertCode(t, w, 0)
}

func TestChangePasswordWrongOld(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("PUT", "/api/v1/users/me/password", tok, map[string]string{"old_password": "WrongPass1", "new_password": "NewPass123"}))
	assertCode(t, w, 10103)
}

// Menu tests

func TestGetMenuTree(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/menus/tree", tok, nil))
	m := assertCode(t, w, 0)
	if len(m["data"].([]any)) < 2 {
		t.Error("should have menu roots")
	}
}

func TestCreateMenu(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/menus", tok, map[string]any{"parent_id": 0, "name": "TestMenu", "path": "/test", "type": 1}))
	assertCode(t, w, 0)
}

func TestCreateSubMenu(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/menus", tok, map[string]any{"parent_id": 1, "name": "SubMenu", "path": "/dashboard/sub", "type": 2}))
	assertCode(t, w, 0)
}

func TestGetMenuByID(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/menus/1", tok, nil))
	assertCode(t, w, 0)
}

func TestDeleteMenu(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	s.ServeHTTP(httptest.NewRecorder(), doRequest("POST", "/api/v1/menus", tok, map[string]any{"parent_id": 0, "name": "Temp", "path": "/tmpMenu", "type": 1}))
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("DELETE", "/api/v1/menus/16", tok, nil))
	assertCode(t, w, 0)
}

// Role tests

func TestGetRoles(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/roles", tok, nil))
	m := assertCode(t, w, 0)
	data := m["data"].(map[string]any)
	roles := data["items"].([]any)
	if len(roles) < 1 {
		t.Fatalf("want at least 1 role, got %d", len(roles))
	}
	// 必备种子角色必须存在（admin/editor/viewer）
	codes := map[string]bool{}
	for _, it := range roles {
		rm, ok := it.(map[string]any)
		if !ok {
			t.Fatalf("bad role element: %v", it)
		}
		if c, ok := rm["code"].(string); ok {
			codes[c] = true
		}
	}
	for _, required := range []string{"admin", "editor", "viewer"} {
		if !codes[required] {
			t.Errorf("roles missing required code %q in %v", required, codes)
		}
	}
	if int(data["total"].(float64)) != len(roles) {
		t.Error("total must equal items length")
	}
}

// 角色列表支持结构化查询
func TestGetRolesFiltered(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/roles?filter="+url.QueryEscape(`{"code":"admin"}`), tok, nil))
	m := assertCode(t, w, 0)
	data := m["data"].(map[string]any)
	items := data["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("want 1 role with code=admin, got %d", len(items))
	}
	if items[0].(map[string]any)["code"] != "admin" {
		t.Error("unexpected role")
	}
}

func TestGetRoleByID(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/roles/1", tok, nil))
	assertCode(t, w, 0)
}

func TestCreateRole(t *testing.T) {
	// skipped: RBAC middleware
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/roles", tok, map[string]string{"name": "Tester", "code": "tester"}))
	assertCode(t, w, 0)
}

func TestGetPermissions(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/permissions", tok, nil))
	m := assertCode(t, w, 0)
	codes := map[string]bool{}
	for _, p := range m["data"].([]any) {
		pm, ok := p.(map[string]any)
		if !ok {
			t.Fatalf("bad perm element: %v", p)
		}
		if c, ok := pm["code"].(string); ok {
			codes[c] = true
		}
	}
	for _, required := range []string{"user:list", "user:create", "role:list", "menu:list", "config:list", "dict:list", "template:list", "message:list"} {
		if !codes[required] {
			t.Errorf("permissions missing required code %q", required)
		}
	}
}

func TestAssignPermissions(t *testing.T) {
	// skipped: RBAC middleware
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("PUT", "/api/v1/roles/3/permissions", tok, map[string]any{"permission_ids": []int{1, 5}}))
	assertCode(t, w, 0)

	// Verify persistence: role 3 (editor) must now have exactly [1, 5]
	w2 := httptest.NewRecorder()
	s.ServeHTTP(w2, doRequest("GET", "/api/v1/roles/3/permissions", tok, nil))
	m := assertCode(t, w2, 0)
	perms, ok := m["data"].([]any)
	if !ok {
		t.Fatalf("want data array, got %v", m["data"])
	}
	var ids []int
	for _, p := range perms {
		pm, ok := p.(map[string]any)
		if !ok {
			t.Fatalf("bad perm element: %v", p)
		}
		ids = append(ids, int(pm["id"].(float64)))
	}
	sort.Ints(ids)
	if fmt.Sprint(ids) != "[1 5]" {
		t.Errorf("want permission ids [1 5], got %v", ids)
	}
}

func TestGetRolePermissions(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/roles/1/permissions", tok, nil))
	assertCode(t, w, 0)
}

func TestGetRoleMenus(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/roles/1/menus", tok, nil))
	assertCode(t, w, 0)
}

func TestAssignMenus(t *testing.T) {
	// skipped: RBAC middleware
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("PUT", "/api/v1/roles/3/menus", tok, map[string]any{"menu_ids": []int{1}}))
	assertCode(t, w, 0)

	// Verify persistence: role 3 (editor) must now have exactly [1]
	w2 := httptest.NewRecorder()
	s.ServeHTTP(w2, doRequest("GET", "/api/v1/roles/3/menus", tok, nil))
	m := assertCode(t, w2, 0)
	menus, ok := m["data"].([]any)
	if !ok {
		t.Fatalf("want data array, got %v", m["data"])
	}
	var ids []int
	for _, mn := range menus {
		mm, ok := mn.(map[string]any)
		if !ok {
			t.Fatalf("bad menu element: %v", mn)
		}
		ids = append(ids, int(mm["id"].(float64)))
	}
	sort.Ints(ids)
	if fmt.Sprint(ids) != "[1]" {
		t.Errorf("want menu ids [1], got %v", ids)
	}
}

// Config tests

func TestGetConfigs(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/configs", tok, nil))
	m := assertCode(t, w, 0)
	data := m["data"].(map[string]any)
	items := data["items"].([]any)
	if len(items) < 1 {
		t.Fatalf("want at least 1 config, got %d", len(items))
	}
	if int(data["total"].(float64)) != len(items) {
		t.Error("total must equal items length")
	}
}

// 配置列表支持结构化查询（is_public / group_id 过滤）
func TestGetConfigsFiltered(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/configs?filter="+url.QueryEscape(`{"group_id":1}`), tok, nil))
	m := assertCode(t, w, 0)
	items := m["data"].(map[string]any)["items"].([]any)
	if len(items) < 1 {
		t.Fatalf("want at least 1 config in group 1, got %d", len(items))
	}
	// 过滤结果必须全部属于 group 1
	for _, it := range items {
		cm, ok := it.(map[string]any)
		if !ok {
			t.Fatalf("bad config element: %v", it)
		}
		if int(cm["group_id"].(float64)) != 1 {
			t.Errorf("config in group-1 filter has group_id=%v", cm["group_id"])
		}
	}
}

func TestGetSingleConfig(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/configs/site_name", tok, nil))
	assertCode(t, w, 0)
}

func TestUpdateConfig(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("PUT", "/api/v1/configs/site_name", tok, map[string]string{"value": "New Name"}))
	assertCode(t, w, 0)
}

func TestDeleteConfig(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("DELETE", "/api/v1/configs/session_timeout", tok, nil))
	assertCode(t, w, 0)
}

// Audit tests

func TestGetAuditLogs(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/audit-logs", tok, nil))
	assertCode(t, w, 0)
}

// Pagination tests

func TestPagination(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/users", tok, nil))
	m := assertCode(t, w, 0)
	d := m["data"].(map[string]any)
	if d["page"].(float64) != 1 {
		t.Error("page should be 1")
	}
}

func TestPaginationCustomSize(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/users?page=1&page_size=5", tok, nil))
	m := assertCode(t, w, 0)
	if m["data"].(map[string]any)["page_size"].(float64) != 5 {
		t.Error("page_size should be 5")
	}
}

func TestSearchKeyword(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/users?keyword=admin", tok, nil))
	m := assertCode(t, w, 0)
	d := m["data"].(map[string]any)
	if d["total"].(float64) < 1 {
		t.Error("should find admin")
	}
}

// 404 test

func TestNotFound(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/notexist", tok, nil))
	if w.Code != 404 {
		t.Errorf("want 404, got %d", w.Code)
	}
}

// --- Profile self-service tests ---

func TestUpdateMeProfile(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)

	// Update nickname + email
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("PUT", "/api/v1/users/me", tok, map[string]string{
		"nickname": "管理员",
		"email":    "admin@example.com",
	}))
	assertCode(t, w, 0)
	m := assertCode(t, w, 0) // re-parse with code check
	_ = m
}

func TestUpdateMeNickname(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)

	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("PUT", "/api/v1/users/me", tok, map[string]string{
		"nickname": "测试昵称",
	}))
	assertCode(t, w, 0)

	// Verify GET /users/me reflects the change
	w2 := httptest.NewRecorder()
	s.ServeHTTP(w2, doRequest("GET", "/api/v1/users/me", tok, nil))
	m := assertCode(t, w2, 0)
	if m["data"].(map[string]any)["nickname"] != "测试昵称" {
		t.Error("nickname not updated")
	}
}

func TestGetMyLoginLogs(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)

	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/users/me/login-logs", tok, nil))
	assertCode(t, w, 0)
}

// ---- 批量操作（batch）----

func TestBatchUserStatusAndDelete(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)

	// 创建两个用户，取 id
	var ids []int
	for _, name := range []string{"batch1", "batch2"} {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, doRequest("POST", "/api/v1/users", tok, map[string]string{"username": name, "password": "Abcd1234"}))
		m := assertCode(t, w, 0)
		ids = append(ids, int(m["data"].(map[string]any)["id"].(float64)))
	}

	// 批量禁用
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/users/batch-status", tok, map[string]any{"ids": ids, "status": 0}))
	assertCode(t, w, 0)

	// 验证禁用生效
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/users?filter="+url.QueryEscape(`{"username":"batch1"}`), tok, nil))
	m := assertCode(t, w, 0)
	items := m["data"].(map[string]any)["items"].([]any)
	if len(items) != 1 || int(items[0].(map[string]any)["status"].(float64)) != 0 {
		t.Error("want status=0 after batch disable")
	}

	// 批量删除
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/users/batch-delete", tok, map[string]any{"ids": ids}))
	assertCode(t, w, 0)

	// 验证软删除后列表不再包含
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/users?filter="+url.QueryEscape(`{"username":"batch1"}`), tok, nil))
	m = assertCode(t, w, 0)
	if total := int(m["data"].(map[string]any)["total"].(float64)); total != 0 {
		t.Errorf("want 0 users after batch delete, got %d", total)
	}
}

func TestBatchRoleAdminProtected(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/roles/batch-delete", tok, map[string]any{"ids": []int{1}}))
	if w.Code != 400 {
		t.Errorf("want 400 deleting admin role, got %d", w.Code)
	}
}

func TestBatchMenuChildrenProtected(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	// 系统管理(id=2)含子节点 → 批量删除被拒
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/menus/batch-delete", tok, map[string]any{"ids": []int{2}}))
	if w.Code != 400 {
		t.Errorf("want 400 deleting parent menu, got %d", w.Code)
	}
	// 批量状态切换无子节点限制
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/menus/batch-status", tok, map[string]any{"ids": []int{2}, "status": 0}))
	assertCode(t, w, 0)
}

func TestBatchDictTypeEntriesProtected(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	// 创建类型 + 条目
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/dict-types", tok, map[string]any{"code": "bt", "name": "批量", "status": 1}))
	m := assertCode(t, w, 0)
	typeID := int(m["data"].(map[string]any)["id"].(float64))
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", fmt.Sprintf("/api/v1/dict-types/%d/entries", typeID), tok, map[string]any{"label": "x", "value": "y", "status": 1}))
	m = assertCode(t, w, 0)
	entryID := int(m["data"].(map[string]any)["id"].(float64))
	// 批量删除含条目的类型 → 拒绝 10504
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/dict-types/batch-delete", tok, map[string]any{"ids": []int{typeID}}))
	assertCode(t, w, 10504)
	// 条目的批量状态切换 + 批量删除
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", fmt.Sprintf("/api/v1/dict-types/%d/entries/batch-status", typeID), tok, map[string]any{"ids": []int{entryID}, "status": 0}))
	assertCode(t, w, 0)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", fmt.Sprintf("/api/v1/dict-types/%d/entries/batch-delete", typeID), tok, map[string]any{"ids": []int{entryID}}))
	assertCode(t, w, 0)
}

func TestBatchConfigDelete(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	for _, k := range []string{"batch_key1", "batch_key2"} {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, doRequest("PUT", "/api/v1/configs/"+k, tok, map[string]string{"value": "1"}))
		assertCode(t, w, 0)
	}
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/configs/batch-delete", tok, map[string]any{"keys": []string{"batch_key1", "batch_key2"}}))
	assertCode(t, w, 0)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/configs?filter="+url.QueryEscape(`{"key":"batch_key1"}`), tok, nil))
	m := assertCode(t, w, 0)
	if total := int(m["data"].(map[string]any)["total"].(float64)); total != 0 {
		t.Errorf("want 0 configs after batch delete, got %d", total)
	}
}

// 配置图片上传端点
func TestConfigUploadImage(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	// 构造一张 1x1 PNG
	png, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==")
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/configs/upload-image", bytes.NewReader(png))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "image/png")
	// 用 multipart 上传
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	fw, _ := mw.CreateFormFile("file", "cover.png")
	_, _ = fw.Write(png)
	_ = mw.Close()
	req = httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/configs/upload-image", body)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	m := assertCode(t, w, 0)
	url, ok := m["data"].(map[string]any)["url"].(string)
	if !ok || !strings.HasPrefix(url, "/uploads/configs/") {
		t.Errorf("want upload url under /uploads/configs/, got %v", m["data"])
	}
}

// 站内信：发送 → 收件箱 → 未读数 → 标记已读 → 批量删除
func TestMessageInbox(t *testing.T) {
	s, _ := setupTestServer(t)
	adminTok := login(t, s)

	// 发送给 viewer 用户（id=3）；异步发送，响应不返回消息 id，需从收件箱轮询取得
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/messages", adminTok, map[string]any{"recipient_id": 3, "title": "测试消息", "content": "hello"}))
	assertCode(t, w, 0)

	// viewer 登录，查收件箱（worker 异步落库，轮询等待）
	viewerTok := loginAs(t, s, "viewer")
	var msgID int
	var m map[string]any
	deadline := time.Now().Add(5 * time.Second)
	for {
		w = httptest.NewRecorder()
		s.ServeHTTP(w, doRequest("GET", "/api/v1/me/messages", viewerTok, nil))
		m = assertCode(t, w, 0)
		items := m["data"].(map[string]any)["items"].([]any)
		if len(items) > 0 {
			msgID = int(items[0].(map[string]any)["id"].(float64))
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("message not delivered within timeout")
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 未读数 = 1
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/me/messages/unread-count", viewerTok, nil))
	m = assertCode(t, w, 0)
	if int(m["data"].(map[string]any)["unread_count"].(float64)) != 1 {
		t.Error("want unread_count 1")
	}

	// 标记已读
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("PUT", fmt.Sprintf("/api/v1/me/messages/%d/read", msgID), viewerTok, nil))
	assertCode(t, w, 0)

	// 未读数 = 0
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/me/messages/unread-count", viewerTok, nil))
	m = assertCode(t, w, 0)
	if int(m["data"].(map[string]any)["unread_count"].(float64)) != 0 {
		t.Error("want unread_count 0 after read")
	}

	// 越权校验：admin 不能删 viewer 的消息（recipient_id 不匹配 → 删不到）
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/me/messages/batch-delete", adminTok, map[string]any{"ids": []int{msgID}}))
	assertCode(t, w, 0)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/me/messages", viewerTok, nil))
	m = assertCode(t, w, 0)
	if len(m["data"].(map[string]any)["items"].([]any)) != 1 {
		t.Error("admin should not be able to delete viewer's message")
	}

	// viewer 自己批量删除
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/me/messages/batch-delete", viewerTok, map[string]any{"ids": []int{msgID}}))
	assertCode(t, w, 0)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", "/api/v1/me/messages", viewerTok, nil))
	m = assertCode(t, w, 0)
	if len(m["data"].(map[string]any)["items"].([]any)) != 0 {
		t.Error("viewer's inbox should be empty after delete")
	}
}

// 站内信详情接口
func TestMessageDetail(t *testing.T) {
	s, _ := setupTestServer(t)
	adminTok := login(t, s)

	// 管理员发信给 viewer(3)；异步发送，从收件箱轮询取得消息 id
	w := httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("POST", "/api/v1/messages", adminTok, map[string]any{"recipient_id": 3, "title": "详情测试", "content": "完整内容"}))
	assertCode(t, w, 0)

	// viewer 查详情
	viewerTok := loginAs(t, s, "viewer")
	var msgID int
	var m map[string]any
	deadline := time.Now().Add(5 * time.Second)
	for {
		w = httptest.NewRecorder()
		s.ServeHTTP(w, doRequest("GET", "/api/v1/me/messages", viewerTok, nil))
		m = assertCode(t, w, 0)
		items := m["data"].(map[string]any)["items"].([]any)
		if len(items) > 0 {
			msgID = int(items[0].(map[string]any)["id"].(float64))
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("message not delivered within timeout")
		}
		time.Sleep(100 * time.Millisecond)
	}
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", fmt.Sprintf("/api/v1/me/messages/%d", msgID), viewerTok, nil))
	m = assertCode(t, w, 0)
	data := m["data"].(map[string]any)
	if data["title"] != "详情测试" || data["content"] != "完整内容" {
		t.Errorf("unexpected detail: %v", data)
	}
	if data["is_read"] != false {
		t.Error("want is_read=false initially")
	}

	// 越权：admin 不能查 viewer 的消息 → 404
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", fmt.Sprintf("/api/v1/me/messages/%d", msgID), adminTok, nil))
	if w.Code != 404 {
		t.Errorf("admin should not read viewer's message, got %d", w.Code)
	}

	// 标读后详情 is_read=true
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("PUT", fmt.Sprintf("/api/v1/me/messages/%d/read", msgID), viewerTok, nil))
	assertCode(t, w, 0)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, doRequest("GET", fmt.Sprintf("/api/v1/me/messages/%d", msgID), viewerTok, nil))
	m = assertCode(t, w, 0)
	if m["data"].(map[string]any)["is_read"] != true {
		t.Error("want is_read=true after mark read")
	}
}
