import request from './request';

export interface WorkTask {
  readonly id: number;
  readonly title: string;
  readonly description: string;
  readonly owner_id: number;
  readonly department_id: number;
  readonly status: string;
  readonly created_at: string;
  readonly updated_at: string;
}

export interface WorkTaskInput {
  readonly title: string;
  readonly description: string;
  readonly department_id: number;
  readonly status: string;
}

export const workTaskApi = {
  list: (params: Record<string, unknown>) => request.get('/tasks', { params }),
  getById: (id: number) => request.get(`/tasks/${id}`),
  create: (data: WorkTaskInput) => request.post('/tasks', data),
  update: (id: number, data: WorkTaskInput) => request.put(`/tasks/${id}`, data),
  delete: (id: number) => request.delete(`/tasks/${id}`),
};
