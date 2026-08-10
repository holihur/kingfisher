-- 站点通知（显示在页面顶部，可关闭；内容变化后重新展示）。公开可读，未登录也可见。
INSERT INTO system_configs (`key`, value, remark, is_public, version, render)
SELECT 'site_notice', '⚠️ 这是一个测试站点，请勿用于生产环境。管理员账号：admin / 密码：Abcd1234', '站点通知（显示在页面顶部，可关闭；内容变化后重新展示）', 1, '1.0.0', 'textarea'
WHERE NOT EXISTS (SELECT 1 FROM system_configs WHERE `key` = 'site_notice');
