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
