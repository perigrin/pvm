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
