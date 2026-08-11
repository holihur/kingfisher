import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Form, Input, Button, Alert } from 'antd';
import { MailOutlined, ArrowLeftOutlined, SafetyCertificateOutlined } from '@ant-design/icons';
import { authApi } from '../../api/auth';
import { configApi } from '../../api/config';
import { useThemeToken } from '../../hooks/useThemeToken';

const ForgotPassword: React.FC = () => {
  const token = useThemeToken();
  const [loading, setLoading] = useState(false);
  const [sent, setSent] = useState(false);
  const [siteName, setSiteName] = useState('Kingfisher');
  const [siteLogo, setSiteLogo] = useState('');
  const [enabled, setEnabled] = useState(true);

  useEffect(() => {
    configApi.getPublic('site_name').then(r => setSiteName((r.data as any)?.value || 'Kingfisher')).catch(() => {});
    configApi.getPublic('site_logo').then(r => setSiteLogo((r.data as any)?.value || '')).catch(() => {});
    configApi.getPublic('password_reset_enabled').then(r => setEnabled((r.data as any)?.value !== 'false')).catch(() => {});
  }, []);

  const onFinish = async (values: { email: string }) => {
    setLoading(true);
    try {
      await authApi.forgotPassword(values.email);
      setSent(true);
    } catch {
      /* interceptor handles */
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ minHeight: '100vh', display: 'flex', background: token.colorBgLayout }}>
      {/* 左品牌区（与登录页一致） */}
      <div style={{
        flex: 1,
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'center',
        padding: 80,
        background: `linear-gradient(135deg, ${token.colorPrimary} 0%, ${token.colorPrimaryHover} 100%)`,
        position: 'relative',
        overflow: 'hidden',
      }}>
        <div style={{ maxWidth: 380, marginLeft: 'auto', marginRight: 80, position: 'relative' }}>
          {siteLogo ? (
            <img src={siteLogo} alt="logo" style={{ height: 48, marginBottom: 20, borderRadius: 6, filter: 'drop-shadow(0 4px 12px rgba(0,0,0,0.2))' }} />
          ) : (
            <div style={{
              width: 56, height: 56, borderRadius: 14, marginBottom: 20,
              background: 'rgba(255,255,255,0.16)', backdropFilter: 'blur(8px)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              fontSize: 26, fontWeight: 700, color: '#fff',
              boxShadow: '0 6px 16px rgba(0,0,0,0.2)',
            }}>
              {(siteName || 'K').charAt(0)}
            </div>
          )}
          <h1 style={{ fontSize: 30, fontWeight: 700, margin: 0, lineHeight: 1.25, letterSpacing: '-0.5px', color: '#fff' }}>{siteName || 'Kingfisher'}</h1>
          <p style={{ marginTop: 14, color: 'rgba(255,255,255,0.85)', fontSize: 15, lineHeight: 1.7 }}>
            找回密码：输入注册邮箱，我们将发送密码重置链接。
          </p>
        </div>
      </div>

      {/* 右表单区 */}
      <div style={{ width: 440, display: 'flex', flexDirection: 'column', justifyContent: 'center', padding: 60, background: token.colorBgContainer }}>
        <div style={{ maxWidth: 340, width: '100%' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
            <SafetyCertificateOutlined style={{ fontSize: 22, color: token.colorPrimary }} />
            <h2 style={{ fontSize: 20, fontWeight: 600, margin: 0 }}>找回密码</h2>
          </div>
          <p style={{ color: token.colorTextTertiary, fontSize: 13, marginBottom: 24 }}>输入注册邮箱，我们将发送重置链接（30 分钟内有效）。</p>

          {!enabled ? (
            <Alert
              type="warning"
              showIcon
              message="找回密码功能未开启"
              description="管理员暂未开启找回密码功能，请联系管理员处理。"
              style={{ marginBottom: 20 }}
            />
          ) : sent ? (
            <Alert
              type="success"
              showIcon
              message="重置邮件已发送"
              description="如果该邮箱已注册，你会收到一封包含重置链接的邮件。请前往邮箱操作。"
              style={{ marginBottom: 20 }}
            />
          ) : (
            <Form onFinish={onFinish} size="large">
              <Form.Item name="email" rules={[{ required: true, message: '请输入邮箱' }, { type: 'email', message: '邮箱格式不正确' }]}>
                <Input prefix={<MailOutlined />} placeholder="注册邮箱" />
              </Form.Item>
              <Form.Item style={{ marginBottom: 0 }}>
                <Button type="primary" htmlType="submit" loading={loading} block>
                  发送重置邮件
                </Button>
              </Form.Item>
            </Form>
          )}

          <div style={{ marginTop: 24, textAlign: 'center', fontSize: 13, color: token.colorTextTertiary }}>
            <Link to="/login" style={{ color: token.colorPrimary }}><ArrowLeftOutlined /> 返回登录</Link>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ForgotPassword;
