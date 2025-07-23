---
category: typed-perl
subcategory: advanced-constraints
type_check: true
tags:
    - constraint-inheritance
    - roles
    - inheritance
---

# Constraint Inheritance

Test constraint inheritance from roles and parent classes

```perl
class Base<T> where T: Defined { }
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
class Base<T> where T:  { }
```

## Typed Perl Output

```perl
class Base<T> where T: Defined { }
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

(none)
