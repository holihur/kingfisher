import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { App, Badge, Button, Descriptions, Dropdown, Empty, Form, Input, Modal, Popconfirm, Select, Space, Tag, Tree, TreeSelect } from 'antd';
import {
  ApartmentOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  DownOutlined,
  EyeOutlined,
  FolderOutlined,
} from '@ant-design/icons';
import type { TreeDataNode } from 'antd';
import DataTable from '../../components/DataTable';
import { useThemeToken } from '../../hooks/useThemeToken';
import { useAuthStore } from '../../stores/auth';
import { departmentApi, type DeptNode } from '../../api/department';
import { roleApi } from '../../api/role';

interface DeptRow {
  id: number;
  parent_id: number;
  name: string;
  sort: number;
  status: number;
  remark?: string;
  role_ids?: number[];
  roles?: { id: number; name: string; code: string }[];
  created_at: string;
  updated_at: string;
}

interface DeptFormValues {
  parent_id?: number;
  name?: string;
  sort?: number;
  status?: number;
  remark?: string;
  role_ids?: number[];
}

/** 部门树 → antd Tree data */
function toTreeData(nodes: DeptNode[]): TreeDataNode[] {
  return nodes.map((n) => ({
    key: n.id,
    title: n.name,
    children: n.children?.length ? toTreeData(n.children) : undefined,
  }));
}

/** TreeSelect 节点类型（children 递归） */
interface DeptSelectNode {
  title: string;
  value: number;
  children?: DeptSelectNode[];
}

/** 部门树 → TreeSelect data（选择上级部门用，禁止选自己/子孙） */
function toSelectTree(nodes: DeptNode[], disabledId?: number): DeptSelectNode[] {
  return nodes
    .filter((n) => n.id !== disabledId)
    .map((n) => {
      const children = n.children?.length ? toSelectTree(n.children, disabledId) : undefined;
      return children === undefined
        ? { title: n.name, value: n.id }
        : { title: n.name, value: n.id, children };
    });
}

/** 找部门名称（列表里上级部门列展示用） */
function findName(nodes: DeptNode[], id: number): string {
  for (const n of nodes) {
    if (n.id === id) return n.name;
    if (n.children?.length) {
      const r = findName(n.children, id);
      if (r) return r;
    }
  }
  return id ? `#${id}` : '—';
}

const DeptManage: React.FC = () => {
  const { message, modal } = App.useApp();
  const token = useThemeToken();
  const permissions = useAuthStore((s) => s.permissions);
  const hasPerm = (code: string) => permissions.includes(code);

  const [tree, setTree] = useState<DeptNode[]>([]);
  const [treeLoading, setTreeLoading] = useState(false);
  /** 左侧树选中的部门（筛选右侧列表为「该部门及其子孙」） */
  const [selectedDeptId, setSelectedDeptId] = useState<number | undefined>(undefined);
  const [listRefreshKey, setListRefreshKey] = useState(0);

  // 详情 / 编辑模态框
  const [detailRow, setDetailRow] = useState<DeptRow | null>(null);
  const [editOpen, setEditOpen] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form] = Form.useForm<DeptFormValues>();
  const [saving, setSaving] = useState(false);

  const [roleOptions, setRoleOptions] = useState<{ label: string; value: number }[]>([]);

  const canCreate = hasPerm('department:create');
  const canUpdate = hasPerm('department:update');
  const canDelete = hasPerm('department:delete');

  const loadTree = useCallback(async () => {
    setTreeLoading(true);
    try {
      const r = await departmentApi.getTree();
      setTree((r.data as DeptNode[]) || []);
    } finally {
      setTreeLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadTree();
    roleApi.getList({ page: 1, page_size: 100 }).then((r) => {
      const data = r.data as Record<string, unknown>;
      const items = (data.items as Record<string, unknown>[]) || [];
      setRoleOptions(items.map((i) => ({ label: i.name as string, value: i.id as number })));
    });
  }, [loadTree]);

  // id → 部门 扁平映射（树节点渲染/名称反查用）
  const deptById = useMemo(() => {
    const m = new Map<number, DeptNode>();
    const walk = (ns: DeptNode[]) => {
      for (const n of ns) {
        m.set(n.id, n);
        if (n.children) walk(n.children);
      }
    };
    walk(tree);
    return m;
  }, [tree]);

  /** 右侧列表请求：选中树节点时按 subtree_id 筛选 */
  const listRequest = useCallback(
    async (params: Record<string, unknown>) => {
      const r = await departmentApi.getList(params);
      const data = r.data as Record<string, unknown>;
      return {
        items: (data.items as DeptRow[]) || [],
        total: (data.total as number) || 0,
      };
    },
    [],
  );

  /** 子树筛选参数（tableParams 变化会自动重载） */
  const listTableParams = useMemo(
    () => (selectedDeptId != null ? { subtree_id: selectedDeptId } : {}),
    [selectedDeptId],
  );

  const openCreate = (parentId = 0) => {
    setEditingId(null);
    setDetailRow(null);
    form.resetFields();
    form.setFieldsValue({ parent_id: parentId || undefined, status: 1, sort: 0 });
    setEditOpen(true);
  };

  const openEdit = (row: DeptRow) => {
    setEditingId(row.id);
    setDetailRow(null);
    form.resetFields();
    form.setFieldsValue({
      parent_id: row.parent_id || undefined,
      name: row.name,
      sort: row.sort,
      status: row.status,
      remark: row.remark || '',
      role_ids: row.role_ids || [],
    });
    setEditOpen(true);
  };

  const handleSave = async () => {
    const values = await form.validateFields();
    const { role_ids, ...fields } = values;
    setSaving(true);
    try {
      if (editingId != null) {
        await departmentApi.update(editingId, fields as Record<string, unknown>);
        await departmentApi.assignRoles(editingId, role_ids || []);
        message.success('部门已更新');
      } else {
        const r = await departmentApi.create(fields as { parent_id: number; name: string });
        const created = r.data as { id: number };
        if (created?.id && role_ids?.length) {
          await departmentApi.assignRoles(created.id, role_ids);
        }
        message.success('部门已创建');
      }
      setEditOpen(false);
      await loadTree();
      setListRefreshKey((k) => k + 1);
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = (id: number) => {
    const node = deptById.get(id);
    modal.confirm({
      title: '删除部门',
      content: `确定删除部门「${node?.name ?? id}」吗？存在子部门时不可删除。`,
      onOk: async () => {
        await departmentApi.delete(id);
        message.success('已删除');
        setDetailRow(null);
        if (selectedDeptId === id) setSelectedDeptId(undefined);
        await loadTree();
        setListRefreshKey((k) => k + 1);
      },
    });
  };

  /** 树节点操作：新增子部门 / 编辑 / 删除 */
  const nodeActions = (data: { key: unknown }) => {
    const id = Number(data.key);
    const items: { key: string; label: React.ReactNode; danger?: boolean }[] = [];
    if (canCreate) items.push({ key: 'add-child', label: <><PlusOutlined /> 新增子部门</> });
    if (canUpdate) items.push({ key: 'edit', label: <><EditOutlined /> 编辑</> });
    if (canDelete) items.push({ key: 'delete', label: <><DeleteOutlined /> 删除</>, danger: true });
    return (
      <Dropdown
        menu={{
          items,
          onClick: ({ key }) => {
            if (key === 'add-child') openCreate(id);
            else if (key === 'edit') openEdit(deptById.get(id) as unknown as DeptRow);
            else if (key === 'delete') handleDelete(id);
          },
        }}
      >
        <Button type="text" size="small" icon={<DownOutlined />} style={{ width: 22, height: 22 }} />
      </Dropdown>
    );
  };

  const titleRender = (data: { key: unknown }) => {
    const id = Number(data.key);
    const node = deptById.get(id);
    return (
      <span style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'space-between', width: '100%', gap: 4 }}>
        <span style={{ flex: 1, minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          <FolderOutlined style={{ marginRight: 6, color: token.colorPrimary }} />
          {node?.name ?? String(data.key)}
        </span>
        {nodeActions(data)}
      </span>
    );
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '部门名称', dataIndex: 'name', width: 160 },
    {
      title: '上级部门',
      dataIndex: 'parent_id',
      width: 140,
      render: (v: unknown) => findName(tree, v as number),
    },
    {
      title: '角色',
      dataIndex: 'role_ids',
      render: (_: unknown, r: DeptRow) => {
        const list = r.roles || [];
        if (!list.length) return <span>-</span>;
        return (
          <Space size={[0, 4]} wrap>
            {list.map((role) => <Tag key={role.id} color="cyan">{role.name}</Tag>)}
          </Space>
        );
      },
    },
    { title: '排序', dataIndex: 'sort', width: 70 },
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      render: (v: unknown) => <Badge status={v === 1 ? 'success' : 'error'} text={v === 1 ? '启用' : '停用'} />,
    },
    { title: '备注', dataIndex: 'remark', ellipsis: true },
    {
      title: '操作',
      key: 'action',
      width: 180,
      render: (_: unknown, r: DeptRow) => [
        <a key="detail" onClick={() => setDetailRow(r)}><EyeOutlined /> 详情</a>,
        ...(canUpdate
          ? [<a key="edit" onClick={() => openEdit(r)} style={{ marginLeft: 12 }}><EditOutlined /> 编辑</a>]
          : []),
        ...(canDelete
          ? [
              <Popconfirm
                key="del"
                title="删除部门"
                description={`确定删除「${r.name}」吗？存在子部门时不可删除。`}
                onConfirm={() => handleDelete(r.id)}
              >
                <a style={{ color: 'red', marginLeft: 12 }}><DeleteOutlined /> 删除</a>
              </Popconfirm>,
            ]
          : []),
      ],
    },
  ];

  // 上级部门 TreeSelect：排除当前编辑的部门自身（防成环）
  const selectTreeData = useMemo(() => toSelectTree(tree, editingId ?? undefined), [tree, editingId]);

  return (
    <div style={{ display: 'flex', gap: 16, alignItems: 'flex-start' }}>
      {/* 左侧：部门树（导航 + 快捷操作） */}
      <div style={{ width: 240, flexShrink: 0, background: token.colorBgContainer, borderRadius: token.borderRadiusLG, padding: 12 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
          <span style={{ fontWeight: 600 }}><ApartmentOutlined /> 部门</span>
          {canCreate && (
            <Button size="small" type="primary" ghost icon={<PlusOutlined />} onClick={() => openCreate(0)}>
              新增
            </Button>
          )}
        </div>
        {tree.length === 0 && !treeLoading ? (
          <Empty description="暂无部门" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : (
          <Tree
            treeData={toTreeData(tree)}
            showLine
            blockNode
            defaultExpandAll
            titleRender={titleRender}
            selectedKeys={selectedDeptId != null ? [selectedDeptId] : []}
            onSelect={(keys) => {
              // 选中节点 → 筛选列表为其子树；再次点击清空
              setSelectedDeptId(keys.length ? Number(keys[0]) : undefined);
              setListRefreshKey((k) => k + 1);
            }}
          />
        )}
      </div>

      {/* 右侧：部门列表 */}
      <div style={{ flex: 1, minWidth: 0 }}>
        <DataTable<DeptRow>
          columns={columns}
          rowKey="id"
          request={listRequest}
          headerTitle={selectedDeptId != null ? `部门列表（${deptById.get(selectedDeptId)?.name ?? ''} 及其子孙）` : '部门列表'}
          tableParams={listTableParams}
          reloadKey={listRefreshKey}
          searchFields={[{ name: 'q', label: '部门名称', type: 'text' }]}
          toolBarRender={
            canCreate ? (
              <Button key="add" type="primary" icon={<PlusOutlined />} onClick={() => openCreate(0)}>
                新增部门
              </Button>
            ) : null
          }
        />
      </div>

      {/* 详情模态框（只读，提供「编辑」入口） */}
      <Modal
        title="部门详情"
        open={detailRow != null}
        onCancel={() => setDetailRow(null)}
        width={520}
        footer={
          <Space>
            {canUpdate && (
              <Button type="primary" icon={<EditOutlined />} onClick={() => detailRow && openEdit(detailRow)}>
                编辑
              </Button>
            )}
            {canDelete && detailRow && (
              <Button danger icon={<DeleteOutlined />} onClick={() => { const id = detailRow.id; setDetailRow(null); handleDelete(id); }}>
                删除
              </Button>
            )}
            <Button onClick={() => setDetailRow(null)}>关闭</Button>
          </Space>
        }
      >
        {detailRow ? (
          <Descriptions column={1} size="small" bordered>
            <Descriptions.Item label="部门名称">{detailRow.name}</Descriptions.Item>
            <Descriptions.Item label="上级部门">{findName(tree, detailRow.parent_id)}</Descriptions.Item>
            <Descriptions.Item label="排序">{detailRow.sort}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Badge status={detailRow.status === 1 ? 'success' : 'error'} text={detailRow.status === 1 ? '启用' : '停用'} />
            </Descriptions.Item>
            <Descriptions.Item label="部门角色">
              {detailRow.roles?.length
                ? detailRow.roles.map((r) => <Tag key={r.id} color="cyan" style={{ marginBottom: 4 }}>{r.name}</Tag>)
                : '—'}
            </Descriptions.Item>
            <Descriptions.Item label="备注">{detailRow.remark || '—'}</Descriptions.Item>
          </Descriptions>
        ) : null}
      </Modal>

      {/* 编辑/新增模态框 */}
      <Modal
        title={editingId != null ? '编辑部门' : '新增部门'}
        open={editOpen}
        onOk={() => void handleSave()}
        confirmLoading={saving}
        onCancel={() => setEditOpen(false)}
        destroyOnHidden
      >
        <Form form={form} layout="vertical">
          <Form.Item name="parent_id" label="上级部门">
            <TreeSelect treeData={selectTreeData} allowClear placeholder="不选则为顶级部门" treeDefaultExpandAll />
          </Form.Item>
          <Form.Item name="name" label="部门名称" rules={[{ required: true, message: '请输入部门名称' }]}>
            <Input maxLength={64} />
          </Form.Item>
          <Form.Item name="sort" label="排序">
            <Input type="number" />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select options={[
              { label: '启用', value: 1 },
              { label: '停用', value: 0 },
            ]} />
          </Form.Item>
          <Form.Item name="remark" label="备注">
            <Input.TextArea rows={2} maxLength={255} />
          </Form.Item>
          <Form.Item
            name="role_ids"
            label="部门角色"
            extra="部门角色会继承给该部门的成员用户，与其直接分配的角色取并集"
          >
            <Select mode="multiple" options={roleOptions} placeholder="可多选" allowClear />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default DeptManage;
