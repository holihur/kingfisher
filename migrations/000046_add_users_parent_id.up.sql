ALTER TABLE users ADD COLUMN parent_id BIGINT NULL;
CREATE INDEX idx_users_parent_id ON users(parent_id);
