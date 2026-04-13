#!/usr/bin/env zsh
# ABOUTME: Zsh user journey smoke test. Runs inside the smoke-test container.
# ABOUTME: Mirrors bash-journey.sh section-for-section with zsh-specific
#          sourcing. Sections: presence, setup, use, local, global, use system,
#          @library / env lifecycle, doctor.

set -u
source "$(dirname "$0")/common.sh"

echo "=== zsh smoke test ==="

############################################################################
# Section 1 — top-level command presence (#432 regression)
############################################################################

root_help=$(pvm --help 2>&1)
smoke_contains "$root_help" "local [version]"  "pvm local is a top-level command"
smoke_contains "$root_help" "global [version]" "pvm global is a top-level command"
smoke_contains "$root_help" "use [version"     "pvm use listed in root help"

############################################################################
# Section 2 — setup
############################################################################

pvm import-system >/dev/null 2>&1 || smoke_fail "import-system failed"

VERSION=$(pvm list 2>/dev/null | grep -oE '5\.[0-9]+\.[0-9]+' | head -1)
[ -n "$VERSION" ] || smoke_fail "no Perl version found after import-system"

smoke_setup_stub_perl "$VERSION"

############################################################################
# Section 3 — pvm use: shell activation (#433 regression)
############################################################################

# shellcheck disable=SC1090
source <(pvm init zsh)

# `pvm use` / `pvm env activate` modify the shell env. Running them under
# a pipe or $(…) puts them in a subshell and the exports are lost.
pvm_use_log=""
pvm_use_run() {
    pvm_use_log=$(mktemp)
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

SECOND_VERSION="5.40.2"
(
    unset PVM_SKIP_NETWORK_CALLS
    pvm install "$SECOND_VERSION" --binary-only >/tmp/install.log 2>&1
)
install_rc=$?
if [ "$install_rc" -ne 0 ]; then
    smoke_pass "skipping two-version auto-switch (install rc=$install_rc; network?)"
    unset PVM_PERL_VERSION
    mkdir -p "$HOME/proj-a"
    echo "$VERSION" > "$HOME/proj-a/.perl-version"
    cd "$HOME/proj-a"
    smoke_equals "$(pvm current --bare 2>/dev/null)" "$VERSION" \
                 "cd fires chpwd hook (single-version fallback): resolves pinned version"
    cd "$HOME"
else
    smoke_pass "pvm install --binary-only fetched $SECOND_VERSION"

    unset PVM_PERL_VERSION
    mkdir -p "$HOME/proj-a" "$HOME/proj-b"
    echo "$VERSION"        > "$HOME/proj-a/.perl-version"
    echo "$SECOND_VERSION" > "$HOME/proj-b/.perl-version"

    cd "$HOME/proj-a"
    smoke_equals "$(pvm current --bare 2>/dev/null)" "$VERSION" \
                 "cd into proj-a: chpwd hook resolves to $VERSION"
    smoke_equals "$(echo "$PATH" | cut -d: -f1)" \
                 "$XDG_DATA_HOME/pvm/versions/$VERSION/bin" \
                 "cd into proj-a: PATH front matches $VERSION's bin"

    cd "$HOME/proj-b"
    smoke_equals "$(pvm current --bare 2>/dev/null)" "$SECOND_VERSION" \
                 "cd into proj-b: chpwd hook resolves to $SECOND_VERSION"
    smoke_equals "$(echo "$PATH" | cut -d: -f1)" \
                 "$XDG_DATA_HOME/pvm/versions/$SECOND_VERSION/bin" \
                 "cd into proj-b: PATH front matches $SECOND_VERSION's bin"

    cd "$HOME/proj-a"
    smoke_equals "$(echo "$PATH" | cut -d: -f1)" \
                 "$XDG_DATA_HOME/pvm/versions/$VERSION/bin" \
                 "cd back to proj-a: PATH front swings back to $VERSION's bin"

    cd "$HOME"
fi
rm -f /tmp/install.log

############################################################################
# Section 4 — pvm local
############################################################################

pvm use "$VERSION" >/dev/null 2>&1
cd "$HOME"

pvm local "$VERSION" 2>&1 | grep -q "Local Perl version set" \
    || smoke_fail "pvm local did not confirm"
smoke_pass "pvm local confirms write"
[ -f "$HOME/.perl-version" ] || smoke_fail "pvm local did not create .perl-version"
smoke_equals "$(cat "$HOME/.perl-version")" "$VERSION" "pvm local wrote $VERSION"

smoke_contains "$(pvm detect-version 2>&1)" ".perl-version" \
               "pvm detect-version sees the pin"

pvm local --unset 2>&1 | grep -q "unset" || smoke_fail "pvm local --unset failed"
[ ! -f "$HOME/.perl-version" ] || smoke_fail "pvm local --unset did not remove file"
smoke_pass "pvm local --unset removes .perl-version"

############################################################################
# Section 5 — pvm global
############################################################################

unset PVM_PERL_VERSION

pvm global "$VERSION" 2>&1 | grep -q "Global Perl version set" \
    || smoke_fail "pvm global did not confirm"
smoke_pass "pvm global confirms write"
smoke_contains "$(pvm current --bare 2>&1)" "$VERSION" \
               "pvm current --bare reads pvm global"

pvm global --unset 2>&1 | grep -q "unset" || smoke_fail "pvm global --unset failed"
smoke_pass "pvm global --unset clears default"

############################################################################
# Section 6 — pvm use system
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
# Section 7 — library environments (regression target: PR #434 I5 was
#            specifically about the zsh template missing PVM_PERL_LIBRARY
#            handling in _pvm_update_perl_path)
############################################################################

mkdir -p "$XDG_DATA_HOME/pvm/environments/testlib/bin"
mkdir -p "$XDG_DATA_HOME/pvm/environments/testlib/lib/perl5"

smoke_contains "$(pvm env list 2>&1)" "testlib" "pvm env list shows testlib"

pvm_use_run use "$VERSION@testlib"
smoke_contains "$(cat "$pvm_use_log")" "Using Perl $VERSION@testlib" \
               "pvm use X@lib prints 'Using Perl X@lib'"
rm -f "$pvm_use_log"
smoke_equals "${PVM_PERL_LIBRARY-}" "testlib" "PVM_PERL_LIBRARY exported"

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

pvm use "$VERSION" >/dev/null 2>&1
[ -z "${PVM_PERL_LIBRARY:-}" ] || smoke_fail "pvm use X (no @) did not clear PVM_PERL_LIBRARY"
smoke_pass "pvm use X (no @) clears PVM_PERL_LIBRARY"

rm_out=$(echo y | pvm env remove testlib 2>&1)
smoke_contains "$rm_out" "has been removed" "pvm env remove confirms removal"
[ ! -d "$XDG_DATA_HOME/pvm/environments/testlib" ] \
    || smoke_fail "pvm env remove did not delete the directory"
smoke_pass "pvm env remove tears the env down"

############################################################################
# Section 8 — doctor sanity
############################################################################

doctor_out=$(pvm self doctor 2>&1)
smoke_exit_eq "$?" "0" "pvm self doctor exits 0 in zsh journey"
if echo "$doctor_out" | grep -q "Multiple pvm binaries in PATH"; then
    smoke_fail "self doctor incorrectly flagged stale installs in clean container"
fi
smoke_pass "self doctor: no false stale-install warning"

echo ""
echo "=== zsh smoke test: all assertions passed ==="
