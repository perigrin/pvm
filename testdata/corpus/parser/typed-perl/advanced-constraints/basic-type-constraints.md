---
category: typed-perl
subcategory: advanced-constraints
type_check: true
tags:
    - type-constraints
    - where-clauses
---

# Basic Type Constraints

Test basic type constraint parsing with where clauses

```perl
class Container<T> where T: Serializable { }
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
class Container<T> where T: Serializable { }
```

## Typed Perl Output

```perl
class Container<T> where T: Serializable { }
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

(none)
