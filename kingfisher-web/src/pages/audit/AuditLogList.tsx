import React, { useRef } from 'react';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import { auditApi } from '../../api/audit';

const AuditLogList: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const columns: ProColumns[] = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '用户', dataIndex: 'username' },
    { title: '操作', dataIndex: 'action' },
    { title: '资源', dataIndex: 'resource' },
    { title: 'IP', dataIndex: 'ip' },
    { title: '时间', dataIndex: 'created_at', valueType: 'dateTime' },
  ];
  return (
    <ProTable
      columns={columns}
      actionRef={actionRef}
      request={async (params) => {
        const r = await auditApi.getList({ page: params.current || 1, page_size: params.pageSize || 20, ...params });
        const data = r.data as Record<string, unknown>;
        return {
          data: (data.items as Record<string, unknown>[]) || [],
          total: (data.total as number) || 0,
          success: true,
        };
      }}
      rowKey="id"
      search={false}
      headerTitle="审计日志"
    />
  );
};

export default AuditLogList;
