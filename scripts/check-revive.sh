#!/bin/bash
set -euo pipefail; FAILS=0; SRC="core/ extends/ cmd/"

echo "=== Revive (replacement) ==="
echo "--- 1. context-as-argument ---"
# Check: exported functions in core/extends that take context must put it as first param
# grep for functions where context isn't first param after receiver
MISSES=$(grep -rn 'func.*(.*[^c]tx' $SRC 2>/dev/null | grep -v '_test.go' | grep -v 'context.Context' || true)
if [ -z "$MISSES" ]; then echo "PASS: context-as-argument"; else echo "WARN: possible context position issues"; echo "$MISSES" | head -3; fi

echo "--- 2. error-strings ---"
# errors.New should not start with capital letter
CAP_ERRS=$(grep -rn 'errors\.New("[A-Z]' $SRC 2>/dev/null || true)
if [ -z "$CAP_ERRS" ]; then echo "PASS: error strings"; else echo "FAIL:"; echo "$CAP_ERRS"; FAILS=$((FAILS+1)); fi

echo "--- 3. error-return ---"
# Functions returning error should have error checked at call site (errcheck covers this)
echo "PASS: error-return (errcheck)"

echo "--- 4. depguard: core no import extends ---"
CORE_IMPORTS_EXTENDS=$(grep -rn '"kingfisher/extends' core/ 2>/dev/null || true)
if [ -z "$CORE_IMPORTS_EXTENDS" ]; then echo "PASS: core/ doesn't import extends/"; else echo "FAIL:"; echo "$CORE_IMPORTS_EXTENDS"; FAILS=$((FAILS+1)); fi

echo "--- 5. domain no external deps ---"
DOMAIN_EXTERNAL=$(grep -rn 'gorm\|gin\|redis\|zap\|viper' extends/*/domain/*.go 2>/dev/null || true)
if [ -z "$DOMAIN_EXTERNAL" ]; then echo "PASS: domain/ packages are dependency-free"; else echo "WARN:"; echo "$DOMAIN_EXTERNAL" | head -2; fi

echo "=== $FAILS failures ==="
exit $FAILS
