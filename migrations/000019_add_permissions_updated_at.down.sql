-- 回滚：权限更新时间字段
ALTER TABLE permissions
    DROP COLUMN updated_at;
