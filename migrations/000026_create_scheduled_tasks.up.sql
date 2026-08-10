-- 周期任务：任务配置表（后台管理 → asynq PeriodicTaskManager 周期性同步调度）
CREATE TABLE IF NOT EXISTS scheduled_tasks (
    id         INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(64)  NOT NULL COMMENT '任务名称',
    task_type  VARCHAR(64)  NOT NULL COMMENT '任务类型（对应 worker 注册的类型，如 message:send）',
    cron_spec  VARCHAR(64)  NOT NULL COMMENT 'cron 表达式（5 段，如 0 9 * * *）',
    payload    TEXT         COMMENT '任务载荷 JSON',
    enabled    TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '是否启用',
    remark     VARCHAR(255) NOT NULL DEFAULT '',
    created_at DATETIME,
    updated_at DATETIME,
    INDEX idx_scheduled_tasks_enabled (enabled)
);

-- 任务管理菜单（系统管理下，Sort 9）
INSERT INTO menus (id, parent_id, name, path, component, icon, sort, permission, status, version, created_at, updated_at)
VALUES (20, 2, '任务管理', '/system/tasks', 'pages/Task/TaskManage', 'ScheduleOutlined', 9, 'task:list', 1, '1.0.0', NOW(), NOW());

-- 周期任务权限
INSERT INTO permissions (id, name, code, resource, action, created_at, updated_at) VALUES
  (28, '查看任务', 'task:list',   'task', 'read',   NOW(), NOW()),
  (29, '创建任务', 'task:create', 'task', 'create', NOW(), NOW()),
  (30, '更新任务', 'task:update', 'task', 'update', NOW(), NOW()),
  (31, '删除任务', 'task:delete', 'task', 'delete', NOW(), NOW());

-- admin 角色授权
INSERT INTO role_permissions (role_id, permission_id) VALUES
  (1, 28), (1, 29), (1, 30), (1, 31);
INSERT INTO role_menus (role_id, menu_id) VALUES (1, 20);

-- 示例：nop 空转任务（测试用，每分钟执行一次，worker 仅打印日志）
INSERT INTO scheduled_tasks (name, task_type, cron_spec, payload, enabled, remark, created_at, updated_at)
VALUES ('nop 测试任务', 'nop:run', '* * * * *', '{"note":"周期任务测试"}', 1, '每分钟空转一次，验证调度器→入队→worker 消费链路', NOW(), NOW());
