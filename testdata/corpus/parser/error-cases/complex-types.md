---
category: error-cases
subcategory: complex-types
tags:
    - complex-expressions
    - error-recovery
    - malformed-types
---

# Complex Malformed Types

Malformed complex type expressions that should produce helpful error messages

<!-- should_error: true -->
<!-- expected_error: error[TSP002] -->

```perl
my ArrayRef[Int| $incomplete_union;
my HashRef[Str,, Int] $double_comma;
my Container[Wrapper[Missing $unclosed_nested;
my (Int|Str& $mixed_operators_without_close;
my Map[Str ArrayRef[Int]] $missing_comma_in_params;
```

# Expected AST

## Text Format

```
Parse error: SYS-007: error[TSP001]: parse error (5 ERROR nodes detected)
Multiple malformed type expressions detected with specific syntax errors.
```

## JSON Format

```json
{
  "error": "SYS-007: error[TSP001]: parse error (5 ERROR nodes detected)",
  "message": "Multiple malformed type expressions with incomplete union types, double commas, unclosed nested parameters, mixed operators, and missing commas",
  "ast": null
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```

## Typed Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```

## Inferred Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```
