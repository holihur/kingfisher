# Kingfisher Spring Boot 版（第一阶段：仅登录）

> 复刻 Go 版本的登录契约，前端 `kingfisher-web` 无需修改即可对接。

## 技术栈

- Java 17+（已验证 Java 20.0.1）
- Spring Boot 3.2.5 + Spring Web / Validation
- MyBatis 3.0.3 + SQLite JDBC 3.45.3.0（复用项目根 `kingfisher.db`）
- JJWT 0.12.5（HS256）
- Spring Security Crypto（BCrypt，兼容 Go 生成的 `$2a$12$` 哈希）
- Redis（可选，`spring-data-redis`，不可用时自动回退内存）

## 目录结构

```
java/
├── pom.xml
├── src/main/java/com/kingfisher/
│   ├── KingfisherApplication.java
│   ├── common/          # ApiResponse / ErrorCode / GlobalExceptionHandler
│   ├── config/          # JwtProperties / CorsProperties / WebConfig
│   ├── security/        # JwtProvider / JwtAuthFilter / TokenBlacklistService / LoginAttemptService
│   └── modules/user/
│       ├── domain/      # User / Role
│       ├── dto/         # LoginRequest / RefreshRequest / LoginResponse
│       ├── mapper/      # UserMapper.java
│       └── service/     # AuthService
│       └── controller/  # AuthController / UserController
└── src/main/resources/
    ├── application.yml
    └── mapper/UserMapper.xml
```

## 快速开始

```bash
cd java

# 编译
mvn compile -DskipTests
# 或打包
mvn package -DskipTests

# 启动（默认 8080，与 Go 版本一致）
java -jar target/kingfisher-0.0.1-SNAPSHOT.jar
# 指定端口
java -jar target/kingfisher-0.0.1-SNAPSHOT.jar --server.port=8081

# 或直接用 Maven 插件
mvn spring-boot:run
# 指定端口
mvn spring-boot:run -Dspring-boot.run.arguments="--server.port=8081"
```

启动后访问：`http://localhost:8080/api/v1/auth/health` 返回 `{"code":0,"data":"ok"}`。

### 数据库

默认复用项目根的 `kingfisher.db`（SQLite，Go 的 `AutoMigrate` 已建表并播种），使用相对路径（以 `java/` 为工作目录）：

```yaml
spring.datasource.url: jdbc:sqlite:${KINGFISHER_DB:../kingfisher.db}
# 等价于 jdbc:sqlite:../kingfisher.db（java/ 的上一级即项目根）
```

可通过环境变量、`.env` 或启动参数覆盖（支持绝对路径）：

```bash
# 方式一：环境变量（推荐，zsh 需 set -a）
KINGFISHER_DB=/path/to/kingfisher.db java -jar target/kingfisher-0.0.1-SNAPSHOT.jar
# 或写入 .env（项目根或 java/.env），Java 启动时自动加载（兼容 export 前缀与引号）
echo "KINGFISHER_DB=/path/to/kingfisher.db" >> .env

# zsh 手动加载 .env（.env 无 export 前缀时需 set -a）
set -a; source .env; set +a
# 或
export $(cat .env | xargs)  # 需 .env 为 KEY=VALUE 格式

# 方式二：启动参数（优先级最高）
java -jar target/kingfisher-0.0.1-SNAPSHOT.jar \
  --spring.datasource.url=jdbc:sqlite:/path/to/kingfisher.db
```

`task java:run` / `task java:dev` 已自动 `set -a; source .env; set +a`，在 zsh/bash 下均可直接读取 `.env`。

种子账号（密码均为 `Abcd1234`）：`admin / editor / viewer / multi`。

### JWT 配置

与 Go `config/config.yaml` 对齐，可通过 `application.yml` 或环境变量覆盖：

```yaml
kingfisher.jwt.secret: change-me-in-production  # 生产请用 32+ 字符
kingfisher.jwt.issuer: kingfisher
kingfisher.jwt.access-ttl: 2h
kingfisher.jwt.refresh-ttl: 168h
```

环境变量：`JWT_SECRET` 会覆盖 `kingfisher.jwt.secret`（通过 Spring 的 relaxed binding）。

### Redis（可选）

`TokenBlacklistService` 与 `LoginAttemptService` 优先使用 Redis，若 `127.0.0.1:6379` 不可用则回退到内存（单机开发足够）。无需强制启动 Redis。

## API 契约（与 Go 一致）

统一响应体：`{code, message, data}`，`code=0` 成功。错误码与 `core/errcode` 对齐，前端强依赖：

| code | 含义 | HTTP |
|------|------|------|
| 0 | success | 200 |
| 10003 | 未认证 | 401 |
| 10103 | 密码错误 | 400 |
| 10104 | Token 过期（触发刷新） | 401 |
| 10105 | Token 无效 | 401 |
| 10107 | 登录失败次数过多 | 429 |

### 登录

```
POST /api/v1/auth/login
Content-Type: application/json

{"username":"admin","password":"Abcd1234"}

=> 200 {code:0, data:{access_token, refresh_token, user, landing_page}}
   400 {code:10103, message:"密码错误"}
   429 {code:10107, message:"登录失败次数过多"}  # 15分钟内>5次错误
```

`user` 字段：`id, username, nickname, email, avatar, status, role_ids, roles[], created_at, updated_at`（`password/session_version` 不返回，`role_ids` 为角色 ID 列表，`roles` 含 `id/name/code/landing_page`）。

`landing_page` 取用户首个角色的 `roles.landing_page`（如 `admin -> /dashboard`）。

### 刷新

```
POST /api/v1/auth/refresh
{"refresh_token":"<refresh_token>"}  # 兼容 refreshToken 驼峰

=> 200 {code:0, data:{access_token:"..."}}
   401 {code:10104} 过期 / {code:10105} 无效
```

前端 `request.ts` 会在收到 `10104` 时自动用 `refresh_token` 换新 `access_token`。

### 退出

```
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
# 或 body: {refresh_token}
=> 200 {code:0}
```
将 token 的 `jti` 加入黑名单至过期。

### 受保护示例

```
GET /api/v1/users/me
Authorization: Bearer <access_token>
=> 200 {code:0, data:{user}}

GET /api/v1/users/me/permissions
=> 200 {code:0, data:["admin", ...]}

# 未认证
=> 401 {code:10003, message:"未认证"}
# Token 过期
=> 401 {code:10104, message:"Token 过期"}
# Token 无效/被撤销/session_version 过旧
=> 401 {code:10105, message:"Token 无效"}
```

CORS 已放行 `http://localhost:5173`（前端 Vite 默认端口）。

## 与 Go 版本的差异（阶段一）

- 仅实现登录/刷新/登出/鉴权过滤，`注册/找回密码` 等暂未实现（返回 404）。
- `login_fail` 限流与 `blacklist` 优先 Redis，失败回退内存（Go 为 Redis 强依赖）。
- `session_version` 校验已实现：修改密码后旧 token 会被 `JwtAuthFilter` 拦截为 `10105`。
- SQLite 同库复用，`ddl-auto: none`，不自动建表。

## 前端联调

前端 `kingfisher-web` 默认 `baseURL: /api/v1`：

```bash
# 后端 8080，前端 5173
cd kingfisher-web
pnpm dev  # VITE_API_BASE_URL 默认代理到 http://localhost:8080
```

若 Spring Boot 改为 8081，需在 `kingfisher-web/.env` 或启动时指定：

```
VITE_API_BASE_URL=http://localhost:8081/api/v1
```

## 验证

```bash
# 健康
curl http://localhost:8080/api/v1/auth/health

# 登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Abcd1234"}'

# 刷新
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'

# 受保护
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <access_token>"
```
