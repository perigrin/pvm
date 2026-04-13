#!/usr/bin/env fish
# ABOUTME: Fish user journey smoke test. Runs inside the smoke-test container.
# ABOUTME: Mirrors bash-journey.sh using fish idioms; especially valuable since
#          fish uncovered multiple template bugs during PR #434 development.

source (dirname (status -f))/common.fish

echo "=== fish smoke test ==="

# Top-level commands exist
set -l local_help (pvm local --help 2>&1 | string collect)
smoke_contains "$local_help" "local [version]" "pvm local is a top-level command"

set -l global_help (pvm global --help 2>&1 | string collect)
smoke_contains "$global_help" "global [version]" "pvm global is a top-level command"

# Import system perl + fake version dir
pvm import-system >/dev/null 2>&1
or smoke_fail "import-system failed"

set -l VERSION (pvm list 2>/dev/null | string match -rg '(5\.[0-9]+\.[0-9]+)' | head -1)
test -n "$VERSION"
or smoke_fail "no Perl version found after import-system"

mkdir -p "$XDG_DATA_HOME/pvm/versions/$VERSION/bin"
printf '#!/bin/sh\necho "This is perl 5, version %s, subversion 0 (v%s) stub"\n' \
    (string replace '5.' '' $VERSION) $VERSION \
    > "$XDG_DATA_HOME/pvm/versions/$VERSION/bin/perl"
chmod +x "$XDG_DATA_HOME/pvm/versions/$VERSION/bin/perl"

# Source pvm init — must not error
pvm init fish | source
or smoke_fail "sourcing pvm init fish failed"
smoke_pass "sourced pvm init fish cleanly"

# Run pvm use; capture output AND stderr to catch template errors
set -l use_out (pvm use "$VERSION" 2>&1 | string collect)
smoke_contains "$use_out" "Using Perl $VERSION" "pvm use prints Using Perl"
# No stray 'Unknown command: unset' or similar fish-parse errors
if string match -q "*Unknown command*" -- $use_out
    smoke_fail "pvm use produced a fish parse error: $use_out"
end
smoke_pass "pvm use output is fish-clean (no Unknown command errors)"

smoke_equals "$PVM_PERL_VERSION" "$VERSION" "PVM_PERL_VERSION exported"

set -l front (string split : -- $PATH | head -1)
smoke_equals "$front" "$XDG_DATA_HOME/pvm/versions/$VERSION/bin" \
    "version bin at front of PATH after pvm use"

set -l perl_out (perl -v 2>&1 | head -1 | string collect)
smoke_contains "$perl_out" "v$VERSION" "perl -v reflects pvm use version"

# Bogus version must exit non-zero (tests $pipestatus[1] plumbing)
pvm use not-a-real-version >/dev/null 2>&1
set -l rc $status
smoke_exit_eq "$rc" "1" "pvm use <bogus> returns non-zero exit"

# Completion must actually register (C2 regression)
set -l pvm_completions (complete -c pvm 2>/dev/null | count)
if test $pvm_completions -lt 5
    smoke_fail "fish completions did not register (found $pvm_completions rules)"
end
smoke_pass "fish completions registered ($pvm_completions rules)"

echo ""
echo "=== fish smoke test: all assertions passed ==="
