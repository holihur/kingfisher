package mcp

import "context"

// Client 进程内 MCP client：组合 call_api 工具与 OpenAPI 端点清单，
// 实现 Executor 接口。方法签名对齐 MCP client 语义（ListTools/CallTool）。
type Client struct {
	caller *CallAPIClient
	spec   *SpecLoader
}

// NewClient 创建进程内 MCP client。
func NewClient(selfBaseURL string) *Client {
	return &Client{
		caller: NewCallAPIClient(selfBaseURL),
		spec:   NewSpecLoader(selfBaseURL),
	}
}

// ListTools 返回可用工具（当前只有 call_api）。
func (c *Client) ListTools() ([]Tool, error) {
	return []Tool{c.caller.Tool()}, nil
}

// CallTool 执行工具。当前仅支持 call_api，透传 userToken 鉴权。
func (c *Client) CallTool(ctx context.Context, name string, args map[string]any, userToken string) (ToolResult, error) {
	if name != callAPIName {
		return ToolResult{Content: []ContentBlock{{Type: "text", Text: "未知工具: " + name}}, IsError: true}, nil
	}
	return c.caller.Call(ctx, args, userToken)
}

// EndpointList 返回系统 OpenAPI 端点清单（供 LLM system prompt 注入）。
func (c *Client) EndpointList() (string, error) {
	return c.spec.EndpointList()
}
