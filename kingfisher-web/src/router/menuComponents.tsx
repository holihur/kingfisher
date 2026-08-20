import { lazy } from 'react';

/**
 * 菜单 component 字符串 → 懒加载组件映射。
 * 数据库菜单的 component 字段（如 pages/Dashboard）在此映射到实际组件，
 * 实现菜单驱动的动态路由。
 */
const componentMap: Record<string, React.LazyExoticComponent<React.ComponentType>> = {
  'pages/Dashboard': lazy(() => import('../pages/dashboard')),
  'pages/User/UserList': lazy(() => import('../pages/user/UserList')),
  'pages/Menu/MenuManage': lazy(() => import('../pages/menu/MenuManage')),
  'pages/Role/RoleList': lazy(() => import('../pages/role/RoleList')),
  'pages/Config/ConfigManage': lazy(() => import('../pages/config/ConfigManage')),
  'pages/Audit/AuditLogList': lazy(() => import('../pages/audit/AuditLogList')),
  'pages/Dict/DictManage': lazy(() => import('../pages/dict/DictManage')),
  'pages/Message/MessageManage': lazy(() => import('../pages/message/MessageManage')),
  'pages/Template/TemplateManage': lazy(() => import('../pages/template/TemplateManage')),
  'pages/Task/TaskManage': lazy(() => import('../pages/task/TaskManage')),
  'pages/WorkTask/WorkTaskManage': lazy(() => import('../pages/worktask/WorkTaskManage')),
  'pages/System/SystemInfo': lazy(() => import('../pages/system/SystemInfo')),
  'pages/Doc/DocManage': lazy(() => import('../pages/doc/DocManage')),
  'pages/Department/DeptManage': lazy(() => import('../pages/department/DeptManage')),
  'pages/Agent/AgentChat': lazy(() => import('../pages/agent/AgentChat')),
  // 特殊：仪表盘2 这类测试菜单若 component 为 pages/Dashboard 也能复用
  'pages/Profile': lazy(() => import('../pages/profile')),
};

/** 根据 component 字符串解析组件；未知返回 null。 */
export function resolveComponent(component?: string): React.LazyExoticComponent<React.ComponentType> | null {
  if (!component) return null;
  return componentMap[component] || null;
}

/** 支持前端注册额外组件（扩展菜单时用）。 */
export function registerComponent(key: string, comp: React.LazyExoticComponent<React.ComponentType>) {
  componentMap[key] = comp;
}
