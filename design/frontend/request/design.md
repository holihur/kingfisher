# Frontend Request — Axios 封装

## 职责

Axios 实例配置、请求/响应拦截器、token 自动刷新、统一错误处理。

## 完整实现

```tsx
// api/request.ts
import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { message } from 'antd';
import { useAuthStore } from '@/stores/auth';

const request = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL,   // http://localhost:8080/api/v1
    timeout: 15000,
    headers: { 'Content-Type': 'application/json' },
});

// 是否正在刷新 token（防止并发请求同时刷新）
let isRefreshing = false;
let pendingRequests: Array<(token: string) => void> = [];

// ---- 请求拦截器 ----
request.interceptors.request.use((config: InternalAxiosRequestConfig) => {
    const token = useAuthStore.getState().token;   // 从 Zustand 取，不依赖 React hook
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

// ---- 响应拦截器 ----
request.interceptors.response.use(
    (response) => {
        const { code, message: msg } = response.data;

        if (code === 0) {
            return response.data;       // 成功——直接返回 data 层
        }

        // 业务错误
        switch (code) {
            case 10104:  // Token 过期——触发刷新
                return handleTokenRefresh(response.config);
            case 10105:  // Token 无效——强制登出
                useAuthStore.getState().logout();
                window.location.href = '/login';
                return Promise.reject(new Error('登录已过期，请重新登录'));
            case 10003:  // 未认证
                useAuthStore.getState().logout();
                window.location.href = '/login';
                return Promise.reject(new Error('请先登录'));
            default:
                message.error(msg || `请求失败 [${code}]`);
                return Promise.reject(new Error(msg));
        }
    },
    (error: AxiosError) => {
        // 网络错误
        if (!error.response) {
            message.error('网络异常，请检查网络连接');
        } else {
            const status = error.response.status;
            const messages: Record<number, string> = {
                400: '请求参数错误',
                401: '未授权',
                403: '无权限访问',
                404: '请求的资源不存在',
                429: '请求过于频繁，请稍后再试',
                500: '服务器内部错误',
            };
            message.error(messages[status] || `服务器错误 (${status})`);
        }
        return Promise.reject(error);
    }
);

// ---- Token 刷新逻辑 ----
async function handleTokenRefresh(failedConfig: InternalAxiosRequestConfig) {
    if (!isRefreshing) {
        isRefreshing = true;
        try {
            const newToken = await useAuthStore.getState().refreshAccessToken();
            // 重放等待中的请求
            pendingRequests.forEach(cb => cb(newToken));
            pendingRequests = [];
            // 重试当前请求
            failedConfig.headers.Authorization = `Bearer ${newToken}`;
            return request(failedConfig);
        } catch {
            useAuthStore.getState().logout();
            window.location.href = '/login';
            return Promise.reject(new Error('登录已过期'));
        } finally {
            isRefreshing = false;
        }
    }

    // 已有刷新在进行中，把请求加入等待队列
    return new Promise((resolve) => {
        pendingRequests.push((token: string) => {
            failedConfig.headers.Authorization = `Bearer ${token}`;
            resolve(request(failedConfig));
        });
    });
}

export default request;
```

## API 调用示例

```tsx
// api/auth.ts
import request from './request';

export const authApi = {
    login: (data: LoginReq) =>
        request.post<LoginReq, ApiResponse<LoginResp>>('/auth/login', data),

    register: (data: RegisterReq) =>
        request.post<RegisterReq, ApiResponse<null>>('/auth/register', data),

    refresh: (refreshToken: string) =>
        request.post<{ refresh_token: string }, ApiResponse<{ access_token: string }>>('/auth/refresh', { refresh_token: refreshToken }),
};

// api/user.ts
import type { ApiResponse, PaginatedData, User, UpdateUserReq } from '@/types/api.generated';
import request from './request';

export const userApi = {
    getList: (params: { page?: number; page_size?: number; keyword?: string }) =>
        request.get<ApiResponse<PaginatedData<User>>>('/users', { params }),

    getById: (id: number) =>
        request.get<ApiResponse<User>>(`/users/${id}`),

    getMe: () =>
        request.get<ApiResponse<User>>('/users/me'),

    getMyPermissions: () =>
        request.get<ApiResponse<string[]>>('/users/me/permissions'),

    create: (data: { username: string; password: string; email?: string; role_id: number }) =>
        request.post<ApiResponse<User>>('/users', data),

    update: (id: number, data: { email?: string; status?: number; role_id?: number }) =>
        request.put<ApiResponse<User>>(`/users/${id}`, data),

    delete: (id: number) =>
        request.delete<ApiResponse<null>>(`/users/${id}`),
};
```

## 类型定义

```tsx
// types/api.ts
export interface ApiResponse<T> {
    code: number;
    message: string;
    data: T;
}

export interface PaginatedData<T> {
    items: T[];
    total: number;
    page: number;
    page_size: number;
}

// types/user.ts
export interface User {
    id: number;
    username: string;
    email: string;
    avatar: string;
    status: number;
    created_at: string;
    updated_at: string;
}

export interface LoginReq { username: string; password: string; }
export interface LoginResp { access_token: string; refresh_token: string; user: User; }
export interface UserListResp extends PaginatedData<User> {}
```

## 设计要点

1. **Token 注入**：请求拦截器自动从 Zustand store 取 token（`getState()` 不依赖 React 上下文）
2. **并发刷新保护**：多个请求同时遇 401 时，只有第一个触发刷新，其余排队等待
3. **响应解包**：拦截器把 `{code, message, data}` 解包为直接返回 `data`，业务代码不再写 `.data.data`
4. **类型安全**：每个 API 函数的泛型 `<LoginReq, ApiResponse<LoginResp>>` 明确入参出参类型
5. **错误分层**：401/403 跳登录，429 提示限流，500 提示服务器错误，网络断开提示断网
