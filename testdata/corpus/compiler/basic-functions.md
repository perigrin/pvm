---
category: compiler
subcategory: basic-functions
tags:
    - function-signatures
    - typed-parameters
    - clean-perl-output
type_check: false
---

# Simple Function Signature

Basic function with typed parameters

```perl
sub Int add (Int $a, Int $b) {
    return $a + $b;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.42.0;
sub add ($a, $b) {
    return $a + $b;
}
```

## Typed Perl Output

```perl
sub Int add (Int $a, Int $b) {
    return $a + $b;
}
```

## Text AST

```
(source
  (subroutine_declaration
    (return_type (simple_type name: (identifier) @type))
    name: (identifier) @function.name
    (signature
      (parameter
        (type_annotation (simple_type name: (identifier) @type))
        (scalar_variable (variable_name (identifier) @parameter)))
      (parameter
        (type_annotation (simple_type name: (identifier) @type))
        (scalar_variable (variable_name (identifier) @parameter))))
    (block
      (return_statement
        (binary_expression
          left: (scalar_variable (variable_name (identifier)))
          operator: "+"
          right: (scalar_variable (variable_name (identifier))))))))
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
      "name": "add",
      "signature": {
        "parameters": [
          {
            "type_annotation": {
              "type": "simple_type",
              "name": "Int"
            },
            "variable": {
              "type": "scalar_variable",
              "name": "a"
            }
          },
          {
            "type_annotation": {
              "type": "simple_type",
              "name": "Int"
            },
            "variable": {
              "type": "scalar_variable",
              "name": "b"
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
              "type": "binary_expression",
              "left": {"type": "scalar_variable", "name": "a"},
              "operator": "+",
              "right": {"type": "scalar_variable", "name": "b"}
            }
          }
        ]
      }
    }
  ]
}
```
