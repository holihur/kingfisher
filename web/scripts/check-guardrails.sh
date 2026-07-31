#!/bin/bash
set -euo pipefail
FAILS=0
SRC="extends/ core/ cmd/"

echo "=== Kingfisher Guardrails ==="

echo "--- 1. No panic() ---"
PANICS=$(grep -rn 'panic(' $SRC 2>/dev/null | grep -v 'main.go' | grep -v 'middleware/recovery.go' | grep -v '_test.go' || true)
if [ -z "$PANICS" ]; then echo "PASS: no panic()"; else echo "FAIL:"; echo "$PANICS"; FAILS=$((FAILS+1)); fi

echo "--- 2. No log.Fatal ---"
FATALS=$(grep -rn 'log\.Fatal' $SRC 2>/dev/null | grep -v 'main.go' || true)
if [ -z "$FATALS" ]; then echo "PASS: no log.Fatal"; else echo "FAIL:"; echo "$FATALS"; FAILS=$((FAILS+1)); fi

echo "--- 3. No fmt.Println ---"
PRINTS=$(grep -rn 'fmt\.Println\|fmt\.Print(' $SRC 2>/dev/null | grep -v 'main.go' | grep -v '_test.go' || true)
if [ -z "$PRINTS" ]; then echo "PASS: no fmt.Println"; else echo "FAIL:"; echo "$PRINTS"; FAILS=$((FAILS+1)); fi

echo "--- 4. No hardcoded secrets ---"
SECRETS=$(grep -rnE 'password.*[:=] *"[^"]+"|secret.*[:=] *"[^"]+"' $SRC 2>/dev/null | grep -v '_test.go' | grep -v 'config.yaml' || true)
if [ -z "$SECRETS" ]; then echo "PASS: no secrets"; else echo "FAIL:"; echo "$SECRETS"; FAILS=$((FAILS+1)); fi

echo "--- 5. go mod tidy ---"
go mod tidy 2>/dev/null
if git diff --exit-code go.mod go.sum >/dev/null 2>&1; then echo "PASS: go mod tidy"; else echo "FAIL: go.mod not tidy"; FAILS=$((FAILS+1)); fi

echo "=== $FAILS failures ==="
exit $FAILS
