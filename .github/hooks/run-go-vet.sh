#!/usr/bin/env bash
set -e

# Improved go vet hook that works with Go modules
# This runs go vet on the entire module rather than on individual directories

# Run go vet for the entire module
echo "Running go vet for the entire module..."
go vet ./...

# If we get this far, all checks passed
exit 0
