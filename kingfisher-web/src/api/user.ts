import request from './request';
export const userApi = {
  getList: (params: { page?: number; page_size?: number; keyword?: string }) => request.get('/users', { params }),
  getById: (id: number) => request.get(`/users/${id}`),
  getMe: () => request.get('/users/me'),
  getMyPermissions: () => request.get('/users/me/permissions'),
  create: (data: { username: string; password: string; email?: string }) => request.post('/users', data),
  update: (id: number, data: Record<string, unknown>) => request.put(`/users/${id}`, data),
  delete: (id: number) => request.delete(`/users/${id}`),
};
