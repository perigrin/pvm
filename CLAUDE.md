# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build/Test Commands

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

### Manual Build Commands (if needed)
- Build: `go build -mod=mod -o build/pvm ./cmd/pvm`
- Test all: `go test -mod=mod ./...`
- Test single package: `go test -mod=mod ./path/to/package`
- Test with coverage: `go test -mod=mod -cover ./...`

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

This process:
- Clones the official tree-sitter-perl grammar from GitHub
- Integrates our custom type annotation grammar extensions
- Generates the extended parser using tree-sitter-cli
- Builds the shared library for go-tree-sitter integration

### PSC-Specific Build Issues
PSC requires tree-sitter integration which has additional dependencies:
- Tree-sitter C library headers
- Extended perl grammar with type annotations
- CGO build flags for header includes

**Current Status**: ✅ Complete! Tree-sitter-typed-perl integration is working with full type annotation support.

## Recurring Build Memories

### CGO Dependencies Management
- We have gone down this build problem path before
- We have cleaned everything up to use the makefile
- The CGO dependencies should all be in tree-sitter-typed-perl now
- Please do not undo that work

## Tree-sitter Integration Principle
- tree-sitter is integral to the system, we MUST NOT work around it

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

[... rest of the file remains the same ...]
