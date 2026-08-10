#!/bin/bash
# Guardrail checks per design/backend/guardrails/design.md
# 任何检查失败 → exit 1。CI 中强制执行。
set -euo pipefail
FAILS=0
SRC="extends/ core/ cmd/"
TMPDIR=$(mktemp -d)

trap 'rm -rf "$TMPDIR"' EXIT

check_clean() {
    local label="$1" pattern="$2" exclude="$3"
    local matches
    matches=$(grep -rn "$pattern" $SRC 2>/dev/null | grep -v "$exclude" | grep -v go || true)
    if [ -z "$matches" ]; then
        echo "PASS: $label"
    else
        echo "FAIL: $label"
        echo "$matches"
        FAILS=$((FAILS + 1))
    fi
}

echo "=== Kingfisher Guardrails ==="

# 1. No panic in business code (non-main, non-recovery)
echo "--- 1. No panic() ---"
PANICS=$(grep -rn 'panic(' $SRC 2>/dev/null | grep -v 'main.go' | grep -v 'middleware/recovery.go' | grep -v 'logger/logger.go' | grep -v '_test.go' || true)
if [ -z "$PANICS" ]; then echo "PASS: no panic()"; else echo "FAIL: panic()"; echo "$PANICS"; FAILS=$((FAILS+1)); fi

# 2. No log.Fatal in business code
echo "--- 2. No log.Fatal ---"
FATALS=$(grep -rn 'log\.Fatal' $SRC 2>/dev/null | grep -v 'main.go' || true)
if [ -z "$FATALS" ]; then echo "PASS: no log.Fatal"; else echo "FAIL: log.Fatal"; echo "$FATALS"; FAILS=$((FAILS+1)); fi

# 3. No fmt.Println/fmt.Print in business code
echo "--- 3. No fmt.Println ---"
PRINTS=$(grep -rn 'fmt\.Println\|fmt\.Print(' $SRC 2>/dev/null | grep -v 'main.go' | grep -v '_test.go' || true)
if [ -z "$PRINTS" ]; then echo "PASS: no fmt.Println"; else echo "FAIL: fmt.Println"; echo "$PRINTS"; FAILS=$((FAILS+1)); fi

# 4. No hardcoded passwords/secrets
echo "--- 4. No hardcoded secrets ---"
SECRETS=$(grep -rnE 'password: *"[^"]+"|Password:[[:space:]]*"[^"]+"|secret: *"[^"]+"|Secret: *"[^"]+"' $SRC 2>/dev/null | grep -v '_test.go' | grep -v 'config.yaml' | grep -v 'design/' | grep -v '#nosec' | grep -v 'bcrypt' || true)
if [ -z "$SECRETS" ]; then echo "PASS: no secrets"; else echo "FAIL: hardcoded secrets"; echo "$SECRETS"; FAILS=$((FAILS+1)); fi

# 5. Ensure go mod tidy is clean
echo "--- 5. go mod tidy ---"
go mod tidy 2>/dev/null
if git diff --exit-code go.mod go.sum > /dev/null 2>&1; then echo "PASS: go mod tidy"; else echo "FAIL: go.mod not tidy — run go mod tidy"; FAILS=$((FAILS+1)); fi

echo ""
echo "=== $FAILS guardrail failures ==="
exit $FAILS
