import request from './request';

export interface DeptNode {
  id: number;
  parent_id: number;
  name: string;
  sort: number;
  status: number;
  remark?: string;
  /** 部门直接挂载的角色 */
  role_ids?: number[];
  roles?: { id: number; name: string; code: string }[];
  children?: DeptNode[];
}

/** 部门管理 API */
export const departmentApi = {
  /** 部门树（含每个部门挂载的角色） */
  getTree: () => request.get('/departments/tree'),
  /** 分页部门列表（filter 支持 subtree_id 子树筛选） */
  getList: (params: Record<string, unknown>) => request.get('/departments', { params }),
  getById: (id: number) => request.get(`/departments/${id}`),
  create: (data: { parent_id: number; name: string; sort?: number; status?: number; remark?: string }) =>
    request.post('/departments', data),
  update: (id: number, data: Record<string, unknown>) => request.put(`/departments/${id}`, data),
  /** 全量替换部门的角色关联 */
  assignRoles: (id: number, role_ids: number[]) => request.put(`/departments/${id}/roles`, { role_ids }),
  delete: (id: number) => request.delete(`/departments/${id}`),
};
