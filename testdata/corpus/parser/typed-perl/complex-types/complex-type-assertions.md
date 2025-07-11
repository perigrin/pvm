---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - type-assertions
    - method-calls
    - complex-types
    - parameterized-types
---

# Complex Type Assertions

Type assertions with complex type expressions

```perl
my $result = $data as ArrayRef[HashRef[Int|Bool]];
my $complex = ($input->process()) as Map[Str, ArrayRef[MyType]];
my $transformed = $obj->convert() as Result[Data[User], Error[String]];
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $result = $data;
my $complex = ($input->process());
my $transformed = $obj->convert();
```

## Typed Perl Output

```perl
my $result = $data as ArrayRef[HashRef[Int|Bool]];
my $complex = ($input->process()) as Map[Str, ArrayRef[MyType]];
my $transformed = $obj->convert() as Result[Data[User], Error[String]];
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

(none)
