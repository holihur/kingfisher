import React, { useEffect, useState } from 'react';import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Breadcrumb, Avatar, Dropdown, Spin, Drawer, Badge, theme as antdTheme } from 'antd';
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
  BookOutlined,
  MenuOutlined,
  QuestionOutlined,
  InboxOutlined,
  MailOutlined,
  FileTextOutlined,
  ScheduleOutlined,
  MonitorOutlined,
  BulbOutlined,
  BulbFilled,
} from '@ant-design/icons';
import { useAuthStore } from '../stores/auth';
import { useTheme } from '../hooks/useTheme';
import { useMenuStore } from '../stores/menu';
import { clearTokens } from '../utils/token';
import { configApi } from '../api/config';
import { messageApi } from '../api/message';

const icons: Record<string, React.ReactNode> = {
  DashboardOutlined: <DashboardOutlined />,
  SettingOutlined: <SettingOutlined />,
  UserOutlined: <UserOutlined />,
  SafetyOutlined: <SafetyOutlined />,
  MenuOutlined: <MenuOutlined />,
  ControlOutlined: <ControlOutlined />,
  AuditOutlined: <AuditOutlined />,
  BookOutlined: <BookOutlined />,
  MailOutlined: <MailOutlined />,
  FileTextOutlined: <FileTextOutlined />,
  ScheduleOutlined: <ScheduleOutlined />,
  MonitorOutlined: <MonitorOutlined />,
};

const MOBILE_BREAKPOINT = 768;

const AdminLayout: React.FC = () => {
  const [mobileDrawer, setMobileDrawer] = useState(false);
  const [isMobile, setIsMobile] = useState(window.innerWidth < MOBILE_BREAKPOINT);
  const [collapsed, setCollapsed] = useState(isMobile);
  const [openKeys, setOpenKeys] = useState<string[]>([]);
  const [siteName, setSiteName] = useState('Kingfisher');
  const [siteLogo, setSiteLogo] = useState('');
  const [unreadCount, setUnreadCount] = useState(0);
  const { theme, toggle: toggleTheme } = useTheme();
  const { token: themeToken } = antdTheme.useToken();
  const { menuTree, loading } = useMenuStore();
  const { userInfo } = useAuthStore();
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    configApi.getPublic('site_name').then(r => {
      const name = (r.data as any)?.value || 'Kingfisher';
      setSiteName(name);
      document.title = name;
    }).catch(() => {});
    configApi.getPublic('site_logo').then(r => setSiteLogo((r.data as any)?.value || '')).catch(() => {});
  }, []);

  useEffect(() => {
    const handleResize = () => {
      const mobile = window.innerWidth < MOBILE_BREAKPOINT;
      setIsMobile(mobile);
      if (mobile) setCollapsed(true);
    };
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // 未读站内信轮询
  useEffect(() => {
    const load = () => {
      messageApi.unreadCount().then((r) => {
        const n = ((r.data as Record<string, unknown>)?.unread_count as number) || 0;
        setUnreadCount(n);
      }).catch(() => {});
    };
    load();
    const timer = setInterval(load, 30000);
    return () => clearInterval(timer);
  }, []);

  // 菜单由路由层 AppRoutes 加载（驱动动态路由），此处不再重复请求

  // 菜单加载完成后 / 路由变化时，展开当前路由对应的祖先菜单
  useEffect(() => {
    if (!menuTree?.length) return;
    const chain = findBreadcrumb(menuTree as any, location.pathname);
    // 面包屑链上除叶子外的所有节点 = 需要展开的子菜单（key 即其 path）
    const ancestors = chain.slice(0, -1).map((m) => m.path).filter(Boolean) as string[];
    if (ancestors.length) {
      setOpenKeys((prev) => Array.from(new Set([...prev, ...ancestors])));
    }
  }, [menuTree, location.pathname]);

  const buildItems = (items: Record<string, unknown>[]): Record<string, unknown>[] =>
    (items || [])
      .filter((m: Record<string, unknown>) => (m.status as number) === 1)
      .sort((a: Record<string, unknown>, b: Record<string, unknown>) => (a.sort as number) - (b.sort as number))
      .map((item: Record<string, unknown>) => {
        const ch = buildItems((item.children as Record<string, unknown>[]) || []);
        return {
          key: (item.path as string) || `menu_${item.id}`,
          icon: icons[item.icon as string] || <QuestionOutlined />,
          label: item.name as string,
          children: ch.length > 0 ? ch : undefined,
        } as Record<string, unknown>;
      });

  const handleLogout = () => {
    clearTokens();
    window.location.href = '/login';
  };

  const menuProps = {
    theme: (theme === 'dark' ? 'dark' : 'light') as 'dark' | 'light',
    mode: 'inline' as const,
    selectedKeys: [location.pathname],
    openKeys,
    onOpenChange: (keys: string[]) => setOpenKeys(keys),
    onClick: ({ key }: { key: string }) => {
      navigate(key);
      setMobileDrawer(false);
    },
    items: buildItems(menuTree as unknown as Record<string, unknown>[]) as any,
  };

  const isSidebarDark = theme === 'dark';
  const sidebarBg = isSidebarDark ? '#001529' : '#fff';
  const sidebarTextColor = isSidebarDark ? '#fff' : 'inherit';
  const sidebarBorderColor = isSidebarDark ? 'rgba(255,255,255,0.1)' : '#f0f0f0';
  const siderContent = loading ? <Spin style={{ display: 'block', marginTop: 40 }} /> : <Menu {...menuProps} />;

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {isMobile ? (
        <Drawer
          open={mobileDrawer}
          onClose={() => setMobileDrawer(false)}
          placement="left"
          width={220}
          styles={{ body: { padding: 0, background: sidebarBg } }}
        >
          <div
            style={{
              height: 64,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 10,
              color: sidebarTextColor,
              fontSize: 18,
              fontWeight: 'bold',
              borderBottom: `1px solid ${sidebarBorderColor}`,
            }}
          >
            {siteLogo ? <img src={siteLogo} alt="logo" style={{ height: 32, borderRadius: 4 }} /> : (
            <span style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <span style={{ width: 32, height: 32, borderRadius: 6, background: isSidebarDark ? 'rgba(255,255,255,0.15)' : 'rgba(0,0,0,0.06)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 16, fontWeight: 600 }}>
                {(siteName || 'K').charAt(0)}
              </span>
              {siteName}
            </span>
          )}
          </div>
          {siderContent}
        </Drawer>
      ) : (
        <Layout.Sider collapsible collapsed={collapsed} onCollapse={setCollapsed} theme={theme === 'dark' ? 'dark' : 'light'} width={220}>
          <div
            style={{
              height: 64,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: collapsed ? 0 : 10,
              color: sidebarTextColor,
              fontSize: collapsed ? 22 : 18,
              fontWeight: 'bold',
              borderBottom: `1px solid ${sidebarBorderColor}`,
            }}
          >
            {siteLogo ? (
              <img src={siteLogo} alt="logo" style={{ height: collapsed ? 28 : 32, borderRadius: 4 }} />
            ) : (
              collapsed ? (
                <span style={{
                  width: 34, height: 34, borderRadius: 6,
                  background: isSidebarDark ? 'rgba(255,255,255,0.15)' : 'rgba(0,0,0,0.06)', display: 'flex',
                  alignItems: 'center', justifyContent: 'center', fontSize: 18,
                }}>
                  {(siteName || 'K').charAt(0)}
                </span>
              ) : (
                <span style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                  <span style={{
                    width: 32, height: 32, borderRadius: 6,
                    background: isSidebarDark ? 'rgba(255,255,255,0.15)' : 'rgba(0,0,0,0.06)', display: 'flex',
                    alignItems: 'center', justifyContent: 'center', fontSize: 16, fontWeight: 600,
                  }}>
                    {(siteName || 'K').charAt(0)}
                  </span>
                  {siteName}
                </span>
              )
            )}
          </div>
          {siderContent}
        </Layout.Sider>
      )}
      <Layout>
        <Layout.Header
          style={{
            background: themeToken.colorBgContainer,
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
              const bc = findBreadcrumb(menuTree as any, location.pathname);
              if (bc.length) {
                return [{ title: '首页' }, ...bc.map(m => ({ title: m.name }))];
              }
              return [{ title: '首页' }, { title: location.pathname.split('/').pop() }];
            })()} />
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <a href="/swagger/index.html" target="_blank" rel="noopener noreferrer" title="API 文档" style={{ color: '#666', fontSize: 17, display: 'flex', alignItems: 'center' }}>
              <FileTextOutlined />
            </a>
            <span onClick={toggleTheme} title={theme === 'dark' ? '切换浅色' : '切换暗色'} style={{ color: '#666', fontSize: 17, cursor: 'pointer', display: 'flex', alignItems: 'center' }}>
              {theme === 'dark' ? <BulbFilled /> : <BulbOutlined />}
            </span>
            <Badge count={unreadCount} size="small" offset={[-2, 2]}>
              <InboxOutlined
                style={{ fontSize: 18, cursor: 'pointer', color: '#666' }}
                onClick={() => navigate('/profile?tab=inbox')}
              />
            </Badge>
            <Dropdown
              menu={{
                  items: [
                    { key: 'profile', icon: <UserOutlined />, label: '个人中心', onClick: () => navigate('/profile') },
                    { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', onClick: handleLogout },
                  ],
                }}
            >
              <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
                <Avatar
                  src={((userInfo as Record<string, unknown>)?.avatar as string) || undefined}
                  style={!((userInfo as Record<string, unknown>)?.avatar as string) ? { background: '#1677ff' } : undefined}
                >
                  {((userInfo as Record<string, unknown>)?.nickname || (userInfo as Record<string, unknown>)?.username || '?')?.toString().charAt(0).toUpperCase()}
                </Avatar>
                <span>{(userInfo as Record<string, unknown>)?.username as string}</span>
              </div>
            </Dropdown>
          </div>
        </Layout.Header>
        <Layout.Content
          style={{
            margin: 16,
            padding: 0,
            minHeight: 'calc(100vh - 64px - 32px)',
          }}
        >
          <Outlet />
        </Layout.Content>
      </Layout>
    </Layout>
  );
};

export default AdminLayout;
