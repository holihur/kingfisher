import React, { Suspense, lazy, useEffect, useMemo } from 'react';
import { Navigate, useLocation, useRoutes } from 'react-router-dom';
import { Spin } from 'antd';
import AdminLayout from '../layouts/AdminLayout';
import { hasToken } from '../utils/token';
import { useAuthStore } from '../stores/auth';
import { useMenuStore } from '../stores/menu';
import { resolveComponent } from './menuComponents';

// 路由级懒加载：按页面分包
const LoginPage = lazy(() => import('../pages/login'));
const NotFound = lazy(() => import('../pages/error/NotFound'));
const Forbidden = lazy(() => import('../pages/error/Forbidden'));
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
  if (!userLoaded) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
        <Spin size="large" />
      </div>
    );
  }
  return <>{children}</>;
}

function PermGuard({ perm, children }: { perm?: string; children: React.ReactNode }) {
  const permissions = useAuthStore((s) => s.permissions);
  // 菜单定义了 permission 时校验；未定义则放行（无权限控制）
  if (perm && !permissions.includes(perm)) {
    return <Navigate to="/403" replace />;
  }
  return <>{children}</>;
}

/** 懒加载 fallback */
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

interface MenuNode {
  id: number;
  parent_id: number;
  path?: string;
  component?: string;
  permission?: string;
  children?: MenuNode[];
}

/** 由菜单树生成动态子路由。 */
function buildMenuRoutes(menuTree: MenuNode[]) {
  const routes: { path: string; element: React.ReactNode }[] = [];
  const walk = (nodes: MenuNode[]) => {
    for (const m of nodes) {
      if (!m.path) {
        // 目录节点，递归子项
        if (m.children?.length) walk(m.children);
        continue;
      }
      // 有 component 则映射组件，否则 403（无组件不可访问）
      const comp = resolveComponent(m.component);
      routes.push({
        path: m.path,
        element: comp ? (
          <PermGuard perm={m.permission}>
            <Lazy>{comp ? React.createElement(comp) : null}</Lazy>
          </PermGuard>
        ) : (
          <Navigate to="/403" replace />
        ),
      });
      if (m.children?.length) walk(m.children);
    }
  };
  walk(menuTree);
  return routes;
}

/** 应用路由：菜单树动态生成 + 固定路由。 */
function AppRoutes() {
  const menuTree = useMenuStore((s) => s.menuTree);
  const fetchMenus = useMenuStore((s) => s.fetchMenus);
  const location = useLocation();

  // 登录后加载菜单（驱动动态路由）；未登录不加载
  useEffect(() => {
    if (hasToken() && menuTree.length === 0) {
      fetchMenus();
    }
  }, [hasToken, menuTree.length, fetchMenus]);

  const menuRoutes = useMemo(() => buildMenuRoutes(menuTree as MenuNode[]), [menuTree]);

  // 菜单未加载时（首屏刷新），展示 loading 等待菜单树
  const routeElement = useRoutes([
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
        { path: 'profile', element: <Lazy><Profile /></Lazy> },
        ...menuRoutes,
        { index: true, element: <Navigate to="/dashboard" replace /> },
      ],
    },
    { path: '*', element: <Lazy><NotFound /></Lazy> },
  ]);

  // 登录态且菜单尚未加载时，等菜单（刷新场景）
  if (hasToken() && menuTree.length === 0 && location.pathname !== '/login') {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
        <Spin size="large" />
      </div>
    );
  }
  return routeElement;
}

export default AppRoutes;
