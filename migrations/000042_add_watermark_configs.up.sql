-- 全局水印配置（安全组 GroupID=2；公开可读，登录后所有页面显示水印）

INSERT INTO system_configs (id, `key`, value, remark, is_public, version, render, render_options, group_id, created_at, updated_at) VALUES
  (NULL, 'watermark_enabled', 'false', '全局水印开关（登录后所有页面显示水印）', 1, '1.0.0', 'switch', '', 2, NOW(), NOW()),
  (NULL, 'watermark_text', 'Kingfisher 内部系统', '水印文字', 1, '1.0.0', 'text', '', 2, NOW(), NOW()),
  (NULL, 'watermark_extra', '{username} {date}', '水印补充内容（支持 {username}/{date} 占位符，留空则仅水印文字）', 1, '1.0.0', 'text', '', 2, NOW(), NOW());
