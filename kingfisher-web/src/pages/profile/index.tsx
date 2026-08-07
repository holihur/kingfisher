import React, { useCallback, useEffect, useState } from 'react';
import { Card, Form, Input, Button, App, Descriptions, Tag, Menu, Upload, Table } from 'antd';
import type { TableProps, UploadProps } from 'antd';
import { UserOutlined, MailOutlined, LockOutlined, UploadOutlined, SolutionOutlined, KeyOutlined, FileTextOutlined } from '@ant-design/icons';
import { userApi } from '../../api/user';
import { useAuthStore } from '../../stores/auth';

interface LoginLog {
  id: number;
  username: string;
  ip: string;
  user_agent: string;
  created_at: string;
}

const Profile: React.FC = () => {
  const { message } = App.useApp();
  const { userInfo, fetchUserInfo, token } = useAuthStore();
  const [profileForm] = Form.useForm();
  const [pwdForm] = Form.useForm();
  const [avatarUrl, setAvatarUrl] = useState<string>('');
  // 左侧菜单当前选中项
  const [activeKey, setActiveKey] = useState('profile');
  // 登录日志本地分页
  const [logs, setLogs] = useState<LoginLog[]>([]);
  const [logTotal, setLogTotal] = useState(0);
  const [logPage, setLogPage] = useState(1);
  const [logLoading, setLogLoading] = useState(false);

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
      <div style={{ width: 200, flexShrink: 0, background: '#fff', borderRadius: 8, padding: 8, boxShadow: '0 1px 2px rgba(0,0,0,0.04)' }}>
        <Menu
          mode="inline"
          selectedKeys={[activeKey]}
          onClick={({ key }) => setActiveKey(key)}
          style={{ border: 'none' }}
          items={[
            { key: 'profile', icon: <SolutionOutlined />, label: '用户资料' },
            { key: 'password', icon: <KeyOutlined />, label: '修改密码' },
            { key: 'logs', icon: <FileTextOutlined />, label: '登录日志' },
          ]}
        />
      </div>

      {/* 右侧内容 */}
      <div style={{ flex: 1, minWidth: 0 }}>
        {activeKey === 'profile' && (
          <Card title="基本资料" style={{ borderRadius: 8, border: 'none', boxShadow: '0 1px 2px rgba(0,0,0,0.04)' }}>
            <div style={{ display: 'flex', gap: 24, flexWrap: 'wrap' }}>
              {/* 头像区域 */}
              <div style={{ textAlign: 'center' }}>
                <div
                  style={{
                    width: 100,
                    height: 100,
                    borderRadius: '50%',
                    background: '#f0f0f0',
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
                    <UserOutlined style={{ fontSize: 40, color: '#bfbfbf' }} />
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
                    {((userInfo as Record<string, unknown>)?.role as Record<string, unknown>)?.name as string || '-'}
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
          </Card>
        )}

        {activeKey === 'password' && (
          <Card title="修改密码" style={{ borderRadius: 8, border: 'none', boxShadow: '0 1px 2px rgba(0,0,0,0.04)' }}>
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
          </Card>
        )}

        {activeKey === 'logs' && (
          <Card title="最近登录记录" style={{ borderRadius: 8, border: 'none', boxShadow: '0 1px 2px rgba(0,0,0,0.04)' }}>
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
          </Card>
        )}
      </div>
    </div>
  );
};

export default Profile;
