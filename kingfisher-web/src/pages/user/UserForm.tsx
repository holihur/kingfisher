import React from 'react';
import { Form, Input, Select } from 'antd';
import type { FormInstance } from 'antd';

interface UserFormProps {
  form: FormInstance;
  editing?: Record<string, unknown> | null;
  roles: { label: string; value: number }[];
  /** 部门多选选项（扁平化后的部门树） */
  departments: { label: string; value: number }[];
}

const UserForm: React.FC<UserFormProps> = ({ form, editing, roles, departments }) => (
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
    <Form.Item
      name="role_ids"
      label="角色"
      extra="直接分配的角色；可留空（若属于某部门，则由部门角色继承）"
    >
      <Select mode="multiple" options={roles} placeholder="可多选" allowClear />
    </Form.Item>
    <Form.Item name="dept_ids" label="部门">
      <Select mode="multiple" options={departments} placeholder="可多选（继承部门角色）" allowClear />
    </Form.Item>
  </Form>
);

export default UserForm;
