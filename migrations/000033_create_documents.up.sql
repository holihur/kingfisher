-- 文档管理：文档表（内容存 Quill 输出的 HTML）
CREATE TABLE IF NOT EXISTS documents (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    dir_id          BIGINT UNSIGNED NOT NULL COMMENT '所属目录 ID',
    title           VARCHAR(128)    NOT NULL COMMENT '文档标题',
    content         TEXT            NULL COMMENT 'Quill 输出的 HTML（当前最新内容）',
    owner_id        BIGINT UNSIGNED NOT NULL COMMENT '作者用户 ID',
    visibility      VARCHAR(16)     NOT NULL DEFAULT 'shared' COMMENT 'shared | private',
    status          VARCHAR(16)     NOT NULL DEFAULT 'draft' COMMENT 'draft | published',
    current_version INT             NOT NULL DEFAULT 1 COMMENT '当前指向 doc_versions 的最新版本号',
    sort            INT             NOT NULL DEFAULT 0,
    published_at    DATETIME        NULL COMMENT '最近一次发布时间',
    created_at      DATETIME,
    updated_at      DATETIME,
    INDEX idx_docs_dir (dir_id),
    INDEX idx_docs_owner (owner_id),
    INDEX idx_docs_status (status),
    CONSTRAINT fk_docs_dir FOREIGN KEY (dir_id) REFERENCES doc_directories(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文档表';
