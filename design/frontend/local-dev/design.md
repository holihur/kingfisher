# Local Dev — 本地开发联调配置

## 问题

前端开发时（`localhost:5173`）调用后端（`localhost:8080`），跨域 + 每次改后端都要重启前端。

## 方案：Vite Proxy + 环境变量分层

```
浏览器 :5173  →  Vite Dev Server  →  proxy /api/*  →  Go :8080
                  (WebSocket HMR)     (免跨域)
```

## Vite 配置

```ts
// vite.config.ts
import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig(({ mode }) => {
    const env = loadEnv(mode, process.cwd(), '');

    return {
        plugins: [react()],
        resolve: {
            alias: {
                '@': path.resolve(__dirname, 'src'),
            },
        },
        server: {
            port: 5173,
            open: true,
            proxy: {
                '/api': {
                    target: env.VITE_API_TARGET || 'http://localhost:8080',
                    changeOrigin: true,
                    // WebSocket 代理（如需 dev 环境实时通知）
                    // ws: true,
                },
                // Swagger 资源
                '/swagger': {
                    target: env.VITE_API_TARGET || 'http://localhost:8080',
                    changeOrigin: true,
                },
            },
        },
        build: {
            outDir: 'dist',
            sourcemap: mode !== 'production',
        },
    };
});
```

## 环境变量

```bash
# .env.development — 本地开发
VITE_API_TARGET=http://localhost:8080

# .env.staging — 联调/预发
VITE_API_TARGET=https://staging-api.example.com

# .env.production — 生产
VITE_API_TARGET=https://api.example.com
```

## Axios baseURL 配置

```tsx
// api/request.ts
import axios from 'axios';

const request = axios.create({
    // 开发环境通过 Vite proxy 转发，baseURL 不需要设完整 URL
    // 生产环境由 nginx 反代，同样不需要
    baseURL: '/api/v1',
    timeout: 15000,
    headers: { 'Content-Type': 'application/json' },
});
```

这样开发和生产都只写 `/api/v1/users`，不需要环境判断。

## 两种开发模式

### 模式 A：前后端各开一个终端（推荐）

```
终端 1: cd kingfisher        && make run                  # Go :8080
终端 2: cd kingfisher-web    && npm run dev               # Vite :5173
浏览器: http://localhost:5173  →  proxy → :8080
```

### 模式 B：docker-compose 一键启动

```yaml
# docker-compose.dev.yaml
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: devpass
      MYSQL_DATABASE: kingfisher
    ports: ["3306:3306"]

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]

  backend:
    build:
      context: ./kingfisher
      dockerfile: Dockerfile
    ports: ["8080:8080"]
    environment:
      MYSQL_HOST: mysql
      REDIS_HOST: redis
      JWT_SECRET: dev-secret
    volumes:
      - ./kingfisher:/app              # 热重载（需 air）
    depends_on: [mysql, redis]

  frontend:
    image: node:20-alpine
    working_dir: /app
    command: sh -c "npm install && npm run dev -- --host"
    ports: ["5173:5173"]
    volumes:
      - ./kingfisher-web:/app
    environment:
      VITE_API_TARGET: http://backend:8080
    depends_on: [backend]
```

```bash
# 一键启动全部
docker-compose -f docker-compose.dev.yaml up -d
```

## 后端 CORS 配置（开发环境）

```yaml
# config/config.dev.yaml
cors:
  allowed_origins:
    - http://localhost:5173
    - http://127.0.0.1:5173
  allowed_methods: ["GET","POST","PUT","DELETE","OPTIONS","PATCH"]
  allowed_headers: ["Authorization","Content-Type","X-Request-ID"]
  allow_credentials: true
  max_age: 12h
```

> 使用 Vite proxy 时 CORS 不触发（同源），但保留配置作为兜底。

## 后端热重载（air）

开发时修改 Go 代码后自动重启：

```bash
# 安装
go install github.com/air-verse/air@latest

# .air.toml（项目根目录）
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/server ./cmd/server"
  bin = "./tmp/server"
  exclude_dir = ["tmp", "vendor", "test", "design", "node_modules"]
  include_ext = ["go", "yaml"]

# 启动
air
```

## nginx 生产反代（参考）

```nginx
server {
    listen 80;
    server_name admin.example.com;

    # 前端静态文件
    root /app/dist;
    index index.html;

    location / {
        try_files $uri /index.html;   # SPA fallback
    }

    # API 反代到后端
    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 30s;
    }
}
```

## 常见联调问题速查

| 问题 | 原因 | 解决 |
|------|------|------|
| 401 Unauthorized | Token 未注入 | 检查 Axios 拦截器 + localStorage |
| CORS error（开发） | Vite proxy 未生效 | 检查 `vite.config.ts` proxy 配置 |
| CORS error（无 proxy） | 后端 CORS 未放行 origin | 添加 `config.dev.yaml` 的 cors.allowed_origins |
| 接口 404 | baseURL 写死 `/api/v1` 但后端没有 | 检查 `VITE_API_TARGET` + 后端路由注册 |
| 登录后白屏 | 路由守卫跳转死循环 | 检查 `AuthGuard` 和 `useAuthStore.isLoggedIn` |
| 修改后端不生效 | 未用 air | 安装 air 或手动 `go run` |
| 前端改了不刷新 | Vite HMR 未连接 | 检查浏览器 console 是否有 WebSocket 错误 |

## 设计要点

- Vite proxy 是首选方案——零跨域问题、cookie 自然携带
- 环境变量不提交含密码的值，`.env.development` 只放 localhost
- 生产用 nginx 反代，同一域名下前后端同源
- `docker-compose.dev.yaml` 用于新人 onboarding——clone 后 `docker-compose -f docker-compose.dev.yaml up -d` 即开即用
