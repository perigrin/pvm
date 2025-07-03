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
$number = "not a number";  # Str assigned to Int
$text = 123;               # Int assigned to Str
```

# Expected Symbol Table

```
=== SYMBOLS ===
scalar number [lexical] at 1:8
scalar text [lexical] at 2:8
=== TYPE ERRORS ===
Type mismatch at line 5: cannot assign Str to Int variable $number
Type mismatch at line 6: cannot assign Int to Str variable $text
```

# Expected Type Analysis

```
Variable $number: declared as Int, assigned Str (ERROR)
Variable $text: declared as Str, assigned Int (ERROR)
```

# Test Notes

This test verifies that the type checker properly detects assignment
mismatches between incompatible types and reports meaningful error messages.
