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
# is the expected substring, third is a human-readable label.
smoke_contains() {
    local actual="$1" expected="$2" label="$3"
    case "$actual" in
        *"$expected"*) smoke_pass "$label" ;;
        *) smoke_fail "$label: expected to contain [$expected], got [$actual]" ;;
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
