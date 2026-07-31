# Frontend Menu — 菜单管理

## 职责

菜单的树形展示与管理。支持目录/菜单/按钮三种类型，拖拽排序（可选），增删改。

## 页面设计

```
┌──────────────────────────────────────────────┐
│  菜单管理                        [ 新增根菜单 ] │
│                                              │
│  ┌────────────────────┐ ┌──────────────────┐ │
│  │ 📁 系统管理         │ │ 菜单名称：系统管理│ │  ← 左：树，右：详情/表单
│  │   📄 用户管理       │ │ 路由路径：/system │ │
│  │     🔘 新增用户     │ │ 图标：Setting    │ │
│  │     🔘 编辑用户     │ │ 类型：目录        │ │
│  │     🔘 删除用户     │ │ 排序：1          │ │
│  │   📄 菜单管理       │ │ 权限标识：-      │ │
│  │   📄 角色管理       │ │                  │ │
│  │   📄 系统配置       │ │ [编辑] [添加子项] │ │
│  │ 📁 其他             │ │                  │ │
│  │   📄 Dashboard      │ │                  │ │
│  └────────────────────┘ └──────────────────┘ │
└──────────────────────────────────────────────┘
```

## 实现

```tsx
const MenuManage: React.FC = () => {
    const [tree, setTree] = useState<MenuItem[]>([]);
    const [selected, setSelected] = useState<MenuItem | null>(null);
    const [modalOpen, setModalOpen] = useState(false);
    const [editing, setEditing] = useState<MenuItem | null>(null);

    const fetchTree = async () => {
        const resp = await menuApi.getTree();
        setTree(resp.data);
    };

    useEffect(() => { fetchTree(); }, []);

    // 将树形数据转成 AntD Table（展开显示层级）
    const flattenForTable = (items: MenuItem[], level = 0): unknown[] => {
        return items.flatMap(item => [
            { ...item, level },
            ...(item.children ? flattenForTable(item.children, level + 1) : [])
        ]);
    };

    const typeMap = { 1: '目录', 2: '菜单', 3: '按钮' };
    const typeColors = { 1: 'blue', 2: 'green', 3: 'orange' };

    const columns = [
        {
            title: '名称', dataIndex: 'name',
            render: (text: string, record: unknown) => (
                <span style={{ paddingLeft: record.level * 24 }}>
                    <Tag color={typeColors[record.type as keyof typeof typeColors]}>
                        {typeMap[record.type as keyof typeof typeMap]}
                    </Tag>
                    {text}
                </span>
            ),
        },
        { title: '路由', dataIndex: 'path' },
        { title: '权限标识', dataIndex: 'permission' },
        { title: '排序', dataIndex: 'sort', width: 80 },
        {
            title: '操作', valueType: 'option',
            render: (_: unknown, record: MenuItem) => [
                <PermissionBtn key="edit" code="menu:update">
                    <a onClick={() => { setEditing(record); setModalOpen(true); }}>编辑</a>
                </PermissionBtn>,
                record.type !== 3 && (
                    <PermissionBtn key="add" code="menu:create">
                        <a onClick={() => { setEditing({ parent_id: record.id } as Partial<MenuItem>); setModalOpen(true); }}>
                            添加子项
                        </a>
                    </PermissionBtn>
                ),
                <PermissionBtn key="del" code="menu:delete">
                    <Popconfirm title="有子节点将无法删除" onConfirm={() => handleDelete(record.id)}>
                        <a style={{ color: 'red' }}>删除</a>
                    </Popconfirm>
                </PermissionBtn>,
            ],
        },
    ];

    return (
        <PageContainer>
            <ProTable
                headerTitle="菜单管理"
                columns={columns}
                dataSource={flattenForTable(tree)}
                rowKey="id"
                search={false, syncToUrl: false}  // 菜单不分页不搜索，URL 仅 expanded 参数同步
                pagination={false}           // 菜单通常不多，不分页
                defaultExpandAllRows
                toolBarRender={() => [
                    <PermissionBtn key="add" code="menu:create">
                        <Button type="primary" icon={<PlusOutlined />}
                                onClick={() => { setEditing(null); setModalOpen(true); }}>
                            新增根菜单
                        </Button>
                    </PermissionBtn>,
                ]}
            />
            <MenuForm
                open={modalOpen}
                initialValues={editing}
                tree={tree}
                onClose={() => { setModalOpen(false); setEditing(null); }}
                onSubmit={fetchTree}
            />
        </PageContainer>
    );
};
```

### MenuForm

```tsx
const MenuForm: React.FC<Props> = ({ open, onClose, onSubmit, initialValues, tree }) => {
    const [form] = Form.useForm();

    return (
        <Modal title={initialValues?.id ? '编辑菜单' : '新增菜单'}
               open={open} onOk={handleSubmit} onCancel={onClose} destroyOnClose>
            <Form form={form} layout="vertical" initialValues={initialValues}>
                <Form.Item name="parent_id" label="上级菜单">
                    <TreeSelect
                        treeData={buildTreeSelectData(tree)}
                        placeholder="不选则为根菜单"
                        allowClear
                        treeDefaultExpandAll
                    />
                </Form.Item>
                <Form.Item name="type" label="类型" rules={[{ required: true }]}>
                    <Select options={[
                        { label: '目录', value: 1 },
                        { label: '菜单', value: 2 },
                        { label: '按钮', value: 3 },
                    ]} />
                </Form.Item>
                <Form.Item name="name" label="名称" rules={[{ required: true }]}>
                    <Input />
                </Form.Item>
                <Form.Item name="path" label="路由路径">
                    <Input placeholder="type=菜单时必填" />
                </Form.Item>
                <Form.Item name="icon" label="图标">
                    <IconSelect />
                </Form.Item>
                <Form.Item name="permission" label="权限标识">
                    <Input placeholder="如 user:create" />
                </Form.Item>
                <Form.Item name="sort" label="排序">
                    <InputNumber min={0} />
                </Form.Item>
            </Form>
        </Modal>
    );
};
```

## 设计要点

- 用展平的 Table 替代 Tree 组件，方便批量操作
- 按 `type` 区分：目录(折叠)、菜单(路由)、按钮(权限点)
- `TreeSelect` 选择父级菜单，支持无限层级
- 删除时后端校验有子节点则拒绝（`ErrMenuHasChildren`），前端也可预防性禁用删除按钮
