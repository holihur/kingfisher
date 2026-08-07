import React, { useEffect } from 'react';
import ProTable, { ProColumns } from '@ant-design/pro-table';
import { auditApi } from '../../api/audit';
import { useTableUrlQuery } from '../../hooks/useTableUrlQuery';
import { buildQueryParams } from '../../utils/query';

const ACTION_VALUES: Record<string, { text: string }> = {
  create: { text: '创建' },
  update: { text: '更新' },
  delete: { text: '删除' },
  login: { text: '登录' },
  logout: { text: '退出' },
  register: { text: '注册' },
};

const AuditLogList: React.FC = () => {
  const { urlParams, page, pageSize, actionRef, formRef, syncFormFromUrl, onSearch, onReset, onPageChange } = useTableUrlQuery();

  // 挂载时用 URL 反填搜索表单
  useEffect(() => {
    syncFormFromUrl();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const columns: ProColumns[] = [
    { title: 'ID', dataIndex: 'id', width: 80, search: false },
    {
      title: '关键词',
      dataIndex: 'q',
      hideInTable: true,
      search: { transform: (v) => ({ q: v }) },
    },
    { title: '用户', dataIndex: 'username', search: false },
    {
      title: '操作',
      dataIndex: 'action',
      valueEnum: ACTION_VALUES,
      render: (_, r) => ACTION_VALUES[r.action as string]?.text || (r.action as string),
    },
    { title: '资源', dataIndex: 'resource' },
    { title: '资源ID', dataIndex: 'resource_id', width: 90, search: false },
    { title: 'IP', dataIndex: 'ip', search: false },
    { title: '时间', dataIndex: 'created_at', valueType: 'dateTime', search: false },
  ];
  return (
    <ProTable
      columns={columns}
      actionRef={actionRef}
      formRef={formRef}
      params={urlParams}
      request={async (params) => {
        const r = await auditApi.getList(buildQueryParams(params));
        const data = r.data as Record<string, unknown>;
        return {
          data: (data.items as Record<string, unknown>[]) || [],
          total: (data.total as number) || 0,
          success: true,
        };
      }}
      rowKey="id"
      onSubmit={onSearch}
      onReset={onReset}
      search={{ labelWidth: 'auto' }}
      pagination={{ current: page, pageSize, showSizeChanger: true, onChange: onPageChange }}
      headerTitle="审计日志"
    />
  );
};

export default AuditLogList;
