-- 回滚：删除 batch_id 列
DROP INDEX IF EXISTS idx_messages_batch_id;
ALTER TABLE messages DROP COLUMN batch_id;
