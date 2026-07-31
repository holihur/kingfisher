import React, { useRef, useState } from 'react';
import { Button, Modal, Form, Input, Select, message, Popconfirm, Tag } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import { useAuthStore } from '../../stores/auth';
import { userApi } from '../../api/user';

const UserList: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Record<string, unknown> | null>(null);
  const [form] = Form.useForm();
  const permissions = useAuthStore((s) => s.permissions);
  const hasPerm = (code: string) => permissions.includes(code);

  const columns: ProColumns[] = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '用户名', dataIndex: 'username' },
    { title: '邮箱', dataIndex: 'email', ellipsis: true },
    {
      title: '状态',
      dataIndex: 'status',
      render: (_, r) => (
        <Tag color={(r.status as number) === 1 ? 'green' : 'red'}>{(r.status as number) === 1 ? '启用' : '禁用'}</Tag>
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
              form.setFieldsValue(r as Record<string, unknown>);
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
      await userApi.create(values);
      message.success('创建成功');
    }
    setModalOpen(false);
    setEditing(null);
    form.resetFields();
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
        search={{ labelWidth: 'auto', syncToUrl: true } as Record<string, unknown>}
        pagination={{ pageSize: 20, showSizeChanger: true, syncToUrl: true } as Record<string, unknown>}
        headerTitle="用户管理"
        toolBarRender={() => [
          hasPerm('user:create') ? (
            <Button
              key="add"
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setEditing(null);
                form.resetFields();
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
          form.resetFields();
        }}
        destroyOnClose
      >
        <Form form={form} layout="vertical" preserve={false}>
          <Form.Item name="username" label="用户名" rules={[{ required: true, min: 3, max: 32 }]}>
            <Input disabled={!!editing?.id} />
          </Form.Item>
          {!editing?.id && (
            <Form.Item name="password" label="密码" rules={[{ required: true, min: 8 }]}>
              <Input.Password />
            </Form.Item>
          )}
          <Form.Item name="email" label="邮箱">
            <Input />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select
              options={[
                { label: '启用', value: 1 },
                { label: '禁用', value: 0 },
              ]}
            />
          </Form.Item>
          <Form.Item name="role_id" label="角色">
            <Select
              options={[
                { label: '管理员', value: 1 },
                { label: '编辑', value: 3 },
                { label: '访客', value: 4 },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
};

export default UserList;
