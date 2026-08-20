DELETE FROM role_menus WHERE menu_id = 25;
DELETE FROM role_permissions WHERE permission_id IN (42, 43, 44, 45);
DELETE FROM menus WHERE id = 25;
DELETE FROM permissions WHERE id IN (42, 43, 44, 45);
DROP TABLE tasks;
