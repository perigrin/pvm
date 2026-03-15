#!/bin/bash
# ABOUTME: Pre-commit hook script for running gosec security scanner
# ABOUTME: Installs gosec if needed and runs security analysis on Go code

set -euo pipefail

# Install gosec if not available
if ! command -v gosec &> /dev/null; then
    echo "Installing gosec..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi

# Run gosec with reasonable settings for pre-commit
echo "Running gosec security scanner..."

# Run gosec but don't fail on findings - just report them
if ! gosec -severity high -confidence high ./...; then
    echo ""
    echo "⚠️  gosec found security issues (shown above)"
    echo "💡 Review the findings and fix critical issues before committing"
    echo "📋 For full scan: make security"
    echo ""
    # Exit with 0 to make it informational rather than blocking
    exit 0
fi

echo "✅ No high-confidence security issues found"
