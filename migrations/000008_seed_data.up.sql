-- Seed data for production MySQL
INSERT INTO roles (id, name, code, description, level, status) VALUES
(1, '超级管理员', 'admin', '拥有全部权限', 0, 1),
(3, '编辑', 'editor', '编辑角色', 1, 1),
(4, '访客', 'viewer', '访客角色', 2, 1);
INSERT INTO permissions (id, name, code, resource, action) VALUES
(1,'查看用户','user:list','user','read'),
(2,'创建用户','user:create','user','create'),
(3,'更新用户','user:update','user','update'),
(4,'删除用户','user:delete','user','delete'),
(5,'查看菜单','menu:list','menu','read'),
(6,'创建菜单','menu:create','menu','create'),
(7,'更新菜单','menu:update','menu','update'),
(8,'删除菜单','menu:delete','menu','delete'),
(9,'查看角色','role:list','role','read'),
(10,'创建角色','role:create','role','create'),
(11,'更新角色','role:update','role','update'),
(12,'删除角色','role:delete','role','delete'),
(13,'查看配置','config:list','config','read'),
(14,'更新配置','config:update','config','update'),
(15,'查看审计','audit:list','audit','read');
-- Admin: all permissions
INSERT INTO role_permissions (role_id, permission_id) VALUES (1,1),(1,2),(1,3),(1,4),(1,5),(1,6),(1,7),(1,8),(1,9),(1,10),(1,11),(1,12),(1,13),(1,14),(1,15);
-- Editor: limited permissions
INSERT INTO role_permissions (role_id, permission_id) VALUES (3,1),(3,2),(3,3),(3,5),(3,6),(3,7),(3,9),(3,13);
-- Viewer: read-only
INSERT INTO role_permissions (role_id, permission_id) VALUES (4,1),(4,5),(4,9),(4,13);
INSERT INTO menus (id, parent_id, name, path, icon, sort, type) VALUES
(1, 0, 'Dashboard', '/dashboard', 'DashboardOutlined', 1, 2),
(2, 0, '系统管理', '', 'SettingOutlined', 2, 1),
(3, 2, '用户管理', '/system/users', 'UserOutlined', 1, 2),
(4, 2, '菜单管理', '/system/menus', 'MenuOutlined', 2, 2),
(5, 2, '角色管理', '/system/roles', 'SafetyOutlined', 3, 2),
(6, 2, '系统配置', '/system/configs', 'ControlOutlined', 4, 2),
(7, 2, '审计日志', '/system/audit', 'AuditOutlined', 5, 2);
-- Admin: all menus
INSERT INTO role_menus (role_id, menu_id) VALUES (1,1),(1,2),(1,3),(1,4),(1,5),(1,6),(1,7);
-- Editor: dashboard + users (no menus/roles/config/audit)
INSERT INTO role_menus (role_id, menu_id) VALUES (3,1),(3,2),(3,3);
-- Viewer: dashboard only
INSERT INTO role_menus (role_id, menu_id) VALUES (4,1);
INSERT INTO system_configs (`key`, value, remark) VALUES ('site_name', 'Kingfisher Admin', '系统名称');
INSERT INTO users (id, username, password, email, role_id) VALUES (1, 'admin', '$2a$12$jDyI8HZp/TVxUrplIqdgNOV/iahF.i3l0YoPHuNLD5kus./WsPTzO', 'admin@example.com', 1);
