#!/usr/bin/env bash
set -eo pipefail

echo "Running Time & Soft-Delete Guard Checks..."

ERRORS=0

# 1. Check for .Unix() calls in internal/ or pkg/ (excluding test files)
UNIX_MATCHES=$(grep -rn --include="*.go" "\.Unix()" internal/ pkg/ | grep -v "_test.go" || true)
if [ -n "$UNIX_MATCHES" ]; then
    echo "ERROR: Forbidden .Unix() usage found outside tests:"
    echo "$UNIX_MATCHES"
    ERRORS=$((ERRORS + 1))
fi

# 2. Check for date.Location() or time.Local in internal/ (excluding test files)
LOC_MATCHES=$(grep -rn --include="*.go" -E "date\.Location\(\)|time\.Local" internal/ | grep -v "_test.go" || true)
if [ -n "$LOC_MATCHES" ]; then
    echo "ERROR: Forbidden implicit timezone usage (date.Location()/time.Local) found:"
    echo "$LOC_MATCHES"
    ERRORS=$((ERRORS + 1))
fi

# 3. Check for Date.prototype.toLocale* without timeZone in TS/TSX files
TOLOCALE_MATCHES=$(grep -rn --include="*.ts" --include="*.tsx" -E "new Date\([^)]*\)\.toLocale(String|DateString|TimeString)\(\)" apps/ packages/ | grep -v "\.test\." | grep -v "\.spec\." | grep -v "node_modules" || true)
if [ -n "$TOLOCALE_MATCHES" ]; then
    echo "ERROR: Forbidden toLocale*() call without timeZone parameter found:"
    echo "$TOLOCALE_MATCHES"
    ERRORS=$((ERRORS + 1))
fi

# 4. Check for deleted_at IS NULL in Go or SQL files
NULL_MATCHES=$(grep -rn --include="*.go" --include="*.sql" "deleted_at IS NULL" internal/ pkg/ db/migrations/ | grep -v "_test.go" || true)
if [ -n "$NULL_MATCHES" ]; then
    echo "ERROR: Forbidden 'deleted_at IS NULL' found in Go code or SQL migrations:"
    echo "$NULL_MATCHES"
    ERRORS=$((ERRORS + 1))
fi

# 5. Check for AutoMigrate anywhere in internal/ or cmd/
AUTOMIGRATE_MATCHES=$(grep -rn --include="*.go" "AutoMigrate" internal/ cmd/ | grep -v "_test.go" || true)
if [ -n "$AUTOMIGRATE_MATCHES" ]; then
    echo "ERROR: Forbidden AutoMigrate call found in application code:"
    echo "$AUTOMIGRATE_MATCHES"
    ERRORS=$((ERRORS + 1))
fi

if [ $ERRORS -gt 0 ]; then
    echo "Guard check failed with $ERRORS error(s)."
    exit 1
else
    echo "All Time & Soft-Delete Guard Checks passed!"
    exit 0
fi
