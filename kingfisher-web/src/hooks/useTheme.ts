import { createContext, useContext } from 'react';

type Theme = 'light' | 'dark';

interface ThemeCtx {
  theme: Theme;
  toggle: () => void;
}

export const ThemeContext = createContext<ThemeCtx>({ theme: 'light', toggle: () => {} });

export function useTheme() {
  return useContext(ThemeContext);
}
