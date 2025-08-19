---
category: error-cases
subcategory: malformed-types
tags: [error-recovery, syntax-errors, type-parsing]
---

# Invalid Union Syntax - Error Recovery

<!-- should_error: false -->
<!-- recovered_behavior: Treats as untyped variable -->
<!-- suggestion: Use '|' for proper union types -->

```perl
my Int||Str $bad_union;
```

### Expected AST

#### Text Format

```
Successfully parsed with error recovery - malformed union handled gracefully.
Variable declared as untyped due to invalid type syntax.
```

#### JSON Format

```json
{
  "status": "success_with_recovery",
  "message": "Invalid union syntax recovered - treated as untyped variable",
  "suggestion": "Use single '|' for union types: Int|Str",
  "ast": "valid_ast_structure"
}
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
use v5.42.0;
my $bad_union;
```

### Typed Perl Output

```perl
use v5.42.0;
my $bad_union;
```

### Inferred Perl Output

```perl
use v5.42.0;
my $bad_union;
```

# Incomplete Type Assertion

<!-- should_error: true -->
<!-- expected_error: error[TSP004] -->
<!-- expected_suggestion: Add the target type after 'as' keyword -->

```perl
my $val = $input as ;
```

### Expected AST

#### Text Format

```
Parse error: SYS-007: error[TSP004]: Incomplete type assertion - missing target type
```

#### JSON Format

```json
{
  "error": "SYS-007: error[TSP004]: Incomplete type assertion - missing target type",
  "message": "Incomplete type assertion - target type required after 'as' keyword",
  "ast": null
}
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```

### Typed Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```

### Inferred Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```

# Invalid Intersection Syntax - Error Recovery

<!-- should_error: false -->
<!-- recovered_behavior: Treats as untyped variable -->
<!-- suggestion: Use '&' for proper intersection types -->

```perl
my Object&&Serializable $obj;
```

### Expected AST

#### Text Format

```
Parse error: SYS-007: error[TSP001]: parse error (1 ERROR nodes detected)
Invalid intersection syntax with double ampersand operator.
```

#### JSON Format

```json
{
  "error": "SYS-007: error[TSP001]: parse error (1 ERROR nodes detected)",
  "message": "Invalid intersection syntax - use single '&' for intersection types",
  "ast": null
}
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```

### Typed Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```

### Inferred Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```
