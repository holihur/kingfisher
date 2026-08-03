import request from './request';
export const configApi = {
  getAll: () => request.get('/configs'),
  get: (key: string) => request.get(`/configs/${key}`),
  set: (key: string, value: string) => request.put(`/configs/${key}`, { value }),
  delete: (key: string) => request.delete(`/configs/${key}`),
};
