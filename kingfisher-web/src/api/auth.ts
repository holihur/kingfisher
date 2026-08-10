import request from './request';
export const authApi = {
  login: (data: { username: string; password: string }) => request.post('/auth/login', data),
  register: (data: { username: string; password: string; email?: string }) => request.post('/auth/register', data),
  refresh: (refreshToken: string) => request.post('/auth/refresh', { refresh_token: refreshToken }),
  forgotPassword: (email: string) => request.post('/auth/forgot-password', { email }),
  resetPassword: (token: string, newPassword: string) => request.post('/auth/reset-password', { token, new_password: newPassword }),
};
