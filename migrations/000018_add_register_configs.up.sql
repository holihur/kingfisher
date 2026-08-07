-- 系统配置：注册开关 + 默认注册角色
INSERT INTO system_configs (`key`, value, remark, is_public, version, render, render_options, group_id)
SELECT 'registration_enabled', 'true', '是否开放注册', 1, '1.0.0', 'switch', NULL,
       (SELECT id FROM config_groups WHERE name = '安全' LIMIT 1)
WHERE NOT EXISTS (SELECT 1 FROM system_configs WHERE `key` = 'registration_enabled');

INSERT INTO system_configs (`key`, value, remark, is_public, version, render, render_options, group_id)
SELECT 'default_register_role_id', '4', '默认注册用户的角色', 0, '1.0.0', 'select',
       '[{"label":"访客","value":"4"},{"label":"编辑","value":"3"},{"label":"超级管理员","value":"1"}]',
       (SELECT id FROM config_groups WHERE name = '安全' LIMIT 1)
WHERE NOT EXISTS (SELECT 1 FROM system_configs WHERE `key` = 'default_register_role_id');
