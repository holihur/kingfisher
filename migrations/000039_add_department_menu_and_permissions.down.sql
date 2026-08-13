-- 部门管理：逆序回滚
DELETE FROM user_departments;
DELETE FROM department_roles;
DELETE FROM departments;
DELETE FROM role_menus WHERE menu_id = 23;
DELETE FROM menus WHERE id = 23;
DELETE FROM role_permissions WHERE permission_id IN (37, 38, 39, 40);
DELETE FROM permissions WHERE id IN (37, 38, 39, 40);
