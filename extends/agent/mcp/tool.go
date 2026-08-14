// Package mcp 自研的 MCP 兼容工具层（零第三方依赖）。
//
// 工具定义与调用结果的格式对齐 MCP 规范：每个工具含
// name / description / input_schema（JSON Schema），调用返回
// {content:[{type:"text",text:"..."}]}。后期如需接入外部标准 MCP server，
// 本包的 Executor 接口作为适配点，agent 核心无需改动。
package mcp

import "context"

// Tool 一个工具的元数据（对齐 MCP tools/list 的元素）。
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// ContentBlock 工具结果中的一个内容块（对齐 MCP CallToolResult.content）。
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToolResult 工具调用结果（对齐 MCP CallToolResult）。
type ToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"is_error"`
}

// Executor Agent 端工具执行器接口，方法签名对齐 MCP client 语义：
// ListTools 获取可用工具；CallTool 执行工具调用。
type Executor interface {
	ListTools() ([]Tool, error)
	// CallTool 执行工具。userToken 为当前登录用户的 Bearer token，
	// 用于内部请求时透传鉴权（RBAC 兜底）。
	CallTool(ctx context.Context, name string, args map[string]any, userToken string) (ToolResult, error)
}
