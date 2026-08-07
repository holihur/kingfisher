-- 系统配置：新增 渲染组件(render) 与 组件配置(render_options) 字段
ALTER TABLE system_configs
    ADD COLUMN render VARCHAR(32) NOT NULL DEFAULT '' COMMENT '前端渲染组件：text|number|switch|select|textarea',
    ADD COLUMN render_options TEXT COMMENT '渲染组件配置（JSON），如 select 的选项 [{"label":"开启","value":"1"}]';

-- 给现有种子配置设置默认渲染组件
UPDATE system_configs SET render = 'text'     WHERE `key` IN ('site_name', 'lockout_duration', 'session_timeout');
UPDATE system_configs SET render = 'textarea' WHERE `key` = 'site_description';
UPDATE system_configs SET render = 'number'   WHERE `key` = 'max_login_attempts';
