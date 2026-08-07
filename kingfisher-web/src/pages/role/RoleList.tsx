import React, { useEffect, useState } from 'react';
import { Button, Modal, Form, Input, Tree, Tabs, Checkbox, Row, Col, App, Popconfirm, Badge, AutoComplete, Tag, Dropdown } from 'antd';
import { PlusOutlined, SafetyOutlined, AppstoreOutlined, EditOutlined, DeleteOutlined, DownOutlined } from '@ant-design/icons';
import DataTable, { SearchField } from '../../components/DataTable';
import { useAuthStore } from '../../stores/auth';
import { roleApi } from '../../api/role';
import { menuApi } from '../../api/menu';
import { formatTime } from '../../utils/format';

interface RoleRow {
  id: number;
  name: string;
  code: string;
  description: string;
  status: number;
  landing_page: string;
  created_at: string;
  updated_at: string;
}

const RoleList: React.FC = () => {
  const { message, modal } = App.useApp();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Record<string, unknown> | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);
  const [permModal, setPermModal] = useState<{ open: boolean; role: Record<string, unknown> | null }>({
    open: false,
    role: null,
  });
  const [menuModal, setMenuModal] = useState<{ open: boolean; role: Record<string, unknown> | null }>({
    open: false,
    role: null,
  });
  const [allPerms, setAllPerms] = useState<Record<string, unknown>[]>([]);
  const [selectedPerms, setSelectedPerms] = useState<number[]>([]);
  const [allMenus, setAllMenus] = useState<Record<string, unknown>[]>([]);
  const [selectedMenus, setSelectedMenus] = useState<number[]>([]);
  // 落地页候选（菜单路径）
  const [landingOptions, setLandingOptions] = useState<{ label: string; value: string }[]>([]);
  const [form] = Form.useForm<Record<string, unknown>>();
  const perms = useAuthStore((s) => s.permissions);

  // 加载落地页候选
  useEffect(() => {
    menuApi.getTree().then((r) => {
      const walk = (nodes: Record<string, unknown>[], out: { label: string; value: string }[]) => {
        (nodes || []).forEach((n) => {
          if (n.path) out.push({ label: `${n.name as string} (${n.path as string})`, value: n.path as string });
          walk(n.children as Record<string, unknown>[], out);
        });
      };
      const opts: { label: string; value: string }[] = [];
      walk(r.data as Record<string, unknown>[], opts);
      setLandingOptions(opts);
    });
  }, []);

  const searchFields: SearchField[] = [{ name: 'q', label: '关键词', type: 'text' }];

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '角色名', dataIndex: 'name' },
    { title: '编码', dataIndex: 'code', render: (_: unknown, r: RoleRow) => <Tag color={r.code === 'admin' ? 'gold' : 'blue'}>{r.code}</Tag> },
    { title: '描述', dataIndex: 'description', ellipsis: true },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: (_: unknown, r: RoleRow) => (
        <Badge status={r.status === 1 ? 'success' : 'error'} text={r.status === 1 ? '启用' : '禁用'} />
      ),
    },
    {
      title: '落地页',
      dataIndex: 'landing_page',
      width: 150,
      ellipsis: true,
      render: (_: unknown, r: RoleRow) =>
        r.landing_page ? <a href={r.landing_page}>{r.landing_page}</a> : '-',
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
      render: (_: unknown, r: RoleRow) => [
        perms.includes('role:update') ? (
          <a
            key="perm"
            onClick={async () => {
              const p = await roleApi.getAllPermissions();
              setAllPerms((p.data as Record<string, unknown>[]) || []);
              const rp = await roleApi.getPermissions(r.id);
              setSelectedPerms(
                ((rp.data as Record<string, unknown>[]) || []).map((i: Record<string, unknown>) => i.id as number)
              );
              setPermModal({ open: true, role: r as unknown as Record<string, unknown> });
            }}
          >
            <SafetyOutlined /> 权限
          </a>
        ) : null,
        perms.includes('role:update') ? (
          <a
            key="menu"
            onClick={async () => {
              const m = await menuApi.getTree();
              setAllMenus((m.data as Record<string, unknown>[]) || []);
              const rm = await roleApi.getMenus(r.id);
              setSelectedMenus(
                ((rm.data as Record<string, unknown>[]) || []).map((i: Record<string, unknown>) => i.id as number)
              );
              setMenuModal({ open: true, role: r as unknown as Record<string, unknown> });
            }}
          >
            <AppstoreOutlined /> 菜单
          </a>
        ) : null,
        perms.includes('role:update') ? (
          <a
            key="ed"
            onClick={() => {
              setEditing(r as unknown as Record<string, unknown>);
              setModalOpen(true);
            }}
          >
            <EditOutlined /> 编辑
          </a>
        ) : null,
        perms.includes('role:delete') ? (
          <Popconfirm
            key="del"
            title="删除角色"
            description={`确定删除角色「${r.name}」吗？`}
            onConfirm={async () => {
              await roleApi.delete(r.id);
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
    const v = await form.validateFields();
    if (editing?.id) {
      await roleApi.update(editing.id as number, v as never);
    } else {
      await roleApi.create(v as never);
    }
    message.success('保存成功');
    setModalOpen(false);
    setEditing(null);
    setRefreshKey((k) => k + 1);
  };

  const grouped = (allPerms || []).reduce(
    (acc: Record<string, Record<string, unknown>[]>, p: Record<string, unknown>) => {
      const k = p.resource as string;
      if (!acc[k]) acc[k] = [];
      acc[k].push(p);
      return acc;
    },
    {}
  );

  // 递归转换菜单树：给每个节点补上 key/title，否则 Tree 组件显示 ---
  const toTreeData = (items: Record<string, unknown>[]): Record<string, unknown>[] =>
    items.map((m) => ({
      ...m,
      key: m.id as number,
      title: m.name as string,
      children: m.children ? toTreeData(m.children as Record<string, unknown>[]) : undefined,
    }));

  return (
    <>
      <DataTable<RoleRow>
        columns={columns}
        rowKey="id"
        request={async (params) => {
          const r = await roleApi.getList(params);
          const data = r.data as Record<string, unknown>;
          return {
            items: (data.items as RoleRow[]) || [],
            total: (data.total as number) || 0,
          };
        }}
        searchFields={searchFields}
        headerTitle="角色管理"
        reloadKey={refreshKey}
        selectable={perms.includes('role:update') || perms.includes('role:delete')}
        batchBarRender={(keys, clear) => {
          const ids = keys as number[];
          const runStatus = async (status: number, label: string) => {
            await roleApi.batchUpdateStatus(ids, status);
            message.success(`已${label}`);
            clear();
            setRefreshKey((k) => k + 1);
          };
          return (
            <Dropdown
              menu={{
                items: [
                  ...(perms.includes('role:update') ? [{ key: 'enable', label: '批量启用' }] : []),
                  ...(perms.includes('role:update') ? [{ key: 'disable', label: '批量禁用' }] : []),
                  ...(perms.includes('role:delete') ? [{ key: 'delete', label: '批量删除', danger: true }] : []),
                ],
                onClick: ({ key }) => {
                  if (key === 'enable') void runStatus(1, '批量启用');
                  else if (key === 'disable') void runStatus(0, '批量禁用');
                  else if (key === 'delete') {
                    modal.confirm({
                      title: '批量删除',
                      content: `确定删除选中的 ${ids.length} 个角色吗？（admin 角色不可删除）`,
                      onOk: async () => {
                        await roleApi.batchDelete(ids);
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
          perms.includes('role:create') ? (
            <Button
              key="add"
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setEditing(null);
                setModalOpen(true);
              }}
            >
              新增角色
            </Button>
          ) : null
        }
      />
      <Modal
        title={editing ? '编辑角色' : '新增角色'}
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
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="角色名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="code" label="编码" rules={[{ required: true }]}>
            <Input disabled={!!editing?.id} />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item
            name="landing_page"
            label="落地页（登录后跳转的页面）"
            extra="可从菜单路径选择，或手动输入路径"
          >
            <AutoComplete
              options={landingOptions}
              placeholder="如 /dashboard"
              allowClear
              filterOption={(inputValue, option) =>
                (option?.value as string)?.toLowerCase().includes(inputValue.toLowerCase()) ?? true
              }
            />
          </Form.Item>
        </Form>
      </Modal>
      <Modal
        title={`分配权限 — ${permModal.role?.name || ''}`}
        open={permModal.open}
        width={700}
        onOk={async () => {
          await roleApi.assignPermissions(permModal.role!.id as number, selectedPerms);
          message.success('权限已更新');
          setPermModal({ open: false, role: null });
        }}
        onCancel={() => setPermModal({ open: false, role: null })}
      >
        <Checkbox.Group value={selectedPerms} onChange={(v) => setSelectedPerms(v as number[])} style={{ width: '100%' }}>
          <Tabs
            items={Object.entries(grouped).map(([res, ps]) => ({
              key: res,
              label: { user: '用户', menu: '菜单', role: '角色', config: '配置', audit: '审计', dict: '字典' }[res] || res,
              children: (
                <Row gutter={[16, 8]}>
                  {ps.map((p) => (
                    <Col span={12} key={p.id as number}>
                      <Checkbox value={p.id}>{p.name as string}</Checkbox>
                    </Col>
                  ))}
                </Row>
              ),
            }))}
          />
        </Checkbox.Group>
      </Modal>
      <Modal
        title={`分配菜单 — ${menuModal.role?.name || ''}`}
        open={menuModal.open}
        onOk={async () => {
          await roleApi.assignMenus(menuModal.role!.id as number, selectedMenus);
          message.success('菜单已更新');
          setMenuModal({ open: false, role: null });
        }}
        onCancel={() => setMenuModal({ open: false, role: null })}
      >
        <Tree
          checkable
          defaultExpandAll
          key={menuModal.role?.id as number}
          checkedKeys={selectedMenus}
          onCheck={(keys) => {
            const arr = Array.isArray(keys) ? keys : (keys as { checked: React.Key[] }).checked;
            setSelectedMenus(arr as number[]);
          }}
          treeData={toTreeData(allMenus as Record<string, unknown>[]) as Record<string, unknown>[]}
        />
      </Modal>
    </>
  );
};

export default RoleList;
