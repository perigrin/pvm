---
category: untyped-perl
subcategory: packages
tags:
    - inheritance
    - parent
    - base
    - isa
    - multiple
---

# Inheritance Patterns

Test inheritance and parent module patterns

```perl
use parent 'BaseClass';
use base qw(Base1 Base2);
use parent qw(Parent::Class Another::Parent);
our @ISA = qw(BaseClass);
push @ISA, 'Mixin::Class';
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
use parent 'BaseClass';
use base qw(Base1 Base2);
use parent qw(Parent::Class Another::Parent);
our @ISA = qw(BaseClass);
push @ISA, 'Mixin::Class';
```

## Typed Perl Output

```perl
use parent 'BaseClass';
use base qw(Base1 Base2);
use parent qw(Parent::Class Another::Parent);
our @ISA = qw(BaseClass);
push @ISA, 'Mixin::Class';
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
