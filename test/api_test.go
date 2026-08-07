package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"

	"kingfisher/core/config"
	"kingfisher/core/database"
	"kingfisher/core/jwt"
	"kingfisher/core/logger"
	"kingfisher/core/router"

	auditTransport "kingfisher/extends/audit/transport"
	configTransport "kingfisher/extends/config/transport"
	menuTransport "kingfisher/extends/menu/transport"
	rbacTransport "kingfisher/extends/rbac/transport"
	userTransport "kingfisher/extends/user/transport"
)

func setupTestServer(t *testing.T) (*gin.Engine, *jwt.JWTManager) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-secret-abc123")

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

	authMw := rbacTransport.AuthMiddleware(jwtMgr)
	allPerms := map[string]bool{
		"user:list": true, "user:create": true, "user:update": true, "user:delete": true,
		"menu:list": true, "menu:create": true, "menu:update": true, "menu:delete": true,
		"role:list": true, "role:create": true, "role:update": true, "role:delete": true,
		"config:list": true, "config:update": true, "audit:list": true,
	}
	rbacMw := func(c *gin.Context) {
		c.Set("permissions", allPerms)
		c.Next()
	}

	mods := []router.Module{
		userTransport.NewUserModule(db, nil, jwtMgr, nil),
		rbacTransport.NewRBACModule(db, nil),
		menuTransport.NewMenuModule(db, nil),
		configTransport.NewConfigModule(db, nil),
		auditTransport.NewAuditModule(db),
	}
	for _, m := range mods {
		_ = m.Init(t.Context())
		router.Register(r, m, authMw, rbacMw)
	}
	return r, jwtMgr
}

func itoa(n int) string { return strconv.Itoa(n) }

func doRequest(method, path, token string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
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

func TestLoginSuccess(t *testing.T) {
	s, _ := setupTestServer(t)
	tok := login(t, s)
	if tok == "" {
		t.Error("empty token")
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
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString("not json"))
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
	if len(items) != 3 {
		t.Fatalf("want 3 public configs (site_name/site_logo/site_description), got %d", len(items))
	}
	for _, it := range items {
		cfg := it.(map[string]any)
		if cfg["is_public"] != true {
			t.Errorf("config %v should be public", cfg["key"])
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
	if len(data["items"].([]any)) != 3 {
		t.Error("want 3 roles")
	}
	if int(data["total"].(float64)) != 3 {
		t.Error("want total 3")
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
	if len(m["data"].([]any)) != 19 {
		t.Error("want 19 perms")
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
	if len(data["items"].([]any)) != 6 {
		t.Error("want 6 configs")
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
	if len(items) != 3 {
		t.Fatalf("want 3 configs in group 1, got %d", len(items))
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
