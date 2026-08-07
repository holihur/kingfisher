import React, { useCallback, useEffect, useState } from 'react';
import { Modal, Form, Input, InputNumber, Switch, Select, message, Tag, Popconfirm, List, Button, Empty } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import ProTable, { ProColumns } from '@ant-design/pro-table';
import { useAuthStore } from '../../stores/auth';
import { dictTypeApi, dictEntryApi } from '../../api/dict';
import { useTableUrlQuery } from '../../hooks/useTableUrlQuery';
import { buildQueryParams } from '../../utils/query';

interface DictType {
  id: number;
  code: string;
  name: string;
  is_public: boolean;
  status: number;
  remark: string;
}

interface DictEntry {
  id: number;
  type_id: number;
  label: string;
  value: string;
  sort: number;
  status: number;
  remark: string;
}

const DictManage: React.FC = () => {
  const { urlParams, page, pageSize, actionRef, formRef, syncFormFromUrl, onSearch, onReset, onPageChange } = useTableUrlQuery();
  const perms = useAuthStore((s) => s.permissions);

  // 字典类型状态
  const [types, setTypes] = useState<DictType[]>([]);
  const [selectedType, setSelectedType] = useState<DictType | null>(null);
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
  }>();

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
    syncFormFromUrl();
    // eslint-disable-next-line react-hooks/exhaustive-deps
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
      });
      message.success('字典类型已更新');
    } else {
      await dictTypeApi.create({
        code: v.code,
        name: v.name,
        is_public: Boolean(v.is_public),
        status: v.status ?? 1,
        remark: v.remark ?? '',
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
    actionRef.current?.reload();
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
      });
      message.success('条目已更新');
    } else {
      await dictEntryApi.create(selectedType.id, {
        label: v.label,
        value: v.value,
        sort: v.sort ?? 0,
        status: v.status ?? 1,
        remark: v.remark ?? '',
      });
      message.success('条目已创建');
    }
    setEntryModal({ open: false, editing: null });
    actionRef.current?.reload();
  };

  const handleEntryDelete = async (e: DictEntry) => {
    if (!selectedType) return;
    await dictEntryApi.delete(selectedType.id, e.id);
    message.success('条目已删除');
    actionRef.current?.reload();
  };

  const entryColumns: ProColumns[] = [
    {
      title: '关键词',
      dataIndex: 'q',
      hideInTable: true,
      search: { transform: (v) => ({ q: v }) },
    },
    { title: '显示名', dataIndex: 'label', width: 120, search: false },
    { title: '值', dataIndex: 'value', width: 120, search: false, render: (_, r) => <Tag>{r.value as string}</Tag> },
    { title: '排序', dataIndex: 'sort', width: 80, search: false },
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      valueEnum: { 1: { text: '启用' }, 0: { text: '禁用' } },
      render: (_, r) =>
        r.status === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>,
    },
    { title: '备注', dataIndex: 'remark', ellipsis: true, search: false },
    {
      title: '操作',
      valueType: 'option',
      render: (_, r) => [
        perms.includes('dict:update') ? (
          <a key="ed" onClick={() => openEditEntry(r as unknown as DictEntry)}>
            编辑
          </a>
        ) : null,
        perms.includes('dict:delete') ? (
          <Popconfirm key="del" title="确认删除？" onConfirm={() => handleEntryDelete(r as unknown as DictEntry)}>
            <a style={{ color: 'red' }}>删除</a>
          </Popconfirm>
        ) : null,
      ],
    },
  ];

  return (
    <div style={{ display: 'flex', gap: 16 }}>
      {/* 左侧：字典类型列表 */}
      <div style={{ width: 240, flexShrink: 0, background: '#fff', borderRadius: 8, padding: 12 }}>
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
        <List
          size="small"
          dataSource={types}
          locale={{ emptyText: <Empty description="暂无字典类型" image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
          renderItem={(t) => (
            <List.Item
              onClick={() => {
                setSelectedType(t);
                // 延迟刷新让 ProTable 用新的 selectedType 重新请求
                setTimeout(() => actionRef.current?.reload(), 0);
              }}
              style={{
                cursor: 'pointer',
                padding: '6px 8px',
                borderRadius: 6,
                background: selectedType?.id === t.id ? '#e6f4ff' : 'transparent',
                border: selectedType?.id === t.id ? '1px solid #91caff' : '1px solid transparent',
              }}
              actions={
                !perms.includes('dict:update') && !perms.includes('dict:delete')
                  ? []
                  : [
                      perms.includes('dict:update') ? (
                        <EditOutlined
                          key="edit"
                          onClick={(e) => {
                            e.stopPropagation();
                            openEditType(t);
                          }}
                        />
                      ) : null,
                      perms.includes('dict:delete') ? (
                        <Popconfirm key="del" title="删除类型？请先确保类型下无条目" onConfirm={() => handleTypeDelete(t)}>
                          <DeleteOutlined style={{ color: 'red' }} onClick={(e) => e.stopPropagation()} />
                        </Popconfirm>
                      ) : null,
                    ]
              }
            >
              <div>
                <div style={{ fontWeight: 500 }}>{t.name}</div>
                <div style={{ fontSize: 12, color: '#888' }}>
                  <Tag style={{ fontSize: 11 }}>{t.code}</Tag>
                  {t.is_public ? <Tag color="green" style={{ fontSize: 11 }}>公开</Tag> : null}
                  {t.status === 0 ? <Tag color="red" style={{ fontSize: 11 }}>禁用</Tag> : null}
                </div>
              </div>
            </List.Item>
          )}
        />
      </div>

      {/* 右侧：字典条目列表 */}
      <div style={{ flex: 1, minWidth: 0 }}>
        <ProTable
          columns={entryColumns}
          actionRef={actionRef}
          formRef={formRef}
          params={urlParams}
          request={async (params) => {
            if (!selectedType) return { data: [], success: true };
            const r = await dictEntryApi.listByTypeId(selectedType.id, buildQueryParams(params));
            const data = r.data as Record<string, unknown>;
            return {
              data: (data.items as unknown[]) || [],
              total: (data.total as number) || 0,
              success: true,
            };
          }}
          rowKey="id"
          onSubmit={onSearch}
          onReset={onReset}
          search={{ labelWidth: 'auto' }}
          pagination={{ current: page, pageSize, showSizeChanger: true, onChange: onPageChange }}
          headerTitle={
            selectedType
              ? `${selectedType.name}（${selectedType.code}）— 字典条目`
              : '请选择左侧字典类型'
          }
          toolBarRender={() =>
            selectedType && perms.includes('dict:create')
              ? [
                  <Button key="add" type="primary" icon={<PlusOutlined />} onClick={openCreateEntry}>
                    新增条目
                  </Button>,
                ]
              : []
          }
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
        </Form>
      </Modal>
    </div>
  );
};

export default DictManage;
