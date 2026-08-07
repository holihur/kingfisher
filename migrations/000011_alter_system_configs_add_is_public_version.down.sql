-- 回滚：系统配置的 is_public / version 字段
ALTER TABLE system_configs
    DROP COLUMN is_public,
    DROP COLUMN version;
