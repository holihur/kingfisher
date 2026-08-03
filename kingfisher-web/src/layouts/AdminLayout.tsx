import React, { useEffect, useState } from 'react';import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Breadcrumb, Avatar, Dropdown, Spin, Drawer } from 'antd';
import { findBreadcrumb } from '../utils/menu';
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
  MenuOutlined,
  QuestionOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '../stores/auth';
import { useMenuStore } from '../stores/menu';

const icons: Record<string, React.ReactNode> = {
  DashboardOutlined: <DashboardOutlined />,
  SettingOutlined: <SettingOutlined />,
  UserOutlined: <UserOutlined />,
  SafetyOutlined: <SafetyOutlined />,
  MenuOutlined: <MenuOutlined />,
  ControlOutlined: <ControlOutlined />,
  AuditOutlined: <AuditOutlined />,
};

const MOBILE_BREAKPOINT = 768;

const AdminLayout: React.FC = () => {
  const [mobileDrawer, setMobileDrawer] = useState(false);
  const [isMobile, setIsMobile] = useState(window.innerWidth < MOBILE_BREAKPOINT);
  const [collapsed, setCollapsed] = useState(isMobile);
  const { menuTree, fetchMenus, loading } = useMenuStore();
  const { userInfo, logout, fetchUserInfo } = useAuthStore();
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    const handleResize = () => {
      const mobile = window.innerWidth < MOBILE_BREAKPOINT;
      setIsMobile(mobile);
      if (mobile) setCollapsed(true);
    };
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    fetchMenus();
    fetchUserInfo();
  }, []);

  const buildItems = (items: Record<string, unknown>[]): Record<string, unknown>[] =>
    (items || [])
      .filter((m: Record<string, unknown>) => (m.status as number) === 1 && (m.type as number) !== 3)
      .sort((a: Record<string, unknown>, b: Record<string, unknown>) => (a.sort as number) - (b.sort as number))
      .map((item: Record<string, unknown>) => {
        const ch = buildItems((item.children as Record<string, unknown>[]) || []);
        return {
          key: (item.path as string) || `menu_${item.id}`,
          icon: icons[item.icon as string] || <QuestionOutlined />,
          label: item.name as string,
          children: ch.length > 0 ? ch : undefined,
        };
      });

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const menuProps = {
    theme: 'dark' as const,
    mode: 'inline' as const,
    selectedKeys: [location.pathname],
    onClick: ({ key }: { key: string }) => {
      navigate(key);
      setMobileDrawer(false);
    },
    items: buildItems(menuTree as Record<string, unknown>[]),
  };

  const siderContent = loading ? <Spin style={{ display: 'block', marginTop: 40 }} /> : <Menu {...menuProps} />;

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {isMobile ? (
        <Drawer
          open={mobileDrawer}
          onClose={() => setMobileDrawer(false)}
          placement="left"
          width={220}
          styles={{ body: { padding: 0, background: '#001529' } }}
        >
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
            Kingfisher
          </div>
          {siderContent}
        </Drawer>
      ) : (
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
          {siderContent}
        </Layout.Sider>
      )}
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
            {isMobile ? (
              <MenuUnfoldOutlined onClick={() => setMobileDrawer(true)} />
            ) : (
              React.createElement(collapsed ? MenuUnfoldOutlined : MenuFoldOutlined, {
                onClick: () => setCollapsed(!collapsed),
              })
            )}
            <Breadcrumb items={(() => {
              const bc = findBreadcrumb(menuTree as Record<string, unknown>[], location.pathname);
              if (bc.length) {
                return [{ title: '首页' }, ...bc.map(m => ({ title: m.name }))];
              }
              return [{ title: '首页' }, { title: location.pathname.split('/').pop() }];
            })()} />
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
