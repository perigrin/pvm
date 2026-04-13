#!/usr/bin/env bash
# ABOUTME: Shell-agnostic advanced-workflow smoke suite. Covers the Jobs To
# ABOUTME: Be Done that working Perl devs hit regularly but that core-commands
#          doesn't: resolution-source attribution, JSON stability, workspace
#          scaffolding, config round-trip, remote write-path, symlink verify.

set -u
source "$(dirname "$0")/common.sh"

echo "=== advanced commands smoke test (shell-agnostic) ==="

# Ensure we have a version to resolve against. This is redundant with
# core-commands.sh's import-system but the journey should be standalone-
# runnable; both are idempotent.
pvm import-system >/dev/null 2>&1 || smoke_fail "import-system failed"
VERSION=$(pvm list 2>/dev/null | grep -oE '5\.[0-9]+\.[0-9]+' | head -1)
[ -n "$VERSION" ] || smoke_fail "no Perl version found after import-system"

############################################################################
# Section 1 — G1: `pvm current` resolution attribution
############################################################################

# Clean slate: no .perl-version, no PVM_PERL_VERSION.
cd "$HOME"
rm -f .perl-version
unset PVM_PERL_VERSION

# Pin via .perl-version and assert --detailed names the file as source.
echo "$VERSION" > "$HOME/.perl-version"
det=$(pvm current --detailed 2>&1)
smoke_contains "$det" ".perl-version" "current --detailed names the resolution source"
smoke_contains "$det" "project_file" "current --detailed reports project_file resolution_method"

# --json must be parseable and carry stable keys.
json=$(pvm current --json 2>&1)
for key in version source path available; do
    case "$json" in
        *"\"$key\""*) smoke_pass "current --json has key [$key]" ;;
        *) smoke_fail "current --json missing key [$key]: $json" ;;
    esac
done
# Try parsing JSON if python3 is available in the container.
if command -v python3 >/dev/null 2>&1; then
    if python3 -c "import json,sys; json.loads(sys.stdin.read())" <<< "$json" >/dev/null 2>&1; then
        smoke_pass "current --json is well-formed JSON"
    else
        smoke_fail "current --json is not valid JSON: $json"
    fi
fi

# G8: PVM_PERL_VERSION overrides .perl-version.
export PVM_PERL_VERSION="$VERSION"
env_src=$(pvm current --detailed 2>&1)
# Source should no longer be project_file; should name env var or environment.
case "$env_src" in
    *environment*|*env_var*|*PVM_PERL_VERSION*)
        smoke_pass "PVM_PERL_VERSION takes precedence over .perl-version" ;;
    *)
        smoke_fail "PVM_PERL_VERSION did not override .perl-version: $env_src" ;;
esac
unset PVM_PERL_VERSION
rm -f "$HOME/.perl-version"

############################################################################
# Section 2 — G10: broken .perl-version degrades gracefully (no crash)
############################################################################

echo "9.9.9-does-not-exist" > "$HOME/.perl-version"
broken_out=$(pvm current 2>&1)
broken_rc=$?
# pvm's current behavior: fall back to system Perl silently rather than
# erroring. Either response is acceptable; a crash or unlimited hang is
# not. Assert exit code is 0 or 1 (not 130, 137, etc.) and output is
# non-empty.
if [ "$broken_rc" -gt 1 ]; then
    smoke_fail "pvm current crashed (exit=$broken_rc) on broken pin: $broken_out"
fi
[ -n "$broken_out" ] || smoke_fail "pvm current produced no output on broken pin"
smoke_pass "pvm current handles broken .perl-version without crashing (exit=$broken_rc)"
rm -f "$HOME/.perl-version"

############################################################################
# Section 3 — G3: workspace init scaffold
############################################################################

ws_root=$(mktemp -d)
cd "$ws_root"
init_out=$(pvm workspace init my-app 2>&1)
smoke_contains "$init_out" "initialized successfully" "workspace init reports success"

for f in .perl-version cpanfile pvm.toml .gitignore; do
    [ -f "my-app/$f" ] || smoke_fail "workspace init did not create $f"
done
smoke_pass "workspace init creates .perl-version + cpanfile + pvm.toml + .gitignore"

for d in lib t; do
    [ -d "my-app/$d" ] || smoke_fail "workspace init did not create $d/"
done
smoke_pass "workspace init creates lib/ and t/"

# .perl-version should carry a 5.x version string.
content=$(cat my-app/.perl-version)
case "$content" in
    5.*) smoke_pass "workspace init .perl-version pins a 5.x version" ;;
    *)   smoke_fail "workspace init .perl-version has unexpected content: $content" ;;
esac

# workspace status inside the project should NOT say "No workspace detected".
cd my-app
stat_out=$(pvm workspace status 2>&1)
case "$stat_out" in
    *"No workspace detected"*)
        smoke_fail "workspace status inside project said 'No workspace detected'" ;;
    *)
        smoke_pass "workspace status recognizes the initialized project" ;;
esac

# workspace templates lists at least the built-in 'minimal' template.
tmpl_out=$(pvm workspace templates 2>&1)
smoke_contains "$tmpl_out" "minimal" "workspace templates lists 'minimal'"

cd "$HOME"
rm -rf "$ws_root"

############################################################################
# Section 4 — G5: config round-trip
############################################################################

# Use a key that's genuinely settable. pvm.build_jobs is a documented
# user-facing knob; the value is an integer.
pvm config set pvm.build_jobs 8 >/dev/null 2>&1
smoke_exit_eq "$?" "0" "config set exits 0"

got=$(pvm config get pvm.build_jobs 2>&1)
case "$got" in
    *8*) smoke_pass "config get round-trips pvm.build_jobs=8" ;;
    *)   smoke_fail "config get returned [$got], expected to contain 8" ;;
esac

smoke_contains "$(pvm config sources 2>&1)" "Configuration Sources" \
               "config sources reports source-list header"

pvm config validate >/dev/null 2>&1
smoke_exit_eq "$?" "0" "config validate exits 0"

pvm config unset pvm.build_jobs >/dev/null 2>&1
smoke_exit_eq "$?" "0" "config unset exits 0"

############################################################################
# Section 5 — G6: remote add + remove write-path
############################################################################

pvm remote add smoketest-remote https://example.com/perl.git >/dev/null 2>&1
smoke_exit_eq "$?" "0" "remote add exits 0 for valid args"

smoke_contains "$(pvm remote list 2>&1)" "smoketest-remote" \
               "remote list shows the newly-added remote"

pvm remote remove smoketest-remote >/dev/null 2>&1
smoke_exit_eq "$?" "0" "remote remove exits 0"

# After removal it should no longer appear.
case "$(pvm remote list 2>&1)" in
    *smoketest-remote*)
        smoke_fail "remote remove did not delete smoketest-remote" ;;
    *)
        smoke_pass "remote remove actually deletes the entry" ;;
esac

# The built-in 'origin' remote must survive.
smoke_contains "$(pvm remote list 2>&1)" "origin" \
               "remote list still shows built-in origin after round-trip"

############################################################################
# Section 6 — G7: self symlinks verify
############################################################################

# verify exits 0 on a fresh container (the symlinks may or may not exist
# depending on how pvm was installed — we're mainly asserting the command
# doesn't crash and produces a parseable status report).
sym_out=$(pvm self symlinks verify 2>&1)
sym_rc=$?
if [ "$sym_rc" -gt 1 ]; then
    smoke_fail "pvm self symlinks verify crashed (exit=$sym_rc): $sym_out"
fi
smoke_contains "$sym_out" "pvm" "self symlinks verify mentions the pvm entry point"
smoke_pass "self symlinks verify completes without crashing (exit=$sym_rc)"

############################################################################
# Section 7 — G9: rehash --dry-run does not mutate
############################################################################

# --dry-run should print what would happen, not actually rewrite PATH.
# Capture PATH before and after and assert they're identical.
path_before="$PATH"
dry_out=$(pvm rehash --dry-run 2>&1)
smoke_exit_eq "$?" "0" "rehash --dry-run exits 0"
smoke_equals "$PATH" "$path_before" "rehash --dry-run does not mutate PATH"

echo ""
echo "=== advanced commands smoke test: all assertions passed ==="
