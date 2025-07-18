---
category: error-cases
subcategory: malformed-types
tags: [error-recovery, syntax-errors, type-parsing]
---

# Invalid Union Syntax

<!-- should_error: true -->
<!-- expected_error: error[TSP003] -->
<!-- expected_suggestion: Change '||' to '|' for union types -->

```perl
my Int||Str $bad_union;
```

### Expected AST

#### Text Format

```
Parse error: SYS-007: error[TSP001]: parse error (1 ERROR nodes detected)
Invalid union syntax with double pipe operator.
```

#### JSON Format

```json
{
  "error": "SYS-007: error[TSP001]: parse error (1 ERROR nodes detected)",
  "message": "Invalid union syntax - use single '|' for union types",
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
Parse error: SYS-007: error[TSP001]: parse error (1 ERROR nodes detected)
Incomplete type assertion - missing target type.
```

#### JSON Format

```json
{
  "error": "SYS-007: error[TSP001]: parse error (1 ERROR nodes detected)",
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

# Invalid Intersection Syntax

<!-- should_error: true -->
<!-- expected_error: error[TSP008] -->
<!-- expected_suggestion: Change '&&' to '&' for intersection types -->

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
