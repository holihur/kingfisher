-- 权限表：新增更新时间字段
ALTER TABLE permissions
    ADD COLUMN updated_at DATETIME(3) NULL;

-- 历史数据：更新时间用创建时间填充
UPDATE permissions SET updated_at = created_at WHERE updated_at IS NULL;
