import request from './request';

export const messageApi = {
  // 管理员：发送 + 已发送列表 + 撤回
  send: (data: { recipient_ids: number[]; title: string; content?: string }) =>
    request.post('/messages', data),
  listSent: (params: Record<string, unknown>) => request.get('/messages', { params }),
  revoke: (id: number) => request.put(`/messages/${id}/revoke`),
  // 个人收件箱
  list: (params: Record<string, unknown>) => request.get('/me/messages', { params }),
  getById: (id: number) => request.get(`/me/messages/${id}`),
  unreadCount: () => request.get('/me/messages/unread-count'),
  markRead: (id: number) => request.put(`/me/messages/${id}/read`),
  batchDelete: (ids: number[]) => request.post('/me/messages/batch-delete', { ids }),
};
