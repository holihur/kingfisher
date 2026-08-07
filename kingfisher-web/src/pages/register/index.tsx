import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Form, Input, Button, Card, App, Alert } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined } from '@ant-design/icons';
import { authApi } from '../../api/auth';
import { configApi } from '../../api/config';

const RegisterPage: React.FC = () => {
  const { message } = App.useApp();
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
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f0f2f5' }}>
      <Card style={{ width: 400, boxShadow: '0 2px 8px rgba(0,0,0,0.15)' }} variant="borderless">
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <h2 style={{ margin: 0 }}>注册账号</h2>
          <p style={{ color: '#999', margin: '8px 0 0' }}>创建新账号以使用系统</p>
        </div>

        {registrationEnabled === false && (
          <Alert type="warning" showIcon message="当前未开放注册" description="请联系管理员开通账号。" style={{ marginBottom: 16 }} />
        )}

        <Form onFinish={onFinish} size="large" disabled={registrationEnabled === false}>
          <Form.Item name="username" rules={[
            { required: true, message: '请输入用户名' },
            { min: 3, message: '至少 3 个字符' },
            { max: 32, message: '最多 32 个字符' },
          ]}>
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>
          <Form.Item name="password" rules={[
            { required: true, message: '请输入密码' },
            { min: 8, message: '至少 8 位' },
          ]}>
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          <Form.Item
            name="confirm_password"
            dependencies={['password']}
            rules={[
              { required: true, message: '请确认密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) return Promise.resolve();
                  return Promise.reject(new Error('两次输入的密码不一致'));
                },
              }),
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="确认密码" />
          </Form.Item>
          <Form.Item name="email" rules={[{ type: 'email', message: '邮箱格式不正确' }]}>
            <Input prefix={<MailOutlined />} placeholder="邮箱（选填）" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block disabled={registrationEnabled === false}>注 册</Button>
          </Form.Item>
        </Form>

        <div style={{ textAlign: 'center', marginTop: 16 }}>已有账号？<a href="/login">去登录</a></div>
      </Card>
    </div>
  );
};

export default RegisterPage;
