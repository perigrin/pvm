#!/bin/bash
# Test script for Issue #198: Perl version doesn't update

echo "=== Testing Issue #198: Perl version doesn't update ==="
echo

# Clean environment
unset PVM_PERL_VERSION
export PATH_ORIGINAL="$PATH"

echo "1. Initial state (before PVM setup):"
echo "   Current PATH contains shims: $(echo $PATH | grep -o '/home/perigrin/.local/share/pvm/shims' || echo 'NO')"
echo "   which perl: $(which perl)"
echo "   perl version: $(perl -v | head -1)"
echo

echo "2. PVM doctor (initial):"
pvm doctor
echo

echo "3. Adding shims to PATH manually:"
export PATH="/home/perigrin/.local/share/pvm/shims:$PATH"
echo "   PATH updated with shims directory"
echo "   which perl: $(which perl)"
echo

echo "4. Setting up shell integration:"
eval "$(pvm init)"
echo

echo "5. PVM status after setup:"
echo "   pvm current: $(pvm current)"
echo "   pvm list: $(pvm list)"
echo

echo "6. Testing version switching:"
# Set to installed version
echo "   Switching to 5.38.0..."
pvm use 5.38.0
echo "   pvm current after switch: $(pvm current)"
echo "   which perl: $(which perl)"
echo "   perl version: $(perl -v | head -1)"
echo

echo "7. Testing .perl-version file behavior:"
# Create test directory with .perl-version
mkdir -p /tmp/test_pvm_dir
echo "5.38.0" > /tmp/test_pvm_dir/.perl-version
cd /tmp/test_pvm_dir
echo "   Changed to directory with .perl-version (5.38.0)"
echo "   pvm current: $(pvm current)"
echo "   which perl: $(which perl)"
echo "   perl version: $(perl -v | head -1)"
echo

echo "8. Final PVM doctor check:"
pvm doctor
echo

echo "=== Test Summary ==="
echo "Issue #198 Status: $([ "$(perl -v | grep -o 'v5\.[0-9]\+\.[0-9]\+')" != "$(pvm current)" ] && echo 'STILL EXISTS' || echo 'RESOLVED')"
