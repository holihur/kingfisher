-- 回滚：字典版本字段
ALTER TABLE dict_types DROP COLUMN version;
ALTER TABLE dict_entries DROP COLUMN version;
