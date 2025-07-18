---
category: error-cases
subcategory: malformed_types.json
tags: [error-recovery, malformed-types]
should_error: true
---

# Malformed Type Expressions

Test cases for type parsing error recovery

```perl
my Malformed[ $bad_type;
```

## Expected AST

### Text Format

```
Parse error: SYS-007: error[TSP001]: parse error (1 ERROR nodes detected)
Malformed parameterized type with missing closing bracket.
```

### JSON Format

```json
{
  "error": "SYS-007: error[TSP001]: parse error (1 ERROR nodes detected)",
  "message": "Malformed parameterized type - missing closing bracket or parameters",
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
