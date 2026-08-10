import { theme as antdTheme } from 'antd';

/**
 * 便捷获取 Ant Design 设计 token（colorPrimary / colorBgContainer / colorTextSecondary...）。
 * 必须在 ConfigProvider 内使用；主题切换后自动响应，避免各页面写死颜色。
 */
export function useThemeToken() {
  return antdTheme.useToken().token;
}
