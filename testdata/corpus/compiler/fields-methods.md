---
category: compiler
subcategory: fields-methods
tags:
    - field-declarations
    - method-signatures
    - attributes
    - multiline-signatures
type_check: false
---

# Field Declarations

Basic field declarations with types

```perl
field Int $count;
field ArrayRef[Str] $items;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.42.0;
field $count;
field $items;
```

## Typed Perl Output

```perl
field Int $count;
field ArrayRef[Str] $items;
```

# Field with Initialization

Field declarations with initial values

```perl
field Int $counter = 0;
field HashRef[Str] $config = {};
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.42.0;
field $counter = 0;
field $config = {};
```

## Typed Perl Output

```perl
field Int $counter = 0;
field HashRef[Str] $config = {};
```

# Method Declaration

Method with typed parameters

```perl
method Bool add_user(UserId $id, HashRef[Str] $data) {
    return 1;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.42.0;
method add_user($id, $data) {
    return 1;
}
```

## Typed Perl Output

```perl
method Bool add_user(UserId $id, HashRef[Str] $data) {
    return 1;
}
```

# Multiline Signature

Function with multiline signature

```perl
sub Bool multiline (
    Int $first,
    Str $second,
    ArrayRef[Int] $third
) {
    return 1;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.42.0;
sub multiline (
    $first,
    $second,
    $third
) {
    return 1;
}
```

## Typed Perl Output

```perl
sub Bool multiline (
    Int $first,
    Str $second,
    ArrayRef[Int] $third
) {
    return 1;
}
```

# Attributes with Signature

Function with attributes and typed signature

```perl
sub Int tagged :lvalue :const (Int $value) {
    return $value;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.42.0;
sub tagged :lvalue :const ($value) {
    return $value;
}
```

## Typed Perl Output

```perl
sub Int tagged :lvalue :const (Int $value) {
    return $value;
}
```

## Text AST

```
(source
  (subroutine_declaration
    (return_type (simple_type name: (identifier)))
    name: (identifier)
    (attributes
      (attribute name: "lvalue")
      (attribute name: "const"))
    (signature
      (parameter
        (type_annotation (simple_type name: (identifier)))
        (scalar_variable (variable_name (identifier)))))
    (block
      (return_statement
        (scalar_variable (variable_name (identifier)))))))
```

## JSON AST

```json
{
  "type": "source",
  "children": [
    {
      "type": "subroutine_declaration",
      "return_type": {
        "type": "simple_type",
        "name": "Int"
      },
      "name": "tagged",
      "attributes": [
        {"name": "lvalue"},
        {"name": "const"}
      ],
      "signature": {
        "parameters": [
          {
            "type_annotation": {
              "type": "simple_type",
              "name": "Int"
            },
            "variable": {
              "type": "scalar_variable",
              "name": "value"
            }
          }
        ]
      },
      "body": {
        "type": "block",
        "statements": [
          {
            "type": "return_statement",
            "expression": {
              "type": "scalar_variable",
              "name": "value"
            }
          }
        ]
      }
    }
  ]
}
```
