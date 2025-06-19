#!/usr/bin/env bash
set -e

# Improved go vet hook that works with Go modules
# This runs go vet on the entire module rather than on individual directories

# Run go vet for the entire module (excluding tree-sitter packages)
echo "Running go vet for the entire module (excluding tree-sitter)..."

# Enable Go build cache for faster subsequent runs
export GOCACHE="$(go env GOCACHE)"

# Run go vet with module cache enabled
go vet -mod=mod $(go list -mod=mod ./... | grep -v treesitter)

# If we get this far, all checks passed
exit 0
