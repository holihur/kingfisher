import request from './request';

export const scheduledTaskApi = {
  list: (params: Record<string, unknown>) => request.get('/scheduled-tasks', { params }),
  getById: (id: number) => request.get(`/scheduled-tasks/${id}`),
  types: () => request.get('/scheduled-tasks/types'),
  create: (data: {
    name: string;
    task_type: string;
    cron_spec: string;
    payload?: string;
    enabled: number;
    remark?: string;
  }) => request.post('/scheduled-tasks', data),
  update: (
    id: number,
    data: {
      name: string;
      task_type: string;
      cron_spec: string;
      payload?: string;
      enabled: number;
      remark?: string;
    },
  ) => request.put(`/scheduled-tasks/${id}`, data),
  delete: (id: number) => request.delete(`/scheduled-tasks/${id}`),
  run: (id: number) => request.post(`/scheduled-tasks/${id}/run`),
  batchDelete: (ids: number[]) => request.post('/scheduled-tasks/batch-delete', { ids }),
  batchUpdateStatus: (ids: number[], enabled: number) => request.post('/scheduled-tasks/batch-status', { ids, enabled }),
};
