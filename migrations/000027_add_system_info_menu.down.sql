-- 撤销系统信息菜单 + 权限
DELETE FROM role_menus WHERE menu_id = 21;
DELETE FROM menus WHERE id = 21;
DELETE FROM role_permissions WHERE permission_id = 32;
DELETE FROM permissions WHERE id = 32;
