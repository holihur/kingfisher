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
  fetchMenus: async () => {
    set({ loading: true });
    const resp = await menuApi.getMyTree();
    set({ menuTree: (resp.data as MenuItem[]) || [], loading: false });
  },
}));
