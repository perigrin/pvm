---
category: typechecker
subcategory: type-errors
tags:
    - type-mismatch
    - error-detection
    - assignment-validation
type_check: true
---

# Type Mismatch Errors

Test cases for detecting type assignment mismatches.

```perl
my Int $number = 42;
my Str $text = "hello";

# These should cause type errors
$number = "not a number";
$text = 123;
```

# Expected Symbol Table

The binder properly extracts type annotations and flags typed variables.
Assignment validation is not yet implemented in the type checker.

```
=== SYMBOLS ===
scalar number :: Int [lexical|typed] at 1:1
scalar text :: Str [lexical|typed] at 1:1
=== TYPE ERRORS ===
No type errors
```

# Expected Type Analysis

```
Variable $number: declared as Int, assigned Str (ERROR)
Variable $text: declared as Str, assigned Int (ERROR)
```

# Test Notes

This test verifies that the type checker properly detects assignment
mismatches between incompatible types and reports meaningful error messages.
