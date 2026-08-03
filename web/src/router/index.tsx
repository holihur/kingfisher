import { createBrowserRouter, Navigate } from 'react-router-dom';
import AdminLayout from '../layouts/AdminLayout';
import LoginPage from '../pages/login';
import Dashboard from '../pages/dashboard';
import UserList from '../pages/user/UserList';
import MenuManage from '../pages/menu/MenuManage';
import RoleList from '../pages/role/RoleList';
import ConfigManage from '../pages/config/ConfigManage';
import AuditLogList from '../pages/audit/AuditLogList';

function AuthGuard({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('kingfisher_token');
  if (!token) {
    const path = window.location.pathname + window.location.search;
    return <Navigate to={`/login?redirect=${encodeURIComponent(path)}`} replace />;
  }
  return <>{children}</>;
}

const router = createBrowserRouter([
  { path: '/login', element: <LoginPage /> },
  {
    path: '/',
    element: (
      <AuthGuard>
        <AdminLayout />
      </AuthGuard>
    ),
    children: [
      { path: 'dashboard', element: <Suspense><Dashboard /></Suspense> },
      { path: 'system/users', element: <Suspense><UserList /></Suspense> },
      { path: 'system/menus', element: <Suspense><MenuManage /></Suspense> },
      { path: 'system/roles', element: <Suspense><RoleList /></Suspense> },
      { path: 'system/configs', element: <Suspense><ConfigManage /></Suspense> },
      { path: 'system/audit', element: <Suspense><AuditLogList /></Suspense> },
      { index: true, element: <Navigate to="/dashboard" replace /> },
    ],
  },
  { path: '*', element: <NotFound /> },
]);

export default router;
