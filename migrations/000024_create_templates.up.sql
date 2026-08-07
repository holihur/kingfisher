-- 模版管理：消息/通知/通用模板表
CREATE TABLE IF NOT EXISTS templates (
    id            INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name          VARCHAR(64)  NOT NULL COMMENT '模板名称',
    code          VARCHAR(64)  NOT NULL COMMENT '模板编码（唯一）',
    template_type VARCHAR(32)  NOT NULL DEFAULT 'general' COMMENT 'general | message | email | sms',
    title         VARCHAR(255) NOT NULL DEFAULT '' COMMENT '标题模板（支持 {{占位符}}）',
    content       TEXT         NULL COMMENT '内容模板（支持 {{占位符}}）',
    status        TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '1=启用, 0=禁用',
    remark        VARCHAR(255) NOT NULL DEFAULT '',
    version       VARCHAR(32)  NOT NULL DEFAULT '1.0.0',
    created_at    DATETIME,
    updated_at    DATETIME,
    UNIQUE KEY uk_templates_code (code),
    INDEX idx_templates_type (template_type)
);

-- 模版管理菜单（系统管理下，Sort 8）
INSERT INTO menus (id, parent_id, name, path, component, icon, sort, permission, status, version, created_at, updated_at)
VALUES (19, 2, '模版管理', '/system/templates', 'pages/Template/TemplateManage', 'FileTextOutlined', 8, 'template:list', 1, '1.0.0', NOW(), NOW());

-- 模版权限
INSERT INTO permissions (id, name, code, resource, action, created_at, updated_at) VALUES
  (24, '查看模版', 'template:list',   'template', 'read',   NOW(), NOW()),
  (25, '创建模版', 'template:create', 'template', 'create', NOW(), NOW()),
  (26, '更新模版', 'template:update', 'template', 'update', NOW(), NOW()),
  (27, '删除模版', 'template:delete', 'template', 'delete', NOW(), NOW());

-- admin 角色授权
INSERT INTO role_permissions (role_id, permission_id) VALUES
  (1, 24), (1, 25), (1, 26), (1, 27);
INSERT INTO role_menus (role_id, menu_id) VALUES (1, 19);

-- 示例模版
INSERT INTO templates (id, name, code, template_type, title, content, status, version, created_at, updated_at) VALUES
  (1, '欢迎消息', 'welcome_message', 'message', '欢迎 {{nickname}}', '你好 {{nickname}}，欢迎使用 Kingfisher！', 1, '1.0.0', NOW(), NOW()),
  (2, '密码重置通知', 'password_reset', 'message', '密码重置', '您的密码已重置，请登录后修改。', 1, '1.0.0', NOW(), NOW());
