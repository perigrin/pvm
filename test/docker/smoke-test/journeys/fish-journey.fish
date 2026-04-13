#!/usr/bin/env fish
# ABOUTME: Fish user journey smoke test. Runs inside the smoke-test container.
# ABOUTME: Mirrors bash-journey.sh using fish idioms. Especially valuable: fish
#          uncovered multiple template bugs during PR #434 and was the shell
#          that motivated the smoke-test container in the first place.

source (dirname (status -f))/common.fish

echo "=== fish smoke test ==="

############################################################################
# Section 1 — top-level command presence (#432 regression)
############################################################################

set -l root_help (pvm --help 2>&1 | string collect)
smoke_contains "$root_help" "local [version]"  "pvm local is a top-level command"
smoke_contains "$root_help" "global [version]" "pvm global is a top-level command"
smoke_contains "$root_help" "use [version"     "pvm use listed in root help"

############################################################################
# Section 2 — setup
############################################################################

pvm import-system >/dev/null 2>&1
or smoke_fail "import-system failed"

set -l requested_version (pvm list 2>/dev/null | string match -rg '(5\.[0-9]+\.[0-9]+)' | head -1)
test -n "$requested_version"
or smoke_fail "no Perl version found after import-system"

smoke_setup_stub_perl "$requested_version"

############################################################################
# Section 3 — pvm use: shell activation (#433 regression)
############################################################################

pvm init fish | source
or smoke_fail "sourcing pvm init fish failed"
smoke_pass "sourced pvm init fish cleanly"

# `pvm use` / `pvm env activate` modify the shell's env. Running them under
# `(…)` command substitution puts them in a subshell and the exports are
# lost. Use a tempfile to inspect output while keeping the env changes.
set -g pvm_use_log ""
function pvm_use_run
    set -g pvm_use_log (mktemp)
    pvm $argv > $pvm_use_log 2>&1
end

pvm_use_run use "$requested_version"
set -l use_out (cat $pvm_use_log | string collect)
smoke_contains "$use_out" "Using Perl $requested_version" "pvm use prints Using Perl"
if string match -q "*Unknown command*" -- $use_out
    smoke_fail "pvm use produced a fish parse error: $use_out"
end
smoke_pass "pvm use output is fish-clean (no Unknown command errors)"
rm -f $pvm_use_log

smoke_equals "$PVM_PERL_VERSION" "$requested_version" "PVM_PERL_VERSION exported"

set -l front (string split : -- $PATH | head -1)
smoke_equals "$front" "$XDG_DATA_HOME/pvm/versions/$requested_version/bin" \
    "version bin at front of PATH after pvm use"

set -l perl_out (perl -v 2>&1 | head -1 | string collect)
smoke_contains "$perl_out" "v$requested_version" "perl -v reflects pvm use version"

pvm use not-a-real-version >/dev/null 2>&1
smoke_exit_eq "$status" "1" "pvm use <bogus> returns non-zero exit"

pvm use not-a-real-version 2>/dev/null
and smoke_fail "pvm use <bogus> followed by `; and` should NOT run the next command"
smoke_pass "pvm use <bogus> short-circuits `; and` chains"

# Completions must register (C2 regression from PR #434). Note: the init
# template gates the auto-registration on PVM_SKIP_NETWORK_CALLS, which is
# set to 1 in this container to suppress MetaCPAN calls. Source the
# completion output manually here — it's the same code path that runs at
# interactive startup, minus the gate.
pvm completion fish 2>/dev/null | source
set -l pvm_completions (complete -c pvm 2>/dev/null | count)
if test $pvm_completions -lt 5
    smoke_fail "fish completions did not register (found $pvm_completions rules)"
end
smoke_pass "fish completions registered ($pvm_completions rules)"

############################################################################
# Section 4 — pvm local
############################################################################

pvm use "$requested_version" >/dev/null 2>&1
cd "$HOME"

pvm local "$requested_version" 2>&1 | grep -q "Local Perl version set"
or smoke_fail "pvm local did not confirm"
smoke_pass "pvm local confirms write"
test -f "$HOME/.perl-version"
or smoke_fail "pvm local did not create .perl-version"
smoke_equals (cat "$HOME/.perl-version") "$requested_version" "pvm local wrote $requested_version"

smoke_contains (pvm detect-version 2>&1 | string collect) ".perl-version" \
    "pvm detect-version sees the pin"

pvm local --unset 2>&1 | grep -q "unset"
or smoke_fail "pvm local --unset failed"
test ! -f "$HOME/.perl-version"
or smoke_fail "pvm local --unset did not remove file"
smoke_pass "pvm local --unset removes .perl-version"

############################################################################
# Section 5 — pvm global
############################################################################

set -e PVM_PERL_VERSION

pvm global "$requested_version" 2>&1 | grep -q "Global Perl version set"
or smoke_fail "pvm global did not confirm"
smoke_pass "pvm global confirms write"
smoke_contains (pvm current --bare 2>&1 | string collect) "$requested_version" \
    "pvm current --bare reads pvm global"

pvm global --unset 2>&1 | grep -q "unset"
or smoke_fail "pvm global --unset failed"
smoke_pass "pvm global --unset clears default"

############################################################################
# Section 6 — pvm use system
############################################################################

pvm use "$requested_version" >/dev/null 2>&1
smoke_equals "$PVM_PERL_VERSION" "$requested_version" "sanity: version active before use system"

pvm_use_run use system
smoke_contains (cat $pvm_use_log | string collect) "Using system Perl" \
    "pvm use system prints 'Using system Perl'"
rm -f $pvm_use_log
if set -q PVM_PERL_VERSION
    smoke_fail "pvm use system did not clear PVM_PERL_VERSION"
end
smoke_pass "pvm use system clears PVM_PERL_VERSION"

############################################################################
# Section 7 — library environments (main fix target of PR #434 I5 for fish)
############################################################################

mkdir -p "$XDG_DATA_HOME/pvm/environments/testlib/bin"
mkdir -p "$XDG_DATA_HOME/pvm/environments/testlib/lib/perl5"

smoke_contains (pvm env list 2>&1 | string collect) "testlib" "pvm env list shows testlib"

pvm_use_run use "$requested_version@testlib"
smoke_contains (cat $pvm_use_log | string collect) "Using Perl $requested_version@testlib" \
    "pvm use X@lib prints 'Using Perl X@lib'"
rm -f $pvm_use_log
smoke_equals "$PVM_PERL_LIBRARY" "testlib" "PVM_PERL_LIBRARY exported"

if contains "$XDG_DATA_HOME/pvm/environments/testlib/bin" $PATH
    smoke_pass "library bin is on PATH after pvm use X@library"
else
    smoke_fail "library bin NOT on PATH after pvm use X@library; PATH=$PATH"
end

set -l perl5lib_parts
if set -q PERL5LIB
    set perl5lib_parts (string split : -- $PERL5LIB)
end
if contains "$XDG_DATA_HOME/pvm/environments/testlib/lib/perl5" $perl5lib_parts
    smoke_pass "library lib/perl5 is on PERL5LIB after pvm use X@library"
else
    smoke_fail "library lib/perl5 NOT on PERL5LIB; PERL5LIB="(set -q PERL5LIB; and echo $PERL5LIB; or echo '<unset>')
end

pvm use "$requested_version" >/dev/null 2>&1
if set -q PVM_PERL_LIBRARY
    smoke_fail "pvm use X (no @) did not clear PVM_PERL_LIBRARY"
end
smoke_pass "pvm use X (no @) clears PVM_PERL_LIBRARY"

set -l rm_out (echo y | pvm env remove testlib 2>&1 | string collect)
smoke_contains "$rm_out" "has been removed" "pvm env remove confirms removal"
test ! -d "$XDG_DATA_HOME/pvm/environments/testlib"
or smoke_fail "pvm env remove did not delete the directory"
smoke_pass "pvm env remove tears the env down"

############################################################################
# Section 8 — doctor sanity
############################################################################

set -l doctor_out (pvm self doctor 2>&1 | string collect)
smoke_exit_eq "$status" "0" "pvm self doctor exits 0 in fish journey"
if string match -q "*Multiple pvm binaries in PATH*" -- $doctor_out
    smoke_fail "self doctor incorrectly flagged stale installs in clean container"
end
smoke_pass "self doctor: no false stale-install warning"

echo ""
echo "=== fish smoke test: all assertions passed ==="
