CREATE TABLE audit_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    username VARCHAR(32) NOT NULL,
    action VARCHAR(16) NOT NULL,
    resource VARCHAR(32) NOT NULL,
    resource_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    detail TEXT,
    ip VARCHAR(45) NOT NULL DEFAULT '',
    user_agent VARCHAR(512) NOT NULL DEFAULT '',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    INDEX idx_user_id (user_id),
    INDEX idx_resource (resource, resource_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;