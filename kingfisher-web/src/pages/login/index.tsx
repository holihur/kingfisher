import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Form, Input, Button, App, AutoComplete, Checkbox, Tag } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useAuthStore } from '../../stores/auth';
import { loadAccounts, saveAccount, getAccountPassword, removeAccount } from '../../utils/remember';
import type { SavedAccount } from '../../utils/remember';

const LoginPage: React.FC = () => {
  const { message } = App.useApp();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [siteName, setSiteName] = useState('');
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

  const options = useMemo(() => accounts.map(a => ({
    value: a.username,
    label: (
      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
        <span><UserOutlined style={{ marginRight: 8 }} />{a.username}</span>
        <a onMouseDown={e => { e.stopPropagation(); e.preventDefault(); removeAccount(a.username); setAccounts(loadAccounts()); }}>删除</a>
      </div>
    ),
  })), [accounts]);

  return (
    <div style={{ minHeight: '100vh', display: 'flex' }}>
      {/* left panel */}
      <div style={{ flex: 1, background: '#fafafa', display: 'flex', flexDirection: 'column', justifyContent: 'center', padding: 80 }}>
        <div style={{ maxWidth: 360, marginLeft: 'auto', marginRight: 80 }}>
          <h1 style={{ fontSize: 28, fontWeight: 600, marginBottom: 12 }}>{siteName || 'Kingfisher'}</h1>
          {siteDescription && <p style={{ color: '#666', fontSize: 14 }}>{siteDescription}</p>}
        </div>
      </div>

      {/* right panel */}
      <div style={{ width: 440, display: 'flex', flexDirection: 'column', justifyContent: 'center', padding: 60, background: '#fff' }}>
        <div style={{ maxWidth: 340 }}>
          <h2 style={{ fontSize: 18, fontWeight: 500, marginBottom: 28 }}>登录</h2>

          {accounts.length > 0 && (
            <div style={{ marginBottom: 16 }}>
              {accounts.slice(0, 5).map(a => (
                <Tag
                  key={a.username}
                  closable
                  onClose={e => { e.preventDefault(); removeAccount(a.username); setAccounts(loadAccounts()); }}
                  onClick={() => { form.setFieldsValue({ username: a.username }); const pwd = getAccountPassword(a.username); if (pwd) form.setFieldsValue({ password: pwd }); }}
                  style={{ cursor: 'pointer', marginBottom: 6 }}
                >
                  {a.username}
                </Tag>
              ))}
            </div>
          )}

          <Form form={form} onFinish={onFinish} size="large">
            <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
              <AutoComplete options={options} style={{ width: '100%' }}>
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
