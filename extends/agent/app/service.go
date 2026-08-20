package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"kingfisher/core/errcode"
	"kingfisher/extends/agent/domain"
	"kingfisher/extends/agent/llm"
	"kingfisher/extends/agent/mcp"
	"kingfisher/extends/agent/port"
)

// SystemPromptFallback Agent 的默认系统提示词（后台配置 agent_system_prompt 留空时使用）。
// 端点清单由 service 动态拼接。
const SystemPromptFallback = `你是 Kingfisher 后台管理系统的 AI 助手，可以通过调用系统接口帮助用户查询和操作系统。
请始终用中文回答，简洁、准确。当需要查询系统数据（如用户、配置、审计、字典等）时，
使用 call_api 工具调用系统接口；接口返回后根据真实数据回答，不要编造。
如果接口返回 403 说明当前用户无权限，请如实告知。`

// Error 携带 errcode 的错误类型，handler 层据此映射 HTTP 错误码。
// Detail 为底层错误信息（可选），随 SSE error 事件透出便于排查。
type Error struct {
	Code   int
	Detail string
}

func (e *Error) Error() string {
	if e.Detail != "" {
		return errcode.Msg(e.Code) + ": " + e.Detail
	}
	return errcode.Msg(e.Code)
}

// AgentService Agent 聊天核心：会话管理 + LLM 工具循环 + SSE 事件生成。
type AgentService struct {
	repo     port.AgentRepository
	llm      *llm.Client
	mcp      *mcp.Client
	system   func(ctx context.Context) (string, error) // system prompt 前缀（不含端点清单；运行时读取可覆盖）
	apiKey   func(ctx context.Context) (string, error)
	enabled  func() bool
	selfHost string
}

// NewAgentService 创建 AgentService。
//
// apiKey 返回 LLM API key（从系统配置/环境变量解析）；enabled 返回模块是否启用；
// system 返回系统提示词（可为空，ChatStream 时回退默认）。
func NewAgentService(
	repo port.AgentRepository,
	llmClient *llm.Client,
	mcpClient *mcp.Client,
	system func(ctx context.Context) (string, error),
	apiKey func(ctx context.Context) (string, error),
	enabled func() bool,
	selfHost string,
) *AgentService {
	return &AgentService{
		repo:     repo,
		llm:      llmClient,
		mcp:      mcpClient,
		system:   system,
		apiKey:   apiKey,
		enabled:  enabled,
		selfHost: selfHost,
	}
}

// ---- 会话管理 ----

// CheckEnabled 校验 agent 是否启用。
func (s *AgentService) CheckEnabled() error {
	if s.enabled != nil && !s.enabled() {
		return &Error{Code: errcode.ErrAgentDisabled}
	}
	return nil
}

// ListConversations 当前用户会话列表。
func (s *AgentService) ListConversations(ctx context.Context, userID uint) ([]domain.Conversation, error) {
	return s.repo.ListConversations(ctx, userID)
}

// CreateConversation 创建会话，标题取首条消息前 N 字（可后续更新）。
func (s *AgentService) CreateConversation(ctx context.Context, userID uint, title string) (*domain.Conversation, error) {
	if title == "" {
		title = "新会话"
	}
	return s.repo.CreateConversation(ctx, userID, title)
}

// GetConversation 校验归属并返回会话。
func (s *AgentService) GetConversation(ctx context.Context, id, userID uint) (*domain.Conversation, error) {
	c, err := s.repo.GetConversation(ctx, id)
	if err != nil {
		return nil, &Error{Code: errcode.ErrAgentConversationNF}
	}
	if c.UserID != userID {
		return nil, &Error{Code: errcode.ErrAgentConversationNF}
	}
	return c, nil
}

// DeleteConversation 删除会话。
func (s *AgentService) DeleteConversation(ctx context.Context, id, userID uint) error {
	if _, err := s.GetConversation(ctx, id, userID); err != nil {
		return err
	}
	return s.repo.DeleteConversation(ctx, id, userID)
}

// ListMessages 某会话的消息历史。
func (s *AgentService) ListMessages(ctx context.Context, conversationID, userID uint) ([]domain.Message, error) {
	if _, err := s.GetConversation(ctx, conversationID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListMessages(ctx, conversationID)
}

// ---- 对话（LLM 工具循环 + SSE） ----

// Emit 聊天事件输出回调（transport 层写 SSE）。
type Emit func(evt SSEEvent)

// ChatStream 流式对话：在会话中追加用户消息，跑 LLM 工具循环，
// 逐段文本/工具事件经 emit 推送，结束后持久化 assistant 消息。
func (s *AgentService) ChatStream(ctx context.Context, conversationID, userID uint, content, userToken string, emit Emit) error {
	if s.enabled != nil && !s.enabled() {
		return &Error{Code: errcode.ErrAgentDisabled}
	}
	c, err := s.GetConversation(ctx, conversationID, userID)
	if err != nil {
		return err
	}

	key, err := s.apiKey(ctx)
	if err != nil || key == "" {
		return &Error{Code: errcode.ErrAgentNoAPIKey}
	}
	// 每次会话刷新 API key（支持运行中更换）。
	s.llm.SetAPIKey(key)

	// 追加用户消息。
	userMsg := &domain.Message{ConversationID: conversationID, Role: "user", Content: content}
	if err := s.repo.AddMessage(ctx, userMsg); err != nil {
		return err
	}
	// 首条消息时更新会话标题。
	if c.Title == "" || c.Title == "新会话" {
		title := content
		if r := []rune(title); len(r) > 20 {
			title = string(r[:20])
		}
		_ = s.repo.RenameConversation(ctx, conversationID, title)
	}

	// 重建历史 + 追加当前消息。
	messages, err := s.rebuildMessages(ctx, conversationID)
	if err != nil {
		return err
	}
	messages = append(messages, llm.Message{Role: "user", Content: content})

	// system prompt：运行时读取（可被后台配置覆盖）→ 前缀 + 端点清单。
	baseSystem := SystemPromptFallback
	if s.system != nil {
		if p, err := s.system(ctx); err == nil && p != "" {
			baseSystem = p
		}
	}
	system, _ := s.mcp.EndpointList()
	if system != "" {
		system = baseSystem + "\n\n可调用以下系统接口（通过 call_api 工具）：\n" + system
	} else {
		system = baseSystem
	}

	// 工具定义。
	tools, _ := s.mcp.ListTools()
	llmTools := make([]llm.Tool, 0, len(tools))
	for _, t := range tools {
		llmTools = append(llmTools, llm.Tool{
			Name: t.Name, Description: t.Description, InputSchema: t.InputSchema,
		})
	}

	emit(SSEEvent{Type: "start"})

	maxRounds := 5
	assistantText := ""

	for round := 0; round < maxRounds; round++ {
		res, err := s.llm.Chat(ctx, llm.Params{
			Model: s.llm.Model(), MaxTokens: s.llm.MaxTokens(),
			System: system, Messages: messages, Tools: llmTools,
		}, func(delta string) error {
			assistantText += delta
			emit(SSEEvent{Type: "text_delta", Delta: delta})
			return nil
		})
		if err != nil {
			return &Error{Code: errcode.ErrAgentLLMError, Detail: err.Error()}
		}

		// 有工具调用：执行 → 追加 tool_result → 继续循环。
		if len(res.ToolUses) > 0 {
			// 持久化 assistant（含 tool_calls）。
			callsJSON, _ := json.Marshal(res.ToolUses)
			asstMsg := &domain.Message{
				ConversationID: conversationID, Role: "assistant",
				Content: res.Text, ToolCalls: string(callsJSON),
			}
			if err := s.repo.AddMessage(ctx, asstMsg); err != nil {
				return err
			}
			messages = append(messages, assistantToLLM(res))

			// 逐个执行工具，收集 tool_result。
			toolResults := make([]llm.Message, 0, len(res.ToolUses))
			for _, tu := range res.ToolUses {
				emit(SSEEvent{Type: "tool_use", Tool: tu.Name, Input: tu.Input})
				result, err := s.mcp.CallTool(ctx, tu.Name, tu.Input, userToken)
				if err != nil {
					result = mcp.ToolResult{Content: []mcp.ContentBlock{{Type: "text", Text: "工具调用异常: " + err.Error()}}, IsError: true}
				}
				// 持久化 tool 消息。
				resultJSON, _ := json.Marshal(result)
				_ = s.repo.AddMessage(ctx, &domain.Message{
					ConversationID: conversationID, Role: "tool",
					ToolResult: string(resultJSON),
				})
				emit(SSEEvent{Type: "tool_result", Tool: tu.Name, Message: resultText(result)})

				toolResults = append(toolResults, llm.Message{
					Role:    "user",
					Content: []map[string]any{{"type": "tool_result", "tool_use_id": tu.ID, "content": resultText(result)}},
				})
			}
			messages = append(messages, toolResults...)
			assistantText = "" // 工具轮文本不并入最终回复
			continue
		}

		// 无工具调用：最终回复。
		if res.Text != "" {
			_ = s.repo.AddMessage(ctx, &domain.Message{
				ConversationID: conversationID, Role: "assistant", Content: res.Text,
			})
			assistantText = res.Text
		}
		emit(SSEEvent{Type: "done", Content: assistantText})
		return nil
	}

	// 达到最大轮数（工具一直调用）也正常收尾。
	if assistantText != "" {
		_ = s.repo.AddMessage(ctx, &domain.Message{
			ConversationID: conversationID, Role: "assistant", Content: assistantText,
		})
	}
	emit(SSEEvent{Type: "done", Content: assistantText, Message: "已达到最大工具调用轮数"})
	return nil
}

// rebuildMessages 从 DB 历史重建 anthropic 格式消息序列。
// assistant(带 tool_calls) 与其后 tool 消息合并为：assistant(tool_use块) + user(tool_result块)。
func (s *AgentService) rebuildMessages(ctx context.Context, conversationID uint) ([]llm.Message, error) {
	hist, err := s.repo.ListMessages(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	var msgs []llm.Message
	i := 0
	for i < len(hist) {
		m := hist[i]
		switch m.Role {
		case "user":
			msgs = append(msgs, llm.Message{Role: "user", Content: m.Content})
			i++
		case "assistant":
			var blocks []map[string]any
			var toolUses []struct {
				ID    string         `json:"id"`
				Name  string         `json:"name"`
				Input map[string]any `json:"input"`
			}
			if m.Content != "" {
				blocks = append(blocks, map[string]any{"type": "text", "text": m.Content})
			}
			if m.ToolCalls != "" {
				_ = json.Unmarshal([]byte(m.ToolCalls), &toolUses)
			}
			i++
			// 收集其后的 tool 结果。
			var toolResults []map[string]any
			for i < len(hist) && hist[i].Role == "tool" {
				var tr struct {
					Content []struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"content"`
					IsError bool `json:"is_error"`
				}
				if json.Unmarshal([]byte(hist[i].ToolResult), &tr) == nil {
					text := ""
					for _, b := range tr.Content {
						text += b.Text
					}
					toolResults = append(toolResults, map[string]any{
						"type": "tool_result", "content": text, "is_error": tr.IsError,
					})
				}
				i++
			}
			// 关键：tool_use 与 tool_result 必须一一配对（DeepSeek 校验）。
			// 仅保留前 min(len(toolUses), len(toolResults)) 对；悬空的 tool_use 丢弃，
			// 否则报 "tool_use ids were found without tool_result blocks"。
			pairs := len(toolUses)
			if len(toolResults) < pairs {
				pairs = len(toolResults)
			}
			for j := 0; j < pairs; j++ {
				blocks = append(blocks, map[string]any{
					"type": "tool_use", "id": toolUses[j].ID, "name": toolUses[j].Name, "input": toolUses[j].Input,
				})
			}
			// 给每个 tool_result 配对对应的 tool_use_id（按序）。
			for j := 0; j < pairs; j++ {
				toolResults[j]["tool_use_id"] = toolUses[j].ID
			}
			// 兜底：assistant 消息不能为空 content（anthropic 格式要求非空）。
			if len(blocks) == 0 {
				blocks = append(blocks, map[string]any{"type": "text", "text": ""})
			}
			msgs = append(msgs, llm.Message{Role: "assistant", Content: blocks})
			if len(toolResults) > 0 {
				msgs = append(msgs, llm.Message{Role: "user", Content: toolResults})
			}
		default:
			i++
		}
	}
	return msgs, nil
}

// assistantToLLM 把本次工具轮 assistant 响应转成 anthropic 消息（供下一轮上下文）。
func assistantToLLM(res *llm.Result) llm.Message {
	blocks := make([]map[string]any, 0, len(res.ToolUses)+1)
	if res.Text != "" {
		blocks = append(blocks, map[string]any{"type": "text", "text": res.Text})
	}
	for _, tu := range res.ToolUses {
		blocks = append(blocks, map[string]any{
			"type": "tool_use", "id": tu.ID, "name": tu.Name, "input": tu.Input,
		})
	}
	return llm.Message{Role: "assistant", Content: blocks}
}

func resultText(r mcp.ToolResult) string {
	var sb strings.Builder
	for _, b := range r.Content {
		sb.WriteString(b.Text)
	}
	return sb.String()
}

// IsAgentError 判断错误是否为 agent 业务错误。
func IsAgentError(err error) bool {
	var e *Error
	return errors.As(err, &e)
}

// ErrorCode 提取 agent 业务错误码。
func ErrorCode(err error) int {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return -1
}
