# Frontend RBAC — 角色权限管理

## 职责

角色列表 CRUD + 权限分配 + 菜单分配。对接后端 `extends/rbac` 模块。

## 页面设计

### RoleList

```
┌──────────────────────────────────────────────┐
│  角色管理                        [ 新增角色 ]  │
│                                              │
│  ┌──────────────────────────────────────────┐│
│  │ ID │ 角色名 │ 编码   │ 描述   │ 操作     ││
│  │────┼───────┼───────┼───────┼───────────││
│  │ 1  │ 管理员 │ admin │ 全部权限│权限 菜单 ││  ← 两个操作
│  │ 2  │ 编辑   │ editor│ 内容管理│权限 菜单 ││
│  │ 3  │ 访客   │ viewer│ 只读  │ 权限 菜单 ││
│  └──────────────────────────────────────────┘│
└──────────────────────────────────────────────┘
```

### 权限分配弹窗

```
┌──────────────────────────────────────┐
│  分配权限 — 管理员                    │
│                                      │
│  ☑ 全部选中                           │
│  ┌──────┬──────┬──────┬──────┐      │
│  │ 用户  │ 菜单  │ 角色  │ 配置  │      │  ← Tabs 按 Resource 分组
│  ├──────┼──────┼──────┼──────┤      │
│  │☑ 查看│☑ 查看│☑ 查看│☑ 查看│      │  ← Checkbox 按 Action 列出
│  │☑ 创建│☑ 创建│☑ 创建│☑ 更新│      │
│  │☑ 更新│☑ 更新│☑ 更新│      │      │
│  │☑ 删除│☑ 删除│☑ 删除│      │      │
│  └──────┴──────┴──────┴──────┘      │
│                        [取消] [确定] │
└──────────────────────────────────────┘
```

## 实现

### RoleList

```tsx
const RoleList: React.FC = () => {
    const [permModal, setPermModal] = useState<{ open: boolean; role: Role | null }>({ open: false, role: null });
    const [menuModal, setMenuModal] = useState<{ open: boolean; role: Role | null }>({ open: false, role: null });

    const columns: ProColumns<Role>[] = [
        { title: 'ID', dataIndex: 'id', width: 80 },
        { title: '角色名', dataIndex: 'name' },
        { title: '编码', dataIndex: 'code', copyable: true },
        { title: '描述', dataIndex: 'description', ellipsis: true },
        { title: '创建时间', dataIndex: 'created_at', valueType: 'dateTime' },
        {
            title: '操作', valueType: 'option',
            render: (_, record) => [
                <PermissionBtn key="perm" code="role:update">
                    <a onClick={() => setPermModal({ open: true, role: record })}>权限</a>
                </PermissionBtn>,
                <PermissionBtn key="menu" code="role:update">
                    <a onClick={() => setMenuModal({ open: true, role: record })}>菜单</a>
                </PermissionBtn>,
                <PermissionBtn key="edit" code="role:update">
                    <a onClick={() => handleEdit(record)}>编辑</a>
                </PermissionBtn>,
                <PermissionBtn key="del" code="role:delete">
                    <Popconfirm title="确认删除？" onConfirm={() => handleDelete(record.id)}>
                        <a style={{ color: 'red' }}>删除</a>
                    </Popconfirm>
                </PermissionBtn>,
            ],
        },
    ];

    return (
        <>
            <ProTable<Role> columns={columns} request={loadRoles} rowKey="id" />
            <PermissionModal {...permModal} onClose={() => setPermModal({ open: false, role: null })} />
            <MenuAssignModal {...menuModal} onClose={() => setMenuModal({ open: false, role: null })} />
        </>
    );
};
```

### PermissionModal

```tsx
const PermissionModal: React.FC<Props> = ({ open, role, onClose }) => {
    const [allPerms, setAllPerms] = useState<Permission[]>([]);
    const [selected, setSelected] = useState<number[]>([]);

    useEffect(() => {
        if (open && role) {
            // 加载全部权限
            roleApi.getAllPermissions().then(r => setAllPerms(r.data));
            // 加载角色已有权限
            roleApi.getRolePermissions(role.id).then(r => setSelected(r.data.map((p: unknown) => p.id)));
        }
    }, [open, role]);

    const handleSubmit = async () => {
        await roleApi.assignPermissions(role!.id, selected);
        message.success('权限分配成功');
        onClose();
    };

    // 按 resource 分组
    const grouped = groupBy(allPerms, 'resource');

    return (
        <Modal title={`分配权限 — ${role?.name}`} open={open} onOk={handleSubmit} onCancel={onClose} width={700}>
            <Tabs items={Object.entries(grouped).map(([resource, perms]) => ({
                key: resource,
                label: resourceLabel[resource] || resource,
                children: (
                    <Checkbox.Group value={selected} onChange={v => setSelected(v as number[])}>
                        <Row gutter={[16, 8]}>
                            {perms.map(p => (
                                <Col span={12} key={p.id}>
                                    <Checkbox value={p.id}>{p.name}</Checkbox>
                                </Col>
                            ))}
                        </Row>
                    </Checkbox.Group>
                ),
            }))} />
        </Modal>
    );
};
```

### MenuAssignModal

```tsx
const MenuAssignModal: React.FC<Props> = ({ open, role, onClose }) => {
    const [allMenus, setAllMenus] = useState<MenuItem[]>([]);
    const [checkedKeys, setCheckedKeys] = useState<number[]>([]);

    useEffect(() => {
        if (open && role) {
            menuApi.getTree().then(r => setAllMenus(r.data));
            roleApi.getRoleMenus(role.id).then(r => setCheckedKeys(r.data.map((m: unknown) => m.id)));
        }
    }, [open, role]);

    const handleSubmit = async () => {
        await roleApi.assignMenus(role!.id, checkedKeys);
        message.success('菜单分配成功');
        onClose();
    };

    return (
        <Modal title={`分配菜单 — ${role?.name}`} open={open} onOk={handleSubmit} onCancel={onClose} width={500}>
            <Tree
                checkable
                defaultExpandAll
                checkedKeys={checkedKeys}
                onCheck={(keys) => setCheckedKeys(keys as number[])}
                treeData={convertToTreeData(allMenus)}
            />
        </Modal>
    );
};
```

## 设计要点

- 权限列表按 Resource（user/menu/role/config）分组，用 Tabs 切换
- 菜单分配用 Tree 组件，支持 `checkable` 选中/取消
- 保存时全量提交选中的 ID 列表（后端 `roleRepo.AssignPermissions` 先删后插）
- `resourceLabel` 中文字段映射：`{user: '用户', menu: '菜单', role: '角色', config: '配置'}`
