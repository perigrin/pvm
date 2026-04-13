#!/usr/bin/env bash
# ABOUTME: Shared smoke-test helpers loaded by per-shell journey scripts.
# ABOUTME: POSIX-ish so it works from bash / zsh wrapper scripts; fish loads
#          its own equivalent in journeys/common.fish.

set -u

# Exit with a visible marker any test harness can grep for.
smoke_fail() {
    printf '\n[SMOKE FAIL] %s\n' "$*" >&2
    exit 1
}

smoke_pass() {
    printf '[SMOKE PASS] %s\n' "$*"
}

# Assert a string contains a substring. First arg is the actual value, second
# is the expected substring, third is a human-readable label. Error messages
# truncate long `actual` values to keep failure output readable (completion
# scripts are hundreds of lines).
smoke_contains() {
    local actual="$1" expected="$2" label="$3"
    case "$actual" in
        *"$expected"*) smoke_pass "$label" ;;
        *)
            local truncated="$actual"
            if [ "${#truncated}" -gt 200 ]; then
                truncated="${truncated:0:200}… (truncated, ${#actual} chars total)"
            fi
            smoke_fail "$label: expected to contain [$expected], got [$truncated]"
            ;;
    esac
}

# Assert two strings are equal.
smoke_equals() {
    local actual="$1" expected="$2" label="$3"
    if [ "$actual" = "$expected" ]; then
        smoke_pass "$label"
    else
        smoke_fail "$label: expected [$expected], got [$actual]"
    fi
}

# Assert an exit code.
smoke_exit_eq() {
    local actual="$1" expected="$2" label="$3"
    if [ "$actual" = "$expected" ]; then
        smoke_pass "$label (exit=$actual)"
    else
        smoke_fail "$label: expected exit $expected, got $actual"
    fi
}

# Second Perl version used by Section 3b (G2) to exercise the two-version
# cd auto-switch. Hoisted here so bumping the pinned binary-fetch target
# is a one-line change instead of editing every journey file.
: "${SMOKE_SECOND_VERSION:=5.40.2}"
export SMOKE_SECOND_VERSION

# Extract the first 5.x.y version pvm reports as installed. Idempotent and
# side-effect-free; used by every per-shell journey and by advanced-commands.
smoke_first_installed_version() {
    pvm list 2>/dev/null | grep -oE '5\.[0-9]+\.[0-9]+' | head -1
}

# Run `pvm use ...` (or any pvm subcommand that modifies the current shell
# environment) and capture output to a tempfile so the parent shell keeps
# its env-var exports. Callers read the tempfile path from $pvm_use_log.
#
# Why this exists: `pvm use X | grep ...` or `$(pvm use X)` run pvm in a
# subshell, which means PVM_PERL_VERSION / PATH exports are discarded
# before the caller can observe them. The tempfile pattern keeps the
# export machinery in the parent shell.
#
# NB: pvm_use_log is intentionally a global — callers inspect it after
# the helper returns. rm -f it once done.
pvm_use_log=""
pvm_use_run() {
    pvm_use_log=$(mktemp)
    pvm "$@" > "$pvm_use_log" 2>&1
}

# Create a stub perl executable under $XDG_DATA_HOME/pvm/versions/$1/bin/perl
# whose `perl -v` output matches the canonical real-perl shape. Emit three
# integers for major/minor/patch so any future regex-based parser finds the
# right components (real perl prints "version N, subversion M", not
# "version N.M, subversion 0").
smoke_setup_stub_perl() {
    local version="$1"
    local bin_dir="$XDG_DATA_HOME/pvm/versions/$version/bin"
    local major minor patch
    IFS=. read -r major minor patch <<< "$version"
    mkdir -p "$bin_dir"
    cat > "$bin_dir/perl" <<EOF
#!/bin/sh
echo "This is perl $major, version $minor, subversion $patch (v$version) stub"
EOF
    chmod +x "$bin_dir/perl"
}
