-- 站内信：消息表
CREATE TABLE IF NOT EXISTS messages (
    id           INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sender_id    INT UNSIGNED NOT NULL DEFAULT 0,
    sender_type  VARCHAR(16)  NOT NULL DEFAULT 'admin' COMMENT 'admin | system',
    recipient_id INT UNSIGNED NOT NULL,
    title        VARCHAR(128) NOT NULL,
    content      TEXT,
    status       VARCHAR(16)  NOT NULL DEFAULT 'sent' COMMENT 'draft | sent',
    is_read      TINYINT(1)   NOT NULL DEFAULT 0,
    read_at      DATETIME     NULL,
    created_at   DATETIME,
    updated_at   DATETIME,
    INDEX idx_messages_recipient (recipient_id),
    INDEX idx_messages_recipient_read (recipient_id, is_read)
);

-- 站内信菜单（系统管理下，Sort 7）
INSERT INTO menus (id, parent_id, name, path, component, icon, sort, permission, status, version, created_at, updated_at)
VALUES (18, 2, '站内信管理', '/system/messages', 'pages/Message/MessageManage', 'MailOutlined', 7, 'message:list', 1, '1.0.0', NOW(), NOW());

-- 站内信权限
INSERT INTO permissions (id, name, code, resource, action, created_at, updated_at) VALUES
  (20, '查看站内信', 'message:list',   'message', 'read',   NOW(), NOW()),
  (21, '发送站内信', 'message:create', 'message', 'create', NOW(), NOW()),
  (22, '更新站内信', 'message:update', 'message', 'update', NOW(), NOW()),
  (23, '删除站内信', 'message:delete', 'message', 'delete', NOW(), NOW());

-- admin 角色授权
INSERT INTO role_permissions (role_id, permission_id) VALUES
  (1, 20), (1, 21), (1, 22), (1, 23);
INSERT INTO role_menus (role_id, menu_id) VALUES (1, 18);

-- 示例消息（系统发给 admin）
INSERT INTO messages (id, sender_id, sender_type, recipient_id, title, content, status, is_read, created_at, updated_at)
VALUES (1, 0, 'system', 1, '欢迎使用 Kingfisher', '这是一个站内信示例。管理员可发送站内信，收件人可在个人中心-收件箱查看。', 'sent', 0, NOW(), NOW());
