package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"kingfisher/core/errcode"
	"kingfisher/extends/agent/domain"
	"kingfisher/extends/agent/llm"
	"kingfisher/extends/agent/mcp"
	"kingfisher/extends/agent/port"
)

// mockAgentRepo 内存实现 port.AgentRepository，测试 AgentService 用。
type mockAgentRepo struct {
	convs   map[uint]*domain.Conversation
	msgs    map[uint][]domain.Message
	convSeq uint
	msgSeq  uint
}

func newMockRepo() *mockAgentRepo {
	return &mockAgentRepo{convs: map[uint]*domain.Conversation{}, msgs: map[uint][]domain.Message{}}
}

func (m *mockAgentRepo) ListConversations(ctx context.Context, userID uint) ([]domain.Conversation, error) {
	var out []domain.Conversation
	for _, c := range m.convs {
		if c.UserID == userID {
			out = append(out, *c)
		}
	}
	return out, nil
}

func (m *mockAgentRepo) GetConversation(ctx context.Context, id uint) (*domain.Conversation, error) {
	c, ok := m.convs[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

func (m *mockAgentRepo) CreateConversation(ctx context.Context, userID uint, title string) (*domain.Conversation, error) {
	m.convSeq++
	c := &domain.Conversation{ID: m.convSeq, UserID: userID, Title: title}
	m.convs[c.ID] = c
	return c, nil
}

func (m *mockAgentRepo) RenameConversation(ctx context.Context, id uint, title string) error {
	c, ok := m.convs[id]
	if !ok {
		return fmt.Errorf("not found")
	}
	c.Title = title
	return nil
}

func (m *mockAgentRepo) DeleteConversation(ctx context.Context, id uint, userID uint) error {
	if _, ok := m.convs[id]; !ok {
		return fmt.Errorf("not found")
	}
	delete(m.convs, id)
	delete(m.msgs, id)
	return nil
}

func (m *mockAgentRepo) ListMessages(ctx context.Context, conversationID uint) ([]domain.Message, error) {
	return m.msgs[conversationID], nil
}

func (m *mockAgentRepo) AddMessage(ctx context.Context, msg *domain.Message) error {
	m.msgSeq++
	msg.ID = m.msgSeq
	m.msgs[msg.ConversationID] = append(m.msgs[msg.ConversationID], *msg)
	return nil
}

// newTestService 组装一个最小 AgentService。
// llm 用不可达地址的 client（测试不触发真实 LLM 调用）；mcp 用真实 client 但测试
// 不调用 ChatStream（不触发 EndpointList 的 HTTP），只测会话管理与历史重建。
func newTestService(repo port.AgentRepository) *AgentService {
	llmClient := llm.NewClient("http://127.0.0.1:1", "test-key", "test-model", 100)
	mcpClient := mcp.NewClient("http://127.0.0.1:1")
	return NewAgentService(repo, llmClient, mcpClient, func(ctx context.Context) (string, error) { return "test-system", nil },
		func(ctx context.Context) (string, error) { return "test-key", nil },
		func() bool { return true },
		"http://127.0.0.1:8080")
}

func TestCheckEnabled(t *testing.T) {
	svc := newTestService(newMockRepo())
	if err := svc.CheckEnabled(); err != nil {
		t.Fatalf("enabled 时应返回 nil，得到 %v", err)
	}

	// 禁用时返回 ErrAgentDisabled
	svc.enabled = func() bool { return false }
	err := svc.CheckEnabled()
	var e *Error
	if !errors.As(err, &e) || e.Code != errcode.ErrAgentDisabled {
		t.Fatalf("禁用时应返回 ErrAgentDisabled，得到 %v", err)
	}
}

func TestCreateAndGetConversation(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)

	c, err := svc.CreateConversation(context.Background(), 1, "测试")
	if err != nil {
		t.Fatalf("创建会话失败: %v", err)
	}
	if c.Title != "测试" {
		t.Fatalf("标题应为「测试」，得到 %q", c.Title)
	}

	// 归属校验：本用户可访问
	got, err := svc.GetConversation(context.Background(), c.ID, 1)
	if err != nil || got.ID != c.ID {
		t.Fatalf("本用户访问会话失败: %v", err)
	}
	// 其他用户不可访问（返回 ErrAgentConversationNF）
	_, err = svc.GetConversation(context.Background(), c.ID, 99)
	var e *Error
	if !errors.As(err, &e) || e.Code != errcode.ErrAgentConversationNF {
		t.Fatalf("他人访问应返回 ErrAgentConversationNF，得到 %v", err)
	}
	// 不存在
	_, err = svc.GetConversation(context.Background(), 999, 1)
	if !errors.As(err, &e) || e.Code != errcode.ErrAgentConversationNF {
		t.Fatalf("不存在应返回 ErrAgentConversationNF，得到 %v", err)
	}
}

func TestDeleteConversationOwnership(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)

	c, _ := svc.CreateConversation(context.Background(), 1, "")
	// 他人删除 → 拒绝
	err := svc.DeleteConversation(context.Background(), c.ID, 2)
	var e *Error
	if !errors.As(err, &e) || e.Code != errcode.ErrAgentConversationNF {
		t.Fatalf("他人删除应被拒，得到 %v", err)
	}
	// 本人删除 → 成功
	if err := svc.DeleteConversation(context.Background(), c.ID, 1); err != nil {
		t.Fatalf("本人删除失败: %v", err)
	}
}

func TestListMessagesOwnership(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)

	c, _ := svc.CreateConversation(context.Background(), 1, "")
	// 他人读消息 → 拒绝
	if _, err := svc.ListMessages(context.Background(), c.ID, 2); err == nil {
		t.Fatal("他人读消息应被拒")
	}
	// 本人读 → 空列表正常
	msgs, err := svc.ListMessages(context.Background(), c.ID, 1)
	if err != nil || len(msgs) != 0 {
		t.Fatalf("本人读空会话应返回空，得到 %v / %v", msgs, err)
	}
}

// TestRebuildMessages 历史重建：user + assistant(带tool_calls) + tool 应正确合并为
// assistant(tool_use块) + user(tool_result块)，且 tool_result 配对 tool_use_id。
func TestRebuildMessages(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)

	c, _ := svc.CreateConversation(context.Background(), 1, "")
	// 注入历史
	_ = repo.AddMessage(context.Background(), &domain.Message{ConversationID: c.ID, Role: "user", Content: "查用户"})
	_ = repo.AddMessage(context.Background(), &domain.Message{
		ConversationID: c.ID, Role: "assistant", Content: "我来查",
		ToolCalls: `[{"id":"tu_1","name":"call_api","input":{"method":"GET","path":"/api/v1/users"}}]`,
	})
	_ = repo.AddMessage(context.Background(), &domain.Message{
		ConversationID: c.ID, Role: "tool",
		ToolResult: `{"content":[{"type":"text","text":"{\"total\":4}"}],"is_error":false}`,
	})

	msgs, err := svc.rebuildMessages(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("rebuildMessages 失败: %v", err)
	}
	// 期望：[user(查用户), assistant(text+tool_use), user(tool_result)] 3 条
	if len(msgs) != 3 {
		t.Fatalf("期望 3 条消息，得到 %d", len(msgs))
	}
	if msgs[0].Role != "user" || msgs[0].Content != "查用户" {
		t.Fatalf("第 0 条应为 user「查用户」，得到 %v", msgs[0])
	}
	asst, ok := msgs[1].Content.([]map[string]any)
	if !ok || len(asst) != 2 {
		t.Fatalf("assistant 应含 text+tool_use 两块，得到 %v", msgs[1].Content)
	}
	// tool_use 块带 id
	if asst[1]["type"] != "tool_use" || asst[1]["id"] != "tu_1" {
		t.Fatalf("assistant 第 2 块应为 tool_use(tu_1)，得到 %v", asst[1])
	}
	blocks := msgs[2].Content.([]map[string]any)
	if len(blocks) != 1 {
		t.Fatalf("user 应含 1 个 tool_result 块，得到 %d", len(blocks))
	}
	if blocks[0]["type"] != "tool_result" {
		t.Fatalf("user 块应为 tool_result，得到 %v", blocks[0]["type"])
	}
	// tool_result 必须带 tool_use_id（DeepSeek 校验缺失会 400）
	if id, _ := blocks[0]["tool_use_id"].(string); id != "tu_1" {
		t.Fatalf("tool_result.tool_use_id 应为 tu_1，得到 %q", id)
	}
}

// TestRebuildMessagesDropsOrphanToolUse 悬空 tool_use（无对应 tool_result）必须被丢弃，
// 否则 DeepSeek 报 "tool_use ids were found without tool_result blocks" 400。
func TestRebuildMessagesDropsOrphanToolUse(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)

	c, _ := svc.CreateConversation(context.Background(), 1, "")
	_ = repo.AddMessage(context.Background(), &domain.Message{ConversationID: c.ID, Role: "user", Content: "删用户"})
	// assistant 带 2 个 tool_use，但只有 1 个 tool_result（另一个悬空/中断）
	_ = repo.AddMessage(context.Background(), &domain.Message{
		ConversationID: c.ID, Role: "assistant", Content: "我来",
		ToolCalls: `[{"id":"tu_1","name":"call_api","input":{"method":"DELETE","path":"/api/v1/users/3"}},
		             {"id":"tu_2","name":"call_api","input":{"method":"DELETE","path":"/api/v1/users/4"}}]`,
	})
	_ = repo.AddMessage(context.Background(), &domain.Message{
		ConversationID: c.ID, Role: "tool",
		ToolResult: `{"content":[{"type":"text","text":"ok"}],"is_error":false}`,
	})

	msgs, err := svc.rebuildMessages(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("rebuildMessages 失败: %v", err)
	}
	// 期望：[user, assistant(仅 1 个 tool_use), user(tool_result)]
	if len(msgs) != 3 {
		t.Fatalf("期望 3 条消息，得到 %d", len(msgs))
	}
	asst := msgs[1].Content.([]map[string]any)
	toolUses := 0
	for _, b := range asst {
		if b["type"] == "tool_use" {
			toolUses++
		}
	}
	if toolUses != 1 {
		t.Fatalf("悬空 tool_use 应被丢弃，实际保留 %d 个 tool_use", toolUses)
	}
	// 配对的 tool_result 带正确 tool_use_id
	blocks := msgs[2].Content.([]map[string]any)
	if id, _ := blocks[0]["tool_use_id"].(string); id != "tu_1" {
		t.Fatalf("tool_result.tool_use_id 应为 tu_1（配对第一个），得到 %q", id)
	}
}

// TestRebuildMessagesAssistantNoToolResult 仅 text、无 tool_use 的 assistant 正常保留。
func TestRebuildMessagesAssistantNoToolResult(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)

	c, _ := svc.CreateConversation(context.Background(), 1, "")
	_ = repo.AddMessage(context.Background(), &domain.Message{ConversationID: c.ID, Role: "user", Content: "你好"})
	_ = repo.AddMessage(context.Background(), &domain.Message{ConversationID: c.ID, Role: "assistant", Content: "你好，有什么可以帮你？"})

	msgs, err := svc.rebuildMessages(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("rebuildMessages 失败: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("普通 assistant 应保留为 2 条消息，得到 %d 条", len(msgs))
	}
	// assistant 内容应为 text 块（数组形式）
	asst, ok := msgs[1].Content.([]map[string]any)
	if !ok || len(asst) != 1 || asst[0]["text"] != "你好，有什么可以帮你？" {
		t.Fatalf("普通 assistant 应含 text 块，得到 %v", msgs[1].Content)
	}
}
