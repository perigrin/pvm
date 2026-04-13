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
        set -l truncated $actual
        if test (string length -- "$actual") -gt 200
            set truncated (string sub -l 200 -- "$actual")"… (truncated)"
        end
        smoke_fail "$label: expected to contain [$expected], got [$truncated]"
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

# Second Perl version used by Section 3b (G2) for the two-version cd
# auto-switch. Hoisted here so bumping the pinned binary-fetch target is
# a one-line change. Accepts override from the environment.
if not set -q SMOKE_SECOND_VERSION
    set -gx SMOKE_SECOND_VERSION "5.40.2"
end

# Extract the first 5.x.y version pvm reports as installed.
function smoke_first_installed_version
    pvm list 2>/dev/null | string match -rg '(5\.[0-9]+\.[0-9]+)' | head -1
end

# Run `pvm use ...` and capture output to a tempfile so env exports in the
# parent shell survive. Callers read $pvm_use_log afterwards and rm -f it.
# See common.sh's pvm_use_run for the rationale.
set -g pvm_use_log ""
function pvm_use_run
    set -g pvm_use_log (mktemp)
    pvm $argv > $pvm_use_log 2>&1
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
