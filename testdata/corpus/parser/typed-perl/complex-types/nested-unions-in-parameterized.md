---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - union-types
    - parameterized-types
    - complex-combinations
---

# Nested Unions In Parameterized

Union types nested within parameterized types

```perl
my ArrayRef[Int|Str|Bool] @complex_array;
my HashRef[ArrayRef[Int]|HashRef[Str]] %nested_complex;
my Map[Str, Int|Undef] %optional_values;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @complex_array;
my %nested_complex;
my %optional_values;
```

## Typed Perl Output

```perl
my ArrayRef[Int|Str|Bool] @complex_array;
my HashRef[ArrayRef[Int]|HashRef[Str]] %nested_complex;
my Map[Str, Int|Undef] %optional_values;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
AST {
  Path: /tmp/nested-unions-in-parameterized.pl
  Source length: 138 characters
  Type Annotations:
    VarAnnotation: @complex_array :: ArrayRef[Int|Str|Bool] at 1:1
    VarAnnotation: %nested_complex :: HashRef[ArrayRef[Int]|HashRef[Str]] at 2:1
    VarAnnotation: %optional_values :: Map[Str, Int|Undef] at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      variable_declaration
        token
        type_expression
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                union_type
                  type_expression
                    union_type
                      type_expression
                        expression_stmt
                          literal
                      expression_stmt
                        literal
                      type_expression
                        expression_stmt
                          literal
                  expression_stmt
                    literal
                  type_expression
                    expression_stmt
                      literal
            expression_stmt
              literal
        array
          expression_stmt
            literal
          token
    token
    expression_statement
      variable_declaration
        token
        type_expression
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                union_type
                  type_expression
                    parameterized_type
                      expression_stmt
                        literal
                      expression_stmt
                        literal
                      type_parameter_list
                        type_expression
                          expression_stmt
                            literal
                      expression_stmt
                        literal
                  expression_stmt
                    literal
                  type_expression
                    parameterized_type
                      expression_stmt
                        literal
                      expression_stmt
                        literal
                      type_parameter_list
                        type_expression
                          expression_stmt
                            literal
                      expression_stmt
                        literal
            expression_stmt
              literal
        hash
          expression_stmt
            literal
          token
    token
    expression_statement
      variable_declaration
        token
        type_expression
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                expression_stmt
                  literal
              expression_stmt
                literal
              type_expression
                union_type
                  type_expression
                    expression_stmt
                      literal
                  expression_stmt
                    literal
                  type_expression
                    expression_stmt
                      literal
            expression_stmt
              literal
        hash
          expression_stmt
            literal
          token
    token
}
```

## JSON AST

```json
{
  "path": "/tmp/nested-unions-in-parameterized.pl",
  "root": {
    "type": "source_file",
    "start": { "Line": 1, "Column": 1, "Offset": 0 },
    "end": { "Line": 3, "Column": 41, "Offset": 138 },
    "children": [
      {
        "type": "expression_statement",
        "start": { "Line": 1, "Column": 1, "Offset": 0 },
        "end": { "Line": 1, "Column": 41, "Offset": 40 },
        "children": [
          {
            "type": "variable_declaration",
            "start": { "Line": 1, "Column": 1, "Offset": 0 },
            "end": { "Line": 1, "Column": 41, "Offset": 40 },
            "children": [
              { "type": "token", "start": { "Line": 1, "Column": 1, "Offset": 0 }, "end": { "Line": 1, "Column": 3, "Offset": 2 }, "text": "my" },
              {
                "type": "type_expression",
                "start": { "Line": 1, "Column": 4, "Offset": 3 },
                "end": { "Line": 1, "Column": 26, "Offset": 25 },
                "children": [
                  {
                    "type": "parameterized_type",
                    "start": { "Line": 1, "Column": 4, "Offset": 3 },
                    "end": { "Line": 1, "Column": 26, "Offset": 25 },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": { "Line": 1, "Column": 4, "Offset": 3 },
                        "end": { "Line": 1, "Column": 12, "Offset": 11 },
                        "children": [{ "type": "literal", "start": { "Line": 1, "Column": 4, "Offset": 3 }, "end": { "Line": 1, "Column": 12, "Offset": 11 }, "value": "ArrayRef", "kind": "string" }]
                      },
                      {
                        "type": "type_parameter_list",
                        "start": { "Line": 1, "Column": 13, "Offset": 12 },
                        "end": { "Line": 1, "Column": 25, "Offset": 24 },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": { "Line": 1, "Column": 13, "Offset": 12 },
                            "end": { "Line": 1, "Column": 25, "Offset": 24 },
                            "children": [
                              {
                                "type": "union_type",
                                "start": { "Line": 1, "Column": 13, "Offset": 12 },
                                "end": { "Line": 1, "Column": 25, "Offset": 24 },
                                "children": [
                                  { "type": "type_expression", "children": [{ "type": "expression_stmt", "children": [{ "type": "literal", "value": "Int", "kind": "string" }] }] },
                                  { "type": "expression_stmt", "children": [{ "type": "literal", "value": "|", "kind": "string" }] },
                                  { "type": "type_expression", "children": [{ "type": "expression_stmt", "children": [{ "type": "literal", "value": "Str", "kind": "string" }] }] },
                                  { "type": "expression_stmt", "children": [{ "type": "literal", "value": "|", "kind": "string" }] },
                                  { "type": "type_expression", "children": [{ "type": "expression_stmt", "children": [{ "type": "literal", "value": "Bool", "kind": "string" }] }] }
                                ]
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  }
                ]
              },
              {
                "type": "array",
                "start": { "Line": 1, "Column": 27, "Offset": 26 },
                "end": { "Line": 1, "Column": 41, "Offset": 40 },
                "children": [
                  { "type": "expression_stmt", "children": [{ "type": "literal", "value": "@", "kind": "string" }] },
                  { "type": "token", "text": "complex_array" }
                ]
              }
            ]
          }
        ]
      }
    ]
  }
}
```

# Expected Type Errors

(none)
