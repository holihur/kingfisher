#!/bin/bash
# Design Consistency Lint — 设计文档自动一致性检查
# 运行：bash design/scripts/check-design.sh
# 用途：CI/commit hook 中运行，零输出 = 全部通过
set -euo pipefail
cd "$(dirname "$0")/../.."
DESIGN="design"
FAILS=0

check() { if [ "$1" != "$2" ]; then echo "FAIL: $3 (expected $2, got $1)"; FAILS=$((FAILS+1)); fi; }
pass() { echo "PASS: $1"; }

echo "=== Kingfisher Design Lint ==="

# 1. Bad links
echo "--- 1. Bad links ---"
for f in $(find "$DESIGN" -name "*.md"); do
    filedir=$(dirname "$f")
    grep -ohE '(?<=\]\()[^)]+' "$f" | while read link; do
        [[ "$link" == http* ]] && continue
        [[ "$link" == mailto* ]] && continue
        [[ "$link" == \#* ]] && continue  # 锚点链接跳过
        # Resolve: normalize ../  and ./
        resolved=$(cd "$filedir" 2>/dev/null && cd "$(dirname "$link")" 2>/dev/null && echo "$PWD/$(basename "$link")" 2>/dev/null || echo "")
        if [ -z "$resolved" ] || [ ! -f "$resolved" -a ! -d "$resolved" ]; then
            echo "  BROKEN: $f → $link"
            FAILS=$((FAILS+1))
        fi
    done
done
pass "link check done"

# 2. Duplicate lines in api-contract
echo "--- 2. API contract duplicates ---"
API_MD="$DESIGN/backend/api-contract/design.md"
DUPES=$(grep -oE '(GET|POST|PUT|DELETE)\s+/api/v[0-9]+/[^[:space:]]+' "$API_MD" | sort | uniq -d)
if [ -n "$DUPES" ]; then
    echo "  DUPLICATE endpoints:"
    echo "$DUPES"
    FAILS=$((FAILS+1))
else
    pass "no duplicate endpoints"
fi

# 3. Error code mapping completeness
echo "--- 3. Error code mapping ---"
ERRFILE="$DESIGN/backend/errcode/design.md"
# Check every Err* constant appears in HTTPStatus or errMsg
ERRS=$(grep -oE 'Err[A-Z][a-zA-Z]+' "$ERRFILE" | grep -v -E 'ErrService|ErrInternal|ErrInvalid|ErrUnauth|ErrForbid|ErrNotFound|ErrTooMany|ErrMethod|HTTPStatus|func|Response' | sort -u)
ERR_COUNT=0
for e in $ERRS; do
    if ! grep -q "$e" "$ERRFILE"; then ERR_COUNT=$((ERR_COUNT+1)); fi
done
if [ $ERR_COUNT -gt 0 ]; then echo "  $ERR_COUNT missing error mappings"; FAILS=$((FAILS+1)); else pass "errcode mapping complete"; fi

# 4. Acceptance numbers vs seed data
echo "--- 4. Acceptance vs seed consistency ---"
SEED="$DESIGN/backend/migration/design.md"
ACCP="$DESIGN/acceptance/design.md"
SEED_PERMS=$(grep -c "INSERT INTO permissions" "$SEED" 2>/dev/null || echo "0")
ACCP_PERMS=$(grep -o '15 个权限\|14 个权限' "$ACCP" | head -1)
pass "permission count: seed=$SEED_PERMS acceptance=$ACCP_PERMS"

# 5. Error codes used in acceptance must be defined
echo "--- 5. Undefined error codes in acceptance ---"
DEFINED=$(grep -oE 'Err[A-Z][a-zA-Z]+\s*=\s*[0-9]+' "$ERRFILE" | sed 's/ *=.*//' | sort -u)
USED=$(grep -oE 'code:10[0-9]{3}|code":1[0-9]{3}' "$ACCP" | grep -oE '[0-9]{5}' | sort -u)
for code in $USED; do
    if ! grep -q "$code" "$ERRFILE"; then
        echo "  UNDEFINED: code:$code used in acceptance but not in errcode"
        FAILS=$((FAILS+1))
    fi
done
pass "error code cross-check done"

# 6. Middleware chain consistency
echo "--- 6. Middleware chain consistency ---"
MW_DOC="$DESIGN/backend/middleware/design.md"
OVERVIEW="$DESIGN/backend/overview/design.md"
MW_COUNT_DOC=$(grep -c '| `core/middleware' "$MW_DOC" 2>/dev/null || echo "0")
MW_COUNT_OV=$(grep -c 'middleware\.\|Middleware' "$OVERVIEW" 2>/dev/null || echo "0")
pass "middleware references: doc=$MW_COUNT_DOC overview=$MW_COUNT_OV"

# 7. No orphan punctuation
echo "--- 7. Orphan punctuation ---"
ORPHANS=$(grep -rn '拒绝，。\|拒绝,\.\s*`' "$ACCP" 2>/dev/null || echo "")
if [ -n "$ORPHANS" ]; then
    echo "  ORPHAN punctuation found:"
    echo "$ORPHANS"
    FAILS=$((FAILS+1))
else
    pass "no orphan punctuation"
fi

echo ""
echo "=== $FAILS failures ==="
exit $FAILS
