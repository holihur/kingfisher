import { App } from 'antd';
import type { MessageInstance } from 'antd/es/message/interface';
import { useEffect } from 'react';

/**
 * 全局消息单例：供非 React 环境（如 axios 拦截器）使用 antd App 上下文的消息实例。
 *
 * 在 <App> 内挂载 <GlobalFeedback/> 后，getMessage() 返回可用的 message。
 * 未挂载时降级为 console，避免崩溃。
 */

let messageInstance: MessageInstance | null = null;

export function getMessage(): MessageInstance {
  if (messageInstance) return messageInstance;
  // 降级：无 App 上下文时用 console 模拟
  const fallback: Record<string, unknown> = {
    success: (content: unknown) => { console.info('[success]', content); return Promise.resolve(content); },
    error: (content: unknown) => { console.error('[error]', content); return Promise.resolve(content); },
    warning: (content: unknown) => { console.warn('[warn]', content); return Promise.resolve(content); },
    info: (content: unknown) => { console.info('[info]', content); return Promise.resolve(content); },
    loading: () => { throw new Error('feedback not ready'); },
    open: () => { throw new Error('feedback not ready'); },
    destroy: () => {},
  };
  return fallback as unknown as MessageInstance;
}

/** 在 <App> 组件内挂载，把 App.useApp() 的 message 存入单例。 */
export function GlobalFeedback() {
  const { message } = App.useApp();
  useEffect(() => {
    messageInstance = message;
    return () => {
      messageInstance = null;
    };
  }, [message]);
  return null;
}
