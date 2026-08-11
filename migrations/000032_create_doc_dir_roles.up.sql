-- 文档管理：目录-角色授权表（复合主键）
-- 语义（默认拒绝）：无授权记录 = 目录对所有人不可见（仅 admin 可见）；有记录 = 仅列出的角色可见
CREATE TABLE IF NOT EXISTS doc_dir_roles (
    dir_id  BIGINT UNSIGNED NOT NULL COMMENT '目录 ID',
    role_id BIGINT UNSIGNED NOT NULL COMMENT '角色 ID',
    PRIMARY KEY (dir_id, role_id),
    CONSTRAINT fk_doc_dir_roles_dir FOREIGN KEY (dir_id) REFERENCES doc_directories(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='目录-角色可见性授权表';
