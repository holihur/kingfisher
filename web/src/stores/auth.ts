import { create } from 'zustand';
import { authApi } from '../api/auth';
import { userApi } from '../api/user';

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
  token: localStorage.getItem('kingfisher_token'),
  refreshToken: localStorage.getItem('kingfisher_refresh'),
  userInfo: null,
  permissions: [],
  isLoggedIn: !!localStorage.getItem('kingfisher_token'),
  login: async (username, password) => {
    const resp = await authApi.login({ username, password });
    const { access_token, refresh_token, user } = resp.data as Record<string, unknown>;
    localStorage.setItem('kingfisher_token', access_token as string);
    localStorage.setItem('kingfisher_refresh', refresh_token as string);
    set({
      token: access_token as string,
      refreshToken: refresh_token as string,
      userInfo: user as Record<string, unknown>,
      isLoggedIn: true,
    });
  },
  logout: () => {
    localStorage.removeItem('kingfisher_token');
    localStorage.removeItem('kingfisher_refresh');
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
