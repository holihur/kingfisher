-- 配置分组表 + 系统配置关联 group_id
CREATE TABLE config_groups (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    sort INT NOT NULL DEFAULT 0,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE INDEX uk_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

ALTER TABLE system_configs
    ADD COLUMN group_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '配置分组（关联 config_groups.id）';

-- 默认分组：站点 + 安全
INSERT INTO config_groups (id, name, sort) VALUES (1, '站点', 1), (2, '安全', 2);

-- 关联种子配置到分组
UPDATE system_configs SET group_id = 1 WHERE `key` IN ('site_name', 'site_description', 'site_logo');
UPDATE system_configs SET group_id = 2 WHERE `key` IN ('max_login_attempts', 'lockout_duration', 'session_timeout');
