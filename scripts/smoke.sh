#!/bin/bash
set -euo pipefail
BASE="${BASE_URL:-http://localhost:8080}"
FAILS=0

check() {
  local expected="${2:-200}" method="${3:-GET}" body="${4:-}"
  local code args=(-s -o /dev/null -w "%{http_code}")
  [ "$method" = "POST" ] && args+=(-X POST -H 'Content-Type: application/json')
  [ -n "$body" ] && args+=(-d "$body")
  code=$(curl "${args[@]}" "$BASE$1" 2>/dev/null || echo "000")
  if [ "$code" != "$expected" ]; then echo "FAIL $1: HTTP $code (want $expected)"; FAILS=$((FAILS+1)); else echo "PASS $1 → $code"; fi
}

echo "=== Kingfisher Smoke Test ==="

# Public
check /health 200
check /version 200
check /api/v1/auth/login 400 POST '{"username":"x","password":"y"}'  # bad creds → 400
check /api/v1/auth/register 400 POST '{}'  # bad body → 400

# With token (requires Redis for full test, skip token-dependent in CI)
TOKEN=$(curl -sf -X POST "$BASE/api/v1/auth/login" -H 'Content-Type: application/json' -d '{"username":"admin","password":"Abcd1234"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['access_token'])" 2>/dev/null || echo "")
if [ -n "$TOKEN" ]; then
  AUTH="-H Authorization: Bearer $TOKEN"
  check /api/v1/users 200 GET
  
  # RBAC test — no auth should get 401
  code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/v1/users" 2>/dev/null || echo "000")
  [ "$code" = "401" ] && echo "PASS /api/v1/users (no auth) → 401" || { echo "FAIL /api/v1/users (no auth): HTTP $code (want 401)"; FAILS=$((FAILS+1)); }
fi

echo "=== $FAILS failures ==="
exit $FAILS
