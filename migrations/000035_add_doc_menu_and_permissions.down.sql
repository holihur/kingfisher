-- 文档管理：逆序回滚
DELETE FROM role_menus WHERE menu_id = 22;
DELETE FROM menus WHERE id = 22;
DELETE FROM role_permissions WHERE permission_id IN (33, 34, 35, 36);
DELETE FROM permissions WHERE id IN (33, 34, 35, 36);
DELETE FROM doc_versions WHERE doc_id IN (1, 2);
DELETE FROM documents WHERE id IN (1, 2);
DELETE FROM doc_dir_roles WHERE dir_id IN (1, 2, 3);
DELETE FROM doc_directories WHERE id IN (1, 2, 3);
