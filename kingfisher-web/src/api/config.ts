import request from './request';
export const configApi = {
  getAll: () => request.get('/configs'),
  get: (key: string) => request.get(`/configs/${key}`),
  set: (key: string, value: string, isPublic?: boolean, version?: string, render?: string, renderOptions?: string, groupId?: number) =>
    request.put(`/configs/${key}`, { value, is_public: isPublic, version, render, render_options: renderOptions, group_id: groupId }),
  delete: (key: string) => request.delete(`/configs/${key}`),
  // 公开配置：无需登录可读
  getPublicAll: () => request.get('/public/configs'),
  getPublic: (key: string) => request.get(`/public/configs/${key}`),
};

export const configGroupApi = {
  list: () => request.get('/config-groups'),
  create: (name: string, sort: number) => request.post('/config-groups', { name, sort }),
  update: (id: number, name: string, sort: number) => request.put(`/config-groups/${id}`, { name, sort }),
  delete: (id: number) => request.delete(`/config-groups/${id}`),
};
