# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build/Test Commands

All test files MUST pass 100% before committing, even tests that appear
unrelated to the current work.

### Basic Commands
- Build all: `make` (builds all components including tree-sitter-typed-perl)
- Build individual components:
  - PVM: `make pvm`
  - PVX: `make pvx`
  - PVI: `make pvi`
  - PSC: `make psc` (requires tree-sitter-typed-perl build first)
- Test all: `make test`
- Lint: `golangci-lint run`
- Clean: `make clean`

### Cross-Platform Build
- Cross-compile all platforms: `make cross-compile`
- Create release archives: `make release`

Supported platforms:
- Linux (AMD64, ARM64)
- macOS (AMD64, ARM64)
- Windows (AMD64)

### Tree-sitter Build (Required for PSC)
PSC uses tree-sitter-perl with custom type annotation extensions. The build process:

1. **Prerequisites**: Node.js and npm must be installed for tree-sitter-cli
2. **Build tree-sitter**: `make tree-sitter` or `./bin/build_tree_sitter.sh`

### PSC-Specific Build Issues
PSC requires tree-sitter integration which has additional dependencies:
- Tree-sitter C library headers
- Extended perl grammar with type annotations
- CGO build flags for header includes

## Recurring Build Memories

### CGO Dependencies Management
- We have gone down this build problem path before
- We have cleaned everything up to use the makefile
- The CGO dependencies should all be in tree-sitter-typed-perl now
- Please do not undo that work

## Tree-sitter Integration Principle
- tree-sitter is integral to the system, we MUST NOT work around it

## Problem-Solving Philosophy
- **NEVER create workarounds** - If you find yourself needing a workaround, STOP
- Think deeply about the root cause of the problem
- If you can't find a proper solution, ask perigrin for advice
- Workarounds create technical debt and mask the real issues that should be fixed

## Tree-sitter-typed-perl Integration

The project uses a custom `tree-sitter-typed-perl` grammar that extends the standard Perl grammar with type annotations:

### Type Annotation Support
- Typed variable declarations: `my Int $var = 42;`
- Typed field declarations: `field Str $name;`
- Type declarations: `type MyType = Int|Str;`
- Union types: `Int|Str`
- Intersection types: `Object&Serializable`
- Negation types: `!Undef`
- Parameterized types: `ArrayRef[Int]`
- Type assertions: `$value as Int`

### Build Process
1. **Grammar Generation**: `tree-sitter generate` creates parser.c and scanner.c
2. **Function Name Updates**: Scanner functions are renamed for typed_perl namespace
3. **Go Bindings**: Located in `tree-sitter-typed-perl/bindings/go/`
4. **Integration**: PSC component uses the bindings for static type checking

The build is completely self-contained with no external dependencies beyond Go and tree-sitter CLI.

## Compiler Architecture

PVM uses a modular compiler architecture in `internal/compiler` package:

### Compiler Package Structure
- **compiler.go**: Core interfaces and registry for different compilation targets
- **clean_perl.go**: Compiles AST to standard Perl without type annotations
- **typed_perl.go**: Compiles AST preserving all type annotations
- **parser_adapter.go**: Adapts parser.AST to compiler AST interface
- **types.go**: AST interface definitions for compiler independence
- **errors.go**: Structured error handling for compilation failures

### Compilation Targets
- `TargetCleanPerl`: Produces standard Perl code compatible with any interpreter
- `TargetTypedPerl`: Preserves type annotations for PSC consumption

### Usage Example
```go
// Parse file
parser, _ := parser.NewParser()
ast, _ := parser.ParseFile("script.pl")

// Compile to clean Perl
registry := compiler.NewCompilerRegistry()
adapter := compiler.NewParserASTAdapter(ast)
cleanCode, _ := registry.Compile(adapter, compiler.TargetCleanPerl)
```

### Integration with PSC
PSC commands (`psc strip`, `psc run`) use the compiler package for:
- Stripping type annotations for execution on standard Perl
- Converting typed Perl to untyped for CPAN distribution
- Future: Adding multiple compilation targets (JavaScript, etc.)

### Perl Version Output
- The compiler outputs `use v5.36;` pragma for compatibility with modern Perl features
- This enables signatures and other features needed by stripped code
- Using v5.36 for broader compatibility with existing Perl installations

## Code Style Guidelines

## Memory: AST Compilation
- you MUST NOT use regular expressions to compile the AST, if it is an ERROR node raise an error, otherwise compile from the AST. If you see regular expression we need to fix the code
