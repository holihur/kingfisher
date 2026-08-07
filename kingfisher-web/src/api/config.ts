import request from './request';
export const configApi = {
  getAll: () => request.get('/configs'),
  get: (key: string) => request.get(`/configs/${key}`),
  set: (key: string, value: string, isPublic?: boolean, version?: string, render?: string, renderOptions?: string) =>
    request.put(`/configs/${key}`, { value, is_public: isPublic, version, render, render_options: renderOptions }),
  delete: (key: string) => request.delete(`/configs/${key}`),
  // 公开配置：无需登录可读
  getPublicAll: () => request.get('/public/configs'),
  getPublic: (key: string) => request.get(`/public/configs/${key}`),
};
