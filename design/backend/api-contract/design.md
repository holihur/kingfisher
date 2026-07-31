# API Contract — 接口契约

## 职责

为每个 API 接口提供完整的请求/响应示例，团队无需翻代码即可对接。

## 契约格式

每个接口按以下模板记录：

```markdown
### POST /api/v1/auth/login
- **描述**: 用户登录
- **Auth**: 否
- **限流**: 5 req/min per IP

**Request**
Content-Type: application/json
{
  "username": "admin",
  "password": "Abcd1234"
}

**Response 200**
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGci...",
    "refresh_token": "eyJhbGci...",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "avatar": "",
      "status": 1,
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    }
  }
}

**Error Responses**
| code | HTTP | message | 场景 |
|------|------|---------|------|
| 10102 | 400 | user not found | 用户名不存在 |
| 10103 | 400 | wrong password | 密码错误 |
| 10106 | 400 | user disabled | 用户已禁用 |
| 10107 | 429 | too many attempts | 登录失败超限，15 分钟后重试 |
| 10001 | 400 | invalid param | 参数格式错误 |
```

## 完整 API 清单

### Auth（extends/user）

```
POST   /api/v1/auth/register     注册
POST   /api/v1/auth/login         登录
POST   /api/v1/auth/refresh       刷新 token
POST   /api/v1/auth/logout        注销（需登录）
```

### User（extends/user）

```
GET    /api/v1/users              用户列表（管理员）
GET    /api/v1/users/:id          用户详情
PUT    /api/v1/users/:id          更新用户
DELETE /api/v1/users/:id          删除用户
GET    /api/v1/users/me           当前用户信息
GET    /api/v1/users/me/permissions  当前用户权限列表
PUT    /api/v1/users/me/password  修改密码
```

### Menu（extends/menu）

```
GET    /api/v1/menus/tree         菜单树
GET    /api/v1/menus/:id          菜单详情
POST   /api/v1/menus              创建菜单
PUT    /api/v1/menus/:id          更新菜单
DELETE /api/v1/menus/:id          删除菜单
```

### Role（extends/rbac）

```
GET    /api/v1/roles              角色列表
GET    /api/v1/roles/:id          角色详情
POST   /api/v1/roles              创建角色
PUT    /api/v1/roles/:id          更新角色
DELETE /api/v1/roles/:id          删除角色
GET    /api/v1/roles/:id/permissions   获取角色权限
PUT    /api/v1/roles/:id/permissions   分配权限
GET    /api/v1/roles/:id/menus         获取角色菜单
GET    /api/v1/roles/:id/permissions   获取角色权限
PUT    /api/v1/roles/:id/menus         分配菜单
GET    /api/v1/roles/:id/menus         获取角色菜单
GET    /api/v1/permissions             全部权限列表
```

### Config（extends/config）

```
GET    /api/v1/configs            全部配置
GET    /api/v1/configs/:key       单个配置
PUT    /api/v1/configs/:key       更新配置
DELETE /api/v1/configs/:key       删除配置
```

## 典型接口详细契约

### GET /api/v1/users

```
Query:
  page      int    默认 1
  page_size int    默认 20，最大 100
  keyword   string 搜索用户名/邮箱（可选）

Response 200:
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "username": "admin",
        "email": "admin@example.com",
        "avatar": "",
        "status": 1,
        "role": { "id": 1, "name": "管理员", "code": "admin" },
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z"
      }
    ],
    "total": 98,
    "page": 1,
    "page_size": 20
  }
}
```

### GET /api/v1/menus/tree

```
Response 200:
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "parent_id": 0,
      "name": "系统管理",
      "path": "/system",
      "component": "",
      "icon": "SettingOutlined",
      "sort": 1,
      "type": 1,
      "permission": "",
      "status": 1,
      "children": [
        {
          "id": 2,
          "parent_id": 1,
          "name": "用户管理",
          "path": "/system/users",
          "component": "pages/User/UserList",
          "icon": "UserOutlined",
          "sort": 1,
          "type": 2,
          "permission": "user:list",
          "status": 1,
          "children": [
            {
              "id": 3,
              "parent_id": 2,
              "name": "新增用户",
              "path": "",
              "component": "",
              "icon": "",
              "sort": 1,
              "type": 3,
              "permission": "user:create",
              "status": 1,
              "children": null
            }
          ]
        }
      ]
    }
  ]
}
```

### PUT /api/v1/roles/:id/permissions

```
Request:
{
  "permission_ids": [1, 2, 3, 5, 7]
}

Response 200:
{
  "code": 0,
  "message": "success"
}
```

### PUT /api/v1/configs/:key

```
Path param: key = "site_name"
Request:
{
  "value": "Kingfisher Admin System"
}

Response 200:
{
  "code": 0,
  "message": "success"
}
```

## 设计要点

- 分页参数名统一为 `page` + `page_size`（不是 `pageSize` + `limit`）
- 时间格式统一为 RFC3339（`2026-01-01T00:00:00Z`）
- 排序参数格式：`sort=created_at&order=desc`
- 错误响应都带 `code` + `message`，前端据此做国际化或中文提示
