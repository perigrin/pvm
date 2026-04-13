#!/usr/bin/env bash
# ABOUTME: Bash user journey smoke test. Runs inside the smoke-test container.
# ABOUTME: Exercises the user-facing workflows a bash user would rely on:
#          local/global/use, env activation, library scoping, and the
#          originally-reported bug scenarios (#432/#433) as regression guards.

set -u
# Bash strips aliases from non-interactive scripts by default. The pvm init
# template installs `alias cd="pvm_cd"` as its auto-switch hook; we need
# expand_aliases ON so the G2 test below actually exercises that hook.
shopt -s expand_aliases
source "$(dirname "$0")/common.sh"

echo "=== bash smoke test ==="

############################################################################
# Section 1 — top-level command presence (#432 regression)
############################################################################

# Root --help must list all three commands with their usage strings.
root_help=$(pvm --help 2>&1)
smoke_contains "$root_help" "local [version]"  "pvm local is a top-level command"
smoke_contains "$root_help" "global [version]" "pvm global is a top-level command"
smoke_contains "$root_help" "use [version"     "pvm use listed in root help"

############################################################################
# Section 2 — setup: import system perl, create a stub bin
############################################################################

pvm import-system >/dev/null 2>&1 || smoke_fail "import-system failed"

VERSION=$(pvm list 2>/dev/null | grep -oE '5\.[0-9]+\.[0-9]+' | head -1)
[ -n "$VERSION" ] || smoke_fail "no Perl version found after import-system"

smoke_setup_stub_perl "$VERSION"

############################################################################
# Section 3 — pvm use: shell activation (#433 regression)
############################################################################

# shellcheck disable=SC1090
source <(pvm init bash)

# `pvm use` and `pvm env activate` modify the current shell's environment.
# Piping to grep or $(…) would run them in a subshell and the env-var
# exports would be lost. This helper captures output via a tempfile so the
# parent shell's env changes survive.
pvm_use_log=""
pvm_use_run() {
    pvm_use_log=$(mktemp)
    # Deliberately no pipe / no $(…) — caller just uses this helper and
    # then reads pvm_use_log to inspect output.
    pvm "$@" > "$pvm_use_log" 2>&1
}

pvm_use_run use "$VERSION"
smoke_contains "$(cat "$pvm_use_log")" "Using Perl $VERSION" "pvm use prints Using Perl"
rm -f "$pvm_use_log"

smoke_equals "${PVM_PERL_VERSION-}" "$VERSION" "PVM_PERL_VERSION exported"

smoke_equals "$(echo "$PATH" | cut -d: -f1)" \
             "$XDG_DATA_HOME/pvm/versions/$VERSION/bin" \
             "version bin at front of PATH after pvm use"

smoke_contains "$(perl -v 2>&1 | head -1)" "v$VERSION" \
               "perl -v reflects pvm use version"

# Bogus version: non-zero exit + short-circuits &&. The full-loop regression
# guard for #433 I1 (exit-code propagation through the shell function).
pvm use not-a-real-version >/dev/null 2>&1
smoke_exit_eq "$?" "1" "pvm use <bogus> returns non-zero exit"

if pvm use not-a-real-version 2>/dev/null; then
    smoke_fail "pvm use <bogus> followed by && should NOT run the next command"
else
    smoke_pass "pvm use <bogus> short-circuits && chains"
fi

############################################################################
# Section 3b — G2: cd auto-switch between TWO installed versions
############################################################################

# Install a second Perl via pvm's binary-fetch path so the registry
# knows two distinct versions, then cd between project dirs pinning
# each. This exercises the real JTBD: dev switches projects, shell
# auto-activates the right interpreter, no manual pvm use needed.
#
# This section needs network (to fetch the binary). The container
# sets PVM_SKIP_NETWORK_CALLS=1 as a default for determinism; only
# UNSET it locally for the install and leave it set for the rest of
# the suite. Skip gracefully if the network is unavailable.
SECOND_VERSION="5.40.2"
(
    unset PVM_SKIP_NETWORK_CALLS
    pvm install "$SECOND_VERSION" --binary-only >/tmp/install.log 2>&1
)
install_rc=$?
if [ "$install_rc" -ne 0 ]; then
    # Likely offline; downgrade to the single-version G2 test (pinning
    # proves the hook fires, we just can't prove inter-version switch).
    smoke_pass "skipping two-version auto-switch (install rc=$install_rc; network?)"
    unset PVM_PERL_VERSION
    mkdir -p "$HOME/proj-a"
    echo "$VERSION" > "$HOME/proj-a/.perl-version"
    cd "$HOME/proj-a"
    smoke_equals "$(pvm current --bare 2>/dev/null)" "$VERSION" \
                 "cd fires hook (single-version fallback): resolves pinned version"
    cd "$HOME"
else
    smoke_pass "pvm install --binary-only fetched $SECOND_VERSION"

    # Now exercise the auto-switch between two distinct REAL versions.
    unset PVM_PERL_VERSION
    mkdir -p "$HOME/proj-a" "$HOME/proj-b"
    echo "$VERSION"        > "$HOME/proj-a/.perl-version"
    echo "$SECOND_VERSION" > "$HOME/proj-b/.perl-version"

    cd "$HOME/proj-a"
    smoke_equals "$(pvm current --bare 2>/dev/null)" "$VERSION" \
                 "cd into proj-a: auto-switch resolves to $VERSION"
    smoke_equals "$(echo "$PATH" | cut -d: -f1)" \
                 "$XDG_DATA_HOME/pvm/versions/$VERSION/bin" \
                 "cd into proj-a: PATH front matches $VERSION's bin"

    cd "$HOME/proj-b"
    smoke_equals "$(pvm current --bare 2>/dev/null)" "$SECOND_VERSION" \
                 "cd into proj-b: auto-switch resolves to $SECOND_VERSION"
    smoke_equals "$(echo "$PATH" | cut -d: -f1)" \
                 "$XDG_DATA_HOME/pvm/versions/$SECOND_VERSION/bin" \
                 "cd into proj-b: PATH front matches $SECOND_VERSION's bin"

    # cd back to the first project; PATH must swing back.
    cd "$HOME/proj-a"
    smoke_equals "$(echo "$PATH" | cut -d: -f1)" \
                 "$XDG_DATA_HOME/pvm/versions/$VERSION/bin" \
                 "cd back to proj-a: PATH front swings back to $VERSION's bin"

    cd "$HOME"
fi
rm -f /tmp/install.log

############################################################################
# Section 4 — pvm local: project-scoped pinning
############################################################################

# Reset the shell's env var so the .perl-version file is what resolves.
pvm use "$VERSION" >/dev/null 2>&1

cd "$HOME"
pvm local "$VERSION" 2>&1 | grep -q "Local Perl version set" \
    || smoke_fail "pvm local did not confirm write"
smoke_pass "pvm local confirms write"
[ -f "$HOME/.perl-version" ] || smoke_fail "pvm local did not create .perl-version"
smoke_equals "$(cat "$HOME/.perl-version")" "$VERSION" "pvm local wrote $VERSION"

# detect-version finds it.
smoke_contains "$(pvm detect-version 2>&1)" ".perl-version" \
               "pvm detect-version sees the pin"

pvm local --unset 2>&1 | grep -q "unset" \
    || smoke_fail "pvm local --unset did not confirm"
[ ! -f "$HOME/.perl-version" ] || smoke_fail "pvm local --unset did not remove the file"
smoke_pass "pvm local --unset removes .perl-version"

############################################################################
# Section 5 — pvm global: user-scoped default
############################################################################

# Clear any shell-level override so global is actually consulted.
unset PVM_PERL_VERSION

pvm global "$VERSION" 2>&1 | grep -q "Global Perl version set" \
    || smoke_fail "pvm global did not confirm"
smoke_pass "pvm global confirms write"
smoke_contains "$(pvm current --bare 2>&1)" "$VERSION" \
               "pvm current --bare reads pvm global"

pvm global --unset 2>&1 | grep -q "unset" \
    || smoke_fail "pvm global --unset did not confirm"
smoke_pass "pvm global --unset clears default"

############################################################################
# Section 6 — pvm use system: clears the env-var override
############################################################################

pvm use "$VERSION" >/dev/null 2>&1
smoke_equals "$PVM_PERL_VERSION" "$VERSION" "sanity: version active before use system"

pvm_use_run use system
smoke_contains "$(cat "$pvm_use_log")" "Using system Perl" \
               "pvm use system prints 'Using system Perl'"
rm -f "$pvm_use_log"
[ -z "${PVM_PERL_VERSION-}" ] || smoke_fail "pvm use system did not clear PVM_PERL_VERSION"
smoke_pass "pvm use system clears PVM_PERL_VERSION"

############################################################################
# Section 7 — library environments (@library syntax + env lifecycle)
############################################################################

# Create a fake library env on disk so pvm env list/activate can see it.
mkdir -p "$XDG_DATA_HOME/pvm/environments/testlib/bin"
mkdir -p "$XDG_DATA_HOME/pvm/environments/testlib/lib/perl5"

smoke_contains "$(pvm env list 2>&1)" "testlib" "pvm env list shows testlib"

# pvm use X@library sets PVM_PERL_LIBRARY; shell integration must honor it.
pvm_use_run use "$VERSION@testlib"
smoke_contains "$(cat "$pvm_use_log")" "Using Perl $VERSION@testlib" \
               "pvm use X@lib prints 'Using Perl X@lib'"
rm -f "$pvm_use_log"
smoke_equals "${PVM_PERL_LIBRARY-}" "testlib" "PVM_PERL_LIBRARY exported by sh-use"

# _pvm_update_perl_path should now have library bin + perl5 on PATH/PERL5LIB.
# (This is PR #434's I5 fix; bash already had the handling, but this locks
# it in as a user-visible journey rather than template text.)
case ":$PATH:" in
    *":$XDG_DATA_HOME/pvm/environments/testlib/bin:"*)
        smoke_pass "library bin is on PATH after pvm use X@library" ;;
    *)
        smoke_fail "library bin NOT on PATH after pvm use X@library; PATH=$PATH" ;;
esac

case ":${PERL5LIB:-}:" in
    *":$XDG_DATA_HOME/pvm/environments/testlib/lib/perl5:"*)
        smoke_pass "library lib/perl5 is on PERL5LIB after pvm use X@library" ;;
    *)
        smoke_fail "library lib/perl5 NOT on PERL5LIB; PERL5LIB=${PERL5LIB:-<unset>}" ;;
esac

# Clear the library by switching to a version without @library.
pvm use "$VERSION" >/dev/null 2>&1
[ -z "${PVM_PERL_LIBRARY:-}" ] || smoke_fail "pvm use X (no @) did not clear PVM_PERL_LIBRARY"
smoke_pass "pvm use X (no @) clears PVM_PERL_LIBRARY"

# pvm env remove is interactive ([y/N]). Pipe 'y' to confirm.
rm_out=$(echo y | pvm env remove testlib 2>&1)
smoke_contains "$rm_out" "has been removed" "pvm env remove confirms removal"
[ ! -d "$XDG_DATA_HOME/pvm/environments/testlib" ] \
    || smoke_fail "pvm env remove did not delete the directory"
smoke_pass "pvm env remove tears the env down"

############################################################################
# Section 8 — doctor sanity in this fully-configured shell
############################################################################

# After all the setup above, self doctor should still report a healthy
# install (exit 0) and NOT falsely flag the running binary as stale.
doctor_out=$(pvm self doctor 2>&1)
smoke_exit_eq "$?" "0" "pvm self doctor exits 0 in bash journey"
# The stale-install warning (PR #436) must not fire here — container has
# exactly one pvm binary at /usr/local/bin/pvm.
if echo "$doctor_out" | grep -q "Multiple pvm binaries in PATH"; then
    smoke_fail "self doctor incorrectly flagged stale installs in clean container"
fi
smoke_pass "self doctor: no false stale-install warning"

echo ""
echo "=== bash smoke test: all assertions passed ==="
