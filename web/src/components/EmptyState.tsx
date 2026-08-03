import React from 'react';
import { Empty, Button } from 'antd';

interface EmptyStateProps { type?: 'default' | 'search' | 'nodata'; description?: string; actionText?: string; onAction?: () => void; }

const EmptyState: React.FC<EmptyStateProps> = ({ type = 'default', description, actionText, onAction }) => {
  const config = {
    default: { description: '暂无数据' },
    search: { description: '没有找到匹配的结果，请尝试其他关键词' },
    nodata: { description: '还没有任何数据' },
  };
  return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={description || config[type].description} style={{ padding: '80px 0' }}>
    {actionText && onAction && <Button type="primary" onClick={onAction}>{actionText}</Button>}
  </Empty>;
};
export default EmptyState;
