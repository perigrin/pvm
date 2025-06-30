---
category: compiler
subcategory: complex-types
tags:
    - parameterized-types
    - union-types
    - intersection-types
    - complex-signatures
type_check: false
---

# Complex Parameterized Types

Function with complex parameterized types

```perl
sub process (ArrayRef[HashRef[Str]] $data) -> Result[Bool] {
    return 1;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
sub process  {
    return 1;
}
```

# Nested Union Types

Function with nested union types

```perl
sub complex (Union[Str, Union[Int, Bool]] $param) -> Str {
    return "result";
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
sub complex ($param)  {
    return "result";
}
```

# Intersection and Negation Types

Function with intersection and negation types

```perl
sub validate (Object&Serializable $obj, !Undef $config) -> Bool {
    return 1;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
sub validate ($obj, $config)  {
    return 1;
}
```

# Complex Typed Variables

Complex type declarations for variables

```perl
my ComplexTypes $self = bless {}, $class;
my ArrayRef[HashRef[Str|Int]] $users = [];
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $self = bless {}, $class;
my $users = [];
```

# For Loop with Complex Types

Complex types in for loop context

```perl
for my UserId $id (keys %$config) {
    my HashRef[Str|Int] $user_info = {};
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
for my  $id (keys %$config) {
    my $user_info = {};
}
```
