#!/bin/bash
# Build script for tree-sitter-perl with type annotations extension
#
# This script builds the tree-sitter-perl library with our custom grammar
# extensions for type annotations. PSC exclusively uses tree-sitter for parsing,
# so this build step is required before using PSC.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TREE_SITTER_DIR="$SCRIPT_DIR/vendor/tree-sitter-perl"
EXTENSION_FILE="$SCRIPT_DIR/internal/parser/treesitter/grammar_extension.js"
LIB_DIR="$SCRIPT_DIR/lib"

echo "Building tree-sitter-perl with type annotations..."

# Create directories
mkdir -p "$TREE_SITTER_DIR"
mkdir -p "$LIB_DIR"

# Clone tree-sitter-perl if needed
if [ ! -d "$TREE_SITTER_DIR/.git" ]; then
  echo "Cloning tree-sitter-perl repository..."
  git clone https://github.com/tree-sitter-perl/tree-sitter-perl "$TREE_SITTER_DIR"
fi

cd "$TREE_SITTER_DIR"

# Make sure we have the latest version
echo "Updating tree-sitter-perl repository..."
git pull

# Copy our grammar extension
echo "Copying grammar extension..."
cp "$EXTENSION_FILE" .

# Check if the extension is already included in grammar.js
if ! grep -q "grammar_extension" grammar.js; then
  echo "Adding extension to grammar.js..."
  echo "// Include type annotation extensions" >> grammar.js
  echo "module.exports.rules = Object.assign(module.exports.rules, require('./grammar_extension').rules);" >> grammar.js
fi

# Install dependencies
echo "Installing dependencies..."
npm install

# Generate parser
echo "Generating parser..."
npx tree-sitter generate

# Build shared library
echo "Building shared library..."
if [[ "$OSTYPE" == "darwin"* ]]; then
  # macOS
  cc -dynamiclib -fPIC -g -O2 -I./src src/*.c -shared -o libtree-sitter-perl.dylib
  cp libtree-sitter-perl.dylib "$LIB_DIR/"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
  # Linux
  cc -fPIC -g -O2 -I./src src/*.c -shared -o libtree-sitter-perl.so
  cp libtree-sitter-perl.so "$LIB_DIR/"
else
  # Windows or other
  echo "Unsupported platform, please build manually"
  exit 1
fi

echo "Build complete!"
echo "Library placed in $LIB_DIR"
