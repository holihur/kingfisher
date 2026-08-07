import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Form, Input, Button, App, AutoComplete, Checkbox, Avatar } from 'antd';
import { UserOutlined, LockOutlined, CloseOutlined } from '@ant-design/icons';
import { useAuthStore } from '../../stores/auth';
import { loadAccounts, saveAccount, getAccountPassword, removeAccount } from '../../utils/remember';
import type { SavedAccount } from '../../utils/remember';

const LoginPage: React.FC = () => {
  const { message } = App.useApp();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [siteName, setSiteName] = useState('');
  const [siteLogo, setSiteLogo] = useState('');
  const [siteDescription, setSiteDescription] = useState('');
  const [rememberAccount, setRememberAccount] = useState(true);
  const [rememberPwd, setRememberPwd] = useState(false);
  const [accounts, setAccounts] = useState<SavedAccount[]>(() => loadAccounts());
  const login = useAuthStore((s) => s.login);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  useEffect(() => {
    import('../../api/config').then((m) => {
      m.configApi.getPublic('site_name').then(r => setSiteName((r.data as any)?.value || 'Kingfisher')).catch(() => {});
      m.configApi.getPublic('site_logo').then(r => setSiteLogo((r.data as any)?.value || '')).catch(() => {});
      m.configApi.getPublic('site_description').then(r => setSiteDescription((r.data as any)?.value || '')).catch(() => {});
    });
  }, []);

  const redirectTo = searchParams.get('redirect');

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      await login(values.username, values.password);
      message.success('登录成功');
      saveAccount(values.username, values.password, rememberAccount, rememberPwd);
      navigate(redirectTo || useAuthStore.getState().landingPage || '/dashboard');
    } catch {
    } finally {
      setLoading(false);
    }
  };

  const selectAccount = (username: string) => {
    form.setFieldsValue({ username });
    const pwd = getAccountPassword(username);
    if (pwd) form.setFieldsValue({ password: pwd });
  };

  const delAccount = (username: string) => {
    removeAccount(username);
    setAccounts(loadAccounts());
  };

  const accountOptions = accounts.map(a => ({
    value: a.username,
    label: (
      <span style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <span><Avatar size={22} icon={<UserOutlined />} style={{ marginRight: 8 }} />{a.username}</span>
        <CloseOutlined style={{ color: '#bbb', fontSize: 10 }} onMouseDown={e => { e.stopPropagation(); e.preventDefault(); delAccount(a.username); }} />
      </span>
    ),
  }));

  return (
    <div style={{ minHeight: '100vh', display: 'flex' }}>
      <div style={{ flex: 1, background: '#fafafa', display: 'flex', flexDirection: 'column', justifyContent: 'center', padding: 80 }}>
        <div style={{ maxWidth: 360, marginLeft: 'auto', marginRight: 80 }}>
          {siteLogo ? (
            <img src={siteLogo} alt="logo" style={{ height: 48, marginBottom: 16, borderRadius: 6 }} />
          ) : (
            <div style={{
              width: 48, height: 48, borderRadius: 8, marginBottom: 16,
              background: '#f0f0f0', display: 'flex', alignItems: 'center', justifyContent: 'center',
              fontSize: 22, fontWeight: 600, color: '#666',
            }}>
              {(siteName || 'K').charAt(0)}
            </div>
          )}
          <h1 style={{ fontSize: 28, fontWeight: 600, marginBottom: 12 }}>{siteName || 'Kingfisher'}</h1>
          {siteDescription && <p style={{ color: '#666', fontSize: 14 }}>{siteDescription}</p>}
        </div>
      </div>

      <div style={{ width: 440, display: 'flex', flexDirection: 'column', justifyContent: 'center', padding: 60, background: '#fff' }}>
        <div style={{ maxWidth: 340 }}>
          <h2 style={{ fontSize: 18, fontWeight: 500, marginBottom: 28 }}>登录</h2>

          {accounts.length > 0 && (
            <div style={{ marginBottom: 20 }}>
              {accounts.slice(0, 3).map(a => (
                <div
                  key={a.username}
                  onClick={() => selectAccount(a.username)}
                  style={{
                    display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                    padding: '10px 12px', marginBottom: 8, borderRadius: 6,
                    border: '1px solid #eee', cursor: 'pointer', background: '#fafafa',
                  }}
                  onMouseEnter={e => ((e.currentTarget as HTMLElement).style.background = '#f5f5f5')}
                  onMouseLeave={e => ((e.currentTarget as HTMLElement).style.background = '#fafafa')}
                >
                  <span style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                    <Avatar size={36} icon={<UserOutlined />} />
                    <div>
                      <div style={{ fontWeight: 500, fontSize: 14 }}>{a.username}</div>
                      <div style={{ fontSize: 11, color: '#999' }}>{a.password ? '已保存密码' : '点击后输入密码'}</div>
                    </div>
                  </span>
                  <CloseOutlined style={{ color: '#ccc', fontSize: 12 }} onClick={e => { e.stopPropagation(); delAccount(a.username); }} />
                </div>
              ))}
            </div>
          )}

          <Form form={form} onFinish={onFinish} size="large">
            <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
              <AutoComplete options={accountOptions} style={{ width: '100%' }}>
                <Input prefix={<UserOutlined />} placeholder="用户名" />
              </AutoComplete>
            </Form.Item>
            <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
              <Input.Password prefix={<LockOutlined />} placeholder="密码" />
            </Form.Item>
            <Form.Item style={{ marginBottom: 24 }}>
              <Checkbox checked={rememberAccount} onChange={e => { setRememberAccount(e.target.checked); if (!e.target.checked) setRememberPwd(false); }}>记住账户</Checkbox>
              <Checkbox checked={rememberPwd} disabled={!rememberAccount} onChange={e => setRememberPwd(e.target.checked)} style={{ marginLeft: 16 }}>记住密码</Checkbox>
            </Form.Item>
            <Form.Item style={{ marginBottom: 0 }}>
              <Button type="primary" htmlType="submit" loading={loading} block>登录</Button>
            </Form.Item>
          </Form>

          <div style={{ marginTop: 24, textAlign: 'center', fontSize: 13, color: '#999' }}>
            <a href="/register">注册账号</a>
          </div>
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
