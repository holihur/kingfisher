import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { message } from 'antd';
import { getToken, getRefreshToken, setTokens, clearTokens } from '../utils/token';

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
});

let isRefreshing = false;
let pendingRequests: Array<(token: string) => void> = [];

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
        message.error(msg || `请求失败 [${code}]`);
        return Promise.reject(new Error(msg || `请求失败 [${code}]`));
    }
  },
  (error: AxiosError<{ message?: string }>) => {
    if (!error.response) {
      message.error('网络异常');
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
    };
    message.error(backendMsg || statusMsgs[error.response.status] || `服务器错误 (${error.response.status})`);
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
