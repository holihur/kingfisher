import React, { useCallback, useState } from 'react';
import { Modal, Form, Input, Select, App, Tag, Popconfirm, Button, Dropdown } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, DownOutlined } from '@ant-design/icons';
import DataTable, { SearchField } from '../../components/DataTable';
import { useAuthStore } from '../../stores/auth';
import { templateApi } from '../../api/template';
import { formatTime } from '../../utils/format';

interface Template {
  id: number;
  name: string;
  code: string;
  template_type: string;
  title: string;
  content: string;
  status: number;
  remark: string;
  version: string;
  created_at: string;
  updated_at: string;
}

const typeOptions = [
  { label: '通用', value: 'general' },
  { label: '站内信', value: 'message' },
  { label: '邮件', value: 'email' },
  { label: '短信', value: 'sms' },
];

const typeLabel = (t: string) => typeOptions.find((o) => o.value === t)?.label ?? t;

const TemplateManage: React.FC = () => {
  const { message, modal } = App.useApp();
  const perms = useAuthStore((s) => s.permissions);

  const [refreshKey, setRefreshKey] = useState(0);
  const [formModal, setFormModal] = useState<{ open: boolean; editing: Template | null }>({
    open: false,
    editing: null,
  });
  const [form] = Form.useForm<{
    name: string;
    code: string;
    template_type: string;
    title: string;
    content: string;
    status: number;
    remark: string;
  }>();

  const reload = useCallback(() => setRefreshKey((k) => k + 1), []);

  const openCreate = () => {
    form.resetFields();
    form.setFieldsValue({ template_type: 'general', status: 1 });
    setFormModal({ open: true, editing: null });
  };

  const openEdit = (t: Template) => {
    form.setFieldsValue({
      name: t.name,
      code: t.code,
      template_type: t.template_type,
      title: t.title,
      content: t.content,
      status: t.status,
      remark: t.remark,
    });
    setFormModal({ open: true, editing: t });
  };

  const handleSubmit = async () => {
    const v = await form.validateFields();
    const data = {
      name: v.name,
      code: v.code,
      template_type: v.template_type ?? 'general',
      title: v.title ?? '',
      content: v.content ?? '',
      status: v.status ?? 1,
      remark: v.remark ?? '',
    };
    if (formModal.editing) {
      await templateApi.update(formModal.editing.id, data);
      message.success('模版已更新');
    } else {
      await templateApi.create({ ...data, version: '1.0.0' });
      message.success('模版已创建');
    }
    setFormModal({ open: false, editing: null });
    reload();
  };

  const handleDelete = async (t: Template) => {
    await templateApi.delete(t.id);
    message.success('模版已删除');
    reload();
  };

  const searchFields: SearchField[] = [
    { name: 'q', label: '关键词', type: 'text' },
    { name: 'template_type', label: '模版类型', type: 'select', options: typeOptions },
    {
      name: 'status',
      label: '状态',
      type: 'select',
      options: [
        { label: '启用', value: 1 },
        { label: '禁用', value: 0 },
      ],
    },
  ];

  const columns = [
    { title: '模版名称', dataIndex: 'name', width: 160 },
    { title: '编码', dataIndex: 'code', width: 140, render: (v: unknown) => <Tag color="blue">{v as string}</Tag> },
    {
      title: '类型',
      dataIndex: 'template_type',
      width: 90,
      render: (v: unknown) => <Tag>{typeLabel(v as string)}</Tag>,
    },
    { title: '标题模板', dataIndex: 'title', ellipsis: true },
    {
      title: '内容模板',
      dataIndex: 'content',
      ellipsis: true,
      render: (v: unknown) => (v ? <span style={{ color: '#595959' }}>{v as string}</span> : '-'),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      render: (v: unknown) => (v === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>),
    },
    { title: '版本', dataIndex: 'version', width: 90, render: (v: unknown) => (v ? <Tag>{v as string}</Tag> : '-') },
    { title: '更新时间', dataIndex: 'updated_at', width: 150, render: (v: unknown) => formatTime(v) },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, r: Template) => [
        perms.includes('template:update') ? (
          <a key="ed" onClick={() => openEdit(r)}>
            <EditOutlined /> 编辑
          </a>
        ) : null,
        perms.includes('template:delete') ? (
          <Popconfirm key="del" title="删除模版" description={`确定删除「${r.name}」吗？`} onConfirm={() => handleDelete(r)}>
            <a style={{ color: 'red' }}><DeleteOutlined /> 删除</a>
          </Popconfirm>
        ) : null,
      ],
    },
  ];

  const request = useCallback(async (params: Record<string, unknown>) => {
    const r = await templateApi.list(params);
    const data = r.data as Record<string, unknown>;
    return {
      items: (data.items as Template[]) || [],
      total: (data.total as number) || 0,
    };
  }, []);

  const runStatus = async (ids: number[], status: number, label: string) => {
    await templateApi.batchUpdateStatus(ids, status);
    message.success(`已${label}`);
    reload();
  };

  return (
    <>
      <DataTable<Template>
        columns={columns}
        rowKey="id"
        request={request}
        searchFields={searchFields}
        reloadKey={refreshKey}
        headerTitle="模版管理"
        toolBarRender={
          perms.includes('template:create') ? (
            <Button key="add" type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              新增模版
            </Button>
          ) : null
        }
        selectable={perms.includes('template:update') || perms.includes('template:delete')}
        batchBarRender={(keys, clear) => {
          const ids = keys as number[];
          return (
            <Dropdown
              menu={{
                items: [
                  ...(perms.includes('template:update') ? [{ key: 'enable', label: '批量启用' }] : []),
                  ...(perms.includes('template:update') ? [{ key: 'disable', label: '批量禁用' }] : []),
                  ...(perms.includes('template:delete') ? [{ key: 'delete', label: '批量删除', danger: true }] : []),
                ],
                onClick: ({ key }) => {
                  if (key === 'enable') void runStatus(ids, 1, '启用').then(clear);
                  else if (key === 'disable') void runStatus(ids, 0, '禁用').then(clear);
                  else if (key === 'delete') {
                    modal.confirm({
                      title: '批量删除',
                      content: `确定删除选中的 ${ids.length} 个模版吗？`,
                      onOk: async () => {
                        await templateApi.batchDelete(ids);
                        message.success('已删除');
                        clear();
                        reload();
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

      {/* 新增/编辑弹窗 */}
      <Modal
        title={formModal.editing ? `编辑模版 — ${formModal.editing.name}` : '新增模版'}
        open={formModal.open}
        onOk={handleSubmit}
        onCancel={() => setFormModal({ open: false, editing: null })}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="模版名称" rules={[{ required: true, message: '请输入模版名称' }]}>
            <Input placeholder="如 欢迎消息" />
          </Form.Item>
          <Form.Item name="code" label="模版编码" rules={[{ required: true, message: '请输入模版编码' }]}>
            <Input placeholder="如 welcome_message" disabled={!!formModal.editing} />
          </Form.Item>
          <Form.Item name="template_type" label="模版类型">
            <Select options={typeOptions} />
          </Form.Item>
          <Form.Item name="title" label="标题模板" tooltip="支持占位符，如 欢迎 {{nickname}}">
            <Input placeholder="如 欢迎 {{nickname}}" />
          </Form.Item>
          <Form.Item name="content" label="内容模板" tooltip="支持占位符，如 你好 {{nickname}}">
            <Input.TextArea rows={4} placeholder="模板内容，可使用 {{占位符}}" />
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
    </>
  );
};

export default TemplateManage;