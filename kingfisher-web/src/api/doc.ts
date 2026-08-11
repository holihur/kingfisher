import request from './request';

export interface DocDirNode {
  id: number;
  parent_id: number;
  name: string;
  sort: number;
  status: number;
  granted_roles?: number[];
  children?: DocDirNode[];
}

export interface DocItem {
  id: number;
  dir_id: number;
  title: string;
  content: string;
  owner_id: number;
  owner_name?: string;
  visibility: 'shared' | 'private';
  status: 'draft' | 'published';
  current_version: number;
  created_at: string;
  updated_at: string;
}

export interface DocVersion {
  id: number;
  doc_id: number;
  version_no: number;
  title: string;
  content: string;
  owner_name?: string;
  note: string;
  created_at: string;
}

/** 目录操作 */
export const docDirApi = {
  getTree: () => request.get('/docs/tree'),
  create: (data: { parent_id: number; name: string; sort?: number }) => request.post('/docs/dirs', data),
  update: (id: number, data: Record<string, unknown>) => request.put(`/docs/dirs/${id}`, data),
  delete: (id: number) => request.delete(`/docs/dirs/${id}`),
  getRoles: (id: number) => request.get(`/docs/dirs/${id}/roles`),
  setRoles: (id: number, role_ids: number[]) => request.put(`/docs/dirs/${id}/roles`, { role_ids }),
};

/** 文档操作 */
export const docApi = {
  list: (dir_id: number, params: Record<string, unknown>) => request.get('/docs', { params: { ...params, dir_id } }),
  getById: (id: number) => request.get(`/docs/${id}`),
  create: (data: { dir_id: number; title: string; content: string; visibility?: string; note?: string }) =>
    request.post('/docs', data),
  update: (id: number, data: { title: string; content: string; note?: string }) => request.put(`/docs/${id}`, data),
  publish: (id: number) => request.put(`/docs/${id}/publish`, {}),
  unpublish: (id: number) => request.put(`/docs/${id}/unpublish`, {}),
  versions: (id: number) => request.get(`/docs/${id}/versions`),
  getVersion: (id: number, no: number) => request.get(`/docs/${id}/versions/${no}`),
  restore: (id: number, version_no: number) => request.post(`/docs/${id}/restore`, { version_no }),
  delete: (id: number) => request.delete(`/docs/${id}`),
};
