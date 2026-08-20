import React, { useCallback, useEffect, useState } from 'react';
import { Modal, Form, Input, InputNumber, Switch, Select, App, Tag, Popconfirm, Button, Empty, Dropdown, Upload, Flex } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, DownOutlined, UploadOutlined } from '@ant-design/icons';
import DataTable, { SearchField } from '../../components/DataTable';
import { useAuthStore } from '../../stores/auth';
import { configApi, configGroupApi } from '../../api/config';
import { useTableUrlQuery } from '../../hooks/useTableUrlQuery';
import { useThemeToken } from '../../hooks/useThemeToken';
import { formatTime } from '../../utils/format';

/** 渲染组件类型：text|number|switch|select|textarea|image */
type RenderType = 'text' | 'number' | 'switch' | 'select' | 'textarea' | 'image';

const RENDER_TYPES: { value: RenderType; label: string }[] = [
  { value: 'text', label: '文本' },
  { value: 'number', label: '数字' },
  { value: 'switch', label: '开关' },
  { value: 'select', label: '下拉选择' },
  { value: 'textarea', label: '多行文本' },
  { value: 'image', label: '图片' },
];

/** 无 render 字段时的兜底：根据 key 名推断。 */
function inferRenderType(key: string): RenderType {
  if (/cover|logo|image|banner|img/i.test(key)) return 'image';
  if (/max_|attempts|timeout|duration|ttl/i.test(key)) return 'number';
  if (/enabled|enable|disabled/i.test(key)) return 'switch';
  return 'text';
}

/** 解析 select 的 options（render_options 为 JSON 字符串）。
 *  支持两种结构：纯数组 [{"label","value"},...]（单选），
 *  或 {multiple:true, options:[...]}（多选）。 */
function parseRenderOptions(raw: string | undefined): { options: { label: string; value: string }[]; multiple: boolean } {
  const empty = { options: [], multiple: false };
  if (!raw) return empty;
  try {
    const parsed = JSON.parse(raw);
    if (Array.isArray(parsed)) {
      return { options: parsed.map(toOpt).filter((o) => o.value !== ''), multiple: false };
    }
    if (parsed && typeof parsed === 'object') {
      const opts = Array.isArray(parsed.options) ? parsed.options.map(toOpt).filter((o) => o.value !== '') : [];
      return { options: opts, multiple: Boolean(parsed.multiple) };
    }
    return empty;
  } catch {
    return empty;
  }
}

function toOpt(o: unknown): { label: string; value: string } {
  const obj = (o ?? {}) as Record<string, unknown>;
  return { label: String(obj?.label ?? ''), value: String(obj?.value ?? '') };
}

/** 表单值 → 存储字符串（switch 存 "true"/"false"；数组如多选存逗号串）。 */
function valueToString(v: unknown, render: RenderType): string {
  if (render === 'switch') return String(Boolean(v));
  if (Array.isArray(v)) return v.join(',');
  return v === null || v === undefined ? '' : String(v);
}

/** 存储字符串 → 表单值（switch 变 boolean；number 转数字；多选逗号串转数组）。 */
function stringToFormValue(v: string | undefined, render: RenderType, multiple = false): unknown {
  if (render === 'switch') return v === 'true';
  if (render === 'number') {
    const n = Number(v);
    return Number.isFinite(n) ? n : v;
  }
  if (multiple && v) return v.split(',').map((s) => s.trim()).filter(Boolean);
  return v;
}

interface ConfigGroup {
  id: number;
  name: string;
  sort: number;
}

interface ConfigRow {
  key: string;
  value: string;
  remark: string;
  is_public: boolean;
  version: string;
  render: string;
  render_options: string;
  group_id: number;
  group_name?: string;
  created_at: string;
  updated_at: string;
}

const ConfigManage: React.FC = () => {
  const { message, modal } = App.useApp();
  const themeToken = useThemeToken();
  const { urlParams, updateUrl } = useTableUrlQuery();
  const [editModal, setEditModal] = useState<{ open: boolean; config: Record<string, unknown> | null; isNew: boolean }>({
    open: false,
    config: null,
    isNew: false,
  });
  const [refreshKey, setRefreshKey] = useState(0);
  const [form] = Form.useForm<Record<string, unknown>>();
  const perms = useAuthStore((s) => s.permissions);
  const token = useAuthStore((s) => s.token);

  /** 图片配置的受控输入：预览当前图 + 上传新图（写入 value） */
  const ImageConfigInput: React.FC<{ value?: string; onChange?: (v: string) => void }> = ({ value, onChange }) => (
    <div>
      {value ? (
        <img src={value} alt="preview" style={{ maxWidth: '100%', maxHeight: 96, borderRadius: 4, marginBottom: 8, display: 'block' }} />
      ) : null}
      <Upload
        name="file"
        action="/api/v1/configs/upload-image"
        headers={{ Authorization: `Bearer ${token}` }}
        accept=".png,.jpg,.jpeg,.gif,.webp"
        showUploadList={false}
        beforeUpload={(file) => {
          if (!file.type.startsWith('image/')) { message.error('仅支持图片文件'); return Upload.LIST_IGNORE; }
          if (file.size > 2 * 1024 * 1024) { message.error('文件大小不能超过 2MB'); return Upload.LIST_IGNORE; }
          return true;
        }}
        onChange={(info) => {
          if (info.file.status === 'done') {
            const url = (info.file.response as { data?: { url?: string } } | undefined)?.data?.url;
            if (url) { onChange?.(url); message.success('图片已上传'); }
          } else if (info.file.status === 'error') {
            message.error('上传失败');
          }
        }}
      >
        <Button icon={<UploadOutlined />}>上传图片</Button>
      </Upload>
    </div>
  );

  // 分组状态（选中的 group_id 来自 URL，刷新/分享可保持）
  const [groups, setGroups] = useState<ConfigGroup[]>([]);
  const selectedGroupId = Number(urlParams.group_id) || 0;
  const [groupModal, setGroupModal] = useState<{ open: boolean; editing: ConfigGroup | null }>({ open: false, editing: null });
  const [groupForm] = Form.useForm<{ name: string; sort: number }>();

  const loadGroups = useCallback(async () => {
    try {
      const r = await configGroupApi.list();
      setGroups((r.data as ConfigGroup[]) || []);
    } catch {
      /* interceptor handles */
    }
  }, []);

  useEffect(() => {
    loadGroups();
  }, [loadGroups]);

  // 编辑弹窗内当前选中的渲染组件（联动 Value 控件与选项编辑区）
  const watchedRender = (Form.useWatch('render', form) as RenderType | undefined) ||
    (editModal.config ? ((editModal.config.render as RenderType) || inferRenderType(editModal.config.key as string)) : 'text');
  const watchedOptions = Form.useWatch('render_options', form) as string | undefined;

  const searchFields: SearchField[] = [
    { name: 'q', label: '关键词', type: 'text' },
    {
      name: 'is_public',
      label: '是否公开',
      type: 'select',
      options: [
        { label: '公开', value: 'true' },
        { label: '私密', value: 'false' },
      ],
    },
    { name: 'version', label: '版本', type: 'text' },
    { name: 'render', label: '渲染组件', type: 'select', options: RENDER_TYPES },
  ];

  const columns = [
    { title: 'Key', dataIndex: 'key', width: 200, render: (_: unknown, r: ConfigRow) => <Tag color="blue">{r.key}</Tag> },
    {
      title: 'Value',
      dataIndex: 'value',
      render: (v: unknown, r: ConfigRow) => {
        const render = (r.render as RenderType) || inferRenderType(r.key);
        // Value 用渲染组件展示：image → 缩略图；switch → 开/关 Tag；select → 选中项 label；其余 → 文本
        if (render === 'image') {
          return v ? <img src={v as string} alt="config" style={{ height: 32, borderRadius: 4 }} /> : <span style={{ color: themeToken.colorTextDisabled }}>-</span>;
        }
        if (render === 'switch') {
          return v === 'true' ? <Tag color="green">开</Tag> : <Tag color="red">关</Tag>;
        }
        if (render === 'select') {
          const parsed = parseRenderOptions(r.render_options);
          // 多选：值存逗号串，拆开逐个匹配 label
          if (parsed.multiple && typeof v === 'string') {
            const parts = (v as string).split(',').map((s) => s.trim()).filter(Boolean);
            if (parts.length) {
              return (
                <Flex gap={4} wrap="wrap">
                  {parts.map((p) => {
                    const opt = parsed.options.find((o) => o.value === p);
                    return opt ? <Tag key={p} color="blue">{opt.label}</Tag> : <Tag key={p}>{p}</Tag>;
                  })}
                </Flex>
              );
            }
            return <span style={{ wordBreak: 'break-all' }}>{v as string}</span>;
          }
          const opt = parsed.options.find((o) => o.value === v);
          return opt ? <Tag color="blue">{opt.label}</Tag> : <span style={{ wordBreak: 'break-all' }}>{v as string}</span>;
        }
        return <span style={{ wordBreak: 'break-all' }}>{v as string}</span>;
      },
    },
    { title: '备注', dataIndex: 'remark', ellipsis: true },
    {
      title: '分组',
      dataIndex: 'group_name',
      width: 90,
      render: (_: unknown, r: ConfigRow) => (r.group_name ? <Tag color="purple">{r.group_name}</Tag> : '-'),
    },
    {
      title: '是否公开',
      dataIndex: 'is_public',
      width: 100,
      render: (_: unknown, r: ConfigRow) =>
        r.is_public ? <Tag color="green">公开</Tag> : <Tag color="default">私密</Tag>,
    },
    { title: '版本', dataIndex: 'version', width: 90 },
    { title: '更新时间', dataIndex: 'updated_at', width: 150, sorter: true, render: (v: unknown) => formatTime(v) },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, r: ConfigRow) => [
        perms.includes('config:update') ? (
          <a
            key="ed"
            onClick={() => {
              const render = (r.render as RenderType) || inferRenderType(r.key);
              form.setFieldsValue({
                ...r,
                is_public: !!r.is_public,
                value: stringToFormValue(r.value, render, parseRenderOptions(r.render_options).multiple),
              });
              setEditModal({ open: true, config: r as unknown as Record<string, unknown>, isNew: false });
            }}
          >
          <EditOutlined /> 编辑
          </a>
        ) : null,
        perms.includes('config:update') ? (
          <Popconfirm
            key="del"
            title="确认删除？"
            onConfirm={async () => {
              await configApi.delete(r.key);
              message.success('已删除');
              setRefreshKey((k) => k + 1);
            }}
          >
            <a style={{ color: 'red' }}><DeleteOutlined /> 删除</a>
          </Popconfirm>
        ) : null,
      ],
    },
  ];

  /** 打开新增配置弹窗：Key 可编辑，其余字段走默认/兜底推断 */
  const openCreate = () => {
    form.resetFields();
    form.setFieldsValue({ is_public: false, group_id: selectedGroupId || undefined });
    setEditModal({ open: true, config: null, isNew: true });
  };

  const handleEdit = async () => {
    const v = await form.validateFields();
    const key = (editModal.config?.key as string) || (v.key as string);
    const render = (v.render as RenderType) || inferRenderType(key);
    await configApi.set(
      key,
      valueToString(v.value, render),
      Boolean(v.is_public),
      String(v.version ?? ''),
      render,
      String(v.render_options ?? ''),
      Number(v.group_id) || 0,
    );
    message.success(editModal.isNew ? '配置已新增' : '更新成功');
    setEditModal({ open: false, config: null, isNew: false });
    setRefreshKey((k) => k + 1);
  };

  /** 根据 render 渲染 Value 编辑控件。 */
  const renderValueInput = (render: RenderType, parsed: { options: { label: string; value: string }[]; multiple: boolean }) => {
    switch (render) {
      case 'number':
        return <InputNumber style={{ width: '100%' }} />;
      case 'switch':
        return <Switch checkedChildren="开" unCheckedChildren="关" />;
      case 'select':
        return (
          <Select
            style={{ width: '100%' }}
            mode={parsed.multiple ? 'multiple' : undefined}
            options={parsed.options.length ? parsed.options : undefined}
            placeholder={parsed.multiple ? '请选择（可多选）' : '请选择（需先配置选项）'}
            allowClear
          />
        );
      case 'textarea':
        return <Input.TextArea rows={3} />;
      case 'image':
        return <ImageConfigInput />;
      default:
        return <Input />;
    }
  };

  // 分组 CRUD
  const openCreateGroup = () => {
    groupForm.resetFields();
    setGroupModal({ open: true, editing: null });
  };
  const openEditGroup = (g: ConfigGroup) => {
    groupForm.setFieldsValue({ name: g.name, sort: g.sort });
    setGroupModal({ open: true, editing: g });
  };
  const handleGroupSubmit = async () => {
    const v = await groupForm.validateFields();
    if (groupModal.editing) {
      await configGroupApi.update(groupModal.editing.id, v.name, v.sort ?? 0);
      message.success('分组已更新');
    } else {
      await configGroupApi.create(v.name, v.sort ?? 0);
      message.success('分组已创建');
    }
    setGroupModal({ open: false, editing: null });
    loadGroups();
  };
  const handleGroupDelete = async (g: ConfigGroup) => {
    await configGroupApi.delete(g.id);
    message.success('分组已删除，其下配置移回未分组');
    if (selectedGroupId === g.id) updateUrl({ group_id: '' });
    loadGroups();
    setRefreshKey((k) => k + 1);
  };

  return (
    <div style={{ display: 'flex', gap: 16 }}>
      {/* 左侧：分组列表 */}
      <div style={{ width: 220, flexShrink: 0, background: themeToken.colorBgContainer, borderRadius: themeToken.borderRadiusLG, padding: 12 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
          <span style={{ fontWeight: 600 }}>配置分组</span>
          {perms.includes('config:update') && (
            <Button size="small" type="primary" ghost icon={<PlusOutlined />} onClick={openCreateGroup}>
              新增
            </Button>
          )}
        </div>
        {groups.length === 0 ? (
          <Empty description="暂无分组" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : (
          <div>
            {[{ id: 0, name: '全部', sort: -1 }, ...groups].map((g) => (
              <div
                key={g.id}
                onClick={() => {
                  // updateUrl 保留现有搜索条件，只切换分组；group_id=0(全部) 时清空该参数
                  updateUrl({ group_id: g.id === 0 ? '' : g.id, page: 1 });
                }}
                style={{
                  cursor: 'pointer',
                  padding: '6px 8px',
                  borderRadius: 6,
                  marginBottom: 4,
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  background: selectedGroupId === g.id ? themeToken.colorPrimaryBg : 'transparent',
                  border: selectedGroupId === g.id ? `1px solid ${themeToken.colorPrimaryBorder}` : '1px solid transparent',
                }}
              >
                <span>{g.name}</span>
                {g.id !== 0 && perms.includes('config:update') && (
                  <span style={{ display: 'inline-flex', gap: 8 }}>
                    <EditOutlined onClick={(e) => { e.stopPropagation(); openEditGroup(g); }} />
                    <Popconfirm title="删除分组？其下配置将移回未分组" onConfirm={() => handleGroupDelete(g)}>
                      <DeleteOutlined style={{ color: 'red' }} onClick={(e) => e.stopPropagation()} />
                    </Popconfirm>
                  </span>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 右侧：配置列表 */}
      <div style={{ flex: 1, minWidth: 0 }}>
        <DataTable<ConfigRow>
          columns={columns}
          rowKey="key"
          request={async (params) => {
            const r = await configApi.getAll(params);
            const data = r.data as Record<string, unknown>;
            return {
              items: (data.items as ConfigRow[]) || [],
              total: (data.total as number) || 0,
            };
          }}
          searchFields={searchFields}
          reloadKey={refreshKey}
          headerTitle={`系统配置${selectedGroupId ? `（${groups.find(g => g.id === selectedGroupId)?.name ?? ''}）` : ''}`}
          selectable={perms.includes('config:update')}
          toolBarRender={
            perms.includes('config:update') ? (
              <Button key="add-config" type="primary" icon={<PlusOutlined />} onClick={openCreate}>
                新增配置
              </Button>
            ) : null
          }
          batchBarRender={(keys, clear) => {
            const k = keys as string[];
            return (
              <Dropdown
                menu={{
                  items: [{ key: 'delete', label: '批量删除', danger: true }],
                  onClick: () => {
                    modal.confirm({
                      title: '批量删除',
                      content: `确定删除选中的 ${k.length} 个配置吗？`,
                      onOk: async () => {
                        await configApi.batchDelete(k);
                        message.success('已删除');
                        clear();
                        setRefreshKey((n) => n + 1);
                      },
                    });
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

      {/* 配置编辑/新增弹窗 */}
      <Modal
        title={editModal.isNew ? '新增配置' : `编辑 — ${editModal.config?.key}`}
        open={editModal.open}
        onOk={handleEdit}
        onCancel={() => setEditModal({ open: false, config: null, isNew: false })}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="key" label="Key" rules={[{ required: true, message: '请输入 Key' }, { pattern: /^[a-z0-9_]+$/i, message: 'Key 仅支持字母、数字、下划线' }]}>
            <Input placeholder="如 new_feature_enabled" disabled={!editModal.isNew} />
          </Form.Item>
          <Form.Item
            name="value"
            label="Value"
            rules={[{ required: true }]}
            valuePropName={watchedRender === 'switch' ? 'checked' : 'value'}
          >
            {editModal.isNew || editModal.config ? renderValueInput(watchedRender, parseRenderOptions(watchedOptions)) : <Input />}
          </Form.Item>
          <Form.Item name="group_id" label="分组">
            <Select
              allowClear
              placeholder="选择分组"
              options={groups.map(g => ({ label: g.name, value: g.id }))}
            />
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

      {/* 分组新增/编辑弹窗 */}
      <Modal
        title={groupModal.editing ? `编辑分组 — ${groupModal.editing.name}` : '新增分组'}
        open={groupModal.open}
        onOk={handleGroupSubmit}
        onCancel={() => setGroupModal({ open: false, editing: null })}
      >
        <Form form={groupForm} layout="vertical">
          <Form.Item name="name" label="分组名" rules={[{ required: true, message: '请输入分组名' }]}>
            <Input placeholder="如 站点 / 安全" />
          </Form.Item>
          <Form.Item name="sort" label="排序">
            <InputNumber style={{ width: '100%' }} placeholder="数字越小越靠前" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ConfigManage;
