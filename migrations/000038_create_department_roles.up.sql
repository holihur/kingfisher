-- 部门管理：department_roles 部门-角色关联表（多对多，一个部门可有多个角色）
CREATE TABLE IF NOT EXISTS department_roles (
    department_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (department_id, role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
