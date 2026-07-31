# Guardrails — 代码层硬性约束

> 防止垃圾代码合入——无论谁写的，无论什么模型生成的。所有约束在 CI 中强制执行，违反即阻断。

---

## Go 约束

### .golangci.yml（error 级别，违反 CI 报红）

```yaml
linters:
  enable:
    # === 硬性约束 ===
    - govet              # Go 官方静态分析
    - errcheck           # 未处理的 error 返回值 → error
    - staticcheck        # 100+ 条静态规则
    - gosec              # 安全漏洞检测
    - depguard           # 禁止跨层 import → 下文详述
    - revive             # 替代 golint，可配置规则
    - goimports          # import 自动整理
    - unconvert          # 不必要的类型转换
    - unparam            # 未使用的函数参数
    - wastedassign       # 写了但从未读的赋值
    - misspell           # 拼写错误
    - nilerr             # return nil, err 应该 return err
    - noctx              # 不带 context 的 HTTP/DB 调用
    - errorlint          # 错误比较用 errors.Is 而非 ==
    - gocritic           # 可读性/性能建议

    # === 禁用（不协商）===
    # - funlen, gocyclo, gocognit （圈复杂度由 code review 判定）
    # - godox （允许 FIXME/TODO 注释）
    # - exhaustruct （GORM model 不强制填充所有字段）
    # - forbidigo （Println 在 debug 模式下允许）

linters-settings:
  depguard:
    rules:
      # 规则 1：core 包绝对不能 import 任何 extends 包
      - list_mode: lax
        files:
          - "!**/core/**/*.go"   # 排除 core 自身
        allow:
          - $gostd
          - "kingfisher/core/**"
      # 等价于：core/ 下的文件只能 import 标准库和 core 自身
      # 如果 import 了 extends → CI 报错

      # 规则 2：domain 包零外部依赖
      - list_mode: lax
        files:
          - "!**/domain/**/*.go"
        allow:
          - $gostd
      # domain 不能 import core、extends、任何第三方库

      # 规则 3：port 包只能 import domain
      - list_mode: lax
        files:
          - "!**/port/**/*.go"
        allow:
          - $gostd
          - "kingfisher/extends/**/domain"

      # 规则 4：adapter 可以 import 外部依赖（GORM、Redis...）
      # （不限制，允许 adapter 使用任意第三方库）

  gosec:
    severity: medium
    excludes:
      - G104  # 少数 err 确实可以忽略（如 defer file.Close()），需加注释

  revive:
    rules:
      - name: exported           # 导出函数必须有注释
        severity: error
      - name: context-as-argument # context 必须是第一个参数
        severity: error
      - name: error-return       # 返回 error 的函数必须检查
        severity: error
      - name: error-strings      # 错误信息不以大写开头、不以标点结尾
        severity: error
      - name: receiver-naming    # receiver 命名规范
        severity: warning
      - name: time-naming        # 不使用 time.Time 作为变量名
        severity: warning

  govet:
    enable-all: true
    disable:
      - fieldalignment   # 结构体对齐优化留给编译器
```

### go.mod 约束

```bash
# CI 中执行，阻断未整理依赖的提交
go mod tidy
git diff --exit-code go.mod go.sum
# 不为零 → go.mod 脏 → CI fail
```

### Wire 编译期约束

```bash
# CI 中每次 push 执行
cd internal/wire && wire
git diff --exit-code internal/wire/wire_gen.go
# 不为零 → Wire 生成代码和提交不一致 → CI fail
```

Wire 代码生成后编译检查：

```bash
go build ./cmd/server
# 编译失败 → 依赖图有循环引用或缺少 Provider → CI fail
```

### 禁止行为清单（CI 强制扫描）

```bash
# 1. 禁止 panic 在非 main.go 文件中
grep -rn 'panic(' internal/ extends/ | grep -v '_test.go' | grep -v 'main.go' && exit 1

# 2. 禁止 hardcode 的密码/secret
gitleaks detect --source . --no-git -v

# 3. 禁止 log.Fatal 在非 main.go 中
grep -rn 'log.Fatal' internal/ extends/ | grep -v 'main.go' && exit 1

# 4. 禁止 fmt.Println 在非 main.go 中（用 logger）
grep -rn 'fmt.Println\|fmt.Print(' internal/ extends/ | grep -v '_test.go' | grep -v 'main.go' && exit 1

# 5. 禁止 time.Now() 和 time.Sleep() 在业务代码中（用 clock interface）
# → 可配置，暂不强制，但新代码应使用 core/clock
```

### 编译约束

```go
// core/router/module.go
type Module interface {
    Name() string
    RegisterPublic(r *gin.RouterGroup)
    RegisterProtected(r *gin.RouterGroup)
}

// extends/user/transport/register.go
var _ core.Module = (*UserModule)(nil)  // ← 编译期检查：UserModule 实现了 Module 接口
// Wire 会二次检查——双重保险
```

---

## 前端约束

### tsconfig.json

```json
{
  "compilerOptions": {
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "noImplicitReturns": true,
    "forceConsistentCasingInFileNames": true,
    "isolatedModules": true
  }
}
```

`strict: true` 强制 TypeScript 严格模式，等价于开启：
- `strictNullChecks` — 不能 `null.foo`
- `strictFunctionTypes` — 函数参数逆变检查
- `strictBindCallApply` — bind/call/apply 参数检查
- `noImplicitAny` — 禁止隐式 any
- `alwaysStrict` — 输出 `"use strict"`

### ESLint

```js
// eslint.config.js
export default [
  {
    rules: {
      // === 硬性约束（CI error）===
      'no-console': ['error', { allow: ['warn', 'error'] }],  // console.log → error
      'no-debugger': 'error',
      'no-alert': 'error',
      'eqeqeq': ['error', 'always'],        // 必须 ===
      'no-var': 'error',                     // 禁止 var
      'prefer-const': 'error',               // 能用 const 不用 let
      '@typescript-eslint/no-explicit-any': 'error',  // 禁止 any
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      '@typescript-eslint/explicit-function-return-type': 'warn',  // 函数返回类型

      // === 安全约束 ===
      'no-eval': 'error',
      'no-implied-eval': 'error',
      'react/no-dangerously-set-innerhtml': 'error',  // 禁止 dangerouslySetInnerHTML

      // === Import 约束 ===
      'import/no-cycle': 'error',           // 禁止循环引用
      'no-restricted-imports': ['error', {
        patterns: [
          {
            group: ['lodash', 'lodash/*'],  // 禁止用 lodash（用 lodash-es 或原生）
            message: 'Use lodash-es or native methods instead'
          }
        ]
      }],
    }
  }
];
```

### 前端禁止行为（CI 强制扫描）

```bash
# 1. 类型检查
npx tsc --noEmit
# 有错误 → CI fail

# 2. ESLint
npx eslint src/ --max-warnings 0
# 有 warning 也 fail

# 3. 构建检查
npm run build
# 构建失败 → CI fail

# 4. 安全审计
npm audit --audit-level=high
# 有 high/critical → CI fail

# 5. 禁止 hardcode API URL
grep -rn 'http://localhost:8080' src/ && exit 1
grep -rn 'http://192.168' src/ && exit 1

# 6. 禁止在组件中直接 fetch/axios（必须用 api/ 封装）
grep -rn 'fetch(' src/pages/ src/components/ src/layouts/ && exit 1
grep -rn 'axios.' src/pages/ src/components/ src/layouts/ && exit 1
```

---

## CI Pipeline（合并所有约束）

```yaml
name: Guardrails
on: [push, pull_request]

jobs:
  backend:
    runs-on: ubuntu-latest
    services:
      mysql: { image: mysql:8.0, env: { MYSQL_ROOT_PASSWORD: test, MYSQL_DATABASE: kingfisher }, ports: ['3306:3306'] }
      redis: { image: redis:7-alpine, ports: ['6379:6379'] }
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.23' }

      # Layer 1: 编译
      - run: go build ./cmd/server

      # Layer 2: go mod tidy 检查
      - run: go mod tidy && git diff --exit-code go.mod go.sum

      # Layer 3: go vet
      - run: go vet ./...

      # Layer 4: golangci-lint
      - uses: golangci/golangci-lint-action@v6
        with: { version: latest, args: --timeout=5m }

      # Layer 5: Wire 检查
      - run: cd internal/wire && wire && git diff --exit-code wire_gen.go

      # Layer 6: 禁止 panic / log.Fatal / fmt.Println / hardcode secret
      - run: |
          scripts/check-no-panic.sh
          scripts/check-no-hardcode-secret.sh
          gitleaks detect --source . --no-git

      # Layer 7: 安全漏洞检查
      - run: govulncheck ./...

      # Layer 8: 测试 + 覆盖率
      - run: go test -v -race -coverprofile=coverage.out ./...
      - run: go tool cover -func=coverage.out | grep total | awk '{if ($3 < 80) exit 1}'

  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: '20' }
      - run: npm ci

      # Layer 1: TypeScript
      - run: npx tsc --noEmit

      # Layer 2: ESLint
      - run: npx eslint src/ --max-warnings 0

      # Layer 3: 构建
      - run: npm run build

      # Layer 4: 安全审计
      - run: npm audit --audit-level=high

      # Layer 5: 禁止直接 fetch/axios、禁止 hardcode URL
      - run: scripts/check-frontend-constraints.sh

      # Layer 6: E2E 测试
      - run: npx playwright test

  deploy-check:
    runs-on: ubuntu-latest
    needs: [backend, frontend]
    steps:
      - run: docker build -f deploy/Dockerfile -t kingfisher:ci .
      - run: |
          SIZE=$(docker images kingfisher:ci --format '{{.Size}}' | grep -oE '[0-9.]+')
          if [ "$(echo "$SIZE > 15" | bc)" = "1" ]; then exit 1; fi
      - run: docker scout quickview kingfisher:ci --exit-code
```

---

## 执行顺序

```
编译 → go mod tidy → vet → lint → Wire → 禁止项扫描 → 安全扫描 → 测试+覆盖率 → 构建 → E2E
```

**任何一环失败，整个 pipeline 阻断。** 即便所有功能测试通过，违反一条 lint 规则也会 fail。

---

## 约束的最高原则

| 原则 | 对应规则 |
|------|----------|
| core 不 import extends | depguard |
| domain 零外部依赖 | depguard |
| 所有 error 必须检查 | errcheck |
| context 必须是第一个参数 | revive |
| 禁止 panic（非 main） | grep check |
| 禁止 any（TypeScript） | ESLint |
| 禁止 console.log | ESLint |
| 禁止直接 fetch/axios 在页面组件 | grep check |
| 构建失败 = 不能合入 | CI |
| 测试覆盖率 < 80% = 不能合入 | CI |
| Wire 代码与提交不一致 = 不能合入 | CI |
| 镜像 > 15MB = 不能合入 | CI |
| 有 critical CVE = 不能合入 | CI |
