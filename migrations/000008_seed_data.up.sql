-- Seed data for production MySQL
INSERT INTO roles (id, name, code, description, level) VALUES (1, '超级管理员', 'admin', '拥有全部权限', 0);
INSERT INTO permissions (id, name, code, resource, action) VALUES (1,'查看用户','user:list','user','read'),(15,'查看审计','audit:list','audit','read');
INSERT INTO role_permissions (role_id, permission_id) VALUES (1,1),(1,15);
INSERT INTO menus (id, parent_id, name, path, type) VALUES (1, 0, 'Dashboard', '/dashboard', 2);
INSERT INTO role_menus (role_id, menu_id) VALUES (1,1);
INSERT INTO system_configs (`key`, value, remark) VALUES ('site_name', 'Kingfisher Admin', '系统名称');
INSERT INTO users (id, username, password, email, role_id) VALUES (1, 'admin', '$2a$12$jDyI8HZp/TVxUrplIqdgNOV/iahF.i3l0YoPHuNLD5kus./WsPTzO', 'admin@example.com', 1);
