-- 种子数据，复刻 Go core/database/gorm.go SeedData
INSERT OR IGNORE INTO permissions (id, name, code, resource, action, created_at, updated_at) VALUES
(1, '查看用户', 'user:list', 'user', 'read', datetime('now'), datetime('now')),
(2, '创建用户', 'user:create', 'user', 'create', datetime('now'), datetime('now')),
(3, '更新用户', 'user:update', 'user', 'update', datetime('now'), datetime('now')),
(4, '删除用户', 'user:delete', 'user', 'delete', datetime('now'), datetime('now')),
(5, '查看菜单', 'menu:list', 'menu', 'read', datetime('now'), datetime('now')),
(6, '创建菜单', 'menu:create', 'menu', 'create', datetime('now'), datetime('now')),
(7, '更新菜单', 'menu:update', 'menu', 'update', datetime('now'), datetime('now')),
(8, '删除菜单', 'menu:delete', 'menu', 'delete', datetime('now'), datetime('now')),
(9, '查看角色', 'role:list', 'role', 'read', datetime('now'), datetime('now')),
(10, '创建角色', 'role:create', 'role', 'create', datetime('now'), datetime('now')),
(11, '更新角色', 'role:update', 'role', 'update', datetime('now'), datetime('now')),
(12, '删除角色', 'role:delete', 'role', 'delete', datetime('now'), datetime('now')),
(13, '查看配置', 'config:list', 'config', 'read', datetime('now'), datetime('now')),
(14, '更新配置', 'config:update', 'config', 'update', datetime('now'), datetime('now')),
(15, '查看审计', 'audit:list', 'audit', 'read', datetime('now'), datetime('now')),
(16, '查看字典', 'dict:list', 'dict', 'read', datetime('now'), datetime('now')),
(17, '创建字典', 'dict:create', 'dict', 'create', datetime('now'), datetime('now')),
(18, '更新字典', 'dict:update', 'dict', 'update', datetime('now'), datetime('now')),
(19, '删除字典', 'dict:delete', 'dict', 'delete', datetime('now'), datetime('now')),
(20, '查看站内信', 'message:list', 'message', 'read', datetime('now'), datetime('now')),
(21, '发送站内信', 'message:create', 'message', 'create', datetime('now'), datetime('now')),
(22, '更新站内信', 'message:update', 'message', 'update', datetime('now'), datetime('now')),
(23, '删除站内信', 'message:delete', 'message', 'delete', datetime('now'), datetime('now')),
(24, '查看模版', 'template:list', 'template', 'read', datetime('now'), datetime('now')),
(25, '创建模版', 'template:create', 'template', 'create', datetime('now'), datetime('now')),
(26, '更新模版', 'template:update', 'template', 'update', datetime('now'), datetime('now')),
(27, '删除模版', 'template:delete', 'template', 'delete', datetime('now'), datetime('now')),
(28, '查看任务', 'task:list', 'task', 'read', datetime('now'), datetime('now')),
(29, '创建任务', 'task:create', 'task', 'create', datetime('now'), datetime('now')),
(30, '更新任务', 'task:update', 'task', 'update', datetime('now'), datetime('now')),
(31, '删除任务', 'task:delete', 'task', 'delete', datetime('now'), datetime('now')),
(32, '查看系统信息', 'system:list', 'system', 'read', datetime('now'), datetime('now')),
(33, '查看文档', 'doc:list', 'doc', 'read', datetime('now'), datetime('now')),
(34, '创建文档', 'doc:create', 'doc', 'create', datetime('now'), datetime('now')),
(35, '更新文档', 'doc:update', 'doc', 'update', datetime('now'), datetime('now')),
(36, '删除文档', 'doc:delete', 'doc', 'delete', datetime('now'), datetime('now')),
(37, '查看部门', 'department:list', 'department', 'read', datetime('now'), datetime('now')),
(38, '创建部门', 'department:create', 'department', 'create', datetime('now'), datetime('now')),
(39, '更新部门', 'department:update', 'department', 'update', datetime('now'), datetime('now')),
(40, '删除部门', 'department:delete', 'department', 'delete', datetime('now'), datetime('now')),
(41, '使用Agent', 'agent:list', 'agent', 'read', datetime('now'), datetime('now')),
(42, '查看业务任务', 'worktask:list', 'worktask', 'read', datetime('now'), datetime('now')),
(43, '创建业务任务', 'worktask:create', 'worktask', 'create', datetime('now'), datetime('now')),
(44, '更新业务任务', 'worktask:update', 'worktask', 'update', datetime('now'), datetime('now')),
(45, '删除业务任务', 'worktask:delete', 'worktask', 'delete', datetime('now'), datetime('now'));

INSERT OR IGNORE INTO roles (id, name, code, level, status, landing_page, created_at, updated_at) VALUES
(1, '超级管理员', 'admin', 0, 1, '/dashboard', datetime('now'), datetime('now')),
(3, '编辑', 'editor', 1, 1, '/system/users', datetime('now'), datetime('now')),
(4, '访客', 'viewer', 2, 1, '/dashboard', datetime('now'), datetime('now'));

INSERT OR IGNORE INTO users (id, username, nickname, password, email, status, session_version, created_at, updated_at) VALUES
(1, 'admin', 'admin', '$2a$12$jDyI8HZp/TVxUrplIqdgNOV/iahF.i3l0YoPHuNLD5kus./WsPTzO', 'admin@example.com', 1, 1, datetime('now'), datetime('now')),
(2, 'editor', 'editor', '$2a$12$jDyI8HZp/TVxUrplIqdgNOV/iahF.i3l0YoPHuNLD5kus./WsPTzO', 'editor@example.com', 1, 1, datetime('now'), datetime('now')),
(3, 'viewer', 'viewer', '$2a$12$jDyI8HZp/TVxUrplIqdgNOV/iahF.i3l0YoPHuNLD5kus./WsPTzO', 'viewer@example.com', 1, 1, datetime('now'), datetime('now')),
(4, 'multi', 'multi', '$2a$12$jDyI8HZp/TVxUrplIqdgNOV/iahF.i3l0YoPHuNLD5kus./WsPTzO', 'multi@example.com', 1, 1, datetime('now'), datetime('now'));

INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES
(1, 1), (2, 3), (3, 4), (4, 1), (4, 3), (4, 4);

INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES
(1, 1), (1, 2), (1, 3), (1, 4), (1, 5), (1, 6), (1, 7), (1, 8),
(1, 9), (1, 10), (1, 11), (1, 12), (1, 13), (1, 14), (1, 15),
(1, 16), (1, 17), (1, 18), (1, 19),
(1, 20), (1, 21), (1, 22), (1, 23),
(1, 24), (1, 25), (1, 26), (1, 27),
(1, 28), (1, 29), (1, 30), (1, 31), (1, 32),
(1, 33), (1, 34), (1, 35), (1, 36),
(1, 37), (1, 38), (1, 39), (1, 40),
(1, 41), (1, 42), (1, 43), (1, 44), (1, 45),
(3, 1), (3, 2), (3, 3), (3, 5), (3, 6), (3, 7), (3, 9), (3, 13), (3, 16),
(3, 33), (3, 34), (3, 35),
(3, 37), (3, 38), (3, 39),
(3, 42), (3, 43), (3, 44),
(3, 41), (4, 41),
(4, 1), (4, 5), (4, 9), (4, 13), (4, 16), (4, 33), (4, 42);

-- config_groups
INSERT OR IGNORE INTO config_groups (id, name, sort, created_at, updated_at) VALUES
(1, '站点', 1, datetime('now'), datetime('now')),
(2, '安全', 2, datetime('now'), datetime('now')),
(3, 'Agent', 3, datetime('now'), datetime('now'));

INSERT OR IGNORE INTO system_configs (id, key, value, remark, is_public, version, render, group_id, created_at, updated_at) VALUES
(1, 'site_name', 'Kingfisher', '系统名称', 1, '1.0.0', 'text', 1, datetime('now'), datetime('now')),
(2, 'registration_enabled', 'true', '是否开放注册', 1, '1.0.0', 'switch', 2, datetime('now'), datetime('now')),
(3, 'default_register_role_id', '4', '默认注册角色', 0, '1.0.0', 'select', 2, datetime('now'), datetime('now')),
(4, 'site_description', 'Kingfisher 后台', '系统描述', 1, '1.0.0', 'textarea', 1, datetime('now'), datetime('now'));

INSERT OR IGNORE INTO dict_types (id, code, name, is_public, status, version, created_at, updated_at) VALUES
(1, 'gender', '性别', 1, 1, '1.0.0', datetime('now'), datetime('now'));
INSERT OR IGNORE INTO dict_entries (id, type_id, label, value, sort, status, version, created_at, updated_at) VALUES
(1, 1, '男', 'male', 1, 1, '1.0.0', datetime('now'), datetime('now')),
(2, 1, '女', 'female', 2, 1, '1.0.0', datetime('now'), datetime('now'));
