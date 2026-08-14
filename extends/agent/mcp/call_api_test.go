package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestServer 启动一个只记录请求的假后端，返回 server + 最后收到的 path。
func newTestServer(t *testing.T) (*httptest.Server, *string) {
	t.Helper()
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		// 验证透传了 Authorization
		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"ok":true}}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &gotPath
}

func TestCallAPIAllowsInternalPath(t *testing.T) {
	srv, gotPath := newTestServer(t)
	c := NewCallAPIClient(srv.URL)

	res, err := c.Call(context.Background(), map[string]any{
		"method": "GET",
		"path":   "/api/v1/users",
	}, "Bearer token123")
	if err != nil {
		t.Fatalf("Call 失败: %v", err)
	}
	if res.IsError {
		t.Fatalf("不应为错误结果: %v", res.Content)
	}
	if *gotPath != "/api/v1/users" {
		t.Fatalf("内部请求路径应为 /api/v1/users，得到 %q", *gotPath)
	}
	// 结果应含 status_code 与 body
	if len(res.Content) == 0 || !strings.Contains(res.Content[0].Text, `"status_code":200`) {
		t.Fatalf("结果应包含 status_code，得到 %v", res.Content)
	}
}

func TestCallAPIRejectsExternalURL(t *testing.T) {
	// 防 SSRF：path 不以 /api/v1 开头的一律拒绝，不能访问外部任意 URL。
	c := NewCallAPIClient("http://127.0.0.1:9")
	res, _ := c.Call(context.Background(), map[string]any{
		"method": "GET",
		"path":   "http://evil.example.com/steal",
	}, "")
	if !res.IsError {
		t.Fatal("外部 URL 应被拒绝（防 SSRF）")
	}
	// 绝对 URL 形式也应拒绝
	res, _ = c.Call(context.Background(), map[string]any{
		"method": "GET",
		"path":   "//evil.example.com/steal",
	}, "")
	if !res.IsError {
		t.Fatal("协议相对 URL 应被拒绝（防 SSRF）")
	}
}

func TestCallAPIRejectsBadMethod(t *testing.T) {
	c := NewCallAPIClient("http://127.0.0.1:9")
	res, _ := c.Call(context.Background(), map[string]any{
		"method": "TRACE",
		"path":   "/api/v1/users",
	}, "")
	if !res.IsError {
		t.Fatal("TRACE 方法应被拒绝")
	}
}

func TestCallAPIWithQueryAndBody(t *testing.T) {
	var gotQuery, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		b := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(b)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	c := NewCallAPIClient(srv.URL)

	_, err := c.Call(context.Background(), map[string]any{
		"method": "POST",
		"path":   "/api/v1/messages",
		"query":  map[string]any{"page": "1", "status": "active"},
		"body":   map[string]any{"title": "测试"},
	}, "")
	if err != nil {
		t.Fatalf("Call 失败: %v", err)
	}
	if !strings.Contains(gotQuery, "page=1") || !strings.Contains(gotQuery, "status=active") {
		t.Fatalf("query 未正确编码: %q", gotQuery)
	}
	// body 应为 JSON
	var parsed map[string]any
	if err := json.Unmarshal([]byte(gotBody), &parsed); err != nil || parsed["title"] != "测试" {
		t.Fatalf("body 未正确序列化: %q", gotBody)
	}
}

func TestToolDefinitionShape(t *testing.T) {
	c := NewCallAPIClient("http://127.0.0.1:9")
	tool := c.Tool()
	if tool.Name != "call_api" {
		t.Fatalf("工具名应为 call_api，得到 %q", tool.Name)
	}
	if tool.InputSchema == nil || tool.InputSchema["type"] != "object" {
		t.Fatalf("input_schema 应为 object JSON Schema，得到 %v", tool.InputSchema)
	}
	req, ok := tool.InputSchema["required"].([]string)
	if !ok || len(req) != 2 {
		t.Fatalf("required 应含 method+path，得到 %v", tool.InputSchema["required"])
	}
}
