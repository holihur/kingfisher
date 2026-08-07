-- 回滚：配置分组表 + group_id 字段
ALTER TABLE system_configs
    DROP COLUMN group_id;

DROP TABLE IF EXISTS config_groups;
