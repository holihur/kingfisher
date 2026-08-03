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
  login: async (username, password) => {
    const resp = await authApi.login({ username, password });
    const { access_token, refresh_token, user } = resp.data as Record<string, unknown>;
    setTokens(access_token as string, refresh_token as string);
    set({
      token: access_token as string,
      refreshToken: refresh_token as string,
      userInfo: user as Record<string, unknown>,
      isLoggedIn: true,
    });
  },
  logout: () => {
    clearTokens();
    set({ token: null, refreshToken: null, userInfo: null, permissions: [], isLoggedIn: false });
  },
  fetchUserInfo: async () => {
    try {
      const userResp = await userApi.getMe();
      const permResp = await userApi.getMyPermissions();
      set({ userInfo: userResp.data as Record<string, unknown>, permissions: (permResp.data as string[]) || [] });
    } catch {
      /* ignore */
    }
  },
}));
