import { create } from 'zustand';
import { authApi } from '../api/auth';
import { userApi } from '../api/user';
import { getToken, getRefreshToken, setTokens, clearTokens, hasToken } from '../utils/token';

interface AuthState {
  token: string | null;
  refreshToken: string | null;
  userInfo: Record<string, unknown> | null;
  permissions: string[];
  isLoggedIn: boolean;
  userLoaded: boolean;
  landingPage: string;
  login: (u: string, p: string) => Promise<void>;
  mfaVerify: (mfa_token: string, method: string, code: string) => Promise<void>;
  logout: () => void;
  fetchUserInfo: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set) => ({
  token: getToken(),
  refreshToken: getRefreshToken(),
  userInfo: null,
  permissions: [],
  isLoggedIn: hasToken(),
  userLoaded: false,
  landingPage: '',
  login: async (username, password) => {
    const resp: Record<string, unknown> = await authApi.login({ username, password }) as unknown as Record<string, unknown>;
    if ((resp as { code?: number }).code === 11201 || (resp as { code?: number }).code === 11206) {
      const data = (resp as { data?: Record<string, unknown> }).data || {};
      const err = new Error((resp as { code?: number }).code === 11206 ? 'mfa_setup_required' : 'mfa_required') as Error & { mfa_token?: string; methods?: string[]; code?: number };
      err.mfa_token = data.mfa_token as string;
      err.methods = (data.methods as string[]) || [];
      err.code = (resp as { code?: number }).code;
      throw err;
    }
    const { access_token, refresh_token, user, landing_page } = (resp.data as Record<string, unknown>) || (resp as Record<string, unknown>);
    if (!access_token) {
      const d = resp.data as Record<string, unknown> | undefined;
      if (d && (d as { mfa_token?: string }).mfa_token) {
        const err = new Error('mfa_required') as Error & { mfa_token?: string; methods?: string[] };
        err.mfa_token = (d as { mfa_token?: string }).mfa_token;
        err.methods = (d as { methods?: string[] }).methods;
        throw err;
      }
      throw new Error('登录失败');
    }
    setTokens(access_token as string, refresh_token as string);
    set({
      token: access_token as string,
      refreshToken: refresh_token as string,
      userInfo: user as Record<string, unknown>,
      isLoggedIn: true,
      landingPage: (landing_page as string) || '',
    });
  },
  mfaVerify: async (mfa_token: string, method: string, code: string) => {
    const resp = await authApi.mfaVerify({ mfa_token, method, code });
    const { access_token, refresh_token, user, landing_page } = resp.data as Record<string, unknown>;
    setTokens(access_token as string, refresh_token as string);
    set({
      token: access_token as string,
      refreshToken: refresh_token as string,
      userInfo: user as Record<string, unknown>,
      isLoggedIn: true,
      landingPage: (landing_page as string) || '',
    });
  },
  logout: () => {
    clearTokens();
    set({ token: null, refreshToken: null, userInfo: null, permissions: [], isLoggedIn: false, userLoaded: false, landingPage: '' });
  },
  fetchUserInfo: async () => {
    try {
      const userResp = await userApi.getMe();
      const permResp = await userApi.getMyPermissions();
      set({ userInfo: userResp.data as Record<string, unknown>, permissions: (permResp.data as string[]) || [] });
    } catch {
      /* ignore */
    } finally {
      // 无论成功与否都标记为已加载，避免一直转圈（失败时权限为空，由 PermGuard 处理）
      set({ userLoaded: true });
    }
  },
}));
