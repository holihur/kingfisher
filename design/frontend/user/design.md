# Frontend User — 用户管理

## 职责

用户列表（搜索+分页）、新增用户、编辑用户、删除用户。对接后端 `extends/user` 模块。

## 页面

### UserList — 列表页

```
┌──────────────────────────────────────────────┐
│  用户管理                                     │
│  ┌──────────────┐ ┌──────┐ ┌──────────────┐ │
│  │ 搜索用户名...  │ │ 搜索  │ │ 重置         │ │  ← 搜索栏
│  └──────────────┘ └──────┘ └──────────────┘ │
│                                  [ 新增用户 ] │  ← 操作按钮（PermissionBtn）
│                                              │
│  ┌──────────────────────────────────────────┐│
│  │ ID │ 用户名 │ 邮箱 │ 角色 │ 状态 │ 操作   ││
│  │────┼───────┼─────┼─────┼─────┼─────────││
│  │ 1  │ admin │ a@x │ 管理员│启用 │编辑 删除││  ← ProTable
│  │ 2  │ editor│ e@x │ 编辑  │启用 │编辑 删除││
│  │ 3  │ viewer│ v@x │ 访客  │禁用 │编辑 删除││
│  └──────────────────────────────────────────┘│
│                    第 1/10 页  共 98 条        │
└──────────────────────────────────────────────┘
```

### 实现

```tsx
const UserList: React.FC = () => {
    const actionRef = useRef<ActionType>();

    const columns: ProColumns<User>[] = [
        { title: 'ID', dataIndex: 'id', width: 80 },
        { title: '用户名', dataIndex: 'username' },
        { title: '邮箱', dataIndex: 'email', ellipsis: true },
        {
            title: '角色', dataIndex: 'role',
            valueEnum: { admin: '管理员', editor: '编辑', viewer: '访客' },
        },
        {
            title: '状态', dataIndex: 'status',
            render: (_, r) => <Badge status={r.status === 1 ? 'success' : 'error'}
                                     text={r.status === 1 ? '启用' : '禁用'} />,
        },
        {
            title: '创建时间', dataIndex: 'created_at', valueType: 'dateTime',
        },
        {
            title: '操作', valueType: 'option',
            render: (_, record) => [
                <PermissionBtn key="edit" code="user:update">
                    <a onClick={() => handleEdit(record)}>编辑</a>
                </PermissionBtn>,
                <PermissionBtn key="del" code="user:delete">
                    <Popconfirm title="确认删除？" onConfirm={() => handleDelete(record.id)}>
                        <a style={{ color: 'red' }}>删除</a>
                    </Popconfirm>
                </PermissionBtn>,
            ],
        },
    ];

    const request = async (params: { current?: number; pageSize?: number; keyword?: string }) => {
        const resp = await userApi.getList({
            page: params.current || 1,
            pageSize: params.pageSize || 20,
            keyword: params.keyword,
        });
        return { data: resp.data.items, total: resp.data.total, success: true };
    };

    return (
        <ProTable<User>
            headerTitle="用户管理"
            actionRef={actionRef}
            columns={columns}
            request={request}
            rowKey="id"
            search={{ labelWidth: 'auto' }}
            pagination={{ pageSize: 20, showSizeChanger: true }}
            toolbar={{
                actions: [
                    <PermissionBtn key="add" code="user:create">
                        <Button type="primary" icon={<PlusOutlined />}
                                onClick={() => setModalOpen(true)}>新增用户</Button>
                    </PermissionBtn>,
                ],
            }}
        />
    );
};
```

### UserForm — 新增/编辑弹窗

```tsx
const UserForm: React.FC<Props> = ({ open, onClose, onSubmit, initialValues }) => {
    const [form] = Form.useForm();
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        if (open && initialValues) form.setFieldsValue(initialValues);
        else if (open) form.resetFields();
    }, [open, initialValues]);

    const handleSubmit = async () => {
        const values = await form.validateFields();
        setLoading(true);
        try {
            if (initialValues?.id) {
                await userApi.update(initialValues.id, values);
                message.success('更新成功');
            } else {
                await userApi.create(values);
                message.success('创建成功');
            }
            onSubmit();
            onClose();
        } catch {
            // 错误已在拦截器中处理
        } finally {
            setLoading(false);
        }
    };

    return (
        <Modal title={initialValues ? '编辑用户' : '新增用户'}
               open={open} onOk={handleSubmit} onCancel={onClose}
               confirmLoading={loading} destroyOnClose>
            <Form form={form} layout="vertical" preserve={false}>
                <Form.Item name="username" label="用户名"
                           rules={[{ required: true }, { min: 3 }, { max: 32 }]}>
                    <Input disabled={!!initialValues} />
                </Form.Item>
                {!initialValues && (
                    <Form.Item name="password" label="密码"
                               rules={[{ required: true }, { min: 8 }]}>
                        <Input.Password />
                    </Form.Item>
                )}
                <Form.Item name="email" label="邮箱"
                           rules={[{ type: 'email', message: '请输入有效邮箱' }]}>
                    <Input />
                </Form.Item>
                <Form.Item name="role" label="角色" rules={[{ required: true }]}>
                    <Select options={[
                        { label: '管理员', value: 'admin' },
                        { label: '编辑', value: 'editor' },
                        { label: '访客', value: 'viewer' },
                    ]} />
                </Form.Item>
            </Form>
        </Modal>
    );
};
```

## 删除确认

```tsx
const handleDelete = async (id: number) => {
    await userApi.delete(id);
    message.success('删除成功');
    actionRef.current?.reload();
};
```

## 设计要点

- ProTable 自带搜索、分页、loading、空状态，不用手写
- 表单用 AntD Form + Modal，`destroyOnClose` 确保关闭后数据重置
- 编辑时 `username` 禁用（不可修改），新增时需要填密码
- 按钮通过 `<PermissionBtn code="...">` 控制显隐（对接 `stores/auth.permissions`）
