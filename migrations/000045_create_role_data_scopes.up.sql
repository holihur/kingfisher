CREATE TABLE role_data_scopes (
    role_id BIGINT NOT NULL,
    resource VARCHAR(64) NOT NULL,
    scope_type VARCHAR(32) NOT NULL,
    PRIMARY KEY (role_id, resource)
);

INSERT INTO role_data_scopes (role_id, resource, scope_type) VALUES
  (1, 'worktask', 'all'),
  (3, 'worktask', 'self'),
  (4, 'worktask', 'self');
