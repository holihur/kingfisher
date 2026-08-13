import React, { useEffect, useState } from 'react';
import { Form, Input, Select, Button, App, Tag, Modal } from 'antd';
import { SendOutlined, RollbackOutlined } from '@ant-design/icons';
import DataTable from '../../components/DataTable';
import { userApi } from '../../api/user';
import { messageApi } from '../../api/message';
import { formatTime } from '../../utils/format';

interface UserOption {
  label: string;
  value: number;
}

interface SentRow {
  id: number;
  recipient_id: number;
  recipient_name?: string;
  title: string;
  content?: string;
  status: 'sent' | 'revoked';
  created_at: string;
}

/** 站内信管理：已发送列表（含撤回）+ 发送弹窗。遵循项目其他管理页（模版/任务）模式。 */
const MessageManage: React.FC = () => {
  const { message, modal } = App.useApp();
  const [form] = Form.useForm();
  const [users, setUsers] = useState<UserOption[]>([]);
  const [searching, setSearching] = useState(false);
  const [sending, setSending] = useState(false);
  const [sendOpen, setSendOpen] = useState(false);
  const [detail, setDetail] = useState<SentRow | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

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
      setSendOpen(false);
      setRefreshKey((k) => k + 1);
    } catch {
      /* interceptor handles */
    } finally {
      setSending(false);
    }
  };

  // 撤回：确认后标记 revoked，收件箱不可见
  const handleRevoke = (row: SentRow) => {
    modal.confirm({
      title: `撤回「${row.title}」？`,
      content: '撤回后收件人将无法再看到该站内信。',
      okText: '撤回',
      okButtonProps: { danger: true },
      onOk: async () => {
        await messageApi.revoke(row.id);
        message.success('已撤回');
        setRefreshKey((k) => k + 1);
      },
    });
  };

  return (
    <>
      <DataTable<SentRow>
        rowKey="id"
        reloadKey={refreshKey}
        headerTitle="已发送站内信"
        toolBarRender={
          <Button type="primary" icon={<SendOutlined />} onClick={() => setSendOpen(true)}>
            发送站内信
          </Button>
        }
        request={async (params) => {
          const resp = await messageApi.listSent(params);
          const data = resp.data as Record<string, unknown>;
          return {
            items: (data.items as SentRow[]) || [],
            total: (data.total as number) || 0,
          };
        }}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 60 },
          { title: '标题', dataIndex: 'title', ellipsis: true },
          {
            title: '内容',
            dataIndex: 'content',
            ellipsis: true,
            render: (v: string) => v || '-',
          },
          {
            title: '收件人',
            dataIndex: 'recipient_name',
            width: 120,
            render: (v: string | undefined, row: SentRow) => v || `#${row.recipient_id}`,
          },
          {
            title: '状态',
            dataIndex: 'status',
            width: 90,
            render: (v: string) => (v === 'revoked' ? <Tag color="red">已撤回</Tag> : <Tag color="green">已发送</Tag>),
          },
          {
            title: '时间',
            dataIndex: 'created_at',
            width: 150,
            render: (v: string) => formatTime(v),
          },
          {
            title: '操作',
            key: 'actions',
            width: 120,
            render: (_: unknown, row: SentRow) => (
              <>
                <a onClick={() => setDetail(row)}>查看</a>
                {row.status === 'revoked' ? (
                  <span style={{ color: 'rgba(0,0,0,0.25)', marginLeft: 8 }}>已撤回</span>
                ) : (
                  <a style={{ color: 'red', marginLeft: 8 }} onClick={() => handleRevoke(row)}>
                    <RollbackOutlined /> 撤回
                  </a>
                )}
              </>
            ),
          },
        ]}
      />

      {/* 详情弹窗 */}
      <Modal
        open={detail !== null}
        title="站内信详情"
        footer={null}
        onCancel={() => setDetail(null)}
        width={520}
        destroyOnHidden
      >
        {detail && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <div>
              <div style={{ color: 'rgba(0,0,0,0.45)', fontSize: 12 }}>标题</div>
              <div style={{ fontWeight: 600, fontSize: 16 }}>{detail.title}</div>
            </div>
            <div>
              <div style={{ color: 'rgba(0,0,0,0.45)', fontSize: 12 }}>收件人</div>
              <div>{detail.recipient_name || `#${detail.recipient_id}`}</div>
            </div>
            <div>
              <div style={{ color: 'rgba(0,0,0,0.45)', fontSize: 12 }}>状态</div>
              <div>
                {detail.status === 'revoked' ? <Tag color="red">已撤回</Tag> : <Tag color="green">已发送</Tag>}
              </div>
            </div>
            <div>
              <div style={{ color: 'rgba(0,0,0,0.45)', fontSize: 12 }}>发送时间</div>
              <div>{formatTime(detail.created_at)}</div>
            </div>
            <div>
              <div style={{ color: 'rgba(0,0,0,0.45)', fontSize: 12 }}>内容</div>
              <div style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>{detail.content || '-'}</div>
            </div>
          </div>
        )}
      </Modal>

      {/* 发送站内信弹窗 */}
      <Modal
        open={sendOpen}
        title="发送站内信"
        okText="发送"
        cancelText="取消"
        confirmLoading={sending}
        onOk={handleSend}
        onCancel={() => setSendOpen(false)}
        destroyOnHidden
      >
        <Form form={form} layout="vertical">
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
        </Form>
      </Modal>
    </>
  );
};

export default MessageManage;
