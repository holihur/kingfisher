// Package llm 实现 Anthropic 兼容格式的 LLM 客户端（零第三方依赖）。
//
// 直接以标准库 net/http + encoding/json 调用 {base_url}/v1/messages 端点，
// 请求/响应均为 anthropic Messages API 格式（DeepSeek 提供兼容端点）。
package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Message 一条对话消息（anthropic 格式）。Content 为字符串（纯文本）
// 或 content block 数组（text / tool_use / tool_result）。
type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// ToolUse assistant 请求的工具调用。
type ToolUse struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

// Tool 工具定义（anthropic 格式）。
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// Params 一次 chat 请求的参数。
type Params struct {
	Model     string
	MaxTokens int
	System    string
	Messages  []Message
	Tools     []Tool
}

// Result 一次 chat 调用的结果。
type Result struct {
	StopReason string
	Text       string
	ToolUses   []ToolUse
}

// OnDelta 流式文本增量回调（用于推送 SSE 打字机效果）；返回非 nil error 中断流。
type OnDelta func(delta string) error

// Client anthropic 兼容端点客户端。
type Client struct {
	baseURL   string
	apiKey    string
	model     string
	maxTokens int
	http      *http.Client
}

// NewClient 创建 LLM 客户端。
func NewClient(baseURL, apiKey, model string, maxTokens int) *Client {
	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		http:      &http.Client{Timeout: 120 * time.Second},
	}
}

// SetAPIKey 运行时更换 API key（支持配置热更新）。
func (c *Client) SetAPIKey(key string) { c.apiKey = key }

// Model 返回模型名。
func (c *Client) Model() string { return c.model }

// MaxTokens 返回单次最大输出 token 数。
func (c *Client) MaxTokens() int { return c.maxTokens }

// chatReq 请求体（anthropic /v1/messages 格式）。
type chatReq struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Stream    bool      `json:"stream"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
	Tools     []Tool    `json:"tools,omitempty"`
}

// Chat 发起流式对话，文本增量经 onDelta 实时回调，返回完整结果。
func (c *Client) Chat(ctx context.Context, p Params, onDelta OnDelta) (*Result, error) {
	req := chatReq{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Stream:    true,
		System:    p.System,
		Messages:  p.Messages,
	}
	if len(p.Tools) > 0 {
		req.Tools = p.Tools
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("accept", "text/event-stream")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("LLM 请求失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("LLM 返回状态 %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	return c.parseStream(ctx, resp.Body, onDelta)
}

// parseStream 逐行解析 SSE 事件流。
//
// 支持的事件：
//   - content_block_start：text / tool_use 块开始（tool_use 带 id/name）
//   - content_block_delta：delta.type=text_delta（文本增量）或 input_json_delta（tool_use 输入增量）
//   - message_delta：delta.stop_reason
func (c *Client) parseStream(ctx context.Context, body io.Reader, onDelta OnDelta) (*Result, error) {
	res := &Result{}
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	// 当前 tool_use 块（按 index 定位）+ 其 input 的 partial_json 累积。
	type pendingTool struct {
		idx          int
		use          ToolUse
		partialInput string
	}
	var cur *pendingTool
	toolsByIndex := map[int]*pendingTool{}

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		line := scanner.Text()
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue // 忽略 event: 行等
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var evt sseEvent
		if err := json.Unmarshal([]byte(data), &evt); err != nil {
			continue
		}
		switch evt.Type {
		case "content_block_start":
			if evt.ContentBlock != nil && evt.ContentBlock.Type == "tool_use" {
				cur = &pendingTool{
					idx: evt.Index,
					use: ToolUse{ID: evt.ContentBlock.ID, Name: evt.ContentBlock.Name},
				}
				toolsByIndex[evt.Index] = cur
			} else {
				cur = nil
			}
		case "content_block_delta":
			if evt.Delta == nil {
				continue
			}
			switch evt.Delta.Type {
			case "text_delta":
				res.Text += evt.Delta.Text
				if onDelta != nil {
					if err := onDelta(evt.Delta.Text); err != nil {
						return nil, err
					}
				}
			case "input_json_delta":
				// 累积 tool_use 的输入 JSON 增量。
				if cur != nil {
					cur.partialInput += evt.Delta.PartialJSON
				}
			}
		case "content_block_stop":
			// 块结束：若当前是 tool_use，解析累积的 input。
			if cur != nil {
				var input map[string]any
				if err := json.Unmarshal([]byte(cur.partialInput), &input); err == nil && input != nil {
					cur.use.Input = input
				} else {
					cur.use.Input = map[string]any{}
				}
				res.ToolUses = append(res.ToolUses, cur.use)
				cur = nil
			}
		case "message_delta":
			if evt.Delta != nil && evt.Delta.StopReason != "" {
				res.StopReason = evt.Delta.StopReason
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取流式响应失败: %w", err)
	}
	if res.StopReason == "" {
		res.StopReason = "end_turn"
	}
	return res, nil
}

// sseEvent anthropic 流式事件（只解析需要的字段）。
type sseEvent struct {
	Type         string           `json:"type"`
	Index        int              `json:"index"`
	Delta        *sseDelta        `json:"delta"`
	ContentBlock *sseContentBlock `json:"content_block"`
}

type sseDelta struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	PartialJSON string `json:"partial_json"`
	StopReason  string `json:"stop_reason"`
}

type sseContentBlock struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}
