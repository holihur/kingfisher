import React, { useRef, useState } from 'react';
import { Modal, Form, Input, InputNumber, Switch, message, Tag, Popconfirm } from 'antd';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import { useAuthStore } from '../../stores/auth';
import { configApi } from '../../api/config';

/** Infer the config value type from the key name. */
function configInputType(key: string): 'number' | 'switch' | 'text' {
  if (/max_|attempts|timeout|duration|ttl/i.test(key)) return 'number';
  if (/enabled|enable|disabled/i.test(key)) return 'switch';
  return 'text';
}

const ConfigManage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [editModal, setEditModal] = useState<{ open: boolean; config: Record<string, unknown> | null }>({
    open: false,
    config: null,
  });
  const [form] = Form.useForm();
  const perms = useAuthStore((s) => s.permissions);

  const columns: ProColumns[] = [
    { title: 'Key', dataIndex: 'key', width: 200, render: (_, r) => <Tag color="blue">{r.key as string}</Tag> },
    {
      title: 'Value',
      dataIndex: 'value',
      render: (v) => <span style={{ wordBreak: 'break-all' }}>{v as string}</span>,
    },
    { title: '备注', dataIndex: 'remark', ellipsis: true },
    {
      title: '操作',
      valueType: 'option',
      render: (_, r) => [
        perms.includes('config:update') ? (
          <a
            key="ed"
            onClick={() => {
              form.setFieldsValue(r);
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
    await configApi.set(editModal.config!.key as string, String(v.value ?? ''));
    message.success('更新成功');
    setEditModal({ open: false, config: null });
    actionRef.current?.reload();
  };

  const renderValueInput = (key: string) => {
    const type = configInputType(key);
    if (type === 'number') {
      return <InputNumber style={{ width: '100%' }} />;
    }
    if (type === 'switch') {
      return <Switch checkedChildren="开" unCheckedChildren="关" />;
    }
    return <Input />;
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
          <Form.Item name="value" label="Value" rules={[{ required: true }]}>
            {editModal.config ? renderValueInput(editModal.config.key as string) : <Input />}
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
