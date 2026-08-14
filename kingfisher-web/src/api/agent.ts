import request from './request';
import { getToken } from '../utils/token';

/** SSE 流式事件（与后端 app.SSEEvent 对齐）。 */
export interface AgentSSEEvent {
  type: 'start' | 'text_delta' | 'tool_use' | 'tool_result' | 'done' | 'error';
  delta?: string;
  tool?: string;
  input?: Record<string, unknown>;
  message?: string;
  code?: number;
  content?: string;
}

export const agentApi = {
  /** 当前用户会话列表。 */
  listConversations: () => request.get('/agent/conversations'),
  /** 创建会话。 */
  createConversation: (title?: string) => request.post('/agent/conversations', { title }),
  /** 删除会话。 */
  deleteConversation: (id: number) => request.delete(`/agent/conversations/${id}`),
  /** 会话消息历史。 */
  listMessages: (id: number) => request.get(`/agent/conversations/${id}/messages`),

  /**
   * 流式对话（SSE）。axios 不适合读流，这里用原生 fetch + ReadableStream。
   * 返回一个 { abort } 句柄用于中断。
   */
  chatStream: (
    conversationId: number,
    content: string,
    onEvent: (evt: AgentSSEEvent) => void,
  ): { abort: () => void } => {
    const controller = new AbortController();
    const baseURL = import.meta.env.VITE_API_BASE_URL || '/api/v1';
    const token = getToken();

    (async () => {
      try {
        const resp = await fetch(`${baseURL}/agent/chat/stream`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            ...(token ? { Authorization: `Bearer ${token}` } : {}),
          },
          body: JSON.stringify({ conversation_id: conversationId, content }),
          signal: controller.signal,
        });
        if (!resp.ok || !resp.body) {
          onEvent({ type: 'error', code: resp.status, message: `请求失败 (${resp.status})` });
          return;
        }
        const reader = resp.body.getReader();
        const decoder = new TextDecoder();
        let buffer = '';
        for (;;) {
          const { done, value } = await reader.read();
          if (done) break;
          buffer += decoder.decode(value, { stream: true });
          // SSE 事件以空行分隔，按块解析，保留末尾未完成部分。
          const parts = buffer.split('\n\n');
          buffer = parts.pop() ?? '';
          for (const part of parts) {
            for (const line of part.split('\n')) {
              if (line.startsWith('data: ')) {
                try {
                  onEvent(JSON.parse(line.slice(6)) as AgentSSEEvent);
                } catch {
                  /* 忽略无法解析的碎片 */
                }
              }
            }
          }
        }
      } catch (err) {
        if ((err as Error).name !== 'AbortError') {
          onEvent({ type: 'error', code: 0, message: '连接中断，请重试' });
        }
      }
    })();

    return { abort: () => controller.abort() };
  },
};
