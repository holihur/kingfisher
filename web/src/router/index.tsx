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
  if (!token) return <Navigate to="/login" replace />;
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
      { path: 'dashboard', element: <Dashboard /> },
      { path: 'system/users', element: <UserList /> },
      { path: 'system/menus', element: <MenuManage /> },
      { path: 'system/roles', element: <RoleList /> },
      { path: 'system/configs', element: <ConfigManage /> },
      { path: 'system/audit', element: <AuditLogList /> },
      { index: true, element: <Navigate to="/dashboard" replace /> },
    ],
  },
  { path: '*', element: <Navigate to="/login" replace /> },
]);

export default router;
