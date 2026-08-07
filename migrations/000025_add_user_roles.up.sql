-- 用户多角色：新增 user_roles 关联表，从 users.role_id 迁移数据
CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (user_id, role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 迁移现有单角色数据到 user_roles
INSERT IGNORE INTO user_roles (user_id, role_id)
SELECT id, role_id FROM users WHERE role_id != 0;

-- 删除旧的单角色列
ALTER TABLE users DROP COLUMN role_id;
