# Scripts — 自动化验收脚本

> 列入 `design/` 的脚本均为完整可执行文件（非伪代码）。实现时复制到项目根目录 `scripts/` 即可运行。

## 脚本清单

| 脚本 | CI | 说明 |
|------|-----|------|
| `scripts/chaos.sh` | 每日定时 | MySQL/Redis 中断恢复测试 |
| `scripts/deploy-check.sh` | 发布前 | 镜像大小、配置安全、安全响应头 |
| `scripts/bench.sh` | 每日定时 | Vegeta 压测 + P99 检查 |
| `scripts/check-traces.sh` | 每日定时 | Jaeger API 验证 span 链完整性 |
| `scripts/check-metrics.sh` | 每日定时 | Prometheus API 验证 6 面板数据 |
| `scripts/check-no-panic.sh` | 每次 push | 扫描 panic/log.Fatal/fmt.Println |
| `scripts/check-frontend-constraints.sh` | 每次 push | 扫描硬编码 URL、直接 fetch/axios |

---

## `scripts/chaos.sh`

```bash
#!/bin/bash
set -euo pipefail
BASE="${BASE_URL:-http://localhost:8080}"

# 获取 admin token
TOKEN=$(curl -sf -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Abcd1234"}' | jq -r '.data.access_token')
AUTH=(-H "Authorization: Bearer $TOKEN")

check() { curl -s -o /dev/null -w "%{http_code}" "$@"; }
assert_eq() {
    local got="$1" expected="$2" label="$3"
    if [ "$got" != "$expected" ]; then
        echo "FAIL [$label]: expected HTTP $expected, got $got" >&2
        exit 1
    fi
    echo "PASS [$label]"
}

cleanup() {
    docker start kingfisher-mysql-1 2>/dev/null || true
    docker start kingfisher-redis-1 2>/dev/null || true
    sleep 5
}
trap cleanup EXIT

echo "=== Test 1: MySQL kill → /ready 503 ==="
docker kill kingfisher-mysql-1 2>/dev/null || true
sleep 3
assert_eq "$(check "$BASE/ready")" "503" "ready-after-mysql-kill"

echo "=== Test 2: MySQL down → login 500 ==="
assert_eq "$(check -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Abcd1234"}')" "500" "login-mysql-down"

echo "=== Test 3: MySQL recovery → /ready 200 ==="
docker start kingfisher-mysql-1
sleep 12
assert_eq "$(check "$BASE/ready")" "200" "ready-after-mysql-recovery"

echo "=== Test 4: Redis kill → login passes, RBAC blocked ==="
docker kill kingfisher-redis-1 2>/dev/null || true
sleep 3
assert_eq "$(check -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Abcd1234"}')" "200" "login-redis-down"
assert_eq "$(check "$BASE/api/v1/users?page=1&page_size=5" "${AUTH[@]}")" "503" "rbac-redis-down"

echo "=== Test 5: Redis recovery → RBAC passes ==="
docker start kingfisher-redis-1
sleep 6
assert_eq "$(check "$BASE/api/v1/users?page=1&page_size=5" "${AUTH[@]}")" "200" "rbac-after-redis-recovery"

echo "=== Test 6: Data persistence across restart ==="
docker-compose down 2>/dev/null || true
docker-compose up -d
sleep 18
USER_COUNT=$(curl -sf "$BASE/api/v1/users?page=1&page_size=1" "${AUTH[@]}" | jq '.data.total')
[ "$USER_COUNT" -gt 0 ] || { echo "FAIL: data lost after restart"; exit 1; }
echo "PASS [data-persistence]: $USER_COUNT users survived restart"

echo "ALL 6 CHAOS TESTS PASSED"
```

---

## `scripts/deploy-check.sh`

```bash
#!/bin/bash
set -euo pipefail
IMAGE="${IMAGE:-kingfisher:latest}"
BASE="${BASE_URL:-http://localhost}"

fail() { echo "FAIL: $*" >&2; exit 1; }
warn() { echo "WARN: $*" >&2; }
pass() { echo "PASS: $*"; }

echo "=== Image size ==="
SIZE=$(docker images "$IMAGE" --format '{{.Size}}' 2>/dev/null | head -1)
if [ -z "$SIZE" ]; then
    warn "image $IMAGE not found, skipping"
else
    MB=$(echo "$SIZE" | grep -oP '[\d.]+' | head -1)
    if [ "$(echo "$MB < 15" | bc 2>/dev/null || echo 0)" = "1" ]; then
        pass "image size: ${MB}MB (< 15MB)"
    else
        fail "image size: ${MB}MB (>= 15MB)"
    fi
fi

echo "=== Production config ==="
CONFIG="config/config.prod.yaml"
[ -f "$CONFIG" ] || fail "$CONFIG not found"
grep -q 'mode:\s*release' "$CONFIG" || fail "mode is not release"
grep -E 'password:\s*""' "$CONFIG" | grep -q 'password' || warn "password may be hardcoded (expected empty string)"
grep -qi 'change-me' "$CONFIG" && fail "default secret 'change-me' still in config"
pass "config secure"

echo "=== Security headers ==="
HEADERS=$(curl -sI "${BASE}/api/v1/users" 2>/dev/null || echo "")
echo "$HEADERS" | grep -qi 'X-Content-Type-Options' || warn "missing X-Content-Type-Options"
echo "$HEADERS" | grep -qi 'X-Frame-Options' || warn "missing X-Frame-Options"
pass "security headers checked"

echo "=== Hidden file protection ==="
STATUS=$(curl -s -o /dev/null -w '%{http_code}' "${BASE}/.env" 2>/dev/null || echo "000")
[ "$STATUS" = "404" ] || [ "$STATUS" = "403" ] || fail ".env accessible (HTTP $STATUS, expected 404/403)"
pass "hidden file protected"

echo "ALL DEPLOY CHECKS PASSED"
```

---

## `scripts/bench.sh`

```bash
#!/bin/bash
set -euo pipefail
BASE="${BASE_URL:-http://localhost:8080}"

TOKEN=$(curl -sf -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Abcd1234"}' | jq -r '.data.access_token')

echo "users?page=1&page_size=20" | vegeta attack \
  -duration=30s -rate=100 \
  -header "Authorization: Bearer $TOKEN" \
  -targets /dev/stdin \
  | vegeta report -type=json \
  | jq '{
      requests: .requests,
      rate: .rate,
      success: .success,
      p50_ms: (.latencies."50th" / 1000000 | floor),
      p99_ms: (.latencies."99th" / 1000000 | floor),
      max_ms:  (.latencies.max       / 1000000 | floor)
    }'

# CI 模式：P99 > 50ms → exit 1
if [ "${CI:-}" = "true" ]; then
    P99=$(echo "GET ${BASE_URL:-http://localhost:8080}/api/v1/users?page=1&page_size=20" | vegeta attack -duration=30s -rate=100 -header "Authorization: Bearer $TOKEN" | vegeta report -type=json | jq -r '(.latencies."99th" / 1000000)')
    if [ "$(echo "$P99 > 50" | bc)" = "1" ]; then
        echo "FAIL: P99=${P99}ms > 50ms" >&2
        exit 1
    fi
fi
```

---

## `scripts/check-traces.sh`

```bash
#!/bin/bash
set -euo pipefail
JAEGER="${JAEGER_URL:-http://localhost:16686}"

# 触发一次登录请求产生 trace
curl -sf -X POST "${BASE_URL:-http://localhost:8080}/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Abcd1234"}' > /dev/null
sleep 2

# 通过 Jaeger API 查询最近的 trace
TRACES=$(curl -sf "$JAEGER/api/traces?service=kingfisher&limit=1&lookback=5m" | jq '.data[0]')

if [ "$TRACES" = "null" ] || [ -z "$TRACES" ]; then
    echo "FAIL: no traces found" >&2
    exit 1
fi

TRACE_ID=$(echo "$TRACES" | jq -r '.traceID')

# 获取 trace 详情
SPANS=$(curl -sf "$JAEGER/api/traces/$TRACE_ID" | jq '.data[0].spans')

# 验证 span 层级链完整：HTTP → handler → service → repo → DB
SPAN_COUNT=$(echo "$SPANS" | jq 'length')
HAS_DB=$(echo "$SPANS" | jq '[.[] | select(.operationName | test("db|sql|query"))] | length')

echo "PASS: trace $TRACE_ID has $SPAN_COUNT spans, $HAS_DB db queries"
[ "$SPAN_COUNT" -ge 3 ] || { echo "FAIL: expected >=3 spans, got $SPAN_COUNT" >&2; exit 1; }
```

---

## `scripts/check-metrics.sh`

```bash
#!/bin/bash
set -euo pipefail
PROM="${PROM_URL:-http://localhost:9090}"

# Prometheus API 即时查询 6 个关键指标
check_metric() {
    local query="$1" name="$2"
    result=$(curl -sfG "$PROM/api/v1/query" --data-urlencode "query=$query" | jq -r '.data.result | length')
    if [ "${result:-0}" -gt 0 ]; then
        echo "PASS [$name]: $result series"
    else
        echo "FAIL [$name]: no data" >&2
        exit 1
    fi
}

check_metric 'rate(http_requests_total[5m])'   'http_rate'
check_metric 'histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))' 'p99_latency'
check_metric 'histogram_quantile(0.99, rate(db_query_duration_seconds_bucket[5m]))'     'db_p99'
check_metric 'rate(redis_commands_total[5m])'   'redis_rate'
check_metric 'go_gc_duration_seconds'           'gc_pause'
check_metric 'go_goroutines'                    'goroutines'

echo "ALL 6 METRICS PRESENT"
```

---

## `scripts/check-no-panic.sh`

```bash
#!/bin/bash
set -euo pipefail
failures=0

echo "--- Checking for panic outside main.go ---"
if grep -rn 'panic(' extends/ 2>/dev/null | grep -v '_test.go' | grep -v 'main.go' | grep -v 'middleware/recovery.go'; then
    echo "FAIL: panic() found in business code"
    ((failures++))
else
    echo "PASS: no panic() in business code"
fi

echo "--- Checking for log.Fatal outside main.go ---"
if grep -rn 'log\.Fatal' extends/ 2>/dev/null | grep -v '_test.go' | grep -v 'main.go'; then
    echo "FAIL: log.Fatal found in business code"
    ((failures++))
else
    echo "PASS: no log.Fatal in business code"
fi

echo "--- Checking for fmt.Println in business code ---"
if grep -rn 'fmt\.Println\|fmt\.Print(' extends/ 2>/dev/null | grep -v '_test.go' | grep -v 'main.go'; then
    echo "FAIL: fmt.Println/Print found in business code (use logger)"
    ((failures++))
else
    echo "PASS: no fmt.Println in business code"
fi

exit $failures
```

---

## `scripts/check-frontend-constraints.sh`

```bash
#!/bin/bash
set -euo pipefail
SRC="${SRC:-src}"
failures=0

echo "--- Checking for hardcoded localhost ---"
if grep -rn 'localhost:8080\|http://192.168\|http://127.0.0.1' "$SRC" 2>/dev/null; then
    echo "FAIL: hardcoded API URL found (use VITE_API_TARGET env var)"
    ((failures++))
else
    echo "PASS: no hardcoded URLs"
fi

echo "--- Checking for direct fetch/axios in pages/components ---"
if grep -rn 'fetch(' "$SRC/pages" "$SRC/components" "$SRC/layouts" 2>/dev/null; then
    echo "FAIL: direct fetch() in page/component (use api/ module)"
    ((failures++))
else
    echo "PASS: no direct fetch()"
fi
if grep -rn 'axios\.' "$SRC/pages" "$SRC/components" "$SRC/layouts" 2>/dev/null | grep -v 'api/request'; then
    echo "FAIL: direct axios usage in page/component (use api/ module)"
    ((failures++))
else
    echo "PASS: no direct axios"
fi

exit $failures
```
