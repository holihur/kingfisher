import React, { useEffect, useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Breadcrumb, Avatar, Dropdown, Spin } from 'antd';
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  LogoutOutlined,
  DashboardOutlined,
  SettingOutlined,
  UserOutlined,
  SafetyOutlined,
  ControlOutlined,
  AuditOutlined,
  QuestionOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '../stores/auth';
import { useMenuStore } from '../stores/menu';

const icons: Record<string, React.ReactNode> = {
  DashboardOutlined: <DashboardOutlined />,
  SettingOutlined: <SettingOutlined />,
  UserOutlined: <UserOutlined />,
  SafetyOutlined: <SafetyOutlined />,
  MenuOutlined: <span />,
  ControlOutlined: <ControlOutlined />,
  AuditOutlined: <AuditOutlined />,
};

const AdminLayout: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const { menuTree, fetchMenus, loading } = useMenuStore();
  const { userInfo, logout, fetchUserInfo } = useAuthStore();
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    fetchMenus();
    fetchUserInfo();
  }, []);

  const buildItems = (items: Record<string, unknown>[]): Record<string, unknown>[] =>
    (items || [])
      .filter((m: Record<string, unknown>) => (m.status as number) === 1)
      .sort((a: Record<string, unknown>, b: Record<string, unknown>) => (a.sort as number) - (b.sort as number))
      .map((item: Record<string, unknown>) => {
        const hasChildren = (item.children as Record<string, unknown>[])?.length > 0;
        const ch = hasChildren ? buildItems(item.children as Record<string, unknown>[]) : undefined;
        return {
          key: (item.path as string) || `menu_${item.id}`,
          icon: icons[item.icon as string] || <QuestionOutlined />,
          label: item.name as string,
          children: ch,
        };
      });

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Layout.Sider collapsible collapsed={collapsed} onCollapse={setCollapsed} theme="dark" width={220}>
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#fff',
            fontSize: 18,
            fontWeight: 'bold',
            borderBottom: '1px solid rgba(255,255,255,0.1)',
          }}
        >
          {collapsed ? 'K' : 'Kingfisher'}
        </div>
        {loading ? (
          <Spin style={{ display: 'block', marginTop: 40 }} />
        ) : (
          <Menu
            theme="dark"
            mode="inline"
            selectedKeys={[location.pathname]}
            items={buildItems(menuTree as Record<string, unknown>[])}
          />
        )}
      </Layout.Sider>
      <Layout>
        <Layout.Header
          style={{
            background: '#fff',
            padding: '0 24px',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            boxShadow: '0 1px 4px rgba(0,0,0,0.08)',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            {React.createElement(collapsed ? MenuUnfoldOutlined : MenuFoldOutlined, {
              onClick: () => setCollapsed(!collapsed),
            })}
            <Breadcrumb items={[{ title: '首页' }, { title: location.pathname.split('/').pop() }]} />
          </div>
          <Dropdown
            menu={{ items: [{ key: 'logout', icon: <LogoutOutlined />, label: '退出登录', onClick: handleLogout }] }}
          >
            <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
              <Avatar src={(userInfo as Record<string, unknown>)?.avatar as string} />
              <span>{(userInfo as Record<string, unknown>)?.username as string}</span>
            </div>
          </Dropdown>
        </Layout.Header>
        <Layout.Content
          style={{
            margin: 24,
            padding: 24,
            background: '#fff',
            borderRadius: 8,
            minHeight: 'calc(100vh - 64px - 48px)',
          }}
        >
          <Outlet />
        </Layout.Content>
      </Layout>
    </Layout>
  );
};

export default AdminLayout;
