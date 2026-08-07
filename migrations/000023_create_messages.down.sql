-- 回滚：站内信
DELETE FROM role_menus WHERE role_id = 1 AND menu_id = 18;
DELETE FROM role_permissions WHERE role_id = 1 AND permission_id IN (20, 21, 22, 23);
DELETE FROM permissions WHERE id IN (20, 21, 22, 23);
DELETE FROM menus WHERE id = 18;
DROP TABLE IF EXISTS messages;
