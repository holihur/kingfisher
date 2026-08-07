-- 回滚：注册开关 + 默认注册角色配置
DELETE FROM system_configs WHERE `key` IN ('registration_enabled', 'default_register_role_id');
