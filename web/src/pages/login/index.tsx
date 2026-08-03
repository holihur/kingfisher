import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Form, Input, Button, Card, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useAuthStore } from '../../stores/auth';

const LoginPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [siteName, setSiteName] = useState('Kingfisher');
  const login = useAuthStore((s) => s.login);
  const navigate = useNavigate();

  const [searchParams] = useSearchParams();

  useEffect(() => {
    import('../../api/config').then(m => m.configApi.get('site_name').then(r => setSiteName((r.data as any)?.value || 'Kingfisher')).catch(() => {}));
  }, []);
  const redirectTo = searchParams.get('redirect') || '/dashboard';

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      await login(values.username, values.password);
      message.success('登录成功');
      navigate(redirectTo);
    } catch {
      /* error handled by interceptor */
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        background: '#f0f2f5',
      }}
    >
      <Card style={{ width: 400, boxShadow: '0 2px 8px rgba(0,0,0,0.15)' }} bordered={false}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <h2 style={{ margin: 0 }}>{siteName}</h2>
          <p style={{ color: '#999', margin: '8px 0 0' }}>后台管理系统</p>
        </div>
        <Form onFinish={onFinish} size="large">
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>
              登 录
            </Button>
          </Form.Item>
        </Form>
      <div style={{ textAlign: "center", marginTop: 16 }}>还没有账号？<a href="/register">去注册</a></div>
      </Card>
    </div>
  );
};

export default LoginPage;
