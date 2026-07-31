# Shared Types — 前后端类型共享 & 联调契约

## 问题

前后端各自定义类型，接口变更时两边不同步，联调才发现字段对不上。

## 方案：以 Swagger JSON 为单一事实来源，前端自动生成类型

```
后端 Go struct + swaggo 注解
         │
         ▼  make swagger
    docs/swagger.json        ← 单一事实来源（提交 git）
         │
         ▼  openapi-typescript / openapi-generator
    src/types/api.generated.ts   ← 自动生成，禁止手改
         │
         ▼  import
    前端 API 层使用
```

## 工具链

### 方案 A：openapi-typescript（推荐）

```bash
# 安装
npm i -D openapi-typescript

# package.json
{
  "scripts": {
    "gen-types": "openapi-typescript http://localhost:8080/swagger/doc.json -o src/types/api.generated.ts"
  }
}
```

### 方案 B：@hey-api/openapi-ts（更完整，同时生成 API 函数）

```bash
npm i -D @hey-api/openapi-ts

npx @hey-api/openapi-ts \
  --input http://localhost:8080/swagger/doc.json \
  --output src/api/generated \
  --client axios           # 生成 Axios 调用函数
```

两种方案对比：

| | openapi-typescript | @hey-api/openapi-ts |
|------|-------------------|---------------------|
| 产物 | 只有类型 | 类型 + API 函数 + 请求/响应类型 |
| 侵入性 | 低（只换类型） | 高（API 调用函数也被替换） |
| 推荐场景 | 初次尝试 | 想把 API 层也自动化 |

**本次选方案 A**（openapi-typescript），渐进式——先用自动生成的类型替换手写类型，API 函数保持手写。

## 后端 Swagger 注解要求

每个 handler 必须带完整的注解，产物才能生成准确的 TS 类型。

```go
// @Summary 用户列表
// @Tags User
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页条数" default(20)
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} response.Response{data=PaginatedUser}
// @Router /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) { ... }

// 涉及的 Swagger 类型定义
type PaginatedUser struct {
    Items    []User `json:"items"`
    Total    int64  `json:"total"`
    Page     int    `json:"page"`
    PageSize int    `json:"page_size"`
}
```

## 生成的 TS 类型示例

```typescript
// src/types/api.generated.ts
export interface User {
    id: number;
    username: string;
    email: string;
    avatar: string;
    /** @enum {integer} 1=启用 0=禁用 */
    status: number;
    role?: Role;
    created_at: string;  // RFC3339
    updated_at: string;
}

export interface PaginatedUser {
    items: User[];
    total: number;
    page: number;
    page_size: number;
}

export interface ApiResponse<T> {
    code: number;
    message: string;
    data: T;
}

// 请求类型
export interface LoginReq {
    username: string;
    password: string;
}

export interface LoginResp {
    access_token: string;
    refresh_token: string;
    user: User;
}
```

## 前端 API 层使用

```tsx
// api/user.ts
import type { ApiResponse, PaginatedUser, User, UpdateUserReq } from '@/types/api.generated';
import request from './request';

export const userApi = {
    getList: (params: { page?: number; page_size?: number; keyword?: string }) =>
        request.get<any, ApiResponse<PaginatedUser>>('/users', { params }),

    getById: (id: number) =>
        request.get<any, ApiResponse<User>>(`/users/${id}`),

    update: (id: number, data: UpdateUserReq) =>
        request.put<any, ApiResponse<User>>(`/users/${id}`, data),
};
```

## 类型校验（运行时）

如果 Swagger 注解不完整，运行时可能出现类型不匹配。加一层 zod 校验（可选）：

```tsx
import { z } from 'zod';

const UserSchema = z.object({
    id: z.number(),
    username: z.string(),
    email: z.string().email().optional(),
    status: z.number().min(0).max(1),
});

// API 响应校验（开发环境）
if (import.meta.env.DEV) {
    UserSchema.parse(resp.data);  // 抛出明确错误
}
```

## CI 集成

```yaml
# .github/workflows/type-check.yaml
- name: Start backend
  run: docker-compose up -d app
- name: Wait for backend
  run: until curl -s http://localhost:8080/health; do sleep 2; done
- name: Generate TS types
  run: npx openapi-typescript http://localhost:8080/swagger/doc.json -o src/types/api.generated.ts
- name: Check diff
  run: git diff --exit-code src/types/api.generated.ts || (echo "Types are stale. Run 'make gen-types'." && exit 1)
```

## 设计要点

1. **单一事实来源**：`docs/swagger.json` 由 `make swagger` 生成，提交 git
2. **CI 检查新鲜度**：PR 中自动检查生成的 TS 类型是否与 Swagger 一致
3. **不要手改 `api.generated.ts`**——所有修改必须来自 Swagger 注解变更
4. 如果某个接口的 Swagger 注解不完整（如缺少 `@Success` 的 model），CI 检查时给出 warning 但不阻断
5. 作为兜底，手写的 `types/` 目录保留通用类型（如 `ApiResponse`），自动生成的只覆盖 API 相关类型
