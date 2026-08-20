import request from './request';
export const roleApi = {
  getList: (params: Record<string, unknown>) => request.get('/roles', { params }),
  getById: (id: number) => request.get(`/roles/${id}`),
  create: (data: Record<string, unknown>) => request.post('/roles', data),
  update: (id: number, data: Record<string, unknown>) => request.put(`/roles/${id}`, data),
  delete: (id: number) => request.delete(`/roles/${id}`),
  getPermissions: (id: number) => request.get(`/roles/${id}/permissions`),
  assignPermissions: (id: number, permIds: number[]) =>
    request.put(`/roles/${id}/permissions`, { permission_ids: permIds }),
  getMenus: (id: number) => request.get(`/roles/${id}/menus`),
  assignMenus: (id: number, menuIds: number[]) => request.put(`/roles/${id}/menus`, { menu_ids: menuIds }),
  getAllPermissions: () => request.get('/permissions'),
  batchDelete: (ids: number[]) => request.post('/roles/batch-delete', { ids }),
  batchUpdateStatus: (ids: number[], status: number) => request.post('/roles/batch-status', { ids, status }),
  getDataScope: (id: number, resource: string) => request.get(`/roles/${id}/data-scope`, { params: { resource } }),
  setDataScope: (id: number, resource: string, scope_type: string) => request.put(`/roles/${id}/data-scope`, { resource, scope_type }),
};
