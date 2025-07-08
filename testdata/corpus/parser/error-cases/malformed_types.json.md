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
