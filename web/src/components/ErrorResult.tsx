import React from 'react';
import { Result, Button } from 'antd';

interface ErrorResultProps { status?: 'error' | '404' | '403' | '500'; message?: string; showRetry?: boolean; onRetry?: () => void; }

const ErrorResult: React.FC<ErrorResultProps> = ({ status = 'error', message, showRetry = true, onRetry }) => {
  const config: Record<string, { title: string; subTitle: string }> = {
    error: { title: '数据加载失败', subTitle: '服务器开小差了，请稍后再试' },
    '404': { title: '页面不存在', subTitle: '请检查 URL 是否正确' },
    '403': { title: '无权限访问', subTitle: '如需开通权限，请联系管理员' },
    '500': { title: '服务器错误', subTitle: '我们已经记录此问题，请稍后重试' },
  };
  const { title, subTitle } = config[status] || config.error;
  return <Result status={status === 'error' ? 'error' : status === '404' ? '404' : status === '403' ? '403' : '500'} title={title} subTitle={message || subTitle} extra={showRetry && onRetry ? <Button type="primary" onClick={onRetry}>重试</Button> : undefined} />;
};
export default ErrorResult;
