#!/bin/bash
# ABOUTME: Quick pre-commit check for basic Go issues
# ABOUTME: Fast alternative to full static analysis for regular commits

set -euo pipefail

echo "Running quick Go checks..."

# Enable Go build cache
export GOCACHE="$(go env GOCACHE)"

# Quick syntax check - just try to build without linking
echo "  → Syntax check..."
if ! go build -o /dev/null ./...; then
    echo "❌ Syntax errors found"
    exit 1
fi

# Quick format check without fixing
echo "  → Format check..."
if [ -n "$(gofmt -l .)" ]; then
    echo "❌ Code formatting issues found. Run 'go fmt ./...' to fix"
    gofmt -l .
    exit 1
fi

# Quick imports check
echo "  → Import check..."
if command -v goimports &> /dev/null; then
    if [ -n "$(goimports -l .)" ]; then
        echo "❌ Import issues found. Run 'goimports -w .' to fix"
        goimports -l .
        exit 1
    fi
fi

echo "✅ Quick checks passed"
