import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'node:path';
import pkg from './package.json' with { type: 'json' };

export default defineConfig({
  plugins: [react()],
  define: {
    // 注入前端版本（构建时从 package.json 读取）
    __APP_VERSION__: JSON.stringify(pkg.version),
  },
  resolve: { alias: { '@': path.resolve(import.meta.dirname, 'src') } },
  server: {
    port: 5173,
    open: true,
    proxy: { '/api': { target: process.env.VITE_API_TARGET || 'http://localhost:8080', changeOrigin: true }, '/swagger': { target: process.env.VITE_API_TARGET || 'http://localhost:8080', changeOrigin: true }, '/uploads': { target: process.env.VITE_API_TARGET || 'http://localhost:8080', changeOrigin: true } },
  },
});
