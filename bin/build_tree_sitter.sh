#!/bin/bash
# ABOUTME: Build script for tree-sitter-perl with type annotations extension
# ABOUTME: Combines base tree-sitter-perl grammar with our custom type annotation extensions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TREE_SITTER_DIR="$SCRIPT_DIR/vendor/tree-sitter-perl"
EXTENSION_FILE="$SCRIPT_DIR/internal/parser/treesitter/grammar_extension.js"
LIB_DIR="$SCRIPT_DIR/lib"
BUILD_DIR="$SCRIPT_DIR/build"

echo "Building tree-sitter-perl with type annotations..."

# Create build and lib directories
mkdir -p "$BUILD_DIR"
mkdir -p "$LIB_DIR"

# Clone tree-sitter-perl if not present (it's not a Go dependency)
if [ ! -d "$TREE_SITTER_DIR" ]; then
  echo "Cloning tree-sitter-perl repository..."
  git clone https://github.com/tree-sitter-perl/tree-sitter-perl "$TREE_SITTER_DIR"
fi

# Copy base tree-sitter-perl to build directory
echo "Copying base tree-sitter-perl grammar to build directory..."
cp -r "$TREE_SITTER_DIR" "$BUILD_DIR/tree-sitter-perl-extended"
cd "$BUILD_DIR/tree-sitter-perl-extended"

# Copy our grammar extension
echo "Integrating type annotation grammar extension..."
cp "$EXTENSION_FILE" .

# Integrate our grammar extensions
echo "Merging type annotation extensions into grammar..."
cp "$SCRIPT_DIR/bin/merge_grammar.js" .
node merge_grammar.js

# Ensure we have node/npm available
if ! command -v node &> /dev/null; then
  echo "Error: Node.js is required to build tree-sitter grammar"
  echo "Please install Node.js and npm"
  exit 1
fi

if ! command -v npm &> /dev/null; then
  echo "Error: npm is required to build tree-sitter grammar"
  echo "Please install Node.js and npm"
  exit 1
fi

# Check for tree-sitter-cli
echo "Checking for tree-sitter-cli..."
if ! command -v tree-sitter &> /dev/null; then
  echo "Installing tree-sitter-cli globally..."
  npm install -g tree-sitter-cli
else
  echo "tree-sitter-cli already available"
fi

# Generate parser from extended grammar
echo "Generating parser from extended grammar..."
if ! command -v tree-sitter &> /dev/null; then
  echo "Error: tree-sitter command not found"
  echo "Please install tree-sitter-cli: npm install -g tree-sitter-cli"
  exit 1
fi

tree-sitter generate

# Build shared library for go-tree-sitter
echo "Building shared library for go-tree-sitter..."
if [[ "$OSTYPE" == "darwin"* ]]; then
  # macOS
  echo "Building for macOS..."
  cc -dynamiclib -fPIC -g -O2 -I./src src/parser.c src/scanner.c -shared -o libtree-sitter-perl.dylib
  cp libtree-sitter-perl.dylib "$LIB_DIR/"
elif [[ "$OSTYPE" == "linux-gnu"* ]] || [[ "$OSTYPE" == "linux" ]]; then
  # Linux
  echo "Building for Linux..."
  cc -fPIC -g -O2 -I./src src/parser.c src/scanner.c -shared -o libtree-sitter-perl.so
  cp libtree-sitter-perl.so "$LIB_DIR/"
elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ "$OS" == "Windows_NT" ]]; then
  # Windows with MSYS2/MinGW
  echo "Building for Windows..."
  gcc -fPIC -g -O2 -I./src src/parser.c src/scanner.c -shared -o libtree-sitter-perl.dll
  cp libtree-sitter-perl.dll "$LIB_DIR/"
else
  # Fallback - try to detect based on environment or use GCC
  echo "Unknown platform, attempting generic build..."
  if command -v gcc &> /dev/null; then
    echo "Using GCC for build..."
    gcc -fPIC -g -O2 -I./src src/parser.c src/scanner.c -shared -o libtree-sitter-perl.so
    cp libtree-sitter-perl.so "$LIB_DIR/"
  else
    echo "Error: No suitable compiler found"
    echo "Please install GCC or appropriate C compiler"
    exit 1
  fi
fi

# Copy source files for go-tree-sitter integration
echo "Setting up sources for go-tree-sitter integration..."
cp src/parser.c "$SCRIPT_DIR/lib/"
cp src/scanner.c "$SCRIPT_DIR/lib/"
cp src/*.h "$SCRIPT_DIR/lib/" 2>/dev/null || true

# Create lib.c for go-tree-sitter with absolute paths
cat > "$SCRIPT_DIR/vendor/github.com/tree-sitter/go-tree-sitter/lib.c" << EOF
// Tree-sitter parser library for go-tree-sitter integration
#include "$SCRIPT_DIR/lib/parser.c"
#include "$SCRIPT_DIR/lib/scanner.c"
EOF

echo "Build complete!"
echo "Extended tree-sitter-perl library placed in $LIB_DIR"
echo "Parser sources copied for go-tree-sitter integration"
echo "Build artifacts in $BUILD_DIR/tree-sitter-perl-extended"
