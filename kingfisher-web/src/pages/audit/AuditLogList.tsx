import React from 'react';
import { Tag } from 'antd';
import DataTable, { SearchField } from '../../components/DataTable';
import { auditApi } from '../../api/audit';
import { formatTime } from '../../utils/format';

const ACTION_VALUES: Record<string, { text: string; color?: string }> = {
  create: { text: '创建', color: 'green' },
  update: { text: '更新', color: 'blue' },
  delete: { text: '删除', color: 'red' },
  login: { text: '登录' },
  logout: { text: '退出' },
  register: { text: '注册' },
};

interface AuditLogRow {
  id: number;
  username: string;
  action: string;
  resource: string;
  resource_id: number;
  ip: string;
  user_agent: string;
  created_at: string;
}

const searchFields: SearchField[] = [
  { name: 'q', label: '关键词', type: 'text' },
  { name: 'resource', label: '资源', type: 'text' },
];

const AuditLogList: React.FC = () => {
  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '用户', dataIndex: 'username' },
    {
      title: '操作',
      dataIndex: 'action',
      render: (_: unknown, r: AuditLogRow) => {
        const v = ACTION_VALUES[r.action];
        return v ? <Tag color={v.color}>{v.text}</Tag> : (r.action as string);
      },
    },
    { title: '资源', dataIndex: 'resource', render: (_: unknown, r: AuditLogRow) => <Tag>{r.resource}</Tag> },
    { title: '资源ID', dataIndex: 'resource_id', width: 90 },
    { title: 'IP', dataIndex: 'ip', render: (_: unknown, r: AuditLogRow) => <code style={{ background: '#f5f5f5', padding: '2px 6px', borderRadius: 4 }}>{r.ip}</code> },
    {
      title: '时间',
      dataIndex: 'created_at',
      width: 160,
      render: (v: unknown) => formatTime(v),
    },
  ];

  return (
    <DataTable<AuditLogRow>
      columns={columns}
      rowKey="id"
      request={async (params) => {
        const r = await auditApi.getList(params);
        const data = r.data as Record<string, unknown>;
        return {
          items: (data.items as AuditLogRow[]) || [],
          total: (data.total as number) || 0,
        };
      }}
      searchFields={searchFields}
      headerTitle="审计日志"
    />
  );
};

export default AuditLogList;
