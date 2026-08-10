import React, { useEffect, useState } from 'react';
import { Button, Modal, Form, Input, InputNumber, TreeSelect, App, Popconfirm, Switch, Tag, Dropdown } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, FileAddOutlined, DownOutlined, DashboardOutlined, SettingOutlined, UserOutlined, MenuOutlined, SafetyOutlined, ControlOutlined, AuditOutlined, BookOutlined } from '@ant-design/icons';
import DataTable from '../../components/DataTable';
import { useAuthStore } from '../../stores/auth';
import { useThemeToken } from '../../hooks/useThemeToken';
import { menuApi } from '../../api/menu';
import { formatTime } from '../../utils/format';

interface MenuItem {
  id: number;
  parent_id: number;
  name: string;
  path: string;
  icon: string;
  sort: number;
  status?: number;
  version?: string;
  children?: MenuItem[];
  level?: number;
}

const MenuManage: React.FC = () => {
  const { message, modal } = App.useApp();
  const token = useThemeToken();
  const [tree, setTree] = useState<MenuItem[]>([]);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<MenuItem | null>(null);
  const [form] = Form.useForm();
  const perms = useAuthStore((s) => s.permissions);


  const fetchTree = async () => {
    const r = await menuApi.getTree();
    setTree((r.data as MenuItem[]) || []);
  };
  useEffect(() => {
    fetchTree();
  }, []);

  const flatten = (items: MenuItem[], lvl = 0): Record<string, unknown>[] =>
    items.flatMap((i) => {
      const { children, ...rest } = i as MenuItem & { children?: MenuItem[] };
      return [{ ...rest, level: lvl }, ...(children ? flatten(children, lvl + 1) : [])];
    });

  const handleSubmit = async () => {
    const v = await form.validateFields();
    if (editing?.id) {
      await menuApi.update(editing.id, v);
    } else {
      await menuApi.create({ ...v, version: v.version || '1.0.0' });
    }
    message.success('保存成功');
    setModalOpen(false);
    setEditing(null);
    fetchTree();
  };

  // 菜单图标字符串 → 图标组件映射
  const iconMap: Record<string, React.ReactNode> = {
    DashboardOutlined: <DashboardOutlined />,
    SettingOutlined: <SettingOutlined />,
    UserOutlined: <UserOutlined />,
    MenuOutlined: <MenuOutlined />,
    SafetyOutlined: <SafetyOutlined />,
    ControlOutlined: <ControlOutlined />,
    AuditOutlined: <AuditOutlined />,
    BookOutlined: <BookOutlined />,
  };

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      render: (_: unknown, r: Record<string, unknown>) => (
        <span style={{ paddingLeft: ((r.level as number) || 0) * 24 }}>
          {r.parent_id ? '↳ ' : ''}
          {iconMap[(r.icon as string) || ''] ? <span style={{ marginRight: 6, color: token.colorTextSecondary }}>{iconMap[(r.icon as string) || '']}</span> : null}
          {r.name as string}
        </span>
      ),
    },
    { title: '路由', dataIndex: 'path' },
    { title: '排序', dataIndex: 'sort', width: 80 },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: (_: unknown, r: Record<string, unknown>) => {
        const enabled = (r.status as number) === 1;
        return (
          <Switch
            size="small"
            checked={enabled}
            checkedChildren="启用"
            unCheckedChildren="禁用"
            disabled={!perms.includes('menu:update')}
            onChange={async (checked) => {
              await menuApi.update(r.id as number, { status: checked ? 1 : 0 });
              message.success(checked ? '已启用' : '已禁用');
              fetchTree();
            }}
          />
        );
      },
    },
    { title: '版本', dataIndex: 'version', width: 90, render: (v: unknown) => (v ? <Tag>{v as string}</Tag> : '-') },
    { title: '更新时间', dataIndex: 'updated_at', width: 150, render: (v: unknown) => formatTime(v) },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, r: Record<string, unknown>) => [
        perms.includes('menu:update') ? (
          <a key="ed" onClick={() => { setEditing(r as unknown as MenuItem); setModalOpen(true); }}><EditOutlined /> 编辑</a>
        ) : null,
        perms.includes('menu:create') ? (
          <a key="add" onClick={() => { setEditing({ parent_id: r.id as number } as MenuItem); setModalOpen(true); }}><FileAddOutlined /> 添加子项</a>
        ) : null,
        perms.includes('menu:delete') ? (
          (r.children as MenuItem[])?.length > 0 ? (
            <a key="del-disabled" style={{ color: token.colorTextDisabled, cursor: 'not-allowed' }} title="有子节点无法删除"><DeleteOutlined /> 删除</a>
          ) : (
            <Popconfirm key="del" title="确认删除？" onConfirm={async () => { try { await menuApi.delete(r.id as number); message.success('已删除'); fetchTree(); } catch { /* error shown by interceptor */ } }}>
              <a style={{ color: 'red' }}><DeleteOutlined /> 删除</a>
            </Popconfirm>
          )
        ) : null,
      ],
    },
  ];

  return (
    <>
      <DataTable<Record<string, unknown>>
        columns={columns}
        dataSource={flatten(tree)}
        rowKey="id"
        headerTitle="菜单管理"
        selectable={perms.includes('menu:update') || perms.includes('menu:delete')}
        batchBarRender={(keys, clear) => {
          const ids = keys as number[];
          const runStatus = async (status: number, label: string) => {
            await menuApi.batchUpdateStatus(ids, status);
            message.success(`已${label}`);
            clear();
            fetchTree();
          };
          return (
            <Dropdown
              menu={{
                items: [
                  ...(perms.includes('menu:update') ? [{ key: 'enable', label: '批量启用' }] : []),
                  ...(perms.includes('menu:update') ? [{ key: 'disable', label: '批量禁用' }] : []),
                  ...(perms.includes('menu:delete') ? [{ key: 'delete', label: '批量删除', danger: true }] : []),
                ],
                onClick: ({ key }) => {
                  if (key === 'enable') void runStatus(1, '批量启用');
                  else if (key === 'disable') void runStatus(0, '批量禁用');
                  else if (key === 'delete') {
                    modal.confirm({
                      title: '批量删除',
                      content: `确定删除选中的 ${ids.length} 个菜单吗？（有子节点的菜单不可删除）`,
                      onOk: async () => {
                        await menuApi.batchDelete(ids);
                        message.success('已删除');
                        clear();
                        fetchTree();
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
          perms.includes('menu:create') ? (
            <Button key="add" type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); setModalOpen(true); }}>
              新增根菜单
            </Button>
          ) : null
        }
      />
      <Modal title={editing?.id ? '编辑菜单' : '新增菜单'} open={modalOpen} onOk={handleSubmit}
        onCancel={() => { setModalOpen(false); setEditing(null); }}
        afterOpenChange={(open) => {
          if (open && editing) {
            form.setFieldsValue(editing as any);
          } else if (open && !editing) {
            form.resetFields();
          }
        }}>
        <Form form={form} layout="vertical">
          <Form.Item name="parent_id" label="上级菜单">
            <TreeSelect treeData={[{ title: '根菜单', value: 0, children: (tree || []).map((m: MenuItem) => ({
              title: m.name, value: m.id,
              children: (m.children || []).map((c: MenuItem) => ({ title: c.name, value: c.id })),
            })) }]} placeholder="不选则为根菜单" allowClear />
          </Form.Item>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="path" label="路由路径">
            <Input />
          </Form.Item>
          <Form.Item name="icon" label="图标">
            <Input placeholder="AntD 图标名" />
          </Form.Item>
          <Form.Item name="sort" label="排序">
            <InputNumber min={0} />
          </Form.Item>
          <Form.Item name="version" label="版本（表示该菜单由哪个版本新增）">
            <Input placeholder="如 1.0.0" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
};

export default MenuManage;
