-- 撤销 Agent 配置分组 + 系统提示词 + 方法白名单配置
DELETE FROM system_configs WHERE `key` IN ('agent_system_prompt', 'agent_allowed_methods');
DELETE FROM config_groups WHERE id = 3;
