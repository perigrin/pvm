#!/bin/bash

# ABOUTME: Demonstration script showing how the repository consistency tests catch regressions
# ABOUTME: This script temporarily modifies config to show test failures, then restores it

set -e

echo "=== Repository Regression Protection Demo ==="
echo

# Save the original file
cp internal/config/types.go internal/config/types.go.backup

echo "1. Running tests with correct configuration..."
go test ./internal/config -run "TestRepositoryConsistency_NoDevReferences" -v
echo "✅ Tests pass with correct configuration"
echo

echo "2. Simulating accidental revert to pvm-dev..."
# Temporarily change repository to pvm-dev to demonstrate test failure
sed -i 's/perigrin\/pvm"/perigrin\/pvm-dev"/g' internal/config/types.go

echo "3. Running tests with reverted (incorrect) configuration..."
set +e  # Allow command to fail
go test ./internal/config -run "TestRepositoryConsistency_NoDevReferences" -v
test_result=$?
set -e

if [ $test_result -eq 0 ]; then
    echo "❌ ERROR: Tests should have failed but didn't!"
else
    echo "✅ Tests correctly detected the regression!"
fi
echo

echo "4. Restoring correct configuration..."
mv internal/config/types.go.backup internal/config/types.go

echo "5. Verifying tests pass again..."
go test ./internal/config -run "TestRepositoryConsistency_NoDevReferences" -v
echo "✅ Tests pass again after restoration"
echo

echo "=== Demo Complete ==="
echo "The repository consistency tests successfully prevent accidental reverts to pvm-dev!"