-- 审计日志增强：结果状态 + 耗时 + 失败原因
ALTER TABLE audit_logs
    ADD COLUMN result VARCHAR(16) NOT NULL DEFAULT 'success' COMMENT '结果：success | failure',
    ADD COLUMN latency INT NOT NULL DEFAULT 0 COMMENT '操作耗时（毫秒）',
    ADD COLUMN message VARCHAR(255) NOT NULL DEFAULT '' COMMENT '结果说明/失败原因',
    ADD INDEX idx_action (action),
    ADD INDEX idx_result (result);
