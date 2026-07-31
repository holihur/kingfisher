# Migration — 数据库迁移 & 种子数据

## 职责

管理所有数据库结构变更（DDL），提供初始种子数据。使用 `golang-migrate`，迁移文件为纯 SQL。

## 迁移文件列表

```
migrations/
├── 000001_create_users.up.sql
├── 000001_create_users.down.sql
├── 000002_create_roles.up.sql
├── 000002_create_roles.down.sql
├── 000003_create_permissions.up.sql
├── 000003_create_permissions.down.sql
├── 000004_create_role_permissions.up.sql
├── 000004_create_role_permissions.down.sql
├── 000005_create_menus.up.sql
├── 000005_create_menus.down.sql
├── 000006_create_role_menus.up.sql
├── 000006_create_role_menus.down.sql
├── 000007_create_system_configs.up.sql
├── 000007_create_system_configs.down.sql
├── 000008_seed_data.up.sql          # 种子数据（预设用户/角色/权限/菜单/配置）
├── 000008_seed_data.down.sql        # 种子回滚（DELETE + 重置自增）
├── 000009_alter_users_add_session_version.up.sql    # 用户表加 session_version
├── 000009_alter_users_add_session_version.down.sql
├── 000010_create_audit_logs.up.sql                   # 审计日志表（SQL 详见 extends/audit/design.md）
└── 000010_create_audit_logs.down.sql
```

## 建表 SQL

### 000001 — users

```sql
CREATE TABLE users (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username   VARCHAR(32)  NOT NULL,
    password   VARCHAR(128) NOT NULL,
    email      VARCHAR(128) NOT NULL DEFAULT '',
    avatar     VARCHAR(255) NOT NULL DEFAULT '',
    status     TINYINT       NOT NULL DEFAULT 1 COMMENT '1=启用 0=禁用',
    role_id    BIGINT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3)   DEFAULT NULL,
    UNIQUE INDEX uk_username (username),
    INDEX idx_deleted_at (deleted_at),
    INDEX idx_role_id (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 000002 — roles

```sql
CREATE TABLE roles (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(32)  NOT NULL,
    code        VARCHAR(32)  NOT NULL,
    description VARCHAR(255) NOT NULL DEFAULT '',
    status      TINYINT      NOT NULL DEFAULT 1,
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE INDEX uk_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 000003 — permissions

```sql
CREATE TABLE permissions (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(64)  NOT NULL,
    code       VARCHAR(64)  NOT NULL,
    resource   VARCHAR(32)  NOT NULL,
    action     VARCHAR(16)  NOT NULL,
    created_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE INDEX uk_code (code),
    INDEX idx_resource (resource)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 000004 — role_permissions（多对多）

```sql
CREATE TABLE role_permissions (
    role_id       BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    INDEX idx_permission_id (permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 000005 — menus

```sql
CREATE TABLE menus (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    parent_id  BIGINT UNSIGNED NOT NULL DEFAULT 0,
    name       VARCHAR(32)  NOT NULL,
    path       VARCHAR(128) NOT NULL DEFAULT '',
    component  VARCHAR(128) NOT NULL DEFAULT '',
    icon       VARCHAR(64)  NOT NULL DEFAULT '',
    sort       INT          NOT NULL DEFAULT 0,
    type       TINYINT      NOT NULL DEFAULT 2 COMMENT '1=目录 2=菜单 3=按钮',
    permission VARCHAR(64)  NOT NULL DEFAULT '',
    status     TINYINT      NOT NULL DEFAULT 1,
    created_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    INDEX idx_parent_id (parent_id),
    INDEX idx_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 000006 — role_menus（多对多）

```sql
CREATE TABLE role_menus (
    role_id BIGINT UNSIGNED NOT NULL,
    menu_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (role_id, menu_id),
    INDEX idx_menu_id (menu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 000007 — system_configs

```sql
CREATE TABLE system_configs (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `key`      VARCHAR(64)  NOT NULL,
    value      TEXT         NOT NULL,
    remark     VARCHAR(255) NOT NULL DEFAULT '',
    created_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE INDEX uk_key (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

## 种子数据（000008）

```sql
-- 预设角色
INSERT INTO roles (id, name, code, description) VALUES
(1, '超级管理员', 'admin',    '拥有系统全部权限'),
(3, '编辑',        'editor',  '内容管理权限，可管理用户和内容'),
(4, '访客',        'viewer',  '只读权限，仅可查看');

-- 预设权限（全部 14 个）
INSERT INTO permissions (id, name, code, resource, action) VALUES
(1,  '查看用户',    'user:list',   'user',   'read'),
(2,  '创建用户',    'user:create', 'user',   'create'),
(3,  '更新用户',    'user:update', 'user',   'update'),
(4,  '删除用户',    'user:delete', 'user',   'delete'),
(5,  '查看菜单',    'menu:list',   'menu',   'read'),
(6,  '创建菜单',    'menu:create', 'menu',   'create'),
(7,  '更新菜单',    'menu:update', 'menu',   'update'),
(8,  '删除菜单',    'menu:delete', 'menu',   'delete'),
(9,  '查看角色',    'role:list',   'role',   'read'),
(10, '创建角色',    'role:create', 'role',   'create'),
(11, '更新角色',    'role:update', 'role',   'update'),
(12, '删除角色',    'role:delete', 'role',   'delete'),
(13, '查看配置',    'config:list', 'config', 'read'),
(14, '更新配置',    'config:update','config', 'update'),
(15, '查看审计',    'audit:list',  'audit',  'read');

-- 管理员拥有全部权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
(1,1),(1,2),(1,3),(1,4),(1,5),(1,6),(1,7),
(1,8),(1,9),(1,10),(1,11),(1,12),(1,13),(1,14),(1,15);

-- 编辑拥有用户+菜单查看编辑权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
(3,1),(3,2),(3,3),(3,5),(3,6),(3,7),(3,9),(3,13);

-- 访客只有查看权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
(4,1),(4,5),(4,9),(4,13);

-- 预设菜单
INSERT INTO menus (id, parent_id, name, path, component, icon, sort, type, permission) VALUES
(1,  0, 'Dashboard', '/dashboard', 'pages/Dashboard',          'DashboardOutlined', 0, 2, ''),
(2,  0, '系统管理',   '/system',    '',                         'SettingOutlined',   1, 1, ''),
(3,  2, '用户管理',   '/system/users', 'pages/User/UserList',   'UserOutlined',      1, 2, 'user:list'),
(4,  3, '新增用户',   '',             '',                       '',                  1, 3, 'user:create'),
(5,  3, '编辑用户',   '',             '',                       '',                  2, 3, 'user:update'),
(6,  3, '删除用户',   '',             '',                       '',                  3, 3, 'user:delete'),
(7,  2, '菜单管理',   '/system/menus', 'pages/Menu/MenuManage', 'MenuOutlined',      2, 2, 'menu:list'),
(8,  7, '新增菜单',   '',             '',                       '',                  1, 3, 'menu:create'),
(9,  7, '编辑菜单',   '',             '',                       '',                  2, 3, 'menu:update'),
(10, 7, '删除菜单',   '',             '',                       '',                  3, 3, 'menu:delete'),
(11, 2, '角色管理',   '/system/roles', 'pages/Role/RoleList',   'SafetyOutlined',    3, 2, 'role:list'),
(12, 11,'新增角色',   '',             '',                       '',                  1, 3, 'role:create'),
(13, 11,'编辑角色',   '',             '',                       '',                  2, 3, 'role:update'),
(14, 11,'删除角色',   '',             '',                       '',                  3, 3, 'role:delete'),
(15, 2, '系统配置',   '/system/configs','pages/Config/ConfigManage','ControlOutlined',4, 2, 'config:list');

-- 管理员看全部菜单
INSERT INTO role_menus (role_id, menu_id) VALUES
(1,1),(1,2),(1,3),(1,4),(1,5),(1,6),(1,7),(1,8),
(1,9),(1,10),(1,11),(1,12),(1,13),(1,14),(1,15);

-- 编辑看 Dashboard + 用户 + 菜单
INSERT INTO role_menus (role_id, menu_id) VALUES
(3,1),(3,3),(3,7);

-- 访客只看 Dashboard
INSERT INTO role_menus (role_id, menu_id) VALUES
(4,1);

-- 预设配置
INSERT INTO system_configs (`key`, value, remark) VALUES
('site_name',   'Kingfisher Admin', '系统名称，显示在登录页和页头'),
('site_logo',   '/logo.png',        '系统 Logo'),
('max_login_attempts', '5',         '最大登录失败次数'),
('lockout_duration',   '15m',       '登录失败锁定时间'),
('session_timeout',    '30m',       '会话超时时间');

-- 预设管理员用户（密码 Abcd1234 的 bcrypt hash）
-- cost=12, 生产部署后应修改密码
INSERT INTO users (id, username, password, email, role_id) VALUES
(1, 'admin', '$2a$12$LJ3m4ys3Lk0TSwHCpNqrIeN5U5Akn5dQUhBvPXFxFG7GqQvHCzB5q', 'admin@example.com', 1);
```

## 迁移命令

```makefile
migrate-up:     # 执行所有未应用的迁移
	go run ./cmd/migrate up

migrate-down:   # 回滚最后一个迁移
	go run ./cmd/migrate down

migrate-new:    # 创建新迁移文件 NAME=xxx
	migrate create -ext sql -dir migrations -seq $(NAME)
```

## 种子数据回滚（000008_seed_data.down.sql）

```sql
-- 按外键依赖倒序清理
DELETE FROM role_menus        WHERE role_id IN (1,3,4);
DELETE FROM role_permissions  WHERE role_id IN (1,3,4);
DELETE FROM users             WHERE id = 1;
DELETE FROM menus             WHERE id BETWEEN 1 AND 15;
DELETE FROM permissions       WHERE id BETWEEN 1 AND 15;
DELETE FROM roles             WHERE id IN (1,3,4);
DELETE FROM system_configs    WHERE `key` IN ('site_name','site_logo','max_login_attempts','lockout_duration','session_timeout');

-- 重置自增（可选，方便重新执行 seed 时 ID 一致）
ALTER TABLE users       AUTO_INCREMENT = 1;
ALTER TABLE roles       AUTO_INCREMENT = 1;
ALTER TABLE permissions AUTO_INCREMENT = 1;
ALTER TABLE menus       AUTO_INCREMENT = 1;
```

## 设计要点

- 迁移 ID 自增序号，不跳号，方便追踪
- 种子数据放在 `000008`，确保在所有建表之后执行
- 初始管理员密码用 bcrypt cost=12，密码 `Abcd1234`——**生产部署后立即修改**
- 角色 ID 1/3/4 留空 2，预留未来扩展
- PII 字段（email）可考虑加密存储，在 adapter 层加解密
