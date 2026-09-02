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
  create: (data: { username: string; password: string; email?: string; role_ids?: number[]; dept_ids?: number[] }) => request.post('/users', data),
  update: (id: number, data: Record<string, unknown>) => request.put(`/users/${id}`, data),
  delete: (id: number) => request.delete(`/users/${id}`),
  batchDelete: (ids: number[]) => request.post('/users/batch-delete', { ids }),
  batchUpdateStatus: (ids: number[], status: number) => request.post('/users/batch-status', { ids, status }),
  getSubAccounts: () => request.get('/users/me/sub-accounts'),
  createSubAccount: (data: { username: string; password: string; email?: string; role_ids: number[] }) =>
    request.post('/users/me/sub-accounts', data),
  updateSubAccount: (id: number, data: { role_ids: number[] }) =>
    request.put(`/users/me/sub-accounts/${id}`, data),
  deleteSubAccount: (id: number) => request.delete(`/users/me/sub-accounts/${id}`),
  adminListSubAccounts: (params: Record<string, unknown>) => request.get('/users/sub-accounts', { params }),
};
