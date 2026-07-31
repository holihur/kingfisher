# Frontend Config — 系统配置管理

## 职责

系统配置的列表展示和编辑。对接后端 `extends/config` 模块。

## 页面设计

```
┌──────────────────────────────────────────────┐
│  系统配置                                     │
│                                              │
│  ┌──────────────────────────────────────────┐│
│  │ Key              │ Value     │ 备注 │ 操作││
│  │──────────────────┼──────────┼─────┼─────││
│  │ site_name        │ Kingfisher│ 站点 │ 编辑││
│  │ site_logo        │ /logo.png │ Logo│ 编辑││
│  │ max_login_attempt│ 5         │ 登录│ 编辑││
│  │ lockout_duration │ 15m       │ 锁定│ 编辑││
│  └──────────────────────────────────────────┘│
└──────────────────────────────────────────────┘
```

## 编辑弹窗（按 Value 类型区分控件）

```
┌──────────────────────────────┐
│  编辑 — site_name            │
│                              │
│  Key:  site_name (不可改)     │
│                              │
│  Value:                      │
│  ┌────────────────────────┐  │
│  │ Kingfisher Admin        │  │  ← Input（纯文本）
│  └────────────────────────┘  │
│                              │
│  备注:                       │
│  ┌────────────────────────┐  │
│  │ 站点名称，显示在页头和登录页│  │
│  └────────────────────────┘  │
│                  [取消] [确定]│
└──────────────────────────────┘
```

## 实现

```tsx
const ConfigManage: React.FC = () => {
    const [editModal, setEditModal] = useState<{ open: boolean; config: SystemConfig | null }>({
        open: false, config: null
    });

    const columns: ProColumns<SystemConfig>[] = [
        {
            title: 'Key', dataIndex: 'key', width: 200,
            render: (text) => <Tag color="blue">{text}</Tag>,
        },
        {
            title: 'Value', dataIndex: 'value',
            render: (text, record) => {
                // 根据 key 类型渲染不同展示
                if (record.key === 'site_logo') return <Avatar src={text} />;
                if (record.key.includes('duration')) return <Tag>{text}</Tag>;
                return <span style={{ wordBreak: 'break-all' }}>{text}</span>;
            },
        },
        { title: '备注', dataIndex: 'remark', ellipsis: true },
        { title: '更新时间', dataIndex: 'updated_at', valueType: 'dateTime', width: 180 },
        {
            title: '操作', valueType: 'option',
            render: (_, record) => [
                <PermissionBtn key="edit" code="config:update">
                    <a onClick={() => setEditModal({ open: true, config: record })}>编辑</a>
                </PermissionBtn>,
            ],
        },
    ];

    return (
        <PageContainer>
            <ProTable<SystemConfig>
                headerTitle="系统配置"
                columns={columns}
                request={async () => {
                    const resp = await configApi.getAll();
                    // 后端返回 map，转成数组
                    const items = Object.entries(resp.data).map(([key, value]) => ({
                        key, value,
                        remark: config.remark,  // 后端返回的完整 SystemConfig 含 remark 字段
                    }));
                    return { data: items, success: true };
                }}
                rowKey="key"
                search={false}  // 配置项少，不分页不搜索，无需 URL 同步
            />
            <ConfigEditModal {...editModal}
                onClose={() => setEditModal({ open: false, config: null })}
            />
        </PageContainer>
    );
};
```

### ConfigEditModal

```tsx
const ConfigEditModal: React.FC<Props> = ({ open, config, onClose }) => {
    const [form] = Form.useForm();
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        if (open && config) form.setFieldsValue(config);
    }, [open, config]);

    const handleSubmit = async () => {
        const values = await form.validateFields();
        setLoading(true);
        await configApi.set(config!.key, values.value);
        message.success('配置已更新');
        setLoading(false);
        onClose();
    };

    // 根据 key 渲染不同输入控件
    const renderValueInput = () => {
        if (config?.key === 'site_logo') {
            return (
                <Form.Item name="value" label="Value" rules={[{ required: true }]}>
                    <Input placeholder="Logo URL" />
                </Form.Item>
            );
        }
        if (config?.key.includes('duration')) {
            return (
                <Form.Item name="value" label="Value" rules={[{ required: true }]}>
                    <Select options={[
                        { label: '5 分钟', value: '5m' },
                        { label: '15 分钟', value: '15m' },
                        { label: '30 分钟', value: '30m' },
                        { label: '1 小时', value: '1h' },
                    ]} />
                </Form.Item>
            );
        }
        return (
            <Form.Item name="value" label="Value" rules={[{ required: true }]}>
                <Input />
            </Form.Item>
        );
    };

    return (
        <Modal title={`编辑配置 — ${config?.key}`} open={open}
               onOk={handleSubmit} onCancel={onClose} confirmLoading={loading}>
            <Form form={form} layout="vertical">
                <Form.Item label="Key">
                    <Input value={config?.key} disabled />
                </Form.Item>
                {renderValueInput()}
                <Form.Item name="remark" label="备注">
                    <Input.TextArea rows={2} />
                </Form.Item>
            </Form>
        </Modal>
    );
};
```

## API

```tsx
// api/config.ts
export const configApi = {
    getAll: () => request.get<never, ApiResponse<Record<string, string>>>('/configs'),
    get: (key: string) => request.get<never, ApiResponse<string>>(`/configs/${key}`),
    set: (key: string, value: string) =>
        request.put<never, ApiResponse<null>>(`/configs/${key}`, { value }),
    delete: (key: string) => request.delete<never, ApiResponse<null>>(`/configs/${key}`),
};
```

## 设计要点

- 配置是键值对，后端返回 `map[string]string`
- 根据 key 名称判断渲染哪种控件（文本/下拉/数字输入）
- 预设配置项不可删除（删除按钮需确认），自定义配置项可删除
- 更新配置后后端自动失效缓存，前端无需额外处理
