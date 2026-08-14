-- Agent 聊天模块回滚：删表 + 撤权限/菜单

DELETE FROM role_menus WHERE role_id IN (1,3,4) AND menu_id = 24;
DELETE FROM role_permissions WHERE role_id IN (1,3,4) AND permission_id = 41;
DELETE FROM permissions WHERE id = 41;
DELETE FROM menus WHERE id = 24;

DROP TABLE IF EXISTS agent_messages;
DROP TABLE IF EXISTS agent_conversations;
