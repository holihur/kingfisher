-- 撤销找回密码开关配置
DELETE FROM system_configs WHERE `key` = 'password_reset_enabled';
