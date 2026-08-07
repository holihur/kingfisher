import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Form, Input, Button, Card, App, AutoComplete, Checkbox, Tag } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useAuthStore } from '../../stores/auth';
import {
  type SavedAccount,
  loadAccounts,
  saveAccount,
  getAccountPassword,
  removeAccount,
} from '../../utils/remember';

const LoginPage: React.FC = () => {
  const { message } = App.useApp();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [siteName, setSiteName] = useState('Kingfisher');
  const [siteDescription, setSiteDescription] = useState('后台管理系统');
  const [rememberAccount, setRememberAccount] = useState(true);
  const [rememberPwd, setRememberPwd] = useState(false);
  const [accounts, setAccounts] = useState<SavedAccount[]>(() => loadAccounts());
  const login = useAuthStore((s) => s.login);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  useEffect(() => {
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
      saveAccount(values.username, values.password, rememberAccount, rememberPwd);
      const landing = useAuthStore.getState().landingPage;
      navigate(redirectTo || landing || '/dashboard');
    } catch {
    } finally {
      setLoading(false);
    }
  };

  const handleSelectAccount = (username: string) => {
    form.setFieldsValue({ username });
    const pwd = getAccountPassword(username);
    if (pwd) form.setFieldsValue({ password: pwd });
  };

  const handleRemoveAccount = (username: string) => {
    removeAccount(username);
    setAccounts(loadAccounts());
  };

  const autoCompleteOptions = useMemo(() =>
    accounts.map(a => ({
      value: a.username,
      label: (
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '4px 0' }}>
          <span><UserOutlined style={{ marginRight: 8 }} />{a.username}</span>
          <span style={{ color: '#999', fontSize: 12 }} onMouseDown={e => { e.stopPropagation(); e.preventDefault(); handleRemoveAccount(a.username); }}>删除</span>
        </div>
      ),
    })), [accounts]);

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f0f2f5' }}>
      <Card style={{ width: 400, boxShadow: '0 2px 8px rgba(0,0,0,0.15)' }} variant="borderless">
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <div style={{ fontSize: 40, marginBottom: 8 }}>🦜</div>
          <h2 style={{ margin: 0 }}>{siteName}</h2>
          <p style={{ color: '#999', margin: '8px 0 0' }}>{siteDescription}</p>
        </div>

        {accounts.length > 0 && (
          <div style={{ marginBottom: 12 }}>
            {accounts.slice(0, 5).map(a => (
              <Tag
                key={a.username}
                closable
                onClose={e => { e.preventDefault(); handleRemoveAccount(a.username); }}
                onClick={() => handleSelectAccount(a.username)}
                style={{ marginBottom: 6, cursor: 'pointer', fontSize: 13, padding: '4px 8px' }}
                color={a.password ? 'green' : 'blue'}
              >
                {a.username}{a.password ? ' (已存密码)' : ''}
              </Tag>
            ))}
          </div>
        )}

        <Form form={form} onFinish={onFinish} size="large">
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <AutoComplete options={autoCompleteOptions} onSelect={handleSelectAccount} style={{ width: '100%' }}>
              <Input prefix={<UserOutlined />} placeholder="用户名" />
            </AutoComplete>
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          <Form.Item style={{ marginBottom: 8 }}>
            <div style={{ display: 'flex', gap: 16 }}>
              <Checkbox checked={rememberAccount} onChange={e => { setRememberAccount(e.target.checked); if (!e.target.checked) setRememberPwd(false); }}>记住账户</Checkbox>
              <Checkbox checked={rememberPwd} disabled={!rememberAccount} onChange={e => setRememberPwd(e.target.checked)}>记住密码</Checkbox>
            </div>
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>登 录</Button>
          </Form.Item>
        </Form>

        <div style={{ textAlign: 'center', marginTop: 16 }}>还没有账号？<a href="/register">去注册</a></div>
      </Card>
    </div>
  );
};

export default LoginPage;
