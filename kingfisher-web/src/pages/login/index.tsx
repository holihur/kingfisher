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
  const [siteLoginCover, setSiteLoginCover] = useState('');
  const [rememberAccount, setRememberAccount] = useState(true);
  const [rememberPwd, setRememberPwd] = useState(false);
  const [accounts, setAccounts] = useState<SavedAccount[]>(() => loadAccounts());
  const [view, setView] = useState<'accounts' | 'form'>(() => (loadAccounts().length > 0 ? 'accounts' : 'form'));
  const login = useAuthStore((s) => s.login);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  useEffect(() => {
    import('../../api/config').then((m) => {
      m.configApi.getPublic('site_name').then(r => setSiteName((r.data as any)?.value || 'Kingfisher')).catch(() => {});
      m.configApi.getPublic('site_logo').then(r => setSiteLogo((r.data as any)?.value || '')).catch(() => {});
      m.configApi.getPublic('site_description').then(r => setSiteDescription((r.data as any)?.value || '')).catch(() => {});
      m.configApi.getPublic('site_login_cover').then(r => setSiteLoginCover((r.data as any)?.value || '')).catch(() => {});
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
    setView('form');
  };

  const delAccount = (username: string) => {
    removeAccount(username);
    const rest = loadAccounts();
    setAccounts(rest);
    if (rest.length === 0) setView('form');
  };

  const accountOptions = accounts.map(a => ({
    value: a.username,
    label: (
      <span style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <span><Avatar size={22} style={{ marginRight: 8, background: '#1677ff' }}>{a.username?.charAt(0)?.toUpperCase()}</Avatar>{a.username}</span>
        <CloseOutlined style={{ color: '#bbb', fontSize: 10 }} onMouseDown={e => { e.stopPropagation(); e.preventDefault(); delAccount(a.username); }} />
      </span>
    ),
  }));

  return (
    <div style={{ minHeight: '100vh', display: 'flex' }}>
      <div className="login-brand-col" style={{
        flex: 1,
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'center',
        padding: 80,
        background: siteLoginCover ? `url(${siteLoginCover}) center/cover no-repeat` : '#fafafa',
        position: 'relative',
      }}>
        {siteLoginCover ? <div style={{ position: 'absolute', inset: 0, background: 'rgba(0,0,0,0.45)' }} /> : null}
        <div className="login-brand" style={{ maxWidth: 360, marginLeft: 'auto', marginRight: 80, position: 'relative', color: siteLoginCover ? '#fff' : 'inherit' }}>
          {siteLogo ? (
            <img src={siteLogo} alt="logo" style={{ height: 48, marginBottom: 20, borderRadius: 6 }} />
          ) : (
            <div style={{
              width: 56, height: 56, borderRadius: 14, marginBottom: 20,
              background: siteLoginCover ? 'rgba(255,255,255,0.16)' : 'linear-gradient(135deg, #1677ff 0%, #69b1ff 100%)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              fontSize: 26, fontWeight: 700, color: '#fff',
              boxShadow: siteLoginCover ? 'none' : '0 6px 16px rgba(22,119,255,0.28)',
            }}>
              {(siteName || 'K').charAt(0)}
            </div>
          )}
          <h1 style={{ fontSize: 30, fontWeight: 700, margin: 0, lineHeight: 1.25, letterSpacing: '-0.5px' }}>{siteName || 'Kingfisher'}</h1>
          {siteDescription && <p style={{ marginTop: 14, color: siteLoginCover ? 'rgba(255,255,255,0.85)' : '#8c8c8c', fontSize: 15, lineHeight: 1.7 }}>{siteDescription}</p>}
        </div>
      </div>

      <div className="login-form" style={{ width: 440, display: 'flex', flexDirection: 'column', justifyContent: 'center', padding: 60, background: '#fff' }}>
        <div style={{ maxWidth: 340 }}>
          {view === 'accounts' && accounts.length > 0 ? (
            <>
              <h2 style={{ fontSize: 18, fontWeight: 500, marginBottom: 24 }}>选择账户</h2>
              <div style={{ marginBottom: 16 }}>
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
                      <Avatar size={36} style={{ background: '#1677ff' }}>{a.username?.charAt(0)?.toUpperCase()}</Avatar>
                      <div>
                        <div style={{ fontWeight: 500, fontSize: 14 }}>{a.username}</div>
                        <div style={{ fontSize: 11, color: '#999' }}>{a.password ? '已保存密码' : '点击后输入密码'}</div>
                      </div>
                    </span>
                    <CloseOutlined style={{ color: '#ccc', fontSize: 12 }} onClick={e => { e.stopPropagation(); delAccount(a.username); }} />
                  </div>
                ))}
              </div>
              <Button block onClick={() => setView('form')}>使用其他账号</Button>
              <div style={{ marginTop: 24, textAlign: 'center', fontSize: 13, color: '#999' }}>
                <a href="/register">注册账号</a>
              </div>
            </>
          ) : (
            <>
              <h2 style={{ fontSize: 18, fontWeight: 500, marginBottom: 28 }}>登录</h2>
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
                {accounts.length > 0 && (
                  <>
                    <a style={{ cursor: 'pointer', color: '#1677ff', marginRight: 16 }} onClick={() => setView('accounts')}>切换账户</a>
                  </>
                )}
                <a href="/register">注册账号</a>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
