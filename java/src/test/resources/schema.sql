-- 测试库 schema，复刻 Go core/database/models.go AutoMigrate
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(32) NOT NULL UNIQUE,
    nickname VARCHAR(64),
    password VARCHAR(128) NOT NULL,
    email VARCHAR(128),
    avatar VARCHAR(255),
    status INTEGER DEFAULT 1,
    session_version INTEGER DEFAULT 1,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY,
    name VARCHAR(32) NOT NULL,
    code VARCHAR(32) NOT NULL UNIQUE,
    description VARCHAR(255),
    status INTEGER DEFAULT 1,
    level INTEGER DEFAULT 2,
    landing_page VARCHAR(128),
    created_at DATETIME,
    updated_at DATETIME
);

CREATE TABLE IF NOT EXISTS permissions (
    id INTEGER PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL UNIQUE,
    resource VARCHAR(32) NOT NULL,
    action VARCHAR(16) NOT NULL,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE TABLE IF NOT EXISTS menus (
    id INTEGER PRIMARY KEY,
    parent_id INTEGER DEFAULT 0,
    name VARCHAR(32) NOT NULL,
    path VARCHAR(128),
    component VARCHAR(128),
    icon VARCHAR(64),
    sort INTEGER DEFAULT 0,
    type INTEGER DEFAULT 2,
    permission VARCHAR(64),
    status INTEGER DEFAULT 1,
    version VARCHAR(32),
    created_at DATETIME,
    updated_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_menus_parent_id ON menus(parent_id);

CREATE TABLE IF NOT EXISTS system_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key VARCHAR(64) NOT NULL UNIQUE,
    value TEXT NOT NULL,
    remark VARCHAR(255),
    is_public INTEGER DEFAULT 0,
    version VARCHAR(32),
    render VARCHAR(32),
    render_options TEXT,
    group_id INTEGER DEFAULT 0,
    created_at DATETIME,
    updated_at DATETIME
);
CREATE TABLE IF NOT EXISTS config_groups (
    id INTEGER PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE,
    sort INTEGER DEFAULT 0,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    username VARCHAR(32) NOT NULL,
    action VARCHAR(16) NOT NULL,
    resource VARCHAR(32) NOT NULL,
    resource_id INTEGER DEFAULT 0,
    detail TEXT,
    result VARCHAR(16) NOT NULL DEFAULT 'success',
    latency INTEGER DEFAULT 0,
    message VARCHAR(255),
    ip VARCHAR(45),
    user_agent VARCHAR(512),
    created_at DATETIME DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_result ON audit_logs(result);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

CREATE TABLE IF NOT EXISTS role_menus (
    role_id INTEGER NOT NULL,
    menu_id INTEGER NOT NULL,
    PRIMARY KEY (role_id, menu_id)
);
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);
CREATE TABLE IF NOT EXISTS role_data_scopes (
    role_id INTEGER NOT NULL,
    resource VARCHAR(64) NOT NULL,
    scope_type VARCHAR(32) NOT NULL,
    PRIMARY KEY (role_id, resource)
);
CREATE TABLE IF NOT EXISTS user_roles (
    user_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS dict_types (
    id INTEGER PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    is_public INTEGER DEFAULT 0,
    status INTEGER DEFAULT 1,
    remark VARCHAR(255),
    version VARCHAR(32),
    created_at DATETIME,
    updated_at DATETIME
);
CREATE TABLE IF NOT EXISTS dict_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type_id INTEGER NOT NULL,
    label VARCHAR(128) NOT NULL,
    value VARCHAR(128) NOT NULL,
    sort INTEGER DEFAULT 0,
    status INTEGER DEFAULT 1,
    remark VARCHAR(255),
    version VARCHAR(32),
    created_at DATETIME,
    updated_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_dict_entries_type_id ON dict_entries(type_id);

CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY,
    sender_id INTEGER,
    sender_type VARCHAR(16) DEFAULT 'admin',
    recipient_id INTEGER NOT NULL,
    batch_id INTEGER DEFAULT 0,
    title VARCHAR(128) NOT NULL,
    content TEXT,
    status VARCHAR(16) DEFAULT 'sent',
    is_read INTEGER DEFAULT 0,
    read_at DATETIME,
    created_at DATETIME,
    updated_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_recipient_id ON messages(recipient_id);
CREATE INDEX IF NOT EXISTS idx_messages_batch_id ON messages(batch_id);
CREATE INDEX IF NOT EXISTS idx_messages_is_read ON messages(is_read);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);

CREATE TABLE IF NOT EXISTS templates (
    id INTEGER PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL UNIQUE,
    template_type VARCHAR(32) DEFAULT 'general',
    title VARCHAR(255) NOT NULL,
    content TEXT,
    status INTEGER DEFAULT 1,
    remark VARCHAR(255),
    version VARCHAR(32),
    created_at DATETIME,
    updated_at DATETIME
);
CREATE TABLE IF NOT EXISTS scheduled_tasks (
    id INTEGER PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    task_type VARCHAR(64) NOT NULL,
    cron_spec VARCHAR(64) NOT NULL,
    payload TEXT,
    enabled INTEGER DEFAULT 1,
    remark VARCHAR(255),
    created_at DATETIME,
    updated_at DATETIME
);
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title VARCHAR(128) NOT NULL,
    description TEXT,
    owner_id INTEGER NOT NULL,
    department_id INTEGER NOT NULL,
    status VARCHAR(32) NOT NULL,
    created_at DATETIME,
    updated_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_tasks_owner_id ON tasks(owner_id);
CREATE INDEX IF NOT EXISTS idx_tasks_department_id ON tasks(department_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);

CREATE TABLE IF NOT EXISTS doc_directories (
    id INTEGER PRIMARY KEY,
    parent_id INTEGER DEFAULT 0,
    name VARCHAR(64) NOT NULL,
    sort INTEGER DEFAULT 0,
    status INTEGER DEFAULT 1,
    version VARCHAR(32),
    visibility VARCHAR(16) DEFAULT 'shared',
    created_at DATETIME,
    updated_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_doc_directories_parent_id ON doc_directories(parent_id);
CREATE TABLE IF NOT EXISTS doc_dir_roles (
    dir_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    PRIMARY KEY (dir_id, role_id)
);
CREATE TABLE IF NOT EXISTS documents (
    id INTEGER PRIMARY KEY,
    dir_id INTEGER NOT NULL,
    title VARCHAR(128) NOT NULL,
    content TEXT,
    owner_id INTEGER NOT NULL,
    visibility VARCHAR(16) DEFAULT 'shared',
    status VARCHAR(16) DEFAULT 'draft',
    current_version INTEGER DEFAULT 1,
    sort INTEGER DEFAULT 0,
    published_at DATETIME,
    created_at DATETIME,
    updated_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_documents_dir_id ON documents(dir_id);
CREATE INDEX IF NOT EXISTS idx_documents_owner_id ON documents(owner_id);
CREATE TABLE IF NOT EXISTS doc_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    doc_id INTEGER NOT NULL,
    version_no INTEGER NOT NULL,
    title VARCHAR(128) NOT NULL,
    content TEXT,
    owner_id INTEGER DEFAULT 0,
    note VARCHAR(255),
    created_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_doc_versions_doc_id ON doc_versions(doc_id);

CREATE TABLE IF NOT EXISTS departments (
    id INTEGER PRIMARY KEY,
    parent_id INTEGER DEFAULT 0,
    name VARCHAR(64) NOT NULL,
    sort INTEGER DEFAULT 0,
    status INTEGER DEFAULT 1,
    remark VARCHAR(255),
    version VARCHAR(32),
    created_at DATETIME,
    updated_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_departments_parent_id ON departments(parent_id);
CREATE TABLE IF NOT EXISTS user_departments (
    user_id INTEGER NOT NULL,
    department_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, department_id)
);
CREATE TABLE IF NOT EXISTS department_roles (
    department_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    PRIMARY KEY (department_id, role_id)
);
CREATE TABLE IF NOT EXISTS agent_conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title VARCHAR(128),
    created_at DATETIME,
    updated_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_agent_conversations_user_id ON agent_conversations(user_id);
CREATE TABLE IF NOT EXISTS agent_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL,
    role VARCHAR(16) NOT NULL,
    content TEXT,
    tool_calls TEXT,
    tool_result TEXT,
    created_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_agent_messages_conversation_id ON agent_messages(conversation_id);
