import request from './request';
export const userApi = {
  getList: (params: Record<string, unknown>) => request.get('/users', { params }),
  getById: (id: number) => request.get(`/users/${id}`),
  getMe: () => request.get('/users/me'),
  getMyPermissions: () => request.get('/users/me/permissions'),
  updateMe: (data: { email?: string; nickname?: string; avatar?: string }) =>
    request.put('/users/me', data),
  changePassword: (data: { old_password: string; new_password: string }) =>
    request.put('/users/me/password', data),
  getMyLoginLogs: (params: Record<string, unknown>) =>
    request.get('/users/me/login-logs', { params }),
  create: (data: { username: string; password: string; email?: string }) => request.post('/users', data),
  update: (id: number, data: Record<string, unknown>) => request.put(`/users/${id}`, data),
  delete: (id: number) => request.delete(`/users/${id}`),
};
