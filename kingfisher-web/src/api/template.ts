import request from './request';

export const templateApi = {
  list: (params: Record<string, unknown>) => request.get('/templates', { params }),
  getById: (id: number) => request.get(`/templates/${id}`),
  create: (data: {
    name: string;
    code: string;
    template_type: string;
    title: string;
    content: string;
    status: number;
    remark: string;
    version?: string;
  }) => request.post('/templates', data),
  update: (
    id: number,
    data: {
      name: string;
      code: string;
      template_type: string;
      title: string;
      content: string;
      status: number;
      remark: string;
      version?: string;
    },
  ) => request.put(`/templates/${id}`, data),
  delete: (id: number) => request.delete(`/templates/${id}`),
  batchDelete: (ids: number[]) => request.post('/templates/batch-delete', { ids }),
  batchUpdateStatus: (ids: number[], status: number) => request.post('/templates/batch-status', { ids, status }),
};
