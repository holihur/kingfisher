import React, { useEffect, useState } from 'react';
import { Form, Input, Select, Button, App, Tag, Modal, Spin } from 'antd';
import { SendOutlined, RollbackOutlined } from '@ant-design/icons';
import DataTable from '../../components/DataTable';
import { userApi } from '../../api/user';
import { messageApi } from '../../api/message';
import { formatTime } from '../../utils/format';

interface UserOption {
  label: string;
  value: number;
}

/** 管理端按批次聚合的一条已发送记录 */
interface SentRow {
  batch_id: number;
  title: string;
  content?: string;
  recipient_count: number;
  recipient_names?: string;
  status: 'sent' | 'partial' | 'revoked';
  read_count?: number;
  unread_count?: number;
  created_at: string;
}

/** 批次详情里单条消息（逐收件人） */
interface BatchMsg {
  id: number;
  recipient_id: number;
  recipient_name?: string;
  title: string;
  content?: string;
  status: 'sent' | 'revoked';
  is_read: boolean;
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
  const [detail, setDetail] = useState<{ batch: SentRow; msgs: BatchMsg[]; loading: boolean } | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  // 打开批次详情：拉取该批次逐收件人消息
  const openDetail = async (row: SentRow) => {
    setDetail({ batch: row, msgs: [], loading: true });
    try {
      const r = await messageApi.listBatchMessages(row.batch_id);
      setDetail({ batch: row, msgs: (r.data as BatchMsg[]) || [], loading: false });
    } catch {
      setDetail({ batch: row, msgs: [], loading: false });
    }
  };

  // 批次详情里对单个收件人撤回
  const handleRevokeOne = (msg: BatchMsg) => {
    modal.confirm({
      title: `撤回发给「${msg.recipient_name || msg.recipient_id}」的消息？`,
      content: '撤回后该收件人将无法再看到。',
      okText: '撤回',
      okButtonProps: { danger: true },
      onOk: async () => {
        await messageApi.revoke(msg.id);
        message.success('已撤回');
        setDetail((d) => (d ? { ...d, msgs: d.msgs.map((m) => (m.id === msg.id ? { ...m, status: 'revoked' } : m)) } : d));
        setRefreshKey((k) => k + 1);
      },
    });
  };

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
        await messageApi.revokeBatch(row.batch_id);
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
          { title: 'ID', dataIndex: 'batch_id', width: 90, render: (v: number) => `#${v}` },
          { title: '标题', dataIndex: 'title', ellipsis: true },
          {
            title: '内容',
            dataIndex: 'content',
            ellipsis: true,
            render: (v: string) => v || '-',
          },
          {
            title: '收件人',
            dataIndex: 'recipient_names',
            ellipsis: true,
            render: (v: string, row: SentRow) => (v ? `${v}（${row.recipient_count}人）` : `${row.recipient_count}人`),
          },
          {
            title: '已读',
            dataIndex: 'read_count',
            width: 80,
            render: (v: number, row: SentRow) => `${v || 0}/${(row.read_count || 0) + (row.unread_count || 0)}`,
          },
          {
            title: '状态',
            dataIndex: 'status',
            width: 100,
            render: (v: string) =>
              v === 'revoked' ? (
                <Tag color="red">已撤回</Tag>
              ) : v === 'partial' ? (
                <Tag color="orange">部分撤回</Tag>
              ) : (
                <Tag color="green">已发送</Tag>
              ),
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
                <a onClick={() => void openDetail(row)}>查看</a>
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

      {/* 批次详情弹窗：批次摘要 + 逐收件人列表（可单个撤回） */}
      <Modal
        open={detail !== null}
        title="批次详情"
        footer={null}
        onCancel={() => setDetail(null)}
        width={620}
        destroyOnHidden
      >
        {detail && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <div>
              <div style={{ color: 'rgba(0,0,0,0.45)', fontSize: 12 }}>标题</div>
              <div style={{ fontWeight: 600, fontSize: 16 }}>{detail.batch.title}</div>
            </div>
            <div>
              <div style={{ color: 'rgba(0,0,0,0.45)', fontSize: 12 }}>内容</div>
              <div style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>{detail.batch.content || '-'}</div>
            </div>
            <div>
              <div style={{ color: 'rgba(0,0,0,0.45)', fontSize: 12, marginBottom: 4 }}>
                收件人明细（{detail.msgs.length}人）
              </div>
              <Spin spinning={detail.loading}>
                {detail.msgs.length === 0 && !detail.loading ? (
                  <div style={{ color: 'rgba(0,0,0,0.25)' }}>暂无明细</div>
                ) : (
                  <div style={{ border: '1px solid rgba(0,0,0,0.06)', borderRadius: 6 }}>
                    {detail.msgs.map((m) => (
                      <div
                        key={m.id}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'space-between',
                          padding: '8px 12px',
                          borderBottom: '1px solid rgba(0,0,0,0.06)',
                        }}
                      >
                        <span>
                          {m.recipient_name || `#${m.recipient_id}`}
                          {m.status === 'revoked' ? (
                            <Tag color="red" style={{ marginLeft: 8 }}>已撤回</Tag>
                          ) : m.is_read ? (
                            <Tag style={{ marginLeft: 8 }}>已读</Tag>
                          ) : (
                            <Tag color="blue" style={{ marginLeft: 8 }}>未读</Tag>
                          )}
                        </span>
                        {m.status !== 'revoked' && (
                          <a style={{ color: 'red' }} onClick={() => handleRevokeOne(m)}>
                            撤回
                          </a>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </Spin>
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
