import React, { useEffect, useState } from 'react';
import { Button, Modal, Form, Input, InputNumber, Select, TreeSelect, Tag, message, Popconfirm } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import ProTable from '@ant-design/pro-table';
import { useAuthStore } from '../../stores/auth';
import { menuApi } from '../../api/menu';

interface MenuItem {
  id: number;
  parent_id: number;
  name: string;
  path: string;
  icon: string;
  sort: number;
  type: number;
  permission: string;
  children?: MenuItem[];
  level?: number;
}

const MenuManage: React.FC = () => {
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
    items.flatMap((i) => [{ ...i, level: lvl }, ...(i.children ? flatten(i.children, lvl + 1) : [])]);

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

  const typeTag = ['blue', 'green', 'orange'];
  const typeLabel = ['目录', '菜单', '按钮'];

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      render: (_: unknown, r: Record<string, unknown>) => (
        <span style={{ paddingLeft: ((r.level as number) || 0) * 24 }}>
          <Tag color={typeTag[((r.type as number) || 2) - 1]}>{typeLabel[((r.type as number) || 2) - 1]}</Tag>
          {r.name as string}
        </span>
      ),
    },
    { title: '路由', dataIndex: 'path' },
    { title: '排序', dataIndex: 'sort', width: 80 },
    {
      title: '操作',
      valueType: 'option',
      render: (_: unknown, r: Record<string, unknown>) => [
        perms.includes('menu:update') ? (
          <a
            key="ed"
            onClick={() => {
              setEditing(r as unknown as MenuItem);
              form.setFieldsValue(r as Record<string, unknown>);
              setModalOpen(true);
            }}
          >
            编辑
          </a>
        ) : null,
        (r.type as number) !== 3 && perms.includes('menu:create') ? (
          <a
            key="add"
            onClick={() => {
              setEditing({ parent_id: r.id as number } as MenuItem);
              form.resetFields();
              form.setFieldValue('parent_id', r.id);
              setModalOpen(true);
            }}
          >
            添加子项
          </a>
        ) : null,
        perms.includes('menu:delete') ? (
          <Popconfirm
            key="del"
            title="有子节点将无法删除"
            onConfirm={async () => {
              await menuApi.delete(r.id as number);
              message.success('已删除');
              fetchTree();
            }}
          >
            <a style={{ color: 'red' }}>删除</a>
          </Popconfirm>
        ) : null,
      ],
    },
  ];

  return (
    <>
      <ProTable
        columns={columns as Record<string, unknown>[]}
        dataSource={flatten(tree)}
        rowKey="id"
        search={false}
        pagination={false}
        headerTitle="菜单管理"
        toolBarRender={() => [
          perms.includes('menu:create') ? (
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
              新增根菜单
            </Button>
          ) : null,
        ]}
      />
      <Modal
        title={editing?.id ? '编辑菜单' : '新增菜单'}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => {
          setModalOpen(false);
          setEditing(null);
        }}
        destroyOnClose
      >
        <Form form={form} layout="vertical" preserve={false}>
          <Form.Item name="parent_id" label="上级菜单">
            <TreeSelect
              treeData={[{ title: '根菜单', value: 0, children: [] }]}
              placeholder="不选则为根菜单"
              allowClear
            />
          </Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select
              options={[
                { label: '目录', value: 1 },
                { label: '菜单', value: 2 },
                { label: '按钮', value: 3 },
              ]}
            />
          </Form.Item>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="path" label="路由路径">
            <Input placeholder="type=菜单时必填" />
          </Form.Item>
          <Form.Item name="icon" label="图标">
            <Input placeholder="AntD 图标名" />
          </Form.Item>
          <Form.Item name="permission" label="权限标识">
            <Input placeholder="如 user:create" />
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
