---
category: untyped-perl
subcategory: subroutines
tags:
    - forward
    - declarations
    - methods
---

# Forward Declarations

Test forward declarations without bodies

```perl
sub no_body;
sub another_forward;
method forward_method;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
sub no_body;
sub another_forward;
method forward_method;
```

## Typed Perl Output

```perl
sub no_body;
sub another_forward;
method forward_method;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
