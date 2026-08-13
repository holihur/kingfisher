-- 部门管理：菜单 + 权限 + 角色授权 + 示例数据

-- 部门管理菜单（系统管理下，Sort 12）
INSERT INTO menus (id, parent_id, name, path, component, icon, sort, permission, status, version, created_at, updated_at)
VALUES (23, 2, '部门管理', '/system/departments', 'pages/Department/DeptManage', 'ApartmentOutlined', 12, 'department:list', 1, '1.0.0', NOW(), NOW());

-- 部门权限
INSERT INTO permissions (id, name, code, resource, action, created_at, updated_at) VALUES
  (37, '查看部门', 'department:list',   'department', 'read',   NOW(), NOW()),
  (38, '创建部门', 'department:create', 'department', 'create', NOW(), NOW()),
  (39, '更新部门', 'department:update', 'department', 'update', NOW(), NOW()),
  (40, '删除部门', 'department:delete', 'department', 'delete', NOW(), NOW());

-- 角色授权：admin 全量；editor 查看+创建+更新；viewer 仅查看
INSERT INTO role_permissions (role_id, permission_id) VALUES
  (1, 37), (1, 38), (1, 39), (1, 40),
  (3, 37), (3, 38), (3, 39),
  (4, 37);
INSERT INTO role_menus (role_id, menu_id) VALUES
  (1, 23), (3, 23), (4, 23);

-- 示例部门树
INSERT INTO departments (id, parent_id, name, sort, status, remark, version, created_at, updated_at) VALUES
  (1, 0, '技术部', 1, 1, '研发与技术支持', '1.0.0', NOW(), NOW()),
  (2, 0, '产品部', 2, 1, '产品规划与设计', '1.0.0', NOW(), NOW()),
  (3, 1, '后端组', 1, 1, '服务端开发', '1.0.0', NOW(), NOW());

-- 部门-角色关联：技术部挂「编辑」角色，产品部挂「访客」角色（演示部门角色继承）
INSERT INTO department_roles (department_id, role_id) VALUES
  (1, 3), (2, 4);

-- 用户-部门关联：admin→技术部+产品部；editor→技术部；viewer→产品部；multi→技术部+产品部
INSERT INTO user_departments (user_id, department_id) VALUES
  (1, 1), (1, 2),
  (2, 1),
  (3, 2),
  (4, 1), (4, 2);
