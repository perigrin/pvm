# ABOUTME: Shared smoke-test helpers for the fish journey script.
# ABOUTME: Mirrors journeys/common.sh in intent; fish has no `set -u` equivalent.

function smoke_fail
    printf '\n[SMOKE FAIL] %s\n' $argv >&2
    exit 1
end

function smoke_pass
    printf '[SMOKE PASS] %s\n' $argv
end

# Assert a string contains a substring.
# Usage: smoke_contains ACTUAL EXPECTED LABEL
function smoke_contains
    set -l actual $argv[1]
    set -l expected $argv[2]
    set -l label $argv[3]
    if string match -q "*$expected*" -- $actual
        smoke_pass $label
    else
        smoke_fail "$label: expected to contain [$expected], got [$actual]"
    end
end

# Assert two strings are equal.
function smoke_equals
    set -l actual $argv[1]
    set -l expected $argv[2]
    set -l label $argv[3]
    if test "$actual" = "$expected"
        smoke_pass $label
    else
        smoke_fail "$label: expected [$expected], got [$actual]"
    end
end

# Assert an exit code.
function smoke_exit_eq
    set -l actual $argv[1]
    set -l expected $argv[2]
    set -l label $argv[3]
    if test "$actual" = "$expected"
        smoke_pass "$label (exit=$actual)"
    else
        smoke_fail "$label: expected exit $expected, got $actual"
    end
end

# Create a stub perl executable under $XDG_DATA_HOME/pvm/versions/$1/bin/perl
# whose `perl -v` output matches the canonical real-perl shape. Emit three
# integers for major/minor/patch so any future regex-based parser finds the
# right components. NB: fish has a read-only $version referring to its own
# version, so we use $requested_version for the argument.
function smoke_setup_stub_perl
    set -l requested_version $argv[1]
    set -l bin_dir "$XDG_DATA_HOME/pvm/versions/$requested_version/bin"
    set -l parts (string split . -- $requested_version)
    set -l major $parts[1]
    set -l minor $parts[2]
    set -l patch $parts[3]
    mkdir -p "$bin_dir"
    printf '#!/bin/sh\necho "This is perl %s, version %s, subversion %s (v%s) stub"\n' \
        $major $minor $patch $requested_version > "$bin_dir/perl"
    chmod +x "$bin_dir/perl"
end
