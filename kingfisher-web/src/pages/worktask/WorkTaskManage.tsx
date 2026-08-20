import React, { useCallback, useState } from 'react';
import { App, Button, Form, Input, Modal, Popconfirm, Select, Tag } from 'antd';
import { DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons';
import DataTable, { SearchField } from '../../components/DataTable';
import { useAuthStore } from '../../stores/auth';
import { formatTime } from '../../utils/format';
import { workTaskApi, type WorkTask, type WorkTaskInput } from '../../api/worktask';

type TaskForm = WorkTaskInput;

const statusOptions = [
  { label: '待处理', value: 'open' },
  { label: '进行中', value: 'in_progress' },
  { label: '已完成', value: 'done' },
];

const statusLabel: Record<string, string> = {
  open: '待处理',
  in_progress: '进行中',
  done: '已完成',
};

const WorkTaskManage: React.FC = () => {
  const { message } = App.useApp();
  const permissions = useAuthStore((state) => state.permissions);
  const [reloadKey, setReloadKey] = useState(0);
  const [editing, setEditing] = useState<WorkTask | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm<TaskForm>();

  const reload = useCallback(() => setReloadKey((value) => value + 1), []);
  const openCreate = () => {
    form.resetFields();
    form.setFieldsValue({ department_id: 0, status: 'open', description: '' });
    setEditing(null);
    setModalOpen(true);
  };
  const openEdit = (task: WorkTask) => {
    form.setFieldsValue({
      title: task.title,
      description: task.description,
      department_id: task.department_id,
      status: task.status,
    });
    setEditing(task);
    setModalOpen(true);
  };
  const submit = async () => {
    const values = await form.validateFields();
    if (editing) {
      await workTaskApi.update(editing.id, values);
      message.success('任务已更新');
    } else {
      await workTaskApi.create(values);
      message.success('任务已创建');
    }
    setModalOpen(false);
    reload();
  };
  const remove = async (task: WorkTask) => {
    await workTaskApi.delete(task.id);
    message.success('任务已删除');
    reload();
  };

  const searchFields: SearchField[] = [
    { name: 'q', label: '关键词', type: 'text' },
    { name: 'status', label: '状态', type: 'select', options: statusOptions },
  ];
  const columns = [
    { title: '任务标题', dataIndex: 'title', width: 220 },
    { title: '说明', dataIndex: 'description', ellipsis: true },
    { title: '负责人', dataIndex: 'owner_id', width: 100 },
    { title: '部门', dataIndex: 'department_id', width: 90 },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (value: string) => <Tag color={value === 'done' ? 'green' : value === 'in_progress' ? 'blue' : 'default'}>{statusLabel[value] ?? value}</Tag>,
    },
    { title: '更新时间', dataIndex: 'updated_at', width: 160, sorter: true, render: (value: string) => formatTime(value) },
    {
      title: '操作',
      key: 'action',
      width: 130,
      render: (_: unknown, task: WorkTask) => [
        permissions.includes('worktask:update') ? <a key="edit" onClick={() => openEdit(task)}><EditOutlined /> 编辑</a> : null,
        permissions.includes('worktask:delete') ? (
          <Popconfirm key="delete" title="确认删除任务？" onConfirm={() => remove(task)}>
            <a style={{ color: 'red' }}><DeleteOutlined /> 删除</a>
          </Popconfirm>
        ) : null,
      ],
    },
  ];
  const request = useCallback(async (params: Record<string, unknown>) => {
    const response = await workTaskApi.list(params);
    const data = response.data as { items?: WorkTask[]; total?: number };
    return { items: data.items ?? [], total: data.total ?? 0 };
  }, []);

  return (
    <>
      <DataTable<WorkTask>
        columns={columns}
        rowKey="id"
        request={request}
        searchFields={searchFields}
        reloadKey={reloadKey}
        headerTitle="业务任务"
        headerSubtitle="按数据权限展示可访问任务"
        toolBarRender={permissions.includes('worktask:create') ? <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建任务</Button> : null}
      />
      <Modal title={editing ? '编辑任务' : '新建任务'} open={modalOpen} onOk={submit} onCancel={() => setModalOpen(false)}>
        <Form form={form} layout="vertical">
          <Form.Item name="title" label="任务标题" rules={[{ required: true, message: '请输入任务标题' }]}><Input /></Form.Item>
          <Form.Item name="description" label="任务说明"><Input.TextArea rows={4} /></Form.Item>
          <Form.Item name="department_id" label="部门 ID" rules={[{ required: true, message: '请输入部门 ID' }]}><Input type="number" /></Form.Item>
          <Form.Item name="status" label="状态" rules={[{ required: true, message: '请选择状态' }]}><Select options={statusOptions} /></Form.Item>
        </Form>
      </Modal>
    </>
  );
};

export default WorkTaskManage;
