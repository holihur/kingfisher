import request from './request';

export const dictTypeApi = {
  list: (params: Record<string, unknown>) => request.get('/dict-types', { params }),
  getById: (id: number) => request.get(`/dict-types/${id}`),
  create: (data: { code: string; name: string; is_public: boolean; status: number; remark: string; version?: string }) =>
    request.post('/dict-types', data),
  update: (id: number, data: { code: string; name: string; is_public: boolean; status: number; remark: string; version?: string }) =>
    request.put(`/dict-types/${id}`, data),
  delete: (id: number) => request.delete(`/dict-types/${id}`),
  batchDelete: (ids: number[]) => request.post('/dict-types/batch-delete', { ids }),
  batchUpdateStatus: (ids: number[], status: number) => request.post('/dict-types/batch-status', { ids, status }),
};

export const dictEntryApi = {
  listByTypeId: (typeId: number, params: Record<string, unknown>) => request.get(`/dict-types/${typeId}/entries`, { params }),
  create: (typeId: number, data: { label: string; value: string; sort: number; status: number; remark: string; version?: string }) =>
    request.post(`/dict-types/${typeId}/entries`, data),
  update: (
    typeId: number,
    entryId: number,
    data: { label: string; value: string; sort: number; status: number; remark: string; version?: string },
  ) => request.put(`/dict-types/${typeId}/entries/${entryId}`, data),
  delete: (typeId: number, entryId: number) => request.delete(`/dict-types/${typeId}/entries/${entryId}`),
  batchDelete: (typeId: number, ids: number[]) => request.post(`/dict-types/${typeId}/entries/batch-delete`, { ids }),
  batchUpdateStatus: (typeId: number, ids: number[], status: number) =>
    request.post(`/dict-types/${typeId}/entries/batch-status`, { ids, status }),
  // 公共 API — 无需认证
  getPublicEntries: (code: string) => request.get(`/public/dicts/${code}/entries`),
};
