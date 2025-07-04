# Tree-sitter Test Corpus

This directory contains test cases for the tree-sitter-typed-perl grammar.

## Test File Format

Each test file contains multiple test cases in the following format:

```
================
Test name
================
Perl code to parse
---
(expected_parse_tree)
```

## Running Tests

To run all tests:
```bash
tree-sitter test
```

To run tests from a specific file:
```bash
tree-sitter test --file-name <filename>
```

To run tests matching a pattern:
```bash
tree-sitter test --include "pattern"
```

## Test Files

- `untyped_perl_fixes` - Tests for untyped Perl constructs that need grammar support
- Other files test various Perl language features

## Helper Script

Use `test_grammar.sh` in the parent directory for quick grammar testing:
```bash
./test_grammar.sh 'perl code here'
```
