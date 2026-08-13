-- 站内信批次：同一次发送给多个收件人共用一个 batch_id，管理端列表按批次聚合
ALTER TABLE messages ADD COLUMN batch_id BIGINT DEFAULT 0;
CREATE INDEX idx_messages_batch_id ON messages (batch_id);
