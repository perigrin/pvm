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
sub Result[Bool] process (ArrayRef[HashRef[Str]] $data) {
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
sub Str complex (Union[Str, Union[Int, Bool]] $param) {
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
sub Bool validate (Object&Serializable $obj, !Undef $config) {
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
for my $id (keys %$config) {
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

```
(source
  (foreach_statement
    (scalar_variable (variable_name (identifier)))
    (type_annotation (simple_type name: (identifier)))
    (expression (hash_dereference (hash_variable)))
    (block
      (variable_declaration
        (scalar_variable (variable_name (identifier)))
        (type_annotation (parameterized_type
          base: (identifier)
          args: (union_type (simple_type) (simple_type))))
        (expression (hash_reference))))))
```

## JSON AST

```json
{
  "type": "source",
  "children": [
    {
      "type": "foreach_statement",
      "variable": {
        "type": "scalar_variable",
        "type_annotation": {
          "type": "simple_type",
          "name": "UserId"
        },
        "name": "id"
      },
      "iterable": {
        "type": "hash_dereference",
        "variable": "config"
      },
      "body": {
        "type": "block",
        "statements": [
          {
            "type": "variable_declaration",
            "variable": {
              "type": "scalar_variable",
              "name": "user_info"
            },
            "type_annotation": {
              "type": "parameterized_type",
              "base": "HashRef",
              "args": [
                {
                  "type": "union_type",
                  "members": ["Str", "Int"]
                }
              ]
            },
            "value": {
              "type": "hash_reference"
            }
          }
        ]
      }
    }
  ]
}
```
