# Cache — 缓存策略

## 职责

定义缓存使用模式、Key 规范、穿透防护。所有缓存操作通过 `core.Cache` 接口。

## 缓存模式

### 1. Cache-Aside（旁路缓存）—— 最常用

```
读: 查缓存 → hit: 返回
            → miss: 查 DB → 写缓存 → 返回
写: 写 DB → 删缓存（不是更新缓存）
```

**使用者**：菜单树、配置项、权限列表

```go
func (s *MenuService) GetTree(ctx context.Context) ([]domain.Menu, error) {
    val, err := s.cache.Get(ctx, "menu:tree")
    if err == nil {
        var menus []domain.Menu
        json.Unmarshal([]byte(val), &menus)
        return menus, nil
    }
    menus, err := s.repo.FindAll(ctx)     // cache miss
    if err != nil { return nil, err }
    tree := BuildTree(menus, 0)
    data, _ := json.Marshal(tree)
    s.cache.Set(ctx, "menu:tree", data, 10*time.Minute)
    return tree, nil
}
```

### 2. Write-Through —— 用于需要立即生效的数据

```
写 DB + 写缓存（同时）
```

**使用者**：系统配置

### 3. 黑名单模式 —— 用于 JWT 注销

```
Set: cache.Set("blacklist:token:{jti}", "1", remainingTTL)
Check: cache.Exists("blacklist:token:{jti}")
```

## 缓存穿透防护

### 空值缓存

```go
// 防止不存在的 key 反复穿透到 DB
user, err := s.repo.FindByID(ctx, id)
if errors.Is(err, gorm.ErrRecordNotFound) {
    s.cache.Set(ctx, "user:"+strconv.Itoa(id), "null", 1*time.Minute)
    return nil, domain.ErrUserNotFound
}
```

### 布隆过滤器（可选，高并发场景）

初始化时把所有存在的 ID 加载到布隆过滤器，查询前先过布隆。

## Key 命名规范

```
{domain}:{identifer}[:{sub}]

示例:
  user:1                    # 用户 ID=1
  menu:tree                 # 菜单树
  menu:role:3               # 角色 3 的菜单
  config:all                # 所有配置
  config:site_name          # 单个配置
  blacklist:token:{jti}     # JWT 黑名单
  ratelimit:{ip}:{path}     # 限流
  user:perms:5              # 用户 5 的权限列表
```

## TTL 策略

| 数据类型 | TTL | 理由 |
|------|-----|------|
| 菜单树 | 10 min | 不常变，允许延迟 |
| 用户信息 | 5 min | 中等频率 |
| 权限列表 | 30 min | 变更后手动失效 |
| 系统配置 | 5 min | 变更后立即失效 |
| 空值 | 1 min | 防穿透，短暂 |
| 限流 | 1 min | 滑动窗口 |
| Token 黑名单 | = token 剩余 TTL | 过期后自动清理 |

## 缓存失效策略

| 操作 | 失效的 Key |
|------|-----------|
| 更新用户 | `user:{id}` |
| 更新菜单 | `menu:tree`, `menu:role:*` |
| 更新角色权限 | `user:perms:*`, `menu:role:{roleID}` |
| 更新系统配置 | `config:all`, `config:{key}` |

## 设计要点

- 所有缓存值用 JSON string 存储（统一序列化）
- 批量失效用 `cache.Delete(ctx, keys...)`
- 不用 `KEYS pattern` 删除（生产禁用），用 `SCAN` 或记录关联关系
