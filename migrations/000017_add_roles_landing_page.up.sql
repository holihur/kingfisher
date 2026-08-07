-- 角色：新增落地页字段（登录后跳转的页面）
ALTER TABLE roles
    ADD COLUMN landing_page VARCHAR(128) NOT NULL DEFAULT '' COMMENT '角色登录后的落地页（如 /dashboard）';

UPDATE roles SET landing_page = '/dashboard' WHERE `code` = 'admin';
UPDATE roles SET landing_page = '/system/users' WHERE `code` = 'editor';
UPDATE roles SET landing_page = '/dashboard' WHERE `code` = 'viewer';
