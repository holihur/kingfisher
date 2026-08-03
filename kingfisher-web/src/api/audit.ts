import request from './request';
export const auditApi = {
  getList: (params: Record<string, unknown>) => request.get('/audit-logs', { params }),
};
