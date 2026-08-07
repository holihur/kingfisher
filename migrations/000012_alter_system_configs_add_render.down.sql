-- 回滚：系统配置的 render / render_options 字段
ALTER TABLE system_configs
    DROP COLUMN render,
    DROP COLUMN render_options;
