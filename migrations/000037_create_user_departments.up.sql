-- 部门管理：user_departments 用户-部门关联表（多对多，一个用户可属于多个部门）
CREATE TABLE IF NOT EXISTS user_departments (
    user_id BIGINT UNSIGNED NOT NULL,
    department_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (user_id, department_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
