-- 回滚：恢复 users.role_id
ALTER TABLE users ADD COLUMN role_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER status;

-- 取每个用户的第一个角色回填
UPDATE users u SET role_id = COALESCE((SELECT ur.role_id FROM user_roles ur WHERE ur.user_id = u.id LIMIT 1), 0);

DROP TABLE IF EXISTS user_roles;
