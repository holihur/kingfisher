-- 字典类型/条目：新增版本字段
ALTER TABLE dict_types ADD COLUMN version VARCHAR(32) NOT NULL DEFAULT '';
ALTER TABLE dict_entries ADD COLUMN version VARCHAR(32) NOT NULL DEFAULT '';

UPDATE dict_types SET version = '1.0.0' WHERE version = '';
UPDATE dict_entries SET version = '1.0.0' WHERE version = '';
