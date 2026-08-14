package mcp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEndpointListFromSpec(t *testing.T) {
	// 假 swagger doc.json，包含一个受保护端点和公开端点。
	spec := `{
	  "paths": {
	    "/users": {"get": {"summary": "用户列表（分页）", "tags": ["User"]}},
	    "/auth/login": {"post": {"summary": "用户登录", "tags": ["Auth"]}},
	    "/users/{id}": {"delete": {"summary": "删除用户"}}
	  }
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(spec))
	}))
	defer srv.Close()

	loader := NewSpecLoader(srv.URL)
	list, err := loader.EndpointList()
	if err != nil {
		t.Fatalf("EndpointList 失败: %v", err)
	}
	for _, want := range []string{
		"GET /api/v1/users - 用户列表（分页）",
		"POST /api/v1/auth/login - 用户登录",
		"DELETE /api/v1/users/{id} - 删除用户",
	} {
		if !strings.Contains(list, want) {
			t.Fatalf("端点清单应包含 %q，实际:\n%s", want, list)
		}
	}
}

func TestEndpointListCaches(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"paths":{"/users":{"get":{"summary":"x"}}}}`))
	}))
	defer srv.Close()

	loader := NewSpecLoader(srv.URL)
	_, _ = loader.EndpointList()
	_, _ = loader.EndpointList()
	if hits != 1 {
		t.Fatalf("30s TTL 内第二次调用应命中缓存，实际拉取 %d 次", hits)
	}
}
