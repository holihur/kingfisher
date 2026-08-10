import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Form, Input, Button, App, Descriptions, Tag, Menu, Upload, Table, Popconfirm, Dropdown, Spin, Empty, Space } from 'antd';
import type { TableProps, UploadProps } from 'antd';
import { UserOutlined, MailOutlined, LockOutlined, UploadOutlined, SolutionOutlined, KeyOutlined, FileTextOutlined, DownOutlined } from '@ant-design/icons';
import PageCard from '../../components/PageCard';
import { useThemeToken } from '../../hooks/useThemeToken';
import { userApi } from '../../api/user';
import { messageApi } from '../../api/message';
import { useAuthStore } from '../../stores/auth';
import { formatTime } from '../../utils/format';

interface LoginLog {
  id: number;
  username: string;
  ip: string;
  user_agent: string;
  created_at: string;
}

interface MessageRow {
  id: number;
  sender_id: number;
  sender_type: string;
  title: string;
  content: string;
  is_read: boolean;
  created_at: string;
}

const Profile: React.FC = () => {
  const { message } = App.useApp();
  const themeToken = useThemeToken();
  const { userInfo, fetchUserInfo, token } = useAuthStore();
  const [profileForm] = Form.useForm();
  const [pwdForm] = Form.useForm();
  const [avatarUrl, setAvatarUrl] = useState<string>('');
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  // 左侧菜单当前选中项（支持 ?tab=inbox 直达收件箱）
  const [activeKey, setActiveKey] = useState(searchParams.get('tab') || 'profile');
  // 消息详情 id（?tab=inbox&id= 直达详情）
  const msgDetailId = searchParams.get('id');
  // 登录日志本地分页
  const [logs, setLogs] = useState<LoginLog[]>([]);
  const [logTotal, setLogTotal] = useState(0);
  const [logPage, setLogPage] = useState(1);
  const [logLoading, setLogLoading] = useState(false);
  // 收件箱
  const [msgs, setMsgs] = useState<MessageRow[]>([]);
  const [msgTotal, setMsgTotal] = useState(0);
  const [msgPage, setMsgPage] = useState(1);
  const [msgLoading, setMsgLoading] = useState(false);
  const [msgRefresh, setMsgRefresh] = useState(0);
  const [selectedMsgKeys, setSelectedMsgKeys] = useState<React.Key[]>([]);
  // 消息详情
  const [msgDetail, setMsgDetail] = useState<MessageRow | null>(null);
  const [msgDetailLoading, setMsgDetailLoading] = useState(false);

  useEffect(() => {
    if (userInfo) {
      const avatar = ((userInfo as Record<string, unknown>).avatar as string) || '';
      setAvatarUrl(avatar);
      profileForm.setFieldsValue({
        nickname: (userInfo as Record<string, unknown>).nickname || '',
        email: (userInfo as Record<string, unknown>).email || '',
      });
    }
  }, [userInfo, profileForm]);

  // 加载收件箱
  useEffect(() => {
    let cancelled = false;
    setMsgLoading(true);
    messageApi
      .list({ page: msgPage, page_size: 10, sort: '-created_at' })
      .then((r) => {
        if (cancelled) return;
        const data = r.data as Record<string, unknown>;
        setMsgs((data?.items as MessageRow[]) || []);
        setMsgTotal((data?.total as number) || 0);
      })
      .catch(() => { if (cancelled) return; setMsgs([]); setMsgTotal(0); })
      .finally(() => { if (!cancelled) setMsgLoading(false); });
    return () => { cancelled = true; };
  }, [msgPage, msgRefresh]);

  // 加载消息详情（有 id 时），未读则自动标记已读
  useEffect(() => {
    if (!msgDetailId) { setMsgDetail(null); return; }
    let cancelled = false;
    setMsgDetailLoading(true);
    messageApi.getById(Number(msgDetailId))
      .then(async (r) => {
        if (cancelled) return;
        const m = r.data as MessageRow;
        setMsgDetail(m);
        if (!m.is_read) {
          await messageApi.markRead(m.id);
          setMsgRefresh((k) => k + 1);
        }
      })
      .catch(() => { if (cancelled) return; setMsgDetail(null); })
      .finally(() => { if (!cancelled) setMsgDetailLoading(false); });
    return () => { cancelled = true; };
  }, [msgDetailId]);

  const inboxColumns: TableProps<MessageRow>['columns'] = [
    {
      title: '标题',
      dataIndex: 'title',
      render: (v: unknown, r: MessageRow) => (
        <a
          onClick={() => navigate(`/profile?tab=inbox&id=${r.id}`)}
          style={{ fontWeight: r.is_read ? 400 : 600 }}
        >
          {v as string}
        </a>
      ),
    },
    {
      title: '发件人',
      dataIndex: 'sender_type',
      width: 90,
      render: (v: unknown) => (v === 'system' ? <Tag color="blue">系统</Tag> : <Tag>管理员</Tag>),
    },
    { title: '时间', dataIndex: 'created_at', width: 160, render: (v: unknown) => formatTime(v) },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, r: MessageRow) => [
        !r.is_read ? (
          <a
            key="read"
            onClick={async () => {
              await messageApi.markRead(r.id);
              message.success('已标记为已读');
              setMsgRefresh((k) => k + 1);
            }}
          >
            标记已读
          </a>
        ) : null,
        <Popconfirm
          key="del"
          title="删除消息？"
          onConfirm={async () => {
            await messageApi.batchDelete([r.id]);
            message.success('已删除');
            setMsgRefresh((k) => k + 1);
          }}
        >
          <a style={{ color: 'red' }}>删除</a>
        </Popconfirm>,
      ],
    },
  ];

  const handleInboxBatch = async (key: string) => {
    const ids = selectedMsgKeys as number[];
    if (key === 'read') {
      for (const id of ids) await messageApi.markRead(id);
      message.success('已标记为已读');
    } else {
      await messageApi.batchDelete(ids);
      message.success('已删除');
    }
    setSelectedMsgKeys([]);
    setMsgRefresh((k) => k + 1);
  };

  const handleProfileSave = useCallback(async () => {
    let v: Record<string, unknown>;
    try {
      v = await profileForm.validateFields();
    } catch (e) {
      // 校验失败：提示第一条错误，不静默
      message.error((e as { errorFields?: { errors?: string[] }[] })?.errorFields?.[0]?.errors?.[0] || '请检查表单填写');
      return;
    }
    await userApi.updateMe({
      email: (v.email as string) || '',
      nickname: (v.nickname as string) || '',
    });
    message.success('资料已更新');
    await fetchUserInfo();
  }, [profileForm, fetchUserInfo, message]);

  const handlePasswordChange = useCallback(async () => {
    let v: Record<string, unknown>;
    try {
      v = await pwdForm.validateFields();
    } catch (e) {
      // 校验失败：提示第一条错误，不静默
      message.error((e as { errorFields?: { errors?: string[] }[] })?.errorFields?.[0]?.errors?.[0] || '请检查表单填写');
      return;
    }
    if (v.new_password !== v.confirm_password) {
      message.error('两次输入的新密码不一致');
      return;
    }
    await userApi.changePassword({
      old_password: v.old_password as string,
      new_password: v.new_password as string,
    });
    message.success('密码已修改，请重新登录');
    pwdForm.resetFields();
  }, [pwdForm, message]);

  const uploadProps: UploadProps = {
    name: 'file',
    action: '/api/v1/users/me/avatar',
    headers: { Authorization: `Bearer ${token}` },
    accept: '.png,.jpg,.jpeg,.gif,.webp',
    showUploadList: false,
    beforeUpload: (file) => {
      const isImage = file.type.startsWith('image/');
      if (!isImage) {
        message.error('仅支持图片文件');
        return Upload.LIST_IGNORE;
      }
      if (file.size > 2 * 1024 * 1024) {
        message.error('文件大小不能超过 2MB');
        return Upload.LIST_IGNORE;
      }
      return true;
    },
    onChange: (info) => {
      if (info.file.status === 'done') {
        const url = (info.file.response as { data?: { url?: string } } | undefined)?.data?.url as string;
        if (url) {
          setAvatarUrl(url);
          message.success('头像已更新');
          fetchUserInfo();
        }
      } else if (info.file.status === 'error') {
        message.error('上传失败');
      }
    },
  };

  const loginLogColumns: TableProps<LoginLog>['columns'] = [
    {
      title: '时间',
      dataIndex: 'created_at',
      width: 180,
      render: (v: unknown) => (v ? new Date(v as string).toLocaleString() : '-'),
    },
    { title: '用户名', dataIndex: 'username', width: 120 },
    { title: 'IP', dataIndex: 'ip', width: 140, render: (_: unknown, r: LoginLog) => <Tag>{r.ip}</Tag> },
    { title: 'UserAgent', dataIndex: 'user_agent', ellipsis: true },
  ];

  // 加载登录日志（Tab 内）
  useEffect(() => {
    let cancelled = false;
    setLogLoading(true);
    userApi
      .getMyLoginLogs({ page: logPage, page_size: 10, sort: '-created_at' })
      .then((r) => {
        if (cancelled) return;
        const data = r.data as Record<string, unknown>;
        setLogs((data?.items as LoginLog[]) || (r.data as LoginLog[]) || []);
        setLogTotal((data?.total as number) || 0);
      })
      .catch(() => {
        if (cancelled) return;
        setLogs([]);
        setLogTotal(0);
      })
      .finally(() => {
        if (!cancelled) setLogLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [logPage]);

  return (
    <div style={{ display: 'flex', gap: 16, width: '100%' }}>
      {/* 左侧菜单 */}
      <div style={{ width: 200, flexShrink: 0, background: themeToken.colorBgContainer, borderRadius: themeToken.borderRadiusLG, padding: 8, boxShadow: themeToken.boxShadowTertiary }}>
        <Menu
          mode="inline"
          selectedKeys={[activeKey]}
          onClick={({ key }) => setActiveKey(key)}
          style={{ border: 'none' }}
          items={[
            { key: 'profile', icon: <SolutionOutlined />, label: '用户资料' },
            { key: 'password', icon: <KeyOutlined />, label: '修改密码' },
            { key: 'inbox', icon: <MailOutlined />, label: '收件箱' },
            { key: 'logs', icon: <FileTextOutlined />, label: '登录日志' },
          ]}
        />
      </div>

      {/* 右侧内容 */}
      <div style={{ flex: 1, minWidth: 0 }}>
        {activeKey === 'profile' && (
          <PageCard title="基本资料">
            <div style={{ display: 'flex', gap: 24, flexWrap: 'wrap' }}>
              {/* 头像区域 */}
              <div style={{ textAlign: 'center' }}>
                <div
                  style={{
                    width: 100,
                    height: 100,
                    borderRadius: '50%',
                    background: themeToken.colorFillTertiary,
                    margin: '0 auto 12px',
                    overflow: 'hidden',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  {avatarUrl ? (
                    <img src={avatarUrl} alt="avatar" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                  ) : (
                    <UserOutlined style={{ fontSize: 40, color: themeToken.colorTextTertiary }} />
                  )}
                </div>
                <Upload {...uploadProps}>
                  <Button icon={<UploadOutlined />} size="small">
                    上传头像
                  </Button>
                </Upload>
              </div>

              {/* 基本信息表单 */}
              <div style={{ flex: 1, minWidth: 280 }}>
                <Descriptions size="small" column={1} style={{ marginBottom: 16 }}>
                  <Descriptions.Item label="用户名">
                    {((userInfo as Record<string, unknown>)?.username as string) || '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="角色">
                    {(() => {
                      const roles = ((userInfo as Record<string, unknown>)?.roles as { name?: string }[]) || [];
                      if (!roles.length) return '-';
                      return roles.map((r) => r.name).join(' / ');
                    })()}
                  </Descriptions.Item>
                </Descriptions>
                <Form form={profileForm} layout="vertical">
                  <Form.Item name="nickname" label="昵称">
                    <Input prefix={<UserOutlined />} placeholder="设置显示昵称" />
                  </Form.Item>
                  <Form.Item name="email" label="邮箱">
                    <Input prefix={<MailOutlined />} placeholder="设置邮箱" />
                  </Form.Item>
                  <Form.Item>
                    <Button type="primary" onClick={handleProfileSave}>保存</Button>
                  </Form.Item>
                </Form>
              </div>
            </div>
          </PageCard>
        )}

        {activeKey === 'password' && (
          <PageCard title="修改密码">
            <Form form={pwdForm} layout="vertical" style={{ maxWidth: 400 }}>
              <Form.Item name="old_password" label="旧密码" rules={[{ required: true, message: '请输入旧密码' }]}>
                <Input.Password prefix={<LockOutlined />} />
              </Form.Item>
              <Form.Item name="new_password" label="新密码" rules={[{ required: true, message: '请输入新密码' }, { min: 8, message: '密码至少8位' }]}>
                <Input.Password prefix={<LockOutlined />} />
              </Form.Item>
              <Form.Item name="confirm_password" label="确认新密码" rules={[{ required: true, message: '请再次输入新密码' }]}>
                <Input.Password prefix={<LockOutlined />} />
              </Form.Item>
              <Form.Item>
                <Button type="primary" danger onClick={handlePasswordChange}>修改密码</Button>
              </Form.Item>
            </Form>
          </PageCard>
        )}

        {activeKey === 'inbox' && (
          msgDetailId ? (
            <PageCard
              title="消息详情"
                            extra={<Button size="small" onClick={() => navigate('/profile?tab=inbox')}>返回列表</Button>}
            >
              {msgDetailLoading ? <Spin /> : msgDetail ? (
                <>
                  <h2 style={{ fontSize: 20, fontWeight: 600, marginBottom: 12 }}>{msgDetail.title}</h2>
                  <Space size={16} style={{ marginBottom: 20, color: themeToken.colorTextTertiary, fontSize: 13 }}>
                    <span>{msgDetail.sender_type === 'system' ? <Tag color="blue">系统</Tag> : <Tag>管理员</Tag>}</span>
                    <span>{formatTime(msgDetail.created_at)}</span>
                  </Space>
                  <div style={{ whiteSpace: 'pre-wrap', lineHeight: 1.8, color: themeToken.colorText }}>
                    {msgDetail.content || '（无内容）'}
                  </div>
                </>
              ) : (
                <Empty description="消息不存在或已被删除" />
              )}
            </PageCard>
          ) : (
            <PageCard
              title="收件箱"
                            extra={
                selectedMsgKeys.length > 0 ? (
                  <Dropdown
                    menu={{
                      items: [
                        { key: 'read', label: '标记已读' },
                        { key: 'delete', label: '批量删除', danger: true },
                      ],
                      onClick: ({ key }) => void handleInboxBatch(key),
                    }}
                  >
                    <Button size="small">批量操作（{selectedMsgKeys.length}）<DownOutlined /></Button>
                  </Dropdown>
                ) : null
              }
            >
              <Table<MessageRow>
                columns={inboxColumns}
                rowKey="id"
                dataSource={msgs}
                loading={msgLoading}
                rowSelection={{
                  selectedRowKeys: selectedMsgKeys,
                  onChange: setSelectedMsgKeys,
                }}
                pagination={{
                  current: msgPage,
                  pageSize: 10,
                  total: msgTotal,
                  showSizeChanger: false,
                  onChange: (p) => setMsgPage(p),
                }}
                size="middle"
              />
            </PageCard>
          )
        )}

        {activeKey === 'logs' && (
          <PageCard title="最近登录记录">
            <Table<LoginLog>
              columns={loginLogColumns}
              rowKey="id"
              dataSource={logs}
              loading={logLoading}
              pagination={{
                current: logPage,
                pageSize: 10,
                total: logTotal,
                showSizeChanger: false,
                onChange: (p) => setLogPage(p),
              }}
              size="middle"
            />
          </PageCard>
        )}
      </div>
    </div>
  );
};

export default Profile;
