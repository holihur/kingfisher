-- Agent 聊天模块：会话/消息表 + 菜单 + 权限 + 角色授权

-- Agent 会话表
CREATE TABLE IF NOT EXISTS agent_conversations (
    id          INTEGER PRIMARY KEY AUTO_INCREMENT,
    user_id     INT NOT NULL,
    title       VARCHAR(128) NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_agent_conversations_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Agent 消息表（role: user/assistant/tool）
CREATE TABLE IF NOT EXISTS agent_messages (
    id               BIGINT PRIMARY KEY AUTO_INCREMENT,
    conversation_id  INT NOT NULL,
    role             VARCHAR(16) NOT NULL,
    content          TEXT,
    tool_calls       TEXT,
    tool_result      TEXT,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_agent_messages_conversation (conversation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Agent 顶级菜单（parent_id=0，Sort 2 位于仪表盘与系统管理之间）
INSERT INTO menus (id, parent_id, name, path, component, icon, sort, permission, status, version, created_at, updated_at)
VALUES (24, 0, 'Agent 助手', '/agent', 'pages/Agent/AgentChat', 'MessageOutlined', 2, 'agent:list', 1, '1.0.0', NOW(), NOW());

-- Agent 权限
INSERT INTO permissions (id, name, code, resource, action, created_at, updated_at) VALUES
  (41, '使用Agent', 'agent:list', 'agent', 'read', NOW(), NOW());

-- 角色授权：所有登录角色均可使用 Agent（工具实际权限由 RBAC 兜底）
INSERT INTO role_permissions (role_id, permission_id) VALUES
  (1, 41), (3, 41), (4, 41);
INSERT INTO role_menus (role_id, menu_id) VALUES
  (1, 24), (3, 24), (4, 24);
