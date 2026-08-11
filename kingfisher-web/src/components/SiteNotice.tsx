import React, { useEffect, useState } from 'react';
import { Alert } from 'antd';
import { configApi } from '../api/config';

// 已关闭通知的存储 key + 内容 hash 对比：
// 关闭后记住内容 hash，仅当通知内容变化时才重新展示。
const DISMISS_KEY = 'site_notice_dismissed_hash';

/** 简单内容 hash（足够区分内容是否变化） */
function hash(str: string): string {
  let h = 0;
  for (let i = 0; i < str.length; i++) {
    h = (h * 31 + str.charCodeAt(i)) | 0;
  }
  return String(h);
}

/**
 * 站点通知：显示在页面顶部，未登录也可见（公开配置）。
 * 可关闭；内容变化后重新展示（localStorage 对比 hash）。
 */
const SiteNotice: React.FC = () => {
  const [notice, setNotice] = useState('');

  useEffect(() => {
    configApi
      .getPublic('site_notice')
      .then((r) => setNotice(((r.data as Record<string, unknown>)?.value as string) || ''))
      .catch(() => {});
  }, []);

  // 内容为空 或 已关闭过相同内容 → 不显示
  if (!notice) return null;
  const currentHash = hash(notice);
  if (localStorage.getItem(DISMISS_KEY) === currentHash) return null;

  return (
    <Alert
      type="info"
      banner
      closable
      message={notice}
      style={{ borderRadius: 0 }}
      onClose={() => localStorage.setItem(DISMISS_KEY, currentHash)}
    />
  );
};

export default SiteNotice;
