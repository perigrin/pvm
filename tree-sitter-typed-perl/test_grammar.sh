#!/bin/bash
# ABOUTME: Helper script to quickly test grammar changes for tree-sitter-typed-perl
# ABOUTME: Provides fast feedback loop for grammar development

set -e

# Check if test input is provided
if [ $# -eq 0 ]; then
    echo "Usage: $0 <perl-code>"
    echo "Example: $0 'our \$Package::qualified;'"
    exit 1
fi

# Create temporary test file
TEMP_FILE=$(mktemp /tmp/test_grammar_XXXXXX.pl)
echo "$1" > "$TEMP_FILE"

echo "Testing: $1"
echo "===================="

# Parse and show the tree
echo "Parse tree:"
tree-sitter parse "$TEMP_FILE"

# Run corpus test if second argument is provided
if [ ! -z "$2" ]; then
    echo ""
    echo "Running corpus test: $2"
    tree-sitter test --include "$2"
fi

# Cleanup
rm -f "$TEMP_FILE"
