import React, { useEffect, useState } from 'react';
import { Card, Form, Input, Select, Button, App } from 'antd';
import { SendOutlined } from '@ant-design/icons';
import { userApi } from '../../api/user';
import { messageApi } from '../../api/message';

interface UserOption {
  label: string;
  value: number;
}

const MessageManage: React.FC = () => {
  const { message } = App.useApp();
  const [form] = Form.useForm();
  const [users, setUsers] = useState<UserOption[]>([]);
  const [searching, setSearching] = useState(false);
  const [sending, setSending] = useState(false);

  // 远程搜索收件人（按 username/nickname 模糊匹配，避免一次性拉全量用户）
  const searchUsers = async (keyword?: string) => {
    setSearching(true);
    try {
      const r = await userApi.getList({ page: 1, page_size: 20, q: keyword || '' });
      const data = r.data as Record<string, unknown>;
      const items = (data.items as Record<string, unknown>[]) || [];
      setUsers(items.map((u) => ({
        label: `${u.username as string}${u.nickname ? `（${u.nickname as string}）` : ''}`,
        value: u.id as number,
      })));
    } catch {
      /* interceptor handles */
    } finally {
      setSearching(false);
    }
  };

  useEffect(() => {
    searchUsers();
  }, []);

  const handleSend = async () => {
    const v = await form.validateFields();
    setSending(true);
    try {
      await messageApi.send({
        recipient_ids: v.recipient_ids as number[],
        title: v.title as string,
        content: (v.content as string) || '',
      });
      message.success(`站内信已发送给 ${(v.recipient_ids as number[]).length} 位用户`);
      form.resetFields();
    } catch {
      /* interceptor handles */
    } finally {
      setSending(false);
    }
  };

  return (
    <Card title="发送站内信" style={{ borderRadius: 8, border: 'none', boxShadow: '0 1px 2px rgba(0,0,0,0.04)' }}>
      <Form form={form} layout="vertical" style={{ maxWidth: 480 }}>
        <Form.Item name="recipient_ids" label="收件人" rules={[{ required: true, message: '请选择收件人' }]}>
          <Select
            mode="multiple"
            showSearch
            placeholder="搜索并选择用户（可多选）"
            options={users}
            filterOption={false}
            onSearch={(kw) => searchUsers(kw)}
            loading={searching}
            notFoundContent={searching ? null : '未找到匹配用户'}
            allowClear
          />
        </Form.Item>
        <Form.Item name="title" label="标题" rules={[{ required: true, message: '请输入标题' }]}>
          <Input placeholder="如：系统维护通知" />
        </Form.Item>
        <Form.Item name="content" label="内容">
          <Input.TextArea rows={4} placeholder="消息内容（可选）" />
        </Form.Item>
        <Form.Item>
          <Button type="primary" icon={<SendOutlined />} loading={sending} onClick={handleSend}>
            发送
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
};

export default MessageManage;
