import request from './request';
export const userApi = {
  getList: (params: Record<string, unknown>) => request.get('/users', { params }),
  getById: (id: number) => request.get(`/users/${id}`),
  getMe: () => request.get('/users/me'),
  getMyPermissions: () => request.get('/users/me/permissions'),
  updateMe: (data: { email?: string; nickname?: string; avatar?: string; phone?: string }) =>
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
  getMFAStatus: () => request.get('/users/me/mfa/status'),
  setupTOTP: () => request.post('/users/me/mfa/totp/setup'),
  verifyTOTP: (code: string) => request.post('/users/me/mfa/totp/verify', { code }),
  disableTOTP: (code: string) => request.delete('/users/me/mfa/totp', { data: { code } } as never),
  sendSMS: (phone?: string) => request.post('/users/me/mfa/sms/send', phone ? { phone } : {}),
  verifySMS: (phone: string, code: string) => request.post('/users/me/mfa/sms/verify', { phone, code }),
  disableSMS: () => request.delete('/users/me/mfa/sms'),
  sendEmailCode: () => request.post('/users/me/mfa/email/send'),
  verifyEmail: (code: string) => request.post('/users/me/mfa/email/verify', { code }),
  disableEmail: () => request.delete('/users/me/mfa/email'),
  adminGetMFAStatus: (id: number) => request.get(`/users/${id}/mfa/status`),
  adminResetMFA: (id: number) => request.delete(`/users/${id}/mfa/reset`),
};
