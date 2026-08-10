import React, { useCallback, useEffect, useState } from 'react';
import { Modal, Form, Input, InputNumber, Switch, Select, App, Tag, Popconfirm, Button, Empty, Checkbox, Dropdown } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, DownOutlined } from '@ant-design/icons';
import DataTable, { SearchField } from '../../components/DataTable';
import { useAuthStore } from '../../stores/auth';
import { dictTypeApi, dictEntryApi } from '../../api/dict';
import { useThemeToken } from '../../hooks/useThemeToken';
import { formatTime } from '../../utils/format';

interface DictType {
  id: number;
  code: string;
  name: string;
  is_public: boolean;
  status: number;
  remark: string;
  version: string;
}

interface DictEntry {
  id: number;
  type_id: number;
  label: string;
  value: string;
  sort: number;
  status: number;
  remark: string;
  version: string;
  created_at: string;
  updated_at: string;
}

const DictManage: React.FC = () => {
  const { message, modal } = App.useApp();
  const token = useThemeToken();
  const perms = useAuthStore((s) => s.permissions);

  // 字典类型状态
  const [types, setTypes] = useState<DictType[]>([]);
  const [selectedType, setSelectedType] = useState<DictType | null>(null);
  const [entryRefreshKey, setEntryRefreshKey] = useState(0);
  const [typeModal, setTypeModal] = useState<{ open: boolean; editing: DictType | null }>({
    open: false,
    editing: null,
  });
  const [typeForm] = Form.useForm<{
    code: string;
    name: string;
    is_public: boolean;
    status: number;
    remark: string;
    version: string;
  }>();

  // 类型批量勾选状态
  const [checkedTypeIds, setCheckedTypeIds] = useState<number[]>([]);

  // 条目状态
  const [entryModal, setEntryModal] = useState<{ open: boolean; editing: DictEntry | null }>({
    open: false,
    editing: null,
  });
  const [entryForm] = Form.useForm<{
    label: string;
    value: string;
    sort: number;
    status: number;
    remark: string;
    version: string;
  }>();

  const loadTypes = useCallback(async () => {
    try {
      const r = await dictTypeApi.list({ page: 1, page_size: 100 });
      const data = r.data as Record<string, unknown>;
      setTypes((data.items as DictType[]) || []);
    } catch {
      /* interceptor handles */
    }
  }, []);

  useEffect(() => {
    loadTypes();
  }, [loadTypes]);

  // ---- 字典类型 CRUD ----
  const openCreateType = () => {
    typeForm.resetFields();
    typeForm.setFieldsValue({ status: 1, is_public: false });
    setTypeModal({ open: true, editing: null });
  };

  const openEditType = (t: DictType) => {
    typeForm.setFieldsValue({
      code: t.code,
      name: t.name,
      is_public: !!t.is_public,
      status: t.status,
      remark: t.remark,
      version: t.version,
    });
    setTypeModal({ open: true, editing: t });
  };

  const handleTypeSubmit = async () => {
    const v = await typeForm.validateFields();
    if (typeModal.editing) {
      await dictTypeApi.update(typeModal.editing.id, {
        code: v.code,
        name: v.name,
        is_public: Boolean(v.is_public),
        status: v.status ?? 1,
        remark: v.remark ?? '',
        version: v.version ?? '',
      });
      message.success('字典类型已更新');
    } else {
      await dictTypeApi.create({
        code: v.code,
        name: v.name,
        is_public: Boolean(v.is_public),
        status: v.status ?? 1,
        remark: v.remark ?? '',
        version: v.version ?? '1.0.0',
      });
      message.success('字典类型已创建');
    }
    setTypeModal({ open: false, editing: null });
    loadTypes();
  };

  const handleTypeDelete = async (t: DictType) => {
    await dictTypeApi.delete(t.id);
    message.success('字典类型已删除');
    if (selectedType?.id === t.id) setSelectedType(null);
    loadTypes();
    setEntryRefreshKey((k) => k + 1);
  };

  // ---- 字典条目 CRUD ----
  const openCreateEntry = () => {
    entryForm.resetFields();
    entryForm.setFieldsValue({ sort: 0, status: 1 });
    setEntryModal({ open: true, editing: null });
  };

  const openEditEntry = (e: DictEntry) => {
    entryForm.setFieldsValue({
      label: e.label,
      value: e.value,
      sort: e.sort,
      status: e.status,
      remark: e.remark,
      version: e.version,
    });
    setEntryModal({ open: true, editing: e });
  };

  const handleEntrySubmit = async () => {
    if (!selectedType) return;
    const v = await entryForm.validateFields();
    if (entryModal.editing) {
      await dictEntryApi.update(selectedType.id, entryModal.editing.id, {
        label: v.label,
        value: v.value,
        sort: v.sort ?? 0,
        status: v.status ?? 1,
        remark: v.remark ?? '',
        version: v.version ?? '',
      });
      message.success('条目已更新');
    } else {
      await dictEntryApi.create(selectedType.id, {
        label: v.label,
        value: v.value,
        sort: v.sort ?? 0,
        status: v.status ?? 1,
        remark: v.remark ?? '',
        version: v.version ?? '1.0.0',
      });
      message.success('条目已创建');
    }
    setEntryModal({ open: false, editing: null });
    setEntryRefreshKey((k) => k + 1);
  };

  const handleEntryDelete = async (e: DictEntry) => {
    if (!selectedType) return;
    await dictEntryApi.delete(selectedType.id, e.id);
    message.success('条目已删除');
    setEntryRefreshKey((k) => k + 1);
  };

  const searchFields: SearchField[] = [{ name: 'q', label: '关键词', type: 'text' }];

  const entryColumns = [
    { title: '显示名', dataIndex: 'label', width: 120 },
    { title: '值', dataIndex: 'value', width: 120, render: (_: unknown, r: DictEntry) => <Tag color="blue">{r.value}</Tag> },
    { title: '排序', dataIndex: 'sort', width: 80 },
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      render: (_: unknown, r: DictEntry) =>
        r.status === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>,
    },
    { title: '备注', dataIndex: 'remark', ellipsis: true },
    { title: '版本', dataIndex: 'version', width: 90, render: (v: unknown) => (v ? <Tag>{v as string}</Tag> : '-') },
    { title: '更新时间', dataIndex: 'updated_at', width: 150, render: (v: unknown) => formatTime(v) },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, r: DictEntry) => [
        perms.includes('dict:update') ? (
          <a key="ed" onClick={() => openEditEntry(r)}>
            <EditOutlined /> 编辑
          </a>
        ) : null,
        perms.includes('dict:delete') ? (
          <Popconfirm key="del" title="删除条目" description={`确定删除「${r.label}」吗？`} onConfirm={() => handleEntryDelete(r)}>
            <a style={{ color: 'red' }}><DeleteOutlined /> 删除</a>
          </Popconfirm>
        ) : null,
      ],
    },
  ];

  const entryRequest = useCallback(
    async (params: Record<string, unknown>) => {
      if (!selectedType) return { items: [] as DictEntry[], total: 0 };
      const r = await dictEntryApi.listByTypeId(selectedType.id, params);
      const data = r.data as Record<string, unknown>;
      return {
        items: (data.items as DictEntry[]) || [],
        total: (data.total as number) || 0,
      };
    },
    [selectedType],
  );

  return (
    <div style={{ display: 'flex', gap: 16 }}>
      {/* 左侧：字典类型列表 */}
      <div style={{ width: 240, flexShrink: 0, background: token.colorBgContainer, borderRadius: token.borderRadiusLG, padding: 12 }}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 8,
          }}
        >
          <span style={{ fontWeight: 600 }}>字典类型</span>
          {perms.includes('dict:create') && (
            <Button size="small" type="primary" ghost icon={<PlusOutlined />} onClick={openCreateType}>
              新增
            </Button>
          )}
        </div>
        {checkedTypeIds.length > 0 && (perms.includes('dict:update') || perms.includes('dict:delete')) ? (
          <div style={{ display: 'flex', gap: 4, marginBottom: 8, alignItems: 'center' }}>
            <span style={{ fontSize: 12, color: token.colorTextTertiary }}>已选 {checkedTypeIds.length} 项</span>
            <Dropdown
              menu={{
                items: [
                  ...(perms.includes('dict:update') ? [{ key: 'enable', label: '批量启用' }] : []),
                  ...(perms.includes('dict:update') ? [{ key: 'disable', label: '批量禁用' }] : []),
                  ...(perms.includes('dict:delete') ? [{ key: 'delete', label: '批量删除', danger: true }] : []),
                ],
                onClick: ({ key }) => {
                  if (key === 'enable') void dictTypeApi.batchUpdateStatus(checkedTypeIds, 1).then(() => {
                    message.success('已批量启用');
                    setCheckedTypeIds([]);
                    loadTypes();
                  });
                  else if (key === 'disable') void dictTypeApi.batchUpdateStatus(checkedTypeIds, 0).then(() => {
                    message.success('已批量禁用');
                    setCheckedTypeIds([]);
                    loadTypes();
                  });
                  else if (key === 'delete') {
                    modal.confirm({
                      title: '批量删除',
                      content: `确定删除选中的 ${checkedTypeIds.length} 个类型吗？（含条目的类型不可删除）`,
                      onOk: async () => {
                        await dictTypeApi.batchDelete(checkedTypeIds);
                        message.success('已删除');
                        if (selectedType && checkedTypeIds.includes(selectedType.id)) setSelectedType(null);
                        setCheckedTypeIds([]);
                        loadTypes();
                        setEntryRefreshKey((k) => k + 1);
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
          </div>
        ) : null}
        {types.length === 0 ? (
          <Empty description="暂无字典类型" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : (
          <div>
            {types.map((t) => (
              <div
                key={t.id}
                onClick={() => {
                  setSelectedType(t);
                  setEntryRefreshKey((k) => k + 1);
                }}
                style={{
                  cursor: 'pointer',
                  padding: '6px 8px',
                  borderRadius: 6,
                  marginBottom: 4,
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  background: selectedType?.id === t.id ? token.colorPrimaryBg : 'transparent',
                  border: selectedType?.id === t.id ? `1px solid ${token.colorPrimaryBorder}` : '1px solid transparent',
                }}
              >
                <Checkbox
                  checked={checkedTypeIds.includes(t.id)}
                  onChange={(e) => {
                    const checked = e.target.checked;
                    setCheckedTypeIds((prev) => (checked ? [...prev, t.id] : prev.filter((id) => id !== t.id)));
                  }}
                  onClick={(e) => e.stopPropagation()}
                  style={{ marginRight: 2 }}
                />
                <div>
                  <div style={{ fontWeight: 500 }}>{t.name}</div>
                  <div style={{ fontSize: 12, color: token.colorTextTertiary }}>
                    <Tag style={{ fontSize: 11 }}>{t.code}</Tag>
                    {t.is_public ? <Tag color="green" style={{ fontSize: 11 }}>公开</Tag> : null}
                    {t.status === 0 ? <Tag color="red" style={{ fontSize: 11 }}>禁用</Tag> : null}
                  </div>
                </div>
                {(perms.includes('dict:update') || perms.includes('dict:delete')) && (
                  <span style={{ display: 'inline-flex', gap: 8 }}>
                    {perms.includes('dict:update') ? (
                      <EditOutlined onClick={(e) => { e.stopPropagation(); openEditType(t); }} />
                    ) : null}
                    {perms.includes('dict:delete') ? (
                      <Popconfirm key="del" title="删除类型？请先确保类型下无条目" onConfirm={() => handleTypeDelete(t)}>
                        <DeleteOutlined style={{ color: 'red' }} onClick={(e) => e.stopPropagation()} />
                      </Popconfirm>
                    ) : null}
                  </span>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 右侧：字典条目列表 */}
      <div style={{ flex: 1, minWidth: 0 }}>
        <DataTable<DictEntry>
          columns={entryColumns}
          rowKey="id"
          request={entryRequest}
          searchFields={searchFields}
          reloadKey={entryRefreshKey}
          headerTitle={
            selectedType
              ? `${selectedType.name}（${selectedType.code}）— 字典条目`
              : '请选择左侧字典类型'
          }
          toolBarRender={
            selectedType && perms.includes('dict:create') ? (
              <Button key="add" type="primary" icon={<PlusOutlined />} onClick={openCreateEntry}>
                新增条目
              </Button>
            ) : null
          }
          selectable={!!selectedType && (perms.includes('dict:update') || perms.includes('dict:delete'))}
          batchBarRender={(keys, clear) => {
            const ids = keys as number[];
            const typeId = selectedType?.id;
            const runStatus = async (status: number, label: string) => {
              if (!typeId) return;
              await dictEntryApi.batchUpdateStatus(typeId, ids, status);
              message.success(`已${label}`);
              clear();
              setEntryRefreshKey((k) => k + 1);
            };
            return (
              <Dropdown
                menu={{
                  items: [
                    ...(perms.includes('dict:update') ? [{ key: 'enable', label: '批量启用' }] : []),
                    ...(perms.includes('dict:update') ? [{ key: 'disable', label: '批量禁用' }] : []),
                    ...(perms.includes('dict:delete') ? [{ key: 'delete', label: '批量删除', danger: true }] : []),
                  ],
                  onClick: ({ key }) => {
                    if (key === 'enable') void runStatus(1, '批量启用');
                    else if (key === 'disable') void runStatus(0, '批量禁用');
                    else if (key === 'delete') {
                      modal.confirm({
                        title: '批量删除',
                        content: `确定删除选中的 ${ids.length} 个条目吗？`,
                        onOk: async () => {
                          if (!typeId) return;
                          await dictEntryApi.batchDelete(typeId, ids);
                          message.success('已删除');
                          clear();
                          setEntryRefreshKey((k) => k + 1);
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
        />
      </div>

      {/* 类型新增/编辑弹窗 */}
      <Modal
        title={typeModal.editing ? `编辑字典类型 — ${typeModal.editing.name}` : '新增字典类型'}
        open={typeModal.open}
        onOk={handleTypeSubmit}
        onCancel={() => setTypeModal({ open: false, editing: null })}
      >
        <Form form={typeForm} layout="vertical">
          <Form.Item name="code" label="编码" rules={[{ required: true, message: '请输入编码' }]}>
            <Input placeholder="如 gender" disabled={!!typeModal.editing} />
          </Form.Item>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="如 性别" />
          </Form.Item>
          <Form.Item name="is_public" label="是否公开" valuePropName="checked">
            <Switch checkedChildren="公开" unCheckedChildren="私密" />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select
              options={[
                { label: '启用', value: 1 },
                { label: '禁用', value: 0 },
              ]}
            />
          </Form.Item>
          <Form.Item name="remark" label="备注">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="version" label="版本（表示该字典类型由哪个版本新增）">
            <Input placeholder="如 1.0.0" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 条目新增/编辑弹窗 */}
      <Modal
        title={entryModal.editing ? '编辑条目' : '新增条目'}
        open={entryModal.open}
        onOk={handleEntrySubmit}
        onCancel={() => setEntryModal({ open: false, editing: null })}
      >
        <Form form={entryForm} layout="vertical">
          <Form.Item name="label" label="显示名" rules={[{ required: true, message: '请输入显示名' }]}>
            <Input placeholder="如 男" />
          </Form.Item>
          <Form.Item name="value" label="值" rules={[{ required: true, message: '请输入值' }]}>
            <Input placeholder="如 male" />
          </Form.Item>
          <Form.Item name="sort" label="排序">
            <InputNumber style={{ width: '100%' }} placeholder="数字越小越靠前" />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select
              options={[
                { label: '启用', value: 1 },
                { label: '禁用', value: 0 },
              ]}
            />
          </Form.Item>
          <Form.Item name="remark" label="备注">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="version" label="版本（表示该字典条目由哪个版本新增）">
            <Input placeholder="如 1.0.0" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default DictManage;
