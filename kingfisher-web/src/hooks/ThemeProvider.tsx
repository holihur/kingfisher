import React, { useState, useCallback, useEffect } from 'react';
import { ThemeContext } from './useTheme';

type Theme = 'light' | 'dark';

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setTheme] = useState<Theme>(() => {
    const saved = localStorage.getItem('kingfisher_theme');
    return (saved === 'dark' ? 'dark' : 'light');
  });

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('kingfisher_theme', theme);
  }, [theme]);

  const toggle = useCallback(() => setTheme((t) => (t === 'light' ? 'dark' : 'light')), []);

  return <ThemeContext value={{ theme, toggle }}>{children}</ThemeContext>;
}
