package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// callAPIName 唯一的工具名：类似 curl 的通用 API 调用工具。
const callAPIName = "call_api"

// maxResponseSize 限制内部请求响应体大小，防止超大返回撑爆上下文。
const maxResponseSize = 1 << 20 // 1MB

// defaultAllowedMethods 默认允许的 HTTP 方法白名单（可在构造时通过 SetAllowedMethods 覆盖）。
var defaultAllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

// CallAPIClient call_api 工具的执行器：构造内部 HTTP 请求打到本服务自身端点，
// 透传当前登录用户的 token，走 Auth+RBAC 中间件完成鉴权。
type CallAPIClient struct {
	selfBaseURL    string
	httpClient     *http.Client
	allowedMethods []string // HTTP 方法白名单（可配置，控制 agent 可用哪些写方法）
}

// NewCallAPIClient 创建 call_api 执行器。selfBaseURL 为本服务自身地址
// （如 http://127.0.0.1:8080），不能由 LLM 输入控制。
func NewCallAPIClient(selfBaseURL string) *CallAPIClient {
	return &CallAPIClient{
		selfBaseURL:    strings.TrimRight(selfBaseURL, "/"),
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		allowedMethods: defaultAllowedMethods,
	}
}

// SetAllowedMethods 覆盖 HTTP 方法白名单（空切片则全部禁止）。
func (c *CallAPIClient) SetAllowedMethods(methods []string) {
	c.allowedMethods = methods
}

// AllowedMethods 返回当前白名单。
func (c *CallAPIClient) AllowedMethods() []string { return c.allowedMethods }

// Tool 返回 call_api 工具定义（MCP 格式）。method 枚举随白名单动态生成。
func (c *CallAPIClient) Tool() Tool {
	methods := c.allowedMethods
	if len(methods) == 0 {
		methods = defaultAllowedMethods
	}
	return Tool{
		Name: callAPIName,
		Description: "调用本系统的 OpenAPI 接口。method 为 HTTP 方法（" + strings.Join(methods, "/") + "）；" +
			"path 必须是系统接口路径（以 /api/v1 开头，如 /api/v1/users，路径参数直接填入如 /api/v1/users/3）；" +
			"query 为查询参数对象（可选）；body 为请求体对象（可选，写操作使用）。" +
			"返回 {status_code, body}。注意：你只能调用 system prompt 中列出的接口，无权限的接口会返回 403。",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"method": map[string]any{"type": "string", "enum": methods, "description": "HTTP 方法"},
				"path":   map[string]any{"type": "string", "description": "系统接口路径，以 /api/v1 开头"},
				"query":  map[string]any{"type": "object", "description": "查询参数（可选）"},
				"body":   map[string]any{"type": "object", "description": "请求体（可选，写操作使用）"},
			},
			"required": []string{"method", "path"},
		},
	}
}

// Call 执行一次内部 API 调用。userToken 为当前用户的 Bearer token。
func (c *CallAPIClient) Call(ctx context.Context, args map[string]any, userToken string) (ToolResult, error) {
	method, _ := args["method"].(string)
	path, _ := args["path"].(string)
	method = strings.ToUpper(method)

	if !containsStr(c.allowedMethods, method) {
		return errResult(fmt.Sprintf("非法 method: %s", method)), nil
	}
	if !strings.HasPrefix(path, "/api/v1/") {
		// 防 SSRF：只允许访问本系统 OpenAPI 端点，拒绝外部任意 URL。
		return errResult(fmt.Sprintf("path 必须以 /api/v1 开头（只能调用系统接口），收到: %s", path)), nil
	}

	target, err := url.Parse(c.selfBaseURL + path)
	if err != nil {
		return errResult(fmt.Sprintf("非法 path: %v", err)), nil
	}

	// query 参数
	if q, ok := args["query"].(map[string]any); ok && len(q) > 0 {
		vals := url.Values{}
		for k, v := range q {
			vals.Set(k, fmt.Sprint(v))
		}
		target.RawQuery = vals.Encode()
	}

	var body io.Reader
	if b, ok := args["body"].(map[string]any); ok && len(b) > 0 {
		data, err := json.Marshal(b)
		if err != nil {
			return errResult(fmt.Sprintf("body 序列化失败: %v", err)), nil
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, target.String(), body)
	if err != nil {
		return errResult(fmt.Sprintf("构造请求失败: %v", err)), nil
	}
	req.Header.Set("Content-Type", "application/json")
	// 透传当前登录用户 token → 经 Auth+RBAC 中间件鉴权，权限天然生效。
	if userToken != "" {
		req.Header.Set("Authorization", userToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errResult(fmt.Sprintf("请求失败: %v", err)), nil
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return errResult(fmt.Sprintf("读取响应失败: %v", err)), nil
	}

	out := map[string]any{
		"status_code": resp.StatusCode,
		"body":        string(data),
	}
	b, _ := json.Marshal(out)
	return ToolResult{Content: []ContentBlock{{Type: "text", Text: string(b)}}}, nil
}

func containsStr(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func errResult(msg string) ToolResult {
	return ToolResult{Content: []ContentBlock{{Type: "text", Text: msg}}, IsError: true}
}
