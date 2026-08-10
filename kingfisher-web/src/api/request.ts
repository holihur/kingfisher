import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { getToken, getRefreshToken, setTokens, clearTokens } from '../utils/token';
import { getMessage } from '../utils/feedback';

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
});

let isRefreshing = false;
let pendingRequests: Array<(token: string) => void> = [];

// 自动重试配置：后端临时不可达（网络错误/502/503/504）时按指数退避重试
const MAX_RETRY = 2; // 初始请求 + 2 次重试
const RETRY_DELAY = 800; // 首次退避间隔 ms，之后翻倍

// 可安全重试的请求方法（幂等；写操作不重试，避免重复提交）
// 注意：axios 内部将 method 统一转为小写，这里用小写匹配
const RETRYABLE_METHODS = new Set(['get', 'head', 'put', 'delete']);

// 标记已重试过，避免递归重试
declare module 'axios' {
  export interface InternalAxiosRequestConfig {
    __retryCount?: number;
  }
}

function shouldRetry(method?: string): boolean {
  return method ? RETRYABLE_METHODS.has(method) : false;
}

function retryDelay(count: number): number {
  return RETRY_DELAY * 2 ** (count - 1);
}

request.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = getToken();
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

request.interceptors.response.use(
  (response) => {
    const { code, message: msg } = response.data;
    if (code === 0) return response.data;
    switch (code) {
      case 10104:
        return handleTokenRefresh(response.config);
      case 10105:
      case 10003:
        clearTokens();
        window.location.href = '/login';
        return Promise.reject(new Error('登录已过期'));
      default:
        getMessage().error(msg || `请求失败 [${code}]`);
        return Promise.reject(new Error(msg || `请求失败 [${code}]`));
    }
  },
  (error: AxiosError<{ message?: string }>) => {
    // 可重试场景：网络错误（后端不可达）或 502/503/504，且方法幂等、未超过最大重试
    const config = error.config as InternalAxiosRequestConfig | undefined;
    const retryable =
      config &&
      shouldRetry(config.method) &&
      ((!error.response) || (error.response.status >= 502 && error.response.status <= 504)) &&
      (config.__retryCount ?? 0) < MAX_RETRY;

    if (retryable) {
      const count = (config.__retryCount ?? 0) + 1;
      config.__retryCount = count;
      const delay = retryDelay(count);
      return new Promise((resolve) => {
        setTimeout(() => resolve(request(config)), delay);
      });
    }

    if (!error.response) {
      getMessage().error('网络异常，请稍后重试');
      return Promise.reject(error);
    }
    // 优先使用后端返回的错误信息
    const backendMsg = error.response.data?.message;
    const statusMsgs: Record<number, string> = {
      400: '请求参数错误',
      401: '未授权',
      403: '无权限',
      404: '资源不存在',
      429: '请求过于频繁',
      500: '服务器内部错误',
      502: '网络异常，请稍后重试',
      503: '网络异常，请稍后重试',
      504: '网络异常，请稍后重试',
    };
    getMessage().error(backendMsg || statusMsgs[error.response.status] || `网络异常，请稍后重试`);
    return Promise.reject(error);
  }
);

async function handleTokenRefresh(config: InternalAxiosRequestConfig) {
  if (!isRefreshing) {
    isRefreshing = true;
    try {
      const rf = getRefreshToken();
      const resp = await axios.post('/api/v1/auth/refresh', { refresh_token: rf });
      const token = resp.data.data.access_token as string;
      setTokens(token, rf || '');
      pendingRequests.forEach((cb) => cb(token));
      pendingRequests = [];
      config.headers.Authorization = `Bearer ${token}`;
      return request(config);
    } catch {
      clearTokens();
      pendingRequests.length = 0;
      window.location.href = '/login';
      return Promise.reject(new Error('登录已过期'));
    } finally {
      isRefreshing = false;
    }
  }
  return new Promise((resolve) => {
    pendingRequests.push((token: string) => {
      config.headers.Authorization = `Bearer ${token}`;
      resolve(request(config));
    });
  });
}

export default request;
