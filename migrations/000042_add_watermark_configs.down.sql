-- 撤销全局水印配置
DELETE FROM system_configs WHERE `key` IN ('watermark_enabled', 'watermark_text', 'watermark_extra');
