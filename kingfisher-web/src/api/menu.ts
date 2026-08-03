import request from './request';
export const menuApi = {
  getTree: () => request.get('/menus/tree'),
  getById: (id: number) => request.get(`/menus/${id}`),
  create: (data: Record<string, unknown>) => request.post('/menus', data),
  update: (id: number, data: Record<string, unknown>) => request.put(`/menus/${id}`, data),
  delete: (id: number) => request.delete(`/menus/${id}`),
};
