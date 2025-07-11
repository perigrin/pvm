---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - parameterized-types
    - union-types
    - parenthesized-unions
    - complex-combinations
---

# Parameterized Unions

Parameterized types within union expressions

```perl
my (ArrayRef[Int]|HashRef[Str]) $param_union;
my (Container[MyType]|Wrapper[OtherType]) $flexible;
my (Result[Data, Error]|Maybe[Value]) $outcome;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $param_union;
my $flexible;
my $outcome;
```

## Typed Perl Output

```perl
my (ArrayRef[Int]|HashRef[Str]) $param_union;
my (Container[MyType]|Wrapper[OtherType]) $flexible;
my (Result[Data, Error]|Maybe[Value]) $outcome;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

(none)
