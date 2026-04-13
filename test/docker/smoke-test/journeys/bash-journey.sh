#!/usr/bin/env bash
# ABOUTME: Bash user journey smoke test. Runs inside the smoke-test container.
# ABOUTME: Exercises the two user-reported bug scenarios (#432 local/global,
#          #433 pvm use PATH refresh) plus the full activation flow.

set -u
source "$(dirname "$0")/common.sh"

echo "=== bash smoke test ==="

# --- issue #432: pvm local must exist at top level --------------------------
local_help=$(pvm local --help 2>&1)
smoke_contains "$local_help" "local [version]" "pvm local is a top-level command"

global_help=$(pvm global --help 2>&1)
smoke_contains "$global_help" "global [version]" "pvm global is a top-level command"

use_help=$(pvm --help 2>&1)
smoke_contains "$use_help" "use [version" "pvm use listed in root help"

# --- prerequisite: import system perl, fake a version dir -------------------
pvm import-system >/dev/null 2>&1 || smoke_fail "import-system failed"

VERSION=$(pvm list 2>/dev/null | grep -oE '5\.[0-9]+\.[0-9]+' | head -1)
[ -n "$VERSION" ] || smoke_fail "no Perl version found after import-system"

mkdir -p "$XDG_DATA_HOME/pvm/versions/$VERSION/bin"
# Minimal stub perl so `perl -v` works
cat > "$XDG_DATA_HOME/pvm/versions/$VERSION/bin/perl" <<EOF
#!/bin/sh
echo "This is perl 5, version ${VERSION#5.}, subversion 0 (v$VERSION) stub"
EOF
chmod +x "$XDG_DATA_HOME/pvm/versions/$VERSION/bin/perl"

# --- issue #433 core: source integration, pvm use, PATH changes -------------
# shellcheck disable=SC1090
source <(pvm init bash)

pvm use "$VERSION" 2>&1 | grep -q "Using Perl $VERSION" \
    || smoke_fail "pvm use did not print 'Using Perl $VERSION'"
smoke_pass "pvm use prints Using Perl"

smoke_equals "$PVM_PERL_VERSION" "$VERSION" "PVM_PERL_VERSION exported"

front=$(echo "$PATH" | cut -d: -f1)
smoke_equals "$front" "$XDG_DATA_HOME/pvm/versions/$VERSION/bin" \
    "version bin at front of PATH after pvm use"

perl_out=$(perl -v 2>&1 | head -1)
smoke_contains "$perl_out" "v$VERSION" "perl -v reflects pvm use version"

# --- issue #433 I1: bogus version must exit non-zero ------------------------
pvm use not-a-real-version >/dev/null 2>&1
rc=$?
smoke_exit_eq "$rc" "1" "pvm use <bogus> returns non-zero exit"

# --- shell integration contract: 'pvm use' is usable via 'A && B' -----------
if pvm use not-a-real-version 2>/dev/null; then
    smoke_fail "pvm use <bogus> followed by && should NOT run the next command"
else
    smoke_pass "pvm use <bogus> short-circuits && chains"
fi

echo ""
echo "=== bash smoke test: all assertions passed ==="
