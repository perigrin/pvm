# PVM Ecosystem Tests

This directory contains tests for the PVM Ecosystem.

## Test Structure

- `e2e/` - End-to-end integration tests (run with `make test`)
- `corpus/` - Test corpus and examples
  - `parser/` - Parser test data (markdown-based tests for typed Perl syntax)
  - `tree-sitter/` - Tree-sitter grammar test corpus
- `build_test.go` - Build system tests
- `helper.go` - Shared test utilities
- `*.typedef.json` - Type definition test files

## Running Tests

```bash
# Run all tests
make test

# Run specific test suites
go test ./test/e2e/          # End-to-end integration tests
go test ./internal/parser/   # Parser tests with testdata
go test ./internal/...       # All internal component tests
```

## Test Data

For typed Perl syntax examples and test cases, see:
- `corpus/parser/` - Markdown-based test files with correct syntax examples
- `corpus/tree-sitter/corpus/` - Tree-sitter grammar tests with comprehensive coverage

### Syntax Examples

The `corpus/parser/typed-perl/` directory contains the authoritative examples of correct typed Perl syntax:

```bash
# Browse syntax examples
ls test/corpus/parser/typed-perl/
# simple-annotations.md - Basic type annotations
# classes-roles.md - Object-oriented features  
# complex-types.md - Advanced type constructs
```
