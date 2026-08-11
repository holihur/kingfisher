import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams, Link } from 'react-router-dom';
import { Form, Input, Button, App, Alert, Card } from 'antd';
import { LockOutlined, ArrowLeftOutlined, KeyOutlined } from '@ant-design/icons';
import { authApi } from '../../api/auth';
import { configApi } from '../../api/config';
import { useThemeToken } from '../../hooks/useThemeToken';

const ResetPassword: React.FC = () => {
  const { message } = App.useApp();
  const token = useThemeToken();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const resetToken = searchParams.get('token') || '';
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);
  const [siteName, setSiteName] = useState('Kingfisher');

  useEffect(() => {
    configApi
      .getPublic('site_name')
      .then((r) => setSiteName((r.data as any)?.value || 'Kingfisher'))
      .catch(() => {});
  }, []);

  const onFinish = async (values: { new_password: string }) => {
    setLoading(true);
    try {
      await authApi.resetPassword(resetToken, values.new_password);
      message.success('密码已重置，请用新密码登录');
      setDone(true);
    } catch {
      /* interceptor handles */
    } finally {
      setLoading(false);
    }
  };

  if (!resetToken) {
    return (
      <div
        style={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: token.colorBgLayout,
        }}
      >
        <Card style={{ borderRadius: token.borderRadiusLG, boxShadow: token.boxShadowTertiary }}>
          <Alert
            type="error"
            showIcon
            title="无效的链接"
            description="重置链接缺少 token 参数，请检查邮件中的完整链接。"
          />
        </Card>
      </div>
    );
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', background: token.colorBgLayout }}>
      {/* 左品牌区 */}
      <div
        style={{
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'center',
          padding: 80,
          background: `linear-gradient(135deg, ${token.colorPrimary} 0%, ${token.colorPrimaryHover} 100%)`,
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        <div style={{ maxWidth: 380, marginLeft: 'auto', marginRight: 80, position: 'relative' }}>
          <div
            style={{
              width: 56,
              height: 56,
              borderRadius: 14,
              marginBottom: 20,
              background: 'rgba(255,255,255,0.16)',
              backdropFilter: 'blur(8px)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: 26,
              fontWeight: 700,
              color: '#fff',
              boxShadow: '0 6px 16px rgba(0,0,0,0.2)',
            }}
          >
            {(siteName || 'K').charAt(0)}
          </div>
          <h1
            style={{
              fontSize: 30,
              fontWeight: 700,
              margin: 0,
              lineHeight: 1.25,
              letterSpacing: '-0.5px',
              color: '#fff',
            }}
          >
            {siteName || 'Kingfisher'}
          </h1>
          <p style={{ marginTop: 14, color: 'rgba(255,255,255,0.85)', fontSize: 15, lineHeight: 1.7 }}>
            设置新密码后，旧会话将失效，需重新登录。
          </p>
        </div>
      </div>

      {/* 右表单区 */}
      <div
        style={{
          width: 440,
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'center',
          padding: 60,
          background: token.colorBgContainer,
        }}
      >
        <div style={{ maxWidth: 340, width: '100%' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
            <KeyOutlined style={{ fontSize: 22, color: token.colorPrimary }} />
            <h2 style={{ fontSize: 20, fontWeight: 600, margin: 0 }}>重置密码</h2>
          </div>
          <p style={{ color: token.colorTextTertiary, fontSize: 13, marginBottom: 24 }}>为你的账号设置一个新密码。</p>

          {done ? (
            <Alert
              type="success"
              showIcon
              title="密码已重置"
              description="现在可以用新密码登录了。"
              action={
                <Button type="link" onClick={() => navigate('/login')}>
                  去登录
                </Button>
              }
              style={{ marginBottom: 20 }}
            />
          ) : (
            <Form onFinish={onFinish} size="large">
              <Form.Item
                name="new_password"
                rules={[
                  { required: true, message: '请输入新密码' },
                  { min: 8, message: '至少 8 位' },
                  { max: 64, message: '最多 64 位' },
                ]}
              >
                <Input.Password prefix={<LockOutlined />} placeholder="新密码（至少 8 位）" />
              </Form.Item>
              <Form.Item
                name="confirm_password"
                dependencies={['new_password']}
                rules={[
                  { required: true, message: '请确认新密码' },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('new_password') === value) return Promise.resolve();
                      return Promise.reject(new Error('两次输入的密码不一致'));
                    },
                  }),
                ]}
              >
                <Input.Password prefix={<LockOutlined />} placeholder="确认新密码" />
              </Form.Item>
              <Form.Item style={{ marginBottom: 0 }}>
                <Button type="primary" htmlType="submit" loading={loading} block>
                  确认重置
                </Button>
              </Form.Item>
            </Form>
          )}

          <div style={{ marginTop: 24, textAlign: 'center', fontSize: 13, color: token.colorTextTertiary }}>
            <Link to="/login" style={{ color: token.colorPrimary }}>
              <ArrowLeftOutlined /> 返回登录
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ResetPassword;
