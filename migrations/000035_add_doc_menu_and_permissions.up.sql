-- 文档管理：菜单 + 权限 + 角色授权 + 示例数据

-- 文档管理菜单（系统管理下，Sort 11）
INSERT INTO menus (id, parent_id, name, path, component, icon, sort, permission, status, version, created_at, updated_at)
VALUES (22, 2, '文档管理', '/system/docs', 'pages/Doc/DocManage', 'FileTextOutlined', 11, 'doc:list', 1, '1.0.0', NOW(), NOW());

-- 文档权限
INSERT INTO permissions (id, name, code, resource, action, created_at, updated_at) VALUES
  (33, '查看文档', 'doc:list',   'doc', 'read',   NOW(), NOW()),
  (34, '创建文档', 'doc:create', 'doc', 'create', NOW(), NOW()),
  (35, '更新文档', 'doc:update', 'doc', 'update', NOW(), NOW()),
  (36, '删除文档', 'doc:delete', 'doc', 'delete', NOW(), NOW());

-- 角色授权：admin 全量；editor 查看+创建+更新；viewer 仅查看
INSERT INTO role_permissions (role_id, permission_id) VALUES
  (1, 33), (1, 34), (1, 35), (1, 36),
  (3, 33), (3, 34), (3, 35),
  (4, 33);
INSERT INTO role_menus (role_id, menu_id) VALUES
  (1, 22), (3, 22), (4, 22);

-- 示例目录（含可见角色授权，演示默认拒绝）
INSERT INTO doc_directories (id, parent_id, name, sort, status, version, created_at, updated_at) VALUES
  (1, 0, '产品文档', 1, 1, '1.0.0', NOW(), NOW()),
  (2, 0, '技术文档', 2, 1, '1.0.0', NOW(), NOW()),
  (3, 1, '内部资料', 1, 1, '1.0.0', NOW(), NOW());

-- 目录可见角色：dir1 全角色；dir2 admin+editor；dir3 仅 admin
INSERT INTO doc_dir_roles (dir_id, role_id) VALUES
  (1, 1), (1, 3), (1, 4),
  (2, 1), (2, 3),
  (3, 1);

-- 示例文档（doc1 已发布共享；doc2 草稿演示 draft 仅作者可见）
INSERT INTO documents (id, dir_id, title, content, owner_id, visibility, status, current_version, sort, created_at, updated_at) VALUES
  (1, 1, '产品使用手册', '<h2>欢迎使用 Kingfisher</h2><p>本文档介绍产品核心功能。</p>', 1, 'shared', 'published', 1, 0, NOW(), NOW()),
  (2, 2, '开发规范', '<p>代码风格与提交流程。</p>', 2, 'shared', 'draft', 1, 0, NOW(), NOW());

-- 示例版本
INSERT INTO doc_versions (id, doc_id, version_no, title, content, owner_id, note, created_at) VALUES
  (1, 1, 1, '产品使用手册', '<h2>欢迎使用 Kingfisher</h2><p>本文档介绍产品核心功能。</p>', 1, '初始版本', NOW()),
  (2, 2, 1, '开发规范', '<p>代码风格与提交流程。</p>', 2, '初始版本', NOW());
