-- 撤销周期任务表 + 菜单 + 权限
DELETE FROM role_menus WHERE menu_id = 20;
DELETE FROM menus WHERE id = 20;
DELETE FROM role_permissions WHERE permission_id IN (28, 29, 30, 31);
DELETE FROM permissions WHERE id IN (28, 29, 30, 31);
DROP TABLE IF EXISTS scheduled_tasks;
