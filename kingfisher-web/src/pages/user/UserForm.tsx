import React from 'react';
import { Form, Input, Select } from 'antd';
import type { FormInstance } from 'antd';

interface UserFormProps {
  form: FormInstance;
  editing?: Record<string, unknown> | null;
  roles: { label: string; value: number }[];
}

const UserForm: React.FC<UserFormProps> = ({ form, editing, roles }) => (
  <Form form={form} layout="vertical" preserve={false}>
    <Form.Item name="username" label="用户名" rules={[{ required: true, min: 3, max: 32 }]}>
      <Input disabled={!!editing?.id} />
    </Form.Item>
    {!editing?.id && (
      <Form.Item name="password" label="密码" rules={[{ required: true, min: 8 }]}>
        <Input.Password />
      </Form.Item>
    )}
    <Form.Item name="email" label="邮箱" rules={[{ type: 'email', message: '邮箱格式不正确' }]}>
      <Input />
    </Form.Item>
    <Form.Item name="status" label="状态">
      <Select options={[
        { label: '启用', value: 1 },
        { label: '禁用', value: 0 },
      ]} />
    </Form.Item>
    <Form.Item name="role_id" label="角色">
      <Select options={roles} />
    </Form.Item>
  </Form>
);

export default UserForm;
