import { Suspense, useEffect } from 'react';
import { createBrowserRouter, Navigate } from 'react-router-dom';
import { Spin } from 'antd';
import AdminLayout from '../layouts/AdminLayout';
import LoginPage from '../pages/login';
import Dashboard from '../pages/dashboard';
import NotFound from '../pages/error/NotFound';
import Forbidden from '../pages/error/Forbidden';
import UserList from '../pages/user/UserList';
import MenuManage from '../pages/menu/MenuManage';
import RoleList from '../pages/role/RoleList';
import ConfigManage from '../pages/config/ConfigManage';
import AuditLogList from '../pages/audit/AuditLogList';
import DictManage from '../pages/dict/DictManage';
import Profile from '../pages/profile';
import RegisterPage from '../pages/register';
import { hasToken } from '../utils/token';
import { useAuthStore } from '../stores/auth';

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

const router = createBrowserRouter([
  { path: '/login', element: <LoginPage /> },
  { path: '/register', element: <RegisterPage /> },
  { path: '/403', element: <Forbidden /> },
  {
    path: '/',
    element: (
      <AuthGuard>
        <AdminLayout />
      </AuthGuard>
    ),
    children: [
      { path: 'dashboard', element: <Suspense><Dashboard /></Suspense> },
      { path: 'profile', element: <Suspense><Profile /></Suspense> },
      { path: 'system/users', element: <Suspense><PermGuard perm="user:list"><UserList /></PermGuard></Suspense> },
      { path: 'system/menus', element: <Suspense><PermGuard perm="menu:list"><MenuManage /></PermGuard></Suspense> },
      { path: 'system/roles', element: <Suspense><PermGuard perm="role:list"><RoleList /></PermGuard></Suspense> },
      { path: 'system/configs', element: <Suspense><PermGuard perm="config:list"><ConfigManage /></PermGuard></Suspense> },
      { path: 'system/audit', element: <Suspense><PermGuard perm="audit:list"><AuditLogList /></PermGuard></Suspense> },
      { path: 'system/dicts', element: <Suspense><PermGuard perm="dict:list"><DictManage /></PermGuard></Suspense> },
      { index: true, element: <Navigate to="/dashboard" replace /> },
    ],
  },
  { path: '*', element: <NotFound /> },
]);

export default router;
