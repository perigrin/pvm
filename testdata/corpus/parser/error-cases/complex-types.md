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

<!-- should_error: false -->
<!-- recovered_behavior: Error recovery handles malformed types -->

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
Successfully parsed with error recovery - malformed type expressions handled gracefully.
Variables with invalid type syntax treated as untyped.
```

## JSON Format

```json
{
  "status": "success_with_recovery",
  "message": "Complex malformed types recovered - incomplete unions, double commas, unclosed parameters handled",
  "ast": "valid_ast_structure"
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.42.0;
my [Int| $incomplete_union;
my $double_comma;
my $missing_comma_in_params;
```

## Typed Perl Output

```perl
my ArrayRef[Int| $incomplete_union;
my HashRef[Str,, Int] $double_comma;
my Container[Wrapper[Missing $unclosed_nested;
my (Int|Str& $mixed_operators_without_close;
my Map[Str ArrayRef[Int]] $missing_comma_in_params;
```

## Inferred Perl Output

```perl
use v5.42.0;
my [Int| $incomplete_union;
my $double_comma;
my $missing_comma_in_params;
```
