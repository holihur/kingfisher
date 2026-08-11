-- 文档管理：版本历史表（append-only，每次保存追加一条；UNIQUE(doc_id, version_no) 兜底并发冲突）
CREATE TABLE IF NOT EXISTS doc_versions (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    doc_id     BIGINT UNSIGNED NOT NULL COMMENT '文档 ID',
    version_no INT             NOT NULL COMMENT '文档内递增版本号：1,2,3…',
    title      VARCHAR(128)    NOT NULL COMMENT '该版本标题快照',
    content    TEXT            NULL COMMENT '该版本 HTML 快照',
    owner_id   BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '保存人用户 ID',
    note       VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '变更说明',
    created_at DATETIME,
    UNIQUE KEY uk_doc_versions_doc_no (doc_id, version_no),
    CONSTRAINT fk_doc_versions_doc FOREIGN KEY (doc_id) REFERENCES documents(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文档版本历史表';
