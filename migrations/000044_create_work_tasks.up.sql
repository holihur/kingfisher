CREATE TABLE tasks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(128) NOT NULL,
    description TEXT,
    owner_id BIGINT NOT NULL,
    department_id BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    INDEX idx_tasks_owner_id (owner_id),
    INDEX idx_tasks_department_id (department_id),
    INDEX idx_tasks_status (status)
);

INSERT INTO permissions (id, name, code, resource, action, created_at, updated_at) VALUES
  (42, '查看业务任务', 'worktask:list', 'worktask', 'read', NOW(), NOW()),
  (43, '创建业务任务', 'worktask:create', 'worktask', 'create', NOW(), NOW()),
  (44, '更新业务任务', 'worktask:update', 'worktask', 'update', NOW(), NOW()),
  (45, '删除业务任务', 'worktask:delete', 'worktask', 'delete', NOW(), NOW());

INSERT INTO menus (id, parent_id, name, path, component, icon, sort, permission, status, version, created_at, updated_at)
VALUES (25, 2, '演示任务', '/system/worktasks', 'pages/WorkTask/WorkTaskManage', 'CheckSquareOutlined', 13, 'worktask:list', 1, '1.0.0', NOW(), NOW());

INSERT INTO role_permissions (role_id, permission_id) VALUES
  (1, 42), (1, 43), (1, 44), (1, 45),
  (3, 42), (3, 43), (3, 44),
  (4, 42);

INSERT INTO role_menus (role_id, menu_id) VALUES (1, 25), (3, 25), (4, 25);
