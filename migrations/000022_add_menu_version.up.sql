-- 菜单：新增版本字段
ALTER TABLE menus ADD COLUMN version VARCHAR(32) NOT NULL DEFAULT '';

UPDATE menus SET version = '1.0.0' WHERE version = '';
