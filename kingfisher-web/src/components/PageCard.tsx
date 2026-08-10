import React from 'react';
import { Card } from 'antd';
import type { CardProps } from 'antd';
import { useThemeToken } from '../hooks/useThemeToken';

/**
 * 统一页面卡片：圆角 / 无边框 / 柔和阴影全部取自设计 token，
 * 自动跟随浅色/暗色主题，替代各页面重复的
 * { borderRadius: 8, border: 'none', boxShadow: '0 1px 2px rgba(0,0,0,0.04)' } 魔数。
 * 透传 antd Card 全部 props；传入的 style 会合并覆盖默认样式。
 */
const PageCard: React.FC<CardProps> = ({ style, ...rest }) => {
  const token = useThemeToken();
  return (
    <Card
      {...rest}
      style={{
        borderRadius: token.borderRadiusLG,
        border: 'none',
        boxShadow: token.boxShadowTertiary,
        ...style,
      }}
    />
  );
};

export default PageCard;
