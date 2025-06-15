# Tree-sitter Typed Perl

A tree-sitter parser for Perl with comprehensive type annotation support.

## Overview

This parser extends the standard Perl grammar to handle type annotations, including:

- Typed variable declarations: `my Int $var = 42;`
- Typed field declarations: `field Str $name;`
- Type declarations: `type MyType = Int|Str;`
- Union types: `Int|Str`
- Intersection types: `Object&Serializable`
- Negation types: `!Undef`
- Parameterized types: `ArrayRef[Int]`
- Type assertions: `$value as Int`
- Method signatures: `method name(Str $param) -> Int`

## Getting Started

### Prerequisites

- Node.js v20 or above
- tree-sitter-cli

### Installation

```bash
# Install dependencies
npm install

# Install tree-sitter CLI locally
npm run dev-install
```

### Building

```bash
# Generate the parser
npm run generate

# Test the parser
npm run test
```

### Integration with Go

This parser includes Go bindings in `bindings/go/`. To use in a Go project:

```go
import tree_sitter_typed_perl "github.com/perigrin/pvm/tree-sitter-typed-perl/bindings/go"
import sitter "github.com/tree-sitter/go-tree-sitter"

func main() {
    language := sitter.NewLanguage(tree_sitter_typed_perl.Language())
    parser := sitter.NewParser()
    parser.SetLanguage(language)
    // ... use parser
}
```

## Development

### Grammar Structure

The grammar is defined in `grammar.js` and includes:

- Core Perl syntax (variables, functions, control structures)
- Type annotation extensions
- Method declarations with type signatures
- Type declarations and definitions

### Testing

Tests are located in `../test/corpus/tree-sitter/corpus/` and follow tree-sitter's test format:

```bash
# Run all tests  
npm test

# Test specific files
npx tree-sitter test ../test/corpus/tree-sitter/corpus/expressions
```

### Debugging Grammar Issues

Use the `debug_grammar.js` script for comprehensive grammar debugging:

```bash
# Debug specific code snippet
./debug_grammar.js "our \$Package::qualified;"

# Debug with verbose output
./debug_grammar.js "my Int \$var = 42;" --verbose

# Debug a file
./debug_grammar.js -f test_file.pl

# Interactive debugging mode
./debug_grammar.js -i

# Show token stream details
./debug_grammar.js "complex code" --tokens
```

The debug script provides:
- Parse tree visualization
- Token stream analysis
- Grammar rule coverage analysis
- Error node identification
- Type annotation feature detection
- Interactive debugging session

### Queries

Syntax highlighting and other editor features are defined in `queries/`:

- `highlights.scm` - syntax highlighting
- `folds.scm` - code folding
- `injections.scm` - language injections

## Build Process

1. **Grammar Generation**: `tree-sitter generate` creates parser.c and scanner.c
2. **Function Renaming**: Scanner functions are updated for typed_perl namespace
3. **Go Bindings**: Located in `bindings/go/` with proper module structure
4. **Testing**: Comprehensive test suite validates all type annotation features

## Integration with PVM

This parser is designed to integrate seamlessly with the PVM (Perl Version Manager) project:

- PSC (Perl Static Checker) uses this parser for type checking
- Build process is automated through the main project's Makefile
- Generated binaries are self-contained with no external dependencies

## Files

- `grammar.js` - Tree-sitter grammar definition
- `src/` - Generated parser source (C code)
- `bindings/` - Language bindings (Go, Node.js, Python, etc.)
- `queries/` - Editor integration queries
- `../test/corpus/tree-sitter/` - Test corpus and highlighting tests

## License

ISC License - see LICENSE file for details.
