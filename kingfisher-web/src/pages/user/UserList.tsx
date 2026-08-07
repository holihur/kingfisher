import React, { useEffect, useState } from 'react';
import { Button, Modal, Form, App, Popconfirm, Badge } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import DataTable, { SearchField } from '../../components/DataTable';
import { useAuthStore } from '../../stores/auth';
import { userApi } from '../../api/user';
import { roleApi } from '../../api/role';
import UserForm from './UserForm';

interface UserRow {
  id: number;
  username: string;
  email: string;
  role_id: number;
  status: number;
  created_at: string;
}

const UserList: React.FC = () => {
  const { message } = App.useApp();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Record<string, unknown> | null>(null);
  const [form] = Form.useForm<Record<string, unknown>>();
  const [refreshKey, setRefreshKey] = useState(0);
  const permissions = useAuthStore((s) => s.permissions);
  const hasPerm = (code: string) => permissions.includes(code);
  const [roleOptions, setRoleOptions] = useState<{ label: string; value: number }[]>([]);
  const [roleNameMap, setRoleNameMap] = useState<Record<number, string>>({});

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

  const searchFields: SearchField[] = [
    { name: 'q', label: '关键词', type: 'text' },
    { name: 'role_id', label: '角色', type: 'select', options: roleOptions },
    {
      name: 'status',
      label: '状态',
      type: 'select',
      options: [
        { label: '启用', value: 1 },
        { label: '禁用', value: 0 },
      ],
    },
  ];

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '用户名', dataIndex: 'username' },
    { title: '邮箱', dataIndex: 'email', ellipsis: true },
    {
      title: '角色',
      dataIndex: 'role_id',
      width: 100,
      render: (_: unknown, r: UserRow) => roleNameMap[r.role_id] || `#${r.role_id}`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      render: (_: unknown, r: UserRow) => (
        <Badge status={r.status === 1 ? 'success' : 'error'} text={r.status === 1 ? '启用' : '禁用'} />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      render: (v: unknown) => (v ? new Date(v as string).toLocaleString() : '-'),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, r: UserRow) => [
        hasPerm('user:update') ? (
          <a
            key="edit"
            onClick={() => {
              setEditing(r as unknown as Record<string, unknown>);
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
              await userApi.delete(r.id);
              message.success('已删除');
              setRefreshKey((k) => k + 1);
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
      await userApi.create(values as never);
      message.success('创建成功');
    }
    setModalOpen(false);
    setEditing(null);
    setRefreshKey((k) => k + 1);
  };

  return (
    <>
      <DataTable<UserRow>
        columns={columns}
        rowKey="id"
        request={async (params) => {
          const resp = await userApi.getList(params);
          const data = resp.data as Record<string, unknown>;
          return {
            items: (data.items as UserRow[]) || [],
            total: (data.total as number) || 0,
          };
        }}
        searchFields={searchFields}
        headerTitle="用户管理"
        reloadKey={refreshKey}
        toolBarRender={
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
          ) : null
        }
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
            form.setFieldsValue(editing as never);
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
