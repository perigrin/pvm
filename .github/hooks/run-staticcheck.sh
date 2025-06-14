#!/bin/bash
# ABOUTME: Pre-commit hook script for running staticcheck static analyzer
# ABOUTME: Installs staticcheck if needed and runs static analysis on Go code

set -euo pipefail

# Install staticcheck if not available
if ! command -v staticcheck &> /dev/null; then
    echo "Installing staticcheck..."
    go install honnef.co/go/tools/cmd/staticcheck@latest
fi

# Run staticcheck with reasonable settings for pre-commit
echo "Running staticcheck static analyzer..."

# Run staticcheck but don't fail on findings - just report them
if ! staticcheck ./...; then
    echo ""
    echo "⚠️  staticcheck found issues (shown above)"
    echo "💡 Review the findings and fix issues before committing"
    echo "📋 For full analysis: make lint"
    echo ""
    # Exit with 0 to make it informational rather than blocking
    exit 0
fi

echo "✅ No staticcheck issues found"
