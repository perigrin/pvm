---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - deep-nesting
    - parameterized-types
    - complex-combinations
---

# Deep Nesting

Deeply nested parameterized types with complex combinations

```perl
my ArrayRef[HashRef[ArrayRef[Int|Str]]] @deep_nested;
my Map[Str, ArrayRef[Tuple[Int, Bool|Str]]] %complex_map;
my Container[Wrapper[Inner[Data[Value]]]] $deeply_nested;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @deep_nested;
my %complex_map;
my $deeply_nested;
```

## Typed Perl Output

```perl
my ArrayRef[HashRef[ArrayRef[Int|Str]]] @deep_nested;
my Map[Str, ArrayRef[Tuple[Int, Bool|Str]]] %complex_map;
my Container[Wrapper[Inner[Data[Value]]]] $deeply_nested;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
AST {
  Path: /tmp/deep-nesting.pl
  Source length: 169 characters
  Type Annotations:
    VarAnnotation: @deep_nested :: ArrayRef[HashRef[ArrayRef[Int|Str]]] at 1:1
    VarAnnotation: %complex_map :: Map[Str, ArrayRef[Tuple[Int, Bool|Str]]] at 2:1
    VarAnnotation: $deeply_nested :: Container[Wrapper[Inner[Data[Value]]]] at 3:1
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
                parameterized_type
                  expression_stmt
                    literal
                  expression_stmt
                    literal
                  type_parameter_list
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
                parameterized_type
                  expression_stmt
                    literal
                  expression_stmt
                    literal
                  type_parameter_list
                    type_expression
                      parameterized_type
                        expression_stmt
                          literal
                        expression_stmt
                          literal
                        type_parameter_list
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
  "path": "/tmp/deep-nesting.pl",
  "root": {
    "type": "source_file",
    "start": { "Line": 1, "Column": 1, "Offset": 0 },
    "end": { "Line": 3, "Column": 53, "Offset": 169 },
    "children": [
      {
        "type": "expression_statement",
        "start": { "Line": 1, "Column": 1, "Offset": 0 },
        "end": { "Line": 1, "Column": 46, "Offset": 45 },
        "children": [
          {
            "type": "variable_declaration",
            "start": { "Line": 1, "Column": 1, "Offset": 0 },
            "end": { "Line": 1, "Column": 46, "Offset": 45 },
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
                                "type": "parameterized_type",
                                "children": [
                                  { "type": "literal", "value": "HashRef", "kind": "string" },
                                  {
                                    "type": "type_parameter_list",
                                    "children": [
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
                                                        "type": "union_type",
                                                        "children": [
                                                          { "type": "literal", "value": "Int", "kind": "string" },
                                                          { "type": "literal", "value": "|", "kind": "string" },
                                                          { "type": "literal", "value": "Str", "kind": "string" }
                                                        ]
                                                      }
                                                    ]
                                                  }
                                                ]
                                              }
                                            ]
                                          }
                                        ]
                                      }
                                    ]
                                  }
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
                  { "type": "token", "text": "deep_nested" }
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
