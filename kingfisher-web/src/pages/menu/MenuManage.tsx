import React, { useEffect, useState } from 'react';
import { Button, Modal, Form, Input, InputNumber, TreeSelect, App, Popconfirm, Switch } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import DataTable from '../../components/DataTable';
import { useAuthStore } from '../../stores/auth';
import { menuApi } from '../../api/menu';
import { formatTime } from '../../utils/format';

interface MenuItem {
  id: number;
  parent_id: number;
  name: string;
  path: string;
  icon: string;
  sort: number;
  children?: MenuItem[];
  level?: number;
}

const MenuManage: React.FC = () => {
  const { message } = App.useApp();
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
      await menuApi.create(v);
    }
    message.success('保存成功');
    setModalOpen(false);
    setEditing(null);
    fetchTree();
  };

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      render: (_: unknown, r: Record<string, unknown>) => (
        <span style={{ paddingLeft: ((r.level as number) || 0) * 24 }}>
          {r.parent_id ? '↳ ' : ''}{r.name as string}
        </span>
      ),
    },
    { title: '路由', dataIndex: 'path' },
    { title: '图标', dataIndex: 'icon', width: 120 },
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
    { title: '更新时间', dataIndex: 'updated_at', width: 150, render: (v: unknown) => formatTime(v) },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, r: Record<string, unknown>) => [
        perms.includes('menu:update') ? (
          <a key="ed" onClick={() => { setEditing(r as unknown as MenuItem); setModalOpen(true); }}>编辑</a>
        ) : null,
        perms.includes('menu:create') ? (
          <a key="add" onClick={() => { setEditing({ parent_id: r.id as number } as MenuItem); setModalOpen(true); }}>添加子项</a>
        ) : null,
        perms.includes('menu:delete') ? (
          (r.children as MenuItem[])?.length > 0 ? (
            <a key="del-disabled" style={{ color: '#ccc', cursor: 'not-allowed' }} title="有子节点无法删除">删除</a>
          ) : (
            <Popconfirm key="del" title="确认删除？" onConfirm={async () => { try { await menuApi.delete(r.id as number); message.success('已删除'); fetchTree(); } catch { /* error shown by interceptor */ } }}>
              <a style={{ color: 'red' }}>删除</a>
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
        </Form>
      </Modal>
    </>
  );
};

export default MenuManage;
