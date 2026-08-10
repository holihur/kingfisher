import React, { useEffect, useState } from 'react';
import { Button, Modal, Form, App, Popconfirm, Badge, Avatar, Tag, Space, Dropdown } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, DownOutlined } from '@ant-design/icons';
import DataTable, { SearchField } from '../../components/DataTable';
import { useAuthStore } from '../../stores/auth';
import { useThemeToken } from '../../hooks/useThemeToken';
import { userApi } from '../../api/user';
import { roleApi } from '../../api/role';
import { formatTime } from '../../utils/format';
import UserForm from './UserForm';

interface UserRow {
  id: number;
  username: string;
  nickname?: string;
  avatar?: string;
  email: string;
  role_ids?: number[];
  roles?: { id: number; name: string }[];
  status: number;
  created_at: string;
  updated_at: string;
}

const UserList: React.FC = () => {
  const { message, modal } = App.useApp();
  const token = useThemeToken();
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
    {
      title: '用户',
      dataIndex: 'username',
      render: (_: unknown, r: UserRow) => (
        <Space>
          <Avatar size="small" src={r.avatar || undefined}>{r.username?.charAt(0)?.toUpperCase()}</Avatar>
          <span>
            {r.username}
            {r.nickname ? <span style={{ color: token.colorTextTertiary, marginLeft: 6, fontSize: 12 }}>({r.nickname})</span> : null}
          </span>
        </Space>
      ),
    },
    { title: '邮箱', dataIndex: 'email', ellipsis: true },
    {
      title: '角色',
      dataIndex: 'role_ids',
      width: 200,
      render: (_: unknown, r: UserRow) => {
        const list = r.roles?.length ? r.roles : (r.role_ids || []).map((id) => ({ id, name: roleNameMap[id] || `#${id}` }));
        if (!list.length) return <span>-</span>;
        return (
          <Space size={[0, 4]} wrap>
            {list.map((role) => (
              <Tag key={role.id} color={role.id === 1 ? 'gold' : role.id === 3 ? 'blue' : 'default'}>{role.name}</Tag>
            ))}
          </Space>
        );
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: (_: unknown, r: UserRow) => (
        <Badge status={r.status === 1 ? 'success' : 'error'} text={r.status === 1 ? '启用' : '禁用'} />
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      width: 150,
      render: (v: unknown) => formatTime(v),
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
            <EditOutlined /> 编辑
          </a>
        ) : null,
        hasPerm('user:delete') ? (
          <Popconfirm
            key="del"
            title="删除用户"
            description={`确定删除用户「${r.username}」吗？`}
            onConfirm={async () => {
              await userApi.delete(r.id);
              message.success('已删除');
              setRefreshKey((k) => k + 1);
            }}
          >
            <a style={{ color: 'red' }}><DeleteOutlined /> 删除</a>
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
        selectable={hasPerm('user:update') || hasPerm('user:delete')}
        batchBarRender={(keys, clear) => {
          const ids = keys as number[];
          const runStatus = async (status: number, label: string) => {
            await userApi.batchUpdateStatus(ids, status);
            message.success(`已${label}`);
            clear();
            setRefreshKey((k) => k + 1);
          };
          return (
            <Dropdown
              menu={{
                items: [
                  ...(hasPerm('user:update') ? [{ key: 'enable', label: '批量启用' }] : []),
                  ...(hasPerm('user:update') ? [{ key: 'disable', label: '批量禁用' }] : []),
                  ...(hasPerm('user:delete') ? [{ key: 'delete', label: '批量删除', danger: true }] : []),
                ],
                onClick: ({ key }) => {
                  if (key === 'enable') void runStatus(1, '批量启用');
                  else if (key === 'disable') void runStatus(0, '批量禁用');
                  else if (key === 'delete') {
                    modal.confirm({
                      title: '批量删除',
                      content: `确定删除选中的 ${ids.length} 个用户吗？`,
                      onOk: async () => {
                        await userApi.batchDelete(ids);
                        message.success('已删除');
                        clear();
                        setRefreshKey((k) => k + 1);
                      },
                    });
                  }
                },
              }}
            >
              <Button size="small">
                批量操作 <DownOutlined />
              </Button>
            </Dropdown>
          );
        }}
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
            // 角色回填：roles 数组 → role_ids（多角色多选）
            const editRow = editing as Record<string, unknown>;
            const rolesArr = editRow.roles as { id: number }[] | undefined;
            const roleIds = rolesArr?.length
              ? rolesArr.map((r) => r.id)
              : ((editRow.role_ids as number[]) || []);
            form.setFieldsValue({ ...editRow, role_ids: roleIds });
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
