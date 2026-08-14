package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeSSEServer 返回固定的 anthropic 兼容 SSE 流，并记录收到的请求。
func fakeSSEServer(t *testing.T, sse string, onReq func(r *http.Request)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if onReq != nil {
			onReq(r)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(sse))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestChatStreamsTextAndCollectsToolUse(t *testing.T) {
	sse := "" +
		"event: message_start\n" +
		"data: {\"type\":\"message_start\"}\n\n" +
		"event: content_block_start\n" +
		"data: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\"}}\n\n" +
		"event: content_block_delta\n" +
		"data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"你\"}}\n\n" +
		"event: content_block_delta\n" +
		"data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"好\"}}\n\n" +
		"event: content_block_start\n" +
		"data: {\"type\":\"content_block_start\",\"index\":1,\"content_block\":{\"type\":\"tool_use\",\"id\":\"tu_1\",\"name\":\"call_api\"}}\n\n" +
		"event: content_block_delta\n" +
		"data: {\"type\":\"content_block_delta\",\"index\":1,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":\"{\\\"method\\\":\\\"GET\\\",\\\"path\\\":\\\"/api/v1/users\\\"}\"}}\n\n" +
		"event: content_block_stop\n" +
		"data: {\"type\":\"content_block_stop\",\"index\":1}\n\n" +
		"event: message_delta\n" +
		"data: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"tool_use\"}}\n\n" +
		"event: message_stop\n" +
		"data: {\"type\":\"message_stop\"}\n\n" +
		"data: [DONE]\n"

	var gotAuth string
	srv := fakeSSEServer(t, sse, func(r *http.Request) {
		gotAuth = r.Header.Get("x-api-key")
	})

	c := NewClient(srv.URL, "sk-test", "model-x", 1000)
	var deltas []string
	res, err := c.Chat(context.Background(), Params{
		Messages: []Message{{Role: "user", Content: "你好"}},
	}, func(d string) error {
		deltas = append(deltas, d)
		return nil
	})
	if err != nil {
		t.Fatalf("Chat 失败: %v", err)
	}
	if gotAuth != "sk-test" {
		t.Fatalf("应透传 x-api-key，得到 %q", gotAuth)
	}
	if res.Text != "你好" {
		t.Fatalf("文本应累积为「你好」，得到 %q", res.Text)
	}
	if len(deltas) != 2 || strings.Join(deltas, "") != "你好" {
		t.Fatalf("delta 回调应为 2 段「你好」，得到 %v", deltas)
	}
	if res.StopReason != "tool_use" {
		t.Fatalf("stop_reason 应为 tool_use，得到 %q", res.StopReason)
	}
	if len(res.ToolUses) != 1 {
		t.Fatalf("应收集 1 个 tool_use，得到 %d", len(res.ToolUses))
	}
	tu := res.ToolUses[0]
	if tu.ID != "tu_1" || tu.Name != "call_api" {
		t.Fatalf("tool_use 元数据错误: %+v", tu)
	}
	if tu.Input["method"] != "GET" || tu.Input["path"] != "/api/v1/users" {
		t.Fatalf("tool_use input 未正确解析: %+v", tu.Input)
	}
}

func TestChatSendsAnthropicHeaders(t *testing.T) {
	var gotVersion, gotContentType, gotAccept string
	srv := fakeSSEServer(t, "data: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"}}\n\ndata: [DONE]\n", func(r *http.Request) {
		gotVersion = r.Header.Get("anthropic-version")
		gotContentType = r.Header.Get("Content-Type")
		gotAccept = r.Header.Get("accept")
	})
	c := NewClient(srv.URL, "sk", "m", 10)

	if _, err := c.Chat(context.Background(), Params{Messages: []Message{{Role: "user", Content: "hi"}}}, nil); err != nil {
		t.Fatalf("Chat 失败: %v", err)
	}
	if gotVersion != "2023-06-01" {
		t.Fatalf("应发送 anthropic-version: 2023-06-01，得到 %q", gotVersion)
	}
	if !strings.Contains(gotContentType, "application/json") {
		t.Fatalf("Content-Type 应为 application/json，得到 %q", gotContentType)
	}
	if !strings.Contains(gotAccept, "text/event-stream") {
		t.Fatalf("accept 应为 text/event-stream，得到 %q", gotAccept)
	}
}

func TestChatReportsHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"bad"}}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "sk", "m", 10)
	_, err := c.Chat(context.Background(), Params{Messages: []Message{{Role: "user", Content: "hi"}}}, nil)
	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Fatalf("非 200 应返回含状态码错误，得到 %v", err)
	}
}
