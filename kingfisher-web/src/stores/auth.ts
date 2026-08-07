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
  /** 用户信息/权限是否已加载（刷新后先拉取再渲染路由，避免权限误判） */
  userLoaded: boolean;
  /** 角色落地页（登录后跳转的页面） */
  landingPage: string;
  login: (u: string, p: string) => Promise<void>;
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
    const resp = await authApi.login({ username, password });
    const { access_token, refresh_token, user, landing_page } = resp.data as Record<string, unknown>;
    setTokens(access_token as string, refresh_token as string);
    set({
      token: access_token as string,
      refreshToken: refresh_token as string,
      userInfo: user as Record<string, unknown>,
      isLoggedIn: true,
      landingPage: (landing_page as string) || '',
      // 注意：不设 userLoaded=true。权限需由 AuthGuard 的 fetchUserInfo 拉取完成后
      // 才放行渲染子路由，避免 PermGuard 在空权限下误判 403。
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
