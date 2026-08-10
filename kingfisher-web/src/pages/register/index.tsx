import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Form, Input, Button, App, Alert } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined } from '@ant-design/icons';
import { authApi } from '../../api/auth';
import { configApi } from '../../api/config';
import { useThemeToken } from '../../hooks/useThemeToken';

const RegisterPage: React.FC = () => {
  const { message } = App.useApp();
  const token = useThemeToken();
  const [loading, setLoading] = useState(false);
  const [registrationEnabled, setRegistrationEnabled] = useState<boolean | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    configApi
      .getPublic('registration_enabled')
      .then((r) => setRegistrationEnabled((r.data as { value?: string })?.value !== 'false'))
      .catch(() => setRegistrationEnabled(true));
  }, []);

  const onFinish = async (values: { username: string; password: string; email?: string }) => {
    setLoading(true);
    try {
      await authApi.register(values);
      message.success('注册成功，请登录');
      navigate('/login');
    } catch {
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ minHeight: '100vh', display: 'flex' }}>
      <div style={{ flex: 1, background: token.colorBgLayout, display: 'flex', flexDirection: 'column', justifyContent: 'center', padding: 80 }}>
        <div style={{ maxWidth: 360, marginLeft: 'auto', marginRight: 80 }}>
          <h1 style={{ fontSize: 28, fontWeight: 600, marginBottom: 12 }}>注册账号</h1>
          <p style={{ color: token.colorTextSecondary, fontSize: 14, marginBottom: 0 }}>创建新账号以使用系统服务</p>
        </div>
      </div>

      <div style={{ width: 440, display: 'flex', flexDirection: 'column', justifyContent: 'center', padding: 60, background: token.colorBgContainer }}>
        <div style={{ maxWidth: 340 }}>
          <h2 style={{ fontSize: 18, fontWeight: 500, marginBottom: 28 }}>注册</h2>

          {registrationEnabled === false && (
            <Alert type="warning" showIcon message="当前未开放注册" description="请联系管理员开通账号。" style={{ marginBottom: 20 }} />
          )}

          <Form onFinish={onFinish} size="large" disabled={registrationEnabled === false}>
            <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }, { min: 3 }, { max: 32 }]}>
              <Input prefix={<UserOutlined />} placeholder="用户名" />
            </Form.Item>
            <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }, { min: 8, message: '至少 8 位' }]}>
              <Input.Password prefix={<LockOutlined />} placeholder="密码" />
            </Form.Item>
            <Form.Item
              name="confirm_password"
              dependencies={['password']}
              rules={[{ required: true, message: '请确认密码' }, ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) return Promise.resolve();
                  return Promise.reject(new Error('两次输入的密码不一致'));
                },
              })]}
            >
              <Input.Password prefix={<LockOutlined />} placeholder="确认密码" />
            </Form.Item>
            <Form.Item name="email" rules={[{ type: 'email', message: '邮箱格式不正确' }]}>
              <Input prefix={<MailOutlined />} placeholder="邮箱（选填）" />
            </Form.Item>
            <Form.Item style={{ marginBottom: 0 }}>
              <Button type="primary" htmlType="submit" loading={loading} block disabled={registrationEnabled === false}>注册</Button>
            </Form.Item>
          </Form>

          <div style={{ marginTop: 24, textAlign: 'center', fontSize: 13, color: token.colorTextTertiary }}>
            <a href="/login">返回登录</a>
          </div>
        </div>
      </div>
    </div>
  );
};

export default RegisterPage;
