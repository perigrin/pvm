# PVM Ecosystem Tests

This directory contains tests for the PVM Ecosystem.

## Parser Tests

The parser implementation currently has several limitations:

1. It uses a simplified parsing approach instead of a full-featured parser like tree-sitter.
2. It can detect basic type annotations but may not correctly handle all complex scenarios.
3. The tests in `internal/parser/parser_test.go` demonstrate the expected behavior, even though some tests are failing with the current implementation.

In a production implementation, we would:

1. Use tree-sitter for robust parsing of Perl code with type annotations.
2. Implement a more comprehensive type checking system.
3. Add more thorough testing for various edge cases.

## Sample Perl Files

The `samples` directory contains sample Perl files with type annotations for testing:

- `sample1.pl`: A simple Perl script with various type annotations.
- `sample2.pl`: A Perl script with intentional type errors to test error detection.

You can run these samples with the PSC command:

```bash
./psc check test/samples/sample1.pl --verbose
./psc check test/samples/sample2.pl --verbose
```

## Test Coverage

Current test coverage includes:

1. Basic type expression parsing
2. Simple type annotation detection
3. Command integration with PSC

Future test improvements would include:

1. More comprehensive type checking
2. Better error reporting and recovery
3. Testing with complex real-world Perl code
4. Testing edge cases and error conditions
5. Integration testing with the PVX component
