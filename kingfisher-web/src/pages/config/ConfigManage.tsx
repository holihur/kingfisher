import React, { useRef, useState } from 'react';
import { Modal, Form, Input, InputNumber, Switch, Select, message, Tag, Popconfirm } from 'antd';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import { useAuthStore } from '../../stores/auth';
import { configApi } from '../../api/config';

/** 渲染组件类型：text|number|switch|select|textarea */
type RenderType = 'text' | 'number' | 'switch' | 'select' | 'textarea';

const RENDER_TYPES: { value: RenderType; label: string }[] = [
  { value: 'text', label: '文本' },
  { value: 'number', label: '数字' },
  { value: 'switch', label: '开关' },
  { value: 'select', label: '下拉选择' },
  { value: 'textarea', label: '多行文本' },
];

/** 无 render 字段时的兜底：根据 key 名推断。 */
function inferRenderType(key: string): RenderType {
  if (/max_|attempts|timeout|duration|ttl/i.test(key)) return 'number';
  if (/enabled|enable|disabled/i.test(key)) return 'switch';
  return 'text';
}

/** 解析 select 的 options（render_options 为 JSON 字符串）。 */
function parseRenderOptions(raw: string | undefined): { label: string; value: string }[] {
  if (!raw) return [];
  try {
    const arr = JSON.parse(raw);
    if (!Array.isArray(arr)) return [];
    return arr
      .map((o) => ({ label: String(o?.label ?? ''), value: String(o?.value ?? '') }))
      .filter((o) => o.value !== '');
  } catch {
    return [];
  }
}

/** 表单值 → 存储字符串（switch 存 "true"/"false"）。 */
function valueToString(v: unknown, render: RenderType): string {
  if (render === 'switch') return String(Boolean(v));
  return v === null || v === undefined ? '' : String(v);
}

/** 存储字符串 → 表单值（switch 变 boolean）。 */
function stringToFormValue(v: string | undefined, render: RenderType): unknown {
  if (render === 'switch') return v === 'true';
  if (render === 'number') {
    const n = Number(v);
    return Number.isFinite(n) ? n : v;
  }
  return v;
}

const ConfigManage: React.FC = () => {
  const actionRef = useRef<ActionType>(null);
  const [editModal, setEditModal] = useState<{ open: boolean; config: Record<string, unknown> | null }>({
    open: false,
    config: null,
  });
  const [form] = Form.useForm<Record<string, unknown>>();
  const perms = useAuthStore((s) => s.permissions);

  // 编辑弹窗内当前选中的渲染组件（联动 Value 控件与选项编辑区）
  const watchedRender = (Form.useWatch('render', form) as RenderType | undefined) ||
    (editModal.config ? ((editModal.config.render as RenderType) || inferRenderType(editModal.config.key as string)) : 'text');
  const watchedOptions = Form.useWatch('render_options', form) as string | undefined;

  const columns: ProColumns[] = [
    { title: 'Key', dataIndex: 'key', width: 200, render: (_, r) => <Tag color="blue">{r.key as string}</Tag> },
    {
      title: 'Value',
      dataIndex: 'value',
      render: (v, r) => {
        const render = (r.render as RenderType) || inferRenderType(r.key as string);
        // Value 用渲染组件展示：switch → 开/关 Tag；select → 选中项 label；其余 → 文本
        if (render === 'switch') {
          return v === 'true' ? <Tag color="green">开</Tag> : <Tag color="red">关</Tag>;
        }
        if (render === 'select') {
          const opt = parseRenderOptions(r.render_options as string).find(o => o.value === v);
          return opt ? <Tag color="blue">{opt.label}</Tag> : <span style={{ wordBreak: 'break-all' }}>{v as string}</span>;
        }
        return <span style={{ wordBreak: 'break-all' }}>{v as string}</span>;
      },
    },
    { title: '备注', dataIndex: 'remark', ellipsis: true },
    {
      title: '是否公开',
      dataIndex: 'is_public',
      width: 100,
      render: (_, r) =>
        r.is_public ? <Tag color="green">公开</Tag> : <Tag color="default">私密</Tag>,
    },
    { title: '版本', dataIndex: 'version', width: 100 },
    {
      title: '操作',
      valueType: 'option',
      render: (_, r) => [
        perms.includes('config:update') ? (
          <a
            key="ed"
            onClick={() => {
              const render = (r.render as RenderType) || inferRenderType(r.key as string);
              form.setFieldsValue({
                ...r,
                is_public: !!r.is_public,
                value: stringToFormValue(r.value as string, render),
              });
              setEditModal({ open: true, config: r });
            }}
          >
            编辑
          </a>
        ) : null,
        perms.includes('config:update') ? (
          <Popconfirm
            key="del"
            title="确认删除？"
            onConfirm={async () => {
              await configApi.delete(r.key as string);
              message.success('已删除');
              actionRef.current?.reload();
            }}
          >
            <a style={{ color: 'red' }}>删除</a>
          </Popconfirm>
        ) : null,
      ],
    },
  ];

  const handleEdit = async () => {
    const v = await form.validateFields();
    const render = (v.render as RenderType) || inferRenderType(editModal.config!.key as string);
    await configApi.set(
      editModal.config!.key as string,
      valueToString(v.value, render),
      Boolean(v.is_public),
      String(v.version ?? ''),
      render,
      String(v.render_options ?? ''),
    );
    message.success('更新成功');
    setEditModal({ open: false, config: null });
    actionRef.current?.reload();
  };

  /** 根据 render 渲染 Value 编辑控件。 */
  const renderValueInput = (render: RenderType, options: { label: string; value: string }[]) => {
    switch (render) {
      case 'number':
        return <InputNumber style={{ width: '100%' }} />;
      case 'switch':
        return <Switch checkedChildren="开" unCheckedChildren="关" />;
      case 'select':
        return (
          <Select
            style={{ width: '100%' }}
            options={options.length ? options : undefined}
            placeholder="请选择（需先配置选项）"
            allowClear
          />
        );
      case 'textarea':
        return <Input.TextArea rows={3} />;
      default:
        return <Input />;
    }
  };

  return (
    <>
      <ProTable
        columns={columns}
        actionRef={actionRef}
        request={async () => {
          const r = await configApi.getAll();
          return { data: (r.data as unknown[]) || [], success: true };
        }}
        rowKey="key"
        search={false}
        headerTitle="系统配置"
      />
      <Modal
        title={`编辑 — ${editModal.config?.key}`}
        open={editModal.open}
        onOk={handleEdit}
        onCancel={() => setEditModal({ open: false, config: null })}
      >
        <Form form={form} layout="vertical">
          <Form.Item label="Key">
            <Input value={(editModal.config?.key as string) || ''} disabled />
          </Form.Item>
          <Form.Item
            name="value"
            label="Value"
            rules={[{ required: true }]}
            valuePropName={watchedRender === 'switch' ? 'checked' : 'value'}
          >
            {editModal.config ? renderValueInput(watchedRender, parseRenderOptions(watchedOptions)) : <Input />}
          </Form.Item>
          <Form.Item name="render" label="渲染组件">
            <Select options={RENDER_TYPES} placeholder="用于编辑此字段的组件" allowClear />
          </Form.Item>
          {watchedRender === 'select' && (
            <Form.Item
              name="render_options"
              label="选项（JSON 数组）"
              extra={'格式：[{"label":"开启","value":"1"},{"label":"关闭","value":"0"}]'}
            >
              <Input.TextArea rows={3} placeholder='[{"label":"选项1","value":"1"}]' />
            </Form.Item>
          )}
          <Form.Item name="is_public" label="是否公开" valuePropName="checked">
            <Switch checkedChildren="公开" unCheckedChildren="私密" />
          </Form.Item>
          <Form.Item name="version" label="版本（表示该配置由哪个版本新增）">
            <Input placeholder="如 1.1.0" />
          </Form.Item>
          <Form.Item name="remark" label="备注">
            <Input.TextArea rows={2} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
};

export default ConfigManage;
