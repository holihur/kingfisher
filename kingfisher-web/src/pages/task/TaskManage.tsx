import React, { useCallback, useEffect, useState } from 'react';
import { Modal, Form, Input, Select, App, Tag, Popconfirm, Button, Dropdown, Alert } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, DownOutlined, PlayCircleOutlined } from '@ant-design/icons';
import DataTable, { SearchField } from '../../components/DataTable';
import { useAuthStore } from '../../stores/auth';
import { useThemeToken } from '../../hooks/useThemeToken';
import { scheduledTaskApi } from '../../api/task';
import { formatTime } from '../../utils/format';

interface ScheduledTask {
  id: number;
  name: string;
  task_type: string;
  cron_spec: string;
  payload: string;
  enabled: number;
  remark: string;
  created_at: string;
  updated_at: string;
}

interface TaskTypeInfo {
  type: string;
  label: string;
  payload_example?: string;
}

const TaskManage: React.FC = () => {
  const { message, modal } = App.useApp();
  const token = useThemeToken();
  const perms = useAuthStore((s) => s.permissions);

  const [taskTypes, setTaskTypes] = useState<TaskTypeInfo[]>([]);
  const [refreshKey, setRefreshKey] = useState(0);
  const [formModal, setFormModal] = useState<{ open: boolean; editing: ScheduledTask | null }>({
    open: false,
    editing: null,
  });
  const [form] = Form.useForm<{
    name: string;
    task_type: string;
    cron_spec: string;
    payload: string;
    enabled: number;
    remark: string;
  }>();
  const watchedType = Form.useWatch('task_type', form);
  const selectedType = taskTypes.find((t) => t.type === watchedType);

  // 动态加载可用任务类型（各模块 worker 声明），新增 worker 后无需改前端
  useEffect(() => {
    scheduledTaskApi
      .types()
      .then((r) => {
        const data = r.data as TaskTypeInfo[] | Record<string, unknown>;
        setTaskTypes((Array.isArray(data) ? data : []) as TaskTypeInfo[]);
      })
      .catch(() => {});
  }, []);

  const taskTypeOptions = taskTypes.map((t) => ({ label: `${t.label} (${t.type})`, value: t.type }));
  const taskTypeLabel = (t: string) => taskTypes.find((o) => o.type === t)?.label ?? t;

  const reload = useCallback(() => setRefreshKey((k) => k + 1), []);

  const openCreate = () => {
    form.resetFields();
    form.setFieldsValue({ task_type: taskTypes[0]?.type, enabled: 1 });
    setFormModal({ open: true, editing: null });
  };

  const openEdit = (t: ScheduledTask) => {
    form.setFieldsValue({
      name: t.name,
      task_type: t.task_type,
      cron_spec: t.cron_spec,
      payload: t.payload,
      enabled: t.enabled,
      remark: t.remark,
    });
    setFormModal({ open: true, editing: t });
  };

  const handleSubmit = async () => {
    const v = await form.validateFields();
    const data = {
      name: v.name,
      task_type: v.task_type,
      cron_spec: v.cron_spec,
      payload: v.payload ?? '',
      enabled: v.enabled ?? 1,
      remark: v.remark ?? '',
    };
    if (formModal.editing) {
      await scheduledTaskApi.update(formModal.editing.id, data);
      message.success('任务已更新，将在下一个同步周期生效');
    } else {
      await scheduledTaskApi.create(data);
      message.success('任务已创建，将在下一个同步周期生效');
    }
    setFormModal({ open: false, editing: null });
    reload();
  };

  const handleDelete = async (t: ScheduledTask) => {
    await scheduledTaskApi.delete(t.id);
    message.success('任务已删除，将在下一个同步周期生效');
    reload();
  };

  const handleRun = async (t: ScheduledTask) => {
    await scheduledTaskApi.run(t.id);
    message.success(`已手动执行「${t.name}」`);
  };

  const searchFields: SearchField[] = [
    { name: 'q', label: '关键词', type: 'text' },
    { name: 'task_type', label: '任务类型', type: 'select', options: taskTypeOptions },
    {
      name: 'enabled',
      label: '状态',
      type: 'select',
      options: [
        { label: '启用', value: 1 },
        { label: '禁用', value: 0 },
      ],
    },
  ];

  const columns = [
    { title: '任务名称', dataIndex: 'name', width: 160 },
    {
      title: '任务类型',
      dataIndex: 'task_type',
      width: 200,
      render: (v: unknown) => <Tag color="blue">{taskTypeLabel(v as string)}</Tag>,
    },
    {
      title: 'cron 表达式',
      dataIndex: 'cron_spec',
      width: 120,
      render: (v: unknown) => <Tag color="geekblue">{v as string}</Tag>,
    },
    {
      title: '载荷',
      dataIndex: 'payload',
      ellipsis: true,
      render: (v: unknown) => (v ? <span style={{ color: token.colorTextSecondary, fontFamily: 'monospace' }}>{v as string}</span> : '-'),
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      width: 80,
      render: (v: unknown) => (v === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>),
    },
    { title: '备注', dataIndex: 'remark', ellipsis: true },
    { title: '更新时间', dataIndex: 'updated_at', width: 150, sorter: true, render: (v: unknown) => formatTime(v) },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, r: ScheduledTask) => [
        perms.includes('task:update') ? (
          <Popconfirm
            key="run"
            title="手动执行任务"
            description={`确定立即执行「${r.name}」吗？将绕过 cron 调度直接入队一次。`}
            onConfirm={() => handleRun(r)}
          >
            <a>
              <PlayCircleOutlined /> 执行
            </a>
          </Popconfirm>
        ) : null,
        perms.includes('task:update') ? (
          <a key="ed" onClick={() => openEdit(r)}>
            <EditOutlined /> 编辑
          </a>
        ) : null,
        perms.includes('task:delete') ? (
          <Popconfirm key="del" title="删除任务" description={`确定删除「${r.name}」吗？`} onConfirm={() => handleDelete(r)}>
            <a style={{ color: 'red' }}><DeleteOutlined /> 删除</a>
          </Popconfirm>
        ) : null,
      ],
    },
  ];

  const request = useCallback(async (params: Record<string, unknown>) => {
    const r = await scheduledTaskApi.list(params);
    const data = r.data as Record<string, unknown>;
    return {
      items: (data.items as ScheduledTask[]) || [],
      total: (data.total as number) || 0,
    };
  }, []);

  const runStatus = async (ids: number[], enabled: number, label: string) => {
    await scheduledTaskApi.batchUpdateStatus(ids, enabled);
    message.success(`已${label}，将在下一个同步周期生效`);
    reload();
  };

  return (
    <>
      <Alert
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
        message="周期任务由 asynq 调度器周期性执行：任务配置从数据库拉取，修改后会在下一个同步周期（默认 60 秒）自动生效，无需重启服务。"
      />
      <DataTable<ScheduledTask>
        columns={columns}
        rowKey="id"
        request={request}
        searchFields={searchFields}
        reloadKey={refreshKey}
        headerTitle="周期任务管理"
        toolBarRender={
          perms.includes('task:create') ? (
            <Button key="add" type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              新增任务
            </Button>
          ) : null
        }
        selectable={perms.includes('task:update') || perms.includes('task:delete')}
        batchBarRender={(keys, clear) => {
          const ids = keys as number[];
          return (
            <Dropdown
              menu={{
                items: [
                  ...(perms.includes('task:update') ? [{ key: 'enable', label: '批量启用' }] : []),
                  ...(perms.includes('task:update') ? [{ key: 'disable', label: '批量禁用' }] : []),
                  ...(perms.includes('task:delete') ? [{ key: 'delete', label: '批量删除', danger: true }] : []),
                ],
                onClick: ({ key }) => {
                  if (key === 'enable') void runStatus(ids, 1, '启用').then(clear);
                  else if (key === 'disable') void runStatus(ids, 0, '禁用').then(clear);
                  else if (key === 'delete') {
                    modal.confirm({
                      title: '批量删除',
                      content: `确定删除选中的 ${ids.length} 个周期任务吗？`,
                      onOk: async () => {
                        await scheduledTaskApi.batchDelete(ids);
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
        title={formModal.editing ? `编辑任务 — ${formModal.editing.name}` : '新增周期任务'}
        open={formModal.open}
        onOk={handleSubmit}
        onCancel={() => setFormModal({ open: false, editing: null })}
        width={560}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="任务名称" rules={[{ required: true, message: '请输入任务名称' }]}>
            <Input placeholder="如 每日 9 点系统通知" />
          </Form.Item>
          <Form.Item name="task_type" label="任务类型" rules={[{ required: true, message: '请选择任务类型' }]}>
            <Select options={taskTypeOptions} />
          </Form.Item>
          <Form.Item
            name="cron_spec"
            label="cron 表达式"
            tooltip="5 段标准格式：分 时 日 月 周。如 0 9 * * * 表示每天 9:00"
            rules={[{ required: true, message: '请输入 cron 表达式' }]}
          >
            <Input placeholder="如 0 9 * * *" />
          </Form.Item>
          <Form.Item
            name="payload"
            label="任务载荷"
            tooltip="JSON 字符串，传递给对应 worker。留空则不带载荷"
          >
            <Input.TextArea rows={3} placeholder={selectedType?.payload_example || '可选，JSON 格式'} />
          </Form.Item>
          <Form.Item name="enabled" label="状态">
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

export default TaskManage;
