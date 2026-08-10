-- 找回密码开关（系统配置，公开可读：登录页据此前置判断入口）
INSERT INTO system_configs (`key`, value, remark, is_public, version, render)
SELECT 'password_reset_enabled', 'true', '是否开启找回密码', 1, '1.0.0', 'switch'
WHERE NOT EXISTS (SELECT 1 FROM system_configs WHERE `key` = 'password_reset_enabled');
