import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Form, Input, Button, Card, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useAuthStore } from '../../stores/auth';

const LoginPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [siteName, setSiteName] = useState('Kingfisher');
  const [siteDescription, setSiteDescription] = useState('后台管理系统');
  const login = useAuthStore((s) => s.login);
  const navigate = useNavigate();

  const [searchParams] = useSearchParams();

  useEffect(() => {
    // site_name / site_description 标记为公开，未登录也能通过公开接口读取；失败时兜底默认值
    import('../../api/config').then((m) => {
      m.configApi.getPublic('site_name').then(r => setSiteName((r.data as any)?.value || 'Kingfisher')).catch(() => {});
      m.configApi.getPublic('site_description').then(r => setSiteDescription((r.data as any)?.value || '后台管理系统')).catch(() => {});
    });
  }, []);
  const redirectTo = searchParams.get('redirect');

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      await login(values.username, values.password);
      message.success('登录成功');
      // 跳转优先级：redirect 参数（用户原本要去的页）> 角色落地页 > 默认 /dashboard
      // 用 getState() 取最新 landing_page，避免闭包捕获旧值
      const landing = useAuthStore.getState().landingPage;
      navigate(redirectTo || landing || '/dashboard');
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
          <div style={{ fontSize: 40, marginBottom: 8 }}>🦜</div>
          <h2 style={{ margin: 0 }}>{siteName}</h2>
          <p style={{ color: '#999', margin: '8px 0 0' }}>{siteDescription}</p>
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
