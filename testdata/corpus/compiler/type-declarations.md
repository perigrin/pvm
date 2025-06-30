---
category: compiler
subcategory: type-declarations
tags:
    - type-definitions
    - custom-types
    - type-removal
type_check: false
---

# Type Declarations

Custom type definitions

```perl
type UserId = Int;
type UserData = HashRef[Str];
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
type UserId = Int;
type UserData = ;
```

# Expected Type Errors

```
(none)
```
