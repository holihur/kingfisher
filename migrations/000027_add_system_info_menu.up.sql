-- 系统信息菜单（系统管理下，Sort 10）
INSERT INTO menus (id, parent_id, name, path, component, icon, sort, permission, status, version, created_at, updated_at)
VALUES (21, 2, '系统信息', '/system/info', 'pages/System/SystemInfo', 'MonitorOutlined', 10, 'system:list', 1, '1.0.0', NOW(), NOW());

-- 系统信息权限
INSERT INTO permissions (id, name, code, resource, action, created_at, updated_at) VALUES
  (32, '查看系统信息', 'system:list', 'system', 'read', NOW(), NOW());

-- admin 角色授权
INSERT INTO role_permissions (role_id, permission_id) VALUES (1, 32);
INSERT INTO role_menus (role_id, menu_id) VALUES (1, 21);
