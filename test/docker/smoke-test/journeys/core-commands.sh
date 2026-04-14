#!/usr/bin/env bash
# ABOUTME: Shell-agnostic smoke suite. Covers commands whose behavior does
# ABOUTME: not depend on which shell invokes them, so we run them once (from
#          bash) instead of three times across bash/zsh/fish.

set -u
source "$(dirname "$0")/common.sh"

echo "=== core commands smoke test (shell-agnostic) ==="

# --- identification --------------------------------------------------------

smoke_contains "$(pvm version 2>&1)" "pvm " "pvm version prints version line"
# Any top-level --help succeeds and mentions the command name.
smoke_contains "$(pvm --help 2>&1)" "Perl Version Manager" "pvm --help shows description"

# Every documented top-level command we advertise in gh-pages/reference/pvm.html
# must exist and at least accept --help without error. Cheap sanity that the
# root AddCommand block doesn't silently lose entries (issue #432 class).
for cmd in install uninstall versions list use global local current available \
           rehash import-system init completion self env remote config; do
    out=$(pvm "$cmd" --help 2>&1)
    rc=$?
    if [ "$rc" -ne 0 ]; then
        smoke_fail "pvm $cmd --help exited $rc (command missing or broken)"
    fi
    smoke_pass "pvm $cmd --help exits 0"
done

# --- version listing & registry --------------------------------------------

pvm import-system >/dev/null 2>&1 || smoke_fail "import-system failed on fresh container"
smoke_pass "import-system ingests container's system Perl"

list_out=$(pvm list 2>&1)
smoke_contains "$list_out" "Installed Perl versions" "pvm list has a header"
smoke_contains "$list_out" "5." "pvm list shows at least one 5.x version"

# versions is an alias for list per reference docs; both must work.
versions_out=$(pvm versions 2>&1)
smoke_contains "$versions_out" "5." "pvm versions (alias for list) shows the same"

# `available` hits the network by default; with PVM_SKIP_NETWORK_CALLS=1 the
# container forces it offline. The command should degrade gracefully rather
# than spin for 10s or crash.
avail_out=$(pvm available 2>&1)
avail_rc=$?
if [ "$avail_rc" -eq 0 ] || [ "$avail_rc" -eq 1 ]; then
    smoke_pass "pvm available returns promptly offline (exit=$avail_rc)"
else
    smoke_fail "pvm available exited with unexpected code $avail_rc: $avail_out"
fi

# --- completion scripts ----------------------------------------------------

for shell in bash zsh fish; do
    comp_out=$(pvm completion "$shell" 2>&1)
    # Shell-appropriate marker — bash/fish use `complete`, zsh uses #compdef.
    case "$shell" in
        bash) marker="_pvm_completion" ;;
        zsh)  marker="#compdef" ;;
        fish) marker="complete -c pvm" ;;
    esac
    if [ -z "$comp_out" ]; then
        smoke_fail "pvm completion $shell emitted empty output"
    fi
    case "$comp_out" in
        *"$marker"*) smoke_pass "pvm completion $shell emits completion rules" ;;
        *) smoke_fail "pvm completion $shell missing marker [$marker]" ;;
    esac
done

# --- config / env / remote: introspection-only commands --------------------

# `config` should list keys without requiring any prior setup. The subcommand
# may be called `list`, `show`, or `get` depending on version — we only assert
# it exits 0 and prints *something*.
cfg_out=$(pvm config 2>&1)
cfg_rc=$?
smoke_exit_eq "$cfg_rc" "0" "pvm config (no args) exits 0"
[ -n "$cfg_out" ] || smoke_fail "pvm config produced no output"
smoke_pass "pvm config produces output"

env_out=$(pvm env list 2>&1)
smoke_contains "$env_out" "environments" "pvm env list mentions environments"

remote_out=$(pvm remote list 2>&1)
# Fresh install has at least the built-in 'origin' remote.
smoke_contains "$remote_out" "origin" "pvm remote list shows built-in origin"

# --- directory-resolver sanity --------------------------------------------

# detect-version walks up from PWD looking for .perl-version. In an empty dir
# it should either print nothing or indicate no file found — exit 0 or 1 are
# both acceptable; a crash is not.
dv_out=$(pvm detect-version 2>&1)
dv_rc=$?
if [ "$dv_rc" -gt 1 ]; then
    smoke_fail "pvm detect-version exited $dv_rc on empty dir: $dv_out"
fi
smoke_pass "pvm detect-version doesn't crash (exit=$dv_rc)"

# --- diagnostic: self doctor -----------------------------------------------

# doctor must exit 0 on the fresh container (warnings are acceptable; errors
# would mean the install is broken before any user journey starts).
doctor_out=$(pvm self doctor 2>&1)
doctor_rc=$?
smoke_exit_eq "$doctor_rc" "0" "pvm self doctor exits 0 on fresh container"
smoke_contains "$doctor_out" "PVM Doctor" "pvm self doctor runs diagnostics"

# --- current-version resolution --------------------------------------------

# After import-system a version exists; `pvm current` must resolve it rather
# than erroring.
cur_out=$(pvm current 2>&1)
cur_rc=$?
smoke_exit_eq "$cur_rc" "0" "pvm current exits 0 post-import"
smoke_contains "$cur_out" "5." "pvm current names a 5.x version"

cur_bare=$(pvm current --bare 2>&1)
smoke_contains "$cur_bare" "5." "pvm current --bare outputs just a version"
# Bare output must be a single line.
line_count=$(printf '%s\n' "$cur_bare" | wc -l)
smoke_equals "$line_count" "1" "pvm current --bare outputs single line"

echo ""
echo "=== core commands smoke test: all assertions passed ==="
