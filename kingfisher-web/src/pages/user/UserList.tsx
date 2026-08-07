import React, { useEffect, useState } from 'react';
import { Button, Modal, Form, message, Popconfirm, Badge } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import ProTable, { ProColumns } from '@ant-design/pro-table';
import { useAuthStore } from '../../stores/auth';
import { userApi } from '../../api/user';
import { roleApi } from '../../api/role';
import { useTableUrlQuery } from '../../hooks/useTableUrlQuery';
import { buildQueryParams } from '../../utils/query';
import UserForm from './UserForm';

const UserList: React.FC = () => {
  const { urlParams, page, pageSize, actionRef, formRef, syncFormFromUrl, onSearch, onReset, onPageChange } = useTableUrlQuery();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Record<string, unknown> | null>(null);
  const [form] = Form.useForm<Record<string, unknown>>();
  const permissions = useAuthStore((s) => s.permissions);
  const hasPerm = (code: string) => permissions.includes(code);
  const [roleOptions, setRoleOptions] = useState<{ label: string; value: number }[]>([]);
  const [roleNameMap, setRoleNameMap] = useState<Record<number, string>>({});

  // 挂载时用 URL 反填搜索表单
  useEffect(() => {
    syncFormFromUrl();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    roleApi.getList({ page: 1, page_size: 100 }).then((r) => {
      const data = r.data as Record<string, unknown>;
      const items = (data.items as Record<string, unknown>[]) || [];
      setRoleOptions(items.map((i) => ({ label: i.name as string, value: i.id as number })));
      const map: Record<number, string> = {};
      items.forEach((i) => { map[i.id as number] = i.name as string; });
      setRoleNameMap(map);
    });
  }, []);

  const columns: ProColumns[] = [
    { title: 'ID', dataIndex: 'id', width: 80, search: false },
    {
      title: '关键词',
      dataIndex: 'q',
      hideInTable: true,
      search: { transform: (v) => ({ q: v }) },
    },
    { title: '用户名', dataIndex: 'username', search: false },
    { title: '邮箱', dataIndex: 'email', ellipsis: true, search: false },
    {
      title: '角色',
      dataIndex: 'role_id',
      width: 100,
      valueEnum: Object.fromEntries(roleOptions.map((o) => [o.value, { text: o.label }])),
      render: (_, r) => roleNameMap[r.role_id as number] || `#${r.role_id}`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      valueEnum: { 1: { text: '启用' }, 0: { text: '禁用' } },
      render: (_, r) => (
        <Badge status={(r.status as number) === 1 ? 'success' : 'error'} text={(r.status as number) === 1 ? '启用' : '禁用'} />
      ),
    },
    { title: '创建时间', dataIndex: 'created_at', valueType: 'dateTime', search: false },
    {
      title: '操作',
      valueType: 'option',
      render: (_, r) => [
        hasPerm('user:update') ? (
          <a
            key="edit"
            onClick={() => {
              setEditing(r as Record<string, unknown>);
              setModalOpen(true);
            }}
          >
            编辑
          </a>
        ) : null,
        hasPerm('user:delete') ? (
          <Popconfirm
            key="del"
            title="确认删除？"
            onConfirm={async () => {
              await userApi.delete(r.id as number);
              message.success('已删除');
              actionRef.current?.reload();
            }}
          >
            <a style={{ color: 'red' }}>删除</a>
          </Popconfirm>
        ) : null,
      ],
    },
  ];

  const handleSubmit = async () => {
    const values = await form.validateFields();
    if (editing?.id) {
      await userApi.update(editing.id as number, values as Record<string, unknown>);
      message.success('更新成功');
    } else {
      await userApi.create(values as any);
      message.success('创建成功');
    }
    setModalOpen(false);
    setEditing(null);
    actionRef.current?.reload();
  };

  return (
    <>
      <ProTable
        columns={columns}
        actionRef={actionRef}
        formRef={formRef}
        params={urlParams}
        request={async (params) => {
          const resp = await userApi.getList(buildQueryParams(params));
          const data = resp.data as Record<string, unknown>;
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
        headerTitle="用户管理"
        toolBarRender={() => [
          hasPerm('user:create') ? (
            <Button
              key="add"
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setEditing(null);
                setModalOpen(true);
              }}
            >
              新增用户
            </Button>
          ) : null,
        ]}
      />
      <Modal
        title={editing ? '编辑用户' : '新增用户'}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => {
          setModalOpen(false);
          setEditing(null);
        }}
        afterOpenChange={(open) => {
          if (open && editing) {
            form.setFieldsValue(editing as any);
          } else if (open && !editing) {
            form.resetFields();
          }
        }}
      >
        <UserForm form={form} editing={editing} roles={roleOptions} />
      </Modal>
    </>
  );
};

export default UserList;
