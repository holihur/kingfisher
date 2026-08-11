-- 文档管理：目录表（树形，parent_id 自引用）
CREATE TABLE IF NOT EXISTS doc_directories (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    parent_id  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父目录 ID，0=顶级',
    name       VARCHAR(64)     NOT NULL COMMENT '目录名称',
    sort       INT             NOT NULL DEFAULT 0 COMMENT '排序',
    status     TINYINT(1)      NOT NULL DEFAULT 1 COMMENT '1=启用, 0=禁用',
    version    VARCHAR(32)     NOT NULL DEFAULT '1.0.0',
    created_at DATETIME,
    updated_at DATETIME,
    INDEX idx_doc_dirs_parent (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文档目录表';
