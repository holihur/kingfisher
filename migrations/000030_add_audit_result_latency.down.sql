-- 撤销审计增强列
ALTER TABLE audit_logs
    DROP INDEX idx_action,
    DROP INDEX idx_result,
    DROP COLUMN message,
    DROP COLUMN latency,
    DROP COLUMN result;
