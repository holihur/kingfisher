import React, { useEffect, useRef, useState } from 'react';
import { Button, Modal, Form, message, Popconfirm, Badge } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import { useAuthStore } from '../../stores/auth';
import { userApi } from '../../api/user';
import { roleApi } from '../../api/role';
import UserForm from './UserForm';

const UserList: React.FC = () => {
  const actionRef = useRef<ActionType>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Record<string, unknown> | null>(null);
  const [form] = Form.useForm<Record<string, unknown>>();
  const permissions = useAuthStore((s) => s.permissions);
  const hasPerm = (code: string) => permissions.includes(code);
  const [roleOptions, setRoleOptions] = useState<{ label: string; value: number }[]>([]);
  const [roleNameMap, setRoleNameMap] = useState<Record<number, string>>({});

  useEffect(() => {
    roleApi.getList().then((r) => {
      const items = (r.data as Record<string, unknown>[]) || [];
      setRoleOptions(items.map((i) => ({ label: i.name as string, value: i.id as number })));
      const map: Record<number, string> = {};
      items.forEach((i) => { map[i.id as number] = i.name as string; });
      setRoleNameMap(map);
    });
  }, []);

  const columns: ProColumns[] = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '用户名', dataIndex: 'username' },
    { title: '邮箱', dataIndex: 'email', ellipsis: true },
    {
      title: '角色',
      dataIndex: 'role_id',
      width: 100,
      render: (_, r) => roleNameMap[r.role_id as number] || `#${r.role_id}`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      render: (_, r) => (
        <Badge status={(r.status as number) === 1 ? 'success' : 'error'} text={(r.status as number) === 1 ? '启用' : '禁用'} />
      ),
    },
    { title: '创建时间', dataIndex: 'created_at', valueType: 'dateTime' },
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
        request={async (params) => {
          const resp = await userApi.getList({
            page: params.current || 1,
            page_size: params.pageSize || 20,
            keyword: params.keyword as string,
          });
          const data = resp.data as Record<string, unknown>;
          return {
            data: (data.items as Record<string, unknown>[]) || [],
            total: (data.total as number) || 0,
            success: true,
          };
        }}
        rowKey="id"
        search={{ labelWidth: 'auto' }}
        pagination={{ pageSize: 20, showSizeChanger: true }}
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
