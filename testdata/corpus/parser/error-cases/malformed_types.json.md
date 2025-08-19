---
category: error-cases
subcategory: malformed_types.json
tags: [error-recovery, malformed-types]
should_error: false
---

# Malformed Type Expressions

Test cases for type parsing error recovery

```perl
my Malformed[ $bad_type;
```

## Expected AST

### Text Format

```
Successfully parsed with error recovery - missing bracket handled gracefully.
Variable declared as untyped due to malformed type annotation.
```

### JSON Format

```json
{
  "status": "success_with_recovery",
  "message": "Malformed parameterized type recovered - treated as untyped variable",
  "ast": "valid_ast_structure"
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.42.0;
my [ $bad_type;
```

## Typed Perl Output

```perl
my Malformed[ $bad_type;
```

## Inferred Perl Output

```perl
use v5.42.0;
my [ $bad_type;
```
