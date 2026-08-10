import { create } from 'zustand';
import { menuApi } from '../api/menu';

interface MenuItem {
  id: number;
  parent_id: number;
  name: string;
  path: string;
  component: string;
  icon: string;
  sort: number;
  type: number;
  permission: string;
  status: number;
  children?: MenuItem[];
}

interface MenuState {
  menuTree: MenuItem[];
  loading: boolean;
  /** 是否已尝试过加载（成功或失败）；避免失败后无限等菜单导致全屏 loading 卡死 */
  loaded: boolean;
  fetchMenus: () => Promise<void>;
}
interface MenuItem {
  id: number;
  parent_id: number;
  name: string;
  path: string;
  component: string;
  icon: string;
  sort: number;
  type: number;
  permission: string;
  status: number;
  children?: MenuItem[];
}
export const useMenuStore = create<MenuState>((set) => ({
  menuTree: [],
  loading: false,
  loaded: false,
  fetchMenus: async () => {
    set({ loading: true });
    try {
      const resp = await menuApi.getMyTree();
      set({ menuTree: (resp.data as MenuItem[]) || [], loading: false, loaded: true });
    } catch {
      // 菜单加载失败不阻塞页面：loading 复位 + 标记已加载完成，避免无限转圈
      set({ menuTree: [], loading: false, loaded: true });
    }
  },
}));
