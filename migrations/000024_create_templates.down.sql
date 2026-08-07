-- 回滚：模版管理
DELETE FROM role_menus WHERE role_id = 1 AND menu_id = 19;
DELETE FROM role_permissions WHERE role_id = 1 AND permission_id IN (24, 25, 26, 27);
DELETE FROM permissions WHERE id IN (24, 25, 26, 27);
DELETE FROM menus WHERE id = 19;
DROP TABLE IF EXISTS templates;
