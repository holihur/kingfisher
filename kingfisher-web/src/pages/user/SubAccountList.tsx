import React, { useEffect, useState } from 'react';
import { Button, Modal, Form, Input, Select, App, Popconfirm, Tag, Space, Card, Alert, Tabs, Checkbox } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { userApi } from '../../api/user';
import { roleApi } from '../../api/role';

interface SubUser {
  id: number;
  username: string;
  email: string;
  status: number;
  role_ids: number[];
  roles?: { id: number; name: string }[];
  parent_id?: number;
}

const SubAccountList: React.FC = () => {
  const { message } = App.useApp();
  const [subs, setSubs] = useState<SubUser[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<SubUser | null>(null);
  const [form] = Form.useForm();
  const [roleOptions, setRoleOptions] = useState<{ label: string; value: number }[]>([]);
  const [myRoleIds, setMyRoleIds] = useState<number[]>([]);
  const [myPerms, setMyPerms] = useState<string[]>([]);
  const [subRoleModal, setSubRoleModal] = useState(false);
  const [roleForm] = Form.useForm();
  const [allPerms, setAllPerms] = useState<Record<string, unknown>[]>([]);
  const [selectedPerms, setSelectedPerms] = useState<number[]>([]);

  const load = async () => {
    setLoading(true);
    try {
      const r = await userApi.getSubAccounts();
      setSubs((r.data as SubUser[]) || []);
    } catch {
      message.error('加载子账户失败');
    } finally {
      setLoading(false);
    }
  };

  const loadRoles = async () => {
    try {
      const me = await userApi.getMe();
      const meData = me.data as Record<string, unknown>;
      const ids = (meData.role_ids as number[]) || [];
      setMyRoleIds(ids);
      const permRes = await userApi.getMyPermissions();
      setMyPerms((permRes.data as string[]) || []);
      const roleRes = await roleApi.getList({ page: 1, page_size: 100 });
      const data = roleRes.data as Record<string, unknown>;
      const items = (data.items as Record<string, unknown>[]) || [];
      setRoleOptions(items.map((i) => ({ label: i.name as string, value: i.id as number })));
    } catch {
      const roleRes = await roleApi.getList({ page: 1, page_size: 100 });
      const data = roleRes.data as Record<string, unknown>;
      const items = (data.items as Record<string, unknown>[]) || [];
      setRoleOptions(items.map((i) => ({ label: i.name as string, value: i.id as number })));
    }
  };

  const loadAllPerms = async () => {
    try {
      const p = await roleApi.getAllPermissions();
      const list = (p.data as Record<string, unknown>[]) || [];
      const filtered = list.filter((perm) => myPerms.includes(perm.code as string));
      setAllPerms(filtered.length ? filtered : list);
    } catch {
      /* ignore */
    }
  };

  useEffect(() => {
    load();
    loadRoles();
  }, []);

  useEffect(() => {
    if (subRoleModal) loadAllPerms();
  }, [subRoleModal, myPerms]);

  const handleCreate = async () => {
    const v = await form.validateFields();
    if (editing) {
      await userApi.updateSubAccount(editing.id, { role_ids: v.role_ids });
      message.success('子账户角色已更新');
    } else {
      await userApi.createSubAccount(v as never);
      message.success('子账户创建成功');
    }
    setModalOpen(false);
    setEditing(null);
    form.resetFields();
    load();
  };

  const handleCreateSubRole = async () => {
    const v = await roleForm.validateFields();
    const r = await roleApi.create({ name: v.name, code: v.code, description: v.description, status: 1 } as never);
    const newRole = r.data as Record<string, unknown>;
    const roleId = newRole.id as number;
    if (selectedPerms.length) {
      await roleApi.assignPermissions(roleId, selectedPerms);
    }
    message.success('子角色创建成功，已限制为父账户权限子集');
    setSubRoleModal(false);
    roleForm.resetFields();
    setSelectedPerms([]);
    loadRoles();
  };

  const grouped = (allPerms || []).reduce(
    (acc: Record<string, Record<string, unknown>[]>, p: Record<string, unknown>) => {
      const k = p.resource as string;
      if (!acc[k]) acc[k] = [];
      acc[k].push(p);
      return acc;
    },
    {},
  );

  return (
    <div>
      <Card
        title="我的子账户"
        extra={
          <Space>
            <Button
              onClick={() => {
                setSubRoleModal(true);
              }}
            >
              创建子角色
            </Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setEditing(null);
                form.resetFields();
                setModalOpen(true);
              }}
            >
              创建子账户
            </Button>
          </Space>
        }
        loading={loading}
      >
        <Alert
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
          message="子账户权限为父账户权限的子集；父账户被取消的权限会自动从子账户移除，新增权限不会自动赋予子账户；仅支持一级父子。"
        />
        {subs.length === 0 ? (
          <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>暂无子账户</div>
        ) : (
          <Space direction="vertical" style={{ width: '100%' }} size={12}>
            {subs.map((u) => (
              <Card key={u.id} size="small">
                <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                  <Space>
                    <strong>{u.username}</strong>
                    <span style={{ color: '#999' }}>{u.email}</span>
                    <Tag color={u.status === 1 ? 'green' : 'red'}>{u.status === 1 ? '启用' : '禁用'}</Tag>
                  </Space>
                  <Space wrap>
                    {(u.roles || u.role_ids || []).map((r: unknown) => {
                      const role = typeof r === 'number' ? { id: r, name: `#${r}` } : (r as { id: number; name: string });
                      return <Tag key={role.id}>{role.name}</Tag>;
                    })}
                    <Button
                      size="small"
                      icon={<EditOutlined />}
                      onClick={() => {
                        setEditing(u);
                        form.setFieldsValue({ role_ids: u.role_ids });
                        setModalOpen(true);
                      }}
                    >
                      设置角色
                    </Button>
                    <Popconfirm
                      title="删除子账户"
                      description={`确定删除子账户「${u.username}」吗？`}
                      onConfirm={async () => {
                        await userApi.deleteSubAccount(u.id);
                        message.success('已删除');
                        load();
                      }}
                    >
                      <Button size="small" danger icon={<DeleteOutlined />}>
                        删除
                      </Button>
                    </Popconfirm>
                  </Space>
                </Space>
              </Card>
            ))}
          </Space>
        )}
      </Card>

      <Modal
        title={editing ? '设置子账户角色' : '创建子账户'}
        open={modalOpen}
        onOk={handleCreate}
        onCancel={() => {
          setModalOpen(false);
          setEditing(null);
        }}
        afterOpenChange={(open) => {
          if (open && editing) {
            form.setFieldsValue({ role_ids: editing.role_ids });
          } else if (open && !editing) {
            form.resetFields();
          }
        }}
      >
        <Form form={form} layout="vertical">
          {!editing && (
            <>
              <Form.Item name="username" label="用户名" rules={[{ required: true, min: 3, max: 32 }]}>
                <Input placeholder="子账户用户名" />
              </Form.Item>
              <Form.Item name="password" label="密码" rules={[{ required: true, min: 8 }]}>
                <Input.Password placeholder="至少8位" />
              </Form.Item>
              <Form.Item name="email" label="邮箱" rules={[{ type: 'email' }]}>
                <Input />
              </Form.Item>
            </>
          )}
          <Form.Item
            name="role_ids"
            label="角色"
            rules={[{ required: true, message: '至少选择一个角色' }]}
            extra="仅可选择父账户拥有的角色子集"
          >
            <Select mode="multiple" options={roleOptions} placeholder="选择角色（父账户权限子集）" />
          </Form.Item>
          {myRoleIds.length === 0 && <Alert type="warning" message="未获取到父账户角色，请检查权限" style={{ marginTop: 8 }} />}
        </Form>
      </Modal>

      <Modal
        title="创建子角色"
        open={subRoleModal}
        onOk={handleCreateSubRole}
        onCancel={() => {
          setSubRoleModal(false);
          setSelectedPerms([]);
        }}
        width={700}
      >
        <Alert type="info" message="子角色的权限必须为父账户权限的子集，已自动过滤" style={{ marginBottom: 12 }} />
        <Form form={roleForm} layout="vertical">
          <Form.Item name="name" label="角色名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="code" label="编码" rules={[{ required: true }]}>
            <Input placeholder="如 sub_editor" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} />
          </Form.Item>
        </Form>
        <div style={{ marginTop: 16 }}>
          <div style={{ fontWeight: 600, marginBottom: 8 }}>权限（已过滤为父账户权限子集）</div>
          <Checkbox.Group value={selectedPerms} onChange={(v) => setSelectedPerms(v as number[])} style={{ width: '100%' }}>
            <Tabs
              tabPosition="left"
              items={Object.entries(grouped).map(([res, ps]) => ({
                key: res,
                label: res,
                children: (
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: '4px 24px' }}>
                    {ps.map((p) => (
                      <Checkbox key={p.id as number} value={p.id} style={{ whiteSpace: 'nowrap' }}>
                        {p.name as string}
                      </Checkbox>
                    ))}
                  </div>
                ),
              }))}
            />
          </Checkbox.Group>
        </div>
      </Modal>
    </div>
  );
};

export default SubAccountList;
