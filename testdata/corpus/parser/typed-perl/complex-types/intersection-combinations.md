---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - intersection-types
    - complex-combinations
    - parameterized-types
---

# Intersection Combinations

Intersection types combined with parameterized and union types

```perl
my ArrayRef[Object&Serializable] @serializable_list;
my HashRef[ArrayRef[Int|Str]&Defined] %defined_arrays;
my Container[Data&Validated&Cached] $safe_container;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @serializable_list;
my %defined_arrays;
my $safe_container;
```

## Typed Perl Output

```perl
my ArrayRef[Object&Serializable] @serializable_list;
my HashRef[ArrayRef[Int|Str]&Defined] %defined_arrays;
my Container[Data&Validated&Cached] $safe_container;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
AST {
  Path: /tmp/intersection-combinations.pl
  Source length: 160 characters
  Type Annotations:
    VarAnnotation: @serializable_list :: ArrayRef[Object&Serializable] at 1:1
    VarAnnotation: %defined_arrays :: HashRef[ArrayRef[Int|Str]&Defined] at 2:1
    VarAnnotation: $safe_container :: Container[Data&Validated&Cached] at 3:1
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
                intersection_type
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
                intersection_type
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
                              expression_stmt
                                literal
                            expression_stmt
                              literal
                            type_expression
                              expression_stmt
                                literal
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
                intersection_type
                  type_expression
                    intersection_type
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
        scalar
          token
          token
    token
}
```

## JSON AST

```json
{
  "path": "/tmp/intersection-combinations.pl",
  "root": {
    "type": "source_file",
    "start": { "Line": 1, "Column": 1, "Offset": 0 },
    "end": { "Line": 3, "Column": 49, "Offset": 160 },
    "children": [
      {
        "type": "expression_statement",
        "start": { "Line": 1, "Column": 1, "Offset": 0 },
        "end": { "Line": 1, "Column": 48, "Offset": 47 },
        "children": [
          {
            "type": "variable_declaration",
            "start": { "Line": 1, "Column": 1, "Offset": 0 },
            "end": { "Line": 1, "Column": 48, "Offset": 47 },
            "children": [
              { "type": "token", "text": "my" },
              {
                "type": "type_expression",
                "children": [
                  {
                    "type": "parameterized_type",
                    "children": [
                      { "type": "literal", "value": "ArrayRef", "kind": "string" },
                      {
                        "type": "type_parameter_list",
                        "children": [
                          {
                            "type": "type_expression",
                            "children": [
                              {
                                "type": "intersection_type",
                                "children": [
                                  { "type": "literal", "value": "Object", "kind": "string" },
                                  { "type": "literal", "value": "&", "kind": "string" },
                                  { "type": "literal", "value": "Serializable", "kind": "string" }
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
                "children": [
                  { "type": "literal", "value": "@", "kind": "string" },
                  { "type": "token", "text": "serializable_list" }
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
