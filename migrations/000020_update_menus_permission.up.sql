-- 菜单补 permission 字段（供前端动态路由权限校验）
UPDATE menus SET permission = 'user:list'  WHERE `path` = '/system/users';
UPDATE menus SET permission = 'menu:list'  WHERE `path` = '/system/menus';
UPDATE menus SET permission = 'role:list'  WHERE `path` = '/system/roles';
UPDATE menus SET permission = 'config:list' WHERE `path` = '/system/configs';
UPDATE menus SET permission = 'audit:list' WHERE `path` = '/system/audit';
UPDATE menus SET permission = 'dict:list'  WHERE `path` = '/system/dicts';
