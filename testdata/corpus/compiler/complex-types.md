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
sub Result[Bool] process (ArrayRef[HashRef[Str]] $data) {
    return 1;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
sub process ($data) {
    return 1;
}
```

## Typed Perl Output

```perl
sub Result[Bool] process (ArrayRef[HashRef[Str]]$data) {
    return 1;
}
```

# Nested Union Types

Function with nested union types

```perl
sub Str complex (Union[Str, Union[Int, Bool]] $param) {
    return "result";
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
sub complex ($param) {
    return "result";
}
```

## Typed Perl Output

```perl
sub Str complex (Union[Str, Union[Int, Bool]]$param) {
    return "result";
}
```

# Intersection and Negation Types

Function with intersection and negation types

```perl
sub Bool validate (Object&Serializable $obj, !Undef $config) {
    return 1;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
sub validate ($obj, $config) {
    return 1;
}
```

## Typed Perl Output

```perl
sub Bool validate (Object&Serializable$obj, !Undef$config) {
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

## Typed Perl Output

```perl
my ComplexTypes $self = bless {}, $class;
my ArrayRef[HashRef[Str|Int]] $users = [];
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

## Typed Perl Output

```perl
for my UserId $id (keys %$config) {
    my HashRef[Str|Int] $user_info = {};
}
```

## Text AST

**Note**: Complex parameterized types, union types, intersection types, and function signatures are not yet supported by the tree-sitter grammar. This syntax would currently produce parse errors and cannot generate a meaningful AST.

```
(parse error - complex type syntax not supported)
```

## JSON AST

```json
{
  "error": "Complex type syntax not yet supported by grammar",
  "note": "This syntax requires grammar extensions for: parameterized types (ArrayRef[Type]), union types (Type|Type), intersection types (Type&Type), negation types (!Type), and complex function signatures"
}
```
