-- 系统配置：新增 是否公开(is_public) 与 版本(version) 字段
ALTER TABLE system_configs
    ADD COLUMN is_public TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否公开：公开项未登录可读',
    ADD COLUMN version VARCHAR(32) NOT NULL DEFAULT '' COMMENT '该配置由哪个版本新增';

-- 初始种子中的对外展示项标记为公开
UPDATE system_configs SET is_public = 1, version = '1.0.0' WHERE `key` IN ('site_name', 'site_logo');
UPDATE system_configs SET version = '1.0.0' WHERE version = '';

-- 系统描述（登录页副标题），公开可读
INSERT INTO system_configs (`key`, value, remark, is_public, version)
SELECT 'site_description', 'Kingfisher 后台管理平台', '系统描述', 1, '1.0.0'
WHERE NOT EXISTS (SELECT 1 FROM system_configs WHERE `key` = 'site_description');
