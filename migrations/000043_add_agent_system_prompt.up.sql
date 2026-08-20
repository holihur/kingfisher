-- Agent 配置分组 + 系统提示词 + 方法白名单配置

-- Agent 配置分组（ID 3）
INSERT INTO config_groups (id, name, sort, created_at, updated_at) VALUES
  (3, 'Agent', 3, NOW(), NOW());

INSERT INTO system_configs (id, `key`, value, remark, is_public, version, render, render_options, group_id, created_at, updated_at) VALUES
  (NULL, 'agent_system_prompt', '', 'Agent 系统提示词（留空用默认；可覆盖以定制 agent 行为）', 1, '1.0.0', 'textarea', '', 3, NOW(), NOW()),
  (NULL, 'agent_allowed_methods', 'GET,POST,PUT,PATCH,DELETE', 'Agent call_api 允许的 HTTP 方法白名单（多选；重启后生效）', 1, '1.0.0', 'select', '{"multiple":true,"options":[{"label":"GET","value":"GET"},{"label":"POST","value":"POST"},{"label":"PUT","value":"PUT"},{"label":"PATCH","value":"PATCH"},{"label":"DELETE","value":"DELETE"}]}', 3, NOW(), NOW());
