import { Suspense, lazy, useEffect } from 'react';
import { createBrowserRouter, Navigate } from 'react-router-dom';
import { Spin } from 'antd';
import AdminLayout from '../layouts/AdminLayout';
import { hasToken } from '../utils/token';
import { useAuthStore } from '../stores/auth';

// 路由级懒加载：按页面分包，避免把所有页面 + antd 打进单个大 chunk
const LoginPage = lazy(() => import('../pages/login'));
const Dashboard = lazy(() => import('../pages/dashboard'));
const NotFound = lazy(() => import('../pages/error/NotFound'));
const Forbidden = lazy(() => import('../pages/error/Forbidden'));
const UserList = lazy(() => import('../pages/user/UserList'));
const MenuManage = lazy(() => import('../pages/menu/MenuManage'));
const RoleList = lazy(() => import('../pages/role/RoleList'));
const ConfigManage = lazy(() => import('../pages/config/ConfigManage'));
const AuditLogList = lazy(() => import('../pages/audit/AuditLogList'));
const DictManage = lazy(() => import('../pages/dict/DictManage'));
const Profile = lazy(() => import('../pages/profile'));
const RegisterPage = lazy(() => import('../pages/register'));

function AuthGuard({ children }: { children: React.ReactNode }) {
  const userLoaded = useAuthStore((s) => s.userLoaded);
  const fetchUserInfo = useAuthStore((s) => s.fetchUserInfo);

  useEffect(() => {
    if (hasToken()) fetchUserInfo();
  }, [fetchUserInfo]);

  if (!hasToken()) {
    const path = window.location.pathname + window.location.search;
    return <Navigate to={`/login?redirect=${encodeURIComponent(path)}`} replace />;
  }
  // 权限尚未加载完成前先展示 loading，避免 PermGuard 在空权限下误判 403
  if (!userLoaded) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
        <Spin size="large" />
      </div>
    );
  }
  return <>{children}</>;
}

function PermGuard({ perm, children }: { perm: string; children: React.ReactNode }) {
  const permissions = useAuthStore((s) => s.permissions);
  if (!permissions.includes(perm)) {
    return <Navigate to="/403" replace />;
  }
  return <>{children}</>;
}

/** 懒加载 fallback：路由分包加载时显示 loading。 */
function Lazy({ children }: { children: React.ReactNode }) {
  return (
    <Suspense
      fallback={
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
          <Spin size="large" />
        </div>
      }
    >
      {children}
    </Suspense>
  );
}

const router = createBrowserRouter([
  { path: '/login', element: <Lazy><LoginPage /></Lazy> },
  { path: '/register', element: <Lazy><RegisterPage /></Lazy> },
  { path: '/403', element: <Lazy><Forbidden /></Lazy> },
  {
    path: '/',
    element: (
      <AuthGuard>
        <AdminLayout />
      </AuthGuard>
    ),
    children: [
      { path: 'dashboard', element: <Lazy><Dashboard /></Lazy> },
      { path: 'profile', element: <Lazy><Profile /></Lazy> },
      { path: 'system/users', element: <Lazy><PermGuard perm="user:list"><UserList /></PermGuard></Lazy> },
      { path: 'system/menus', element: <Lazy><PermGuard perm="menu:list"><MenuManage /></PermGuard></Lazy> },
      { path: 'system/roles', element: <Lazy><PermGuard perm="role:list"><RoleList /></PermGuard></Lazy> },
      { path: 'system/configs', element: <Lazy><PermGuard perm="config:list"><ConfigManage /></PermGuard></Lazy> },
      { path: 'system/audit', element: <Lazy><PermGuard perm="audit:list"><AuditLogList /></PermGuard></Lazy> },
      { path: 'system/dicts', element: <Lazy><PermGuard perm="dict:list"><DictManage /></PermGuard></Lazy> },
      { index: true, element: <Navigate to="/dashboard" replace /> },
    ],
  },
  { path: '*', element: <Lazy><NotFound /></Lazy> },
]);

export default router;
