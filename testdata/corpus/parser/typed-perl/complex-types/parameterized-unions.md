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

# Expected AST

## Text AST

```
AST {
  Path: /tmp/parameterized-unions.pl
  Source length: 146 characters
  Type Annotations:
    VarAnnotation: $param_union :: ArrayRef[Int]|HashRef[Str] at 1:1
    VarAnnotation: $flexible :: Container[MyType]|Wrapper[OtherType] at 2:1
    VarAnnotation: $outcome :: Result[Data, Error]|Maybe[Value] at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      variable_declaration
        token
        token
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
        token
        scalar
          token
          token
    token
    expression_statement
      variable_declaration
        token
        token
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
        token
        scalar
          token
          token
    token
    expression_statement
      variable_declaration
        token
        token
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
        token
        scalar
          token
          token
    token
}
```

## JSON AST

```json
{
  "path": "/tmp/parameterized-unions.pl",
  "root": {
    "type": "source_file",
    "start": { "Line": 1, "Column": 1, "Offset": 0 },
    "end": { "Line": 3, "Column": 48, "Offset": 146 },
    "children": [
      {
        "type": "expression_statement",
        "start": { "Line": 1, "Column": 1, "Offset": 0 },
        "end": { "Line": 1, "Column": 45, "Offset": 44 },
        "children": [
          {
            "type": "variable_declaration",
            "start": { "Line": 1, "Column": 1, "Offset": 0 },
            "end": { "Line": 1, "Column": 45, "Offset": 44 },
            "children": [
              { "type": "token", "start": { "Line": 1, "Column": 1, "Offset": 0 }, "end": { "Line": 1, "Column": 3, "Offset": 2 }, "text": "my" },
              { "type": "token", "start": { "Line": 1, "Column": 4, "Offset": 3 }, "end": { "Line": 1, "Column": 5, "Offset": 4 }, "text": "(" },
              {
                "type": "type_expression",
                "start": { "Line": 1, "Column": 5, "Offset": 4 },
                "end": { "Line": 1, "Column": 31, "Offset": 30 },
                "children": [
                  {
                    "type": "union_type",
                    "start": { "Line": 1, "Column": 5, "Offset": 4 },
                    "end": { "Line": 1, "Column": 31, "Offset": 30 },
                    "children": [
                      {
                        "type": "type_expression",
                        "start": { "Line": 1, "Column": 5, "Offset": 4 },
                        "end": { "Line": 1, "Column": 18, "Offset": 17 },
                        "children": [
                          {
                            "type": "parameterized_type",
                            "start": { "Line": 1, "Column": 5, "Offset": 4 },
                            "end": { "Line": 1, "Column": 18, "Offset": 17 },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": { "Line": 1, "Column": 5, "Offset": 4 },
                                "end": { "Line": 1, "Column": 13, "Offset": 12 },
                                "children": [{ "type": "literal", "start": { "Line": 1, "Column": 5, "Offset": 4 }, "end": { "Line": 1, "Column": 13, "Offset": 12 }, "value": "ArrayRef", "kind": "string" }]
                              },
                              {
                                "type": "expression_stmt",
                                "start": { "Line": 1, "Column": 13, "Offset": 12 },
                                "end": { "Line": 1, "Column": 14, "Offset": 13 },
                                "children": [{ "type": "literal", "start": { "Line": 1, "Column": 13, "Offset": 12 }, "end": { "Line": 1, "Column": 14, "Offset": 13 }, "value": "[", "kind": "string" }]
                              },
                              {
                                "type": "type_parameter_list",
                                "start": { "Line": 1, "Column": 14, "Offset": 13 },
                                "end": { "Line": 1, "Column": 17, "Offset": 16 },
                                "children": [
                                  {
                                    "type": "type_expression",
                                    "start": { "Line": 1, "Column": 14, "Offset": 13 },
                                    "end": { "Line": 1, "Column": 17, "Offset": 16 },
                                    "children": [
                                      {
                                        "type": "expression_stmt",
                                        "start": { "Line": 1, "Column": 14, "Offset": 13 },
                                        "end": { "Line": 1, "Column": 17, "Offset": 16 },
                                        "children": [{ "type": "literal", "start": { "Line": 1, "Column": 14, "Offset": 13 }, "end": { "Line": 1, "Column": 17, "Offset": 16 }, "value": "Int", "kind": "string" }]
                                      }
                                    ]
                                  }
                                ]
                              },
                              {
                                "type": "expression_stmt",
                                "start": { "Line": 1, "Column": 17, "Offset": 16 },
                                "end": { "Line": 1, "Column": 18, "Offset": 17 },
                                "children": [{ "type": "literal", "start": { "Line": 1, "Column": 17, "Offset": 16 }, "end": { "Line": 1, "Column": 18, "Offset": 17 }, "value": "]", "kind": "string" }]
                              }
                            ]
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": { "Line": 1, "Column": 18, "Offset": 17 },
                        "end": { "Line": 1, "Column": 19, "Offset": 18 },
                        "children": [{ "type": "literal", "start": { "Line": 1, "Column": 18, "Offset": 17 }, "end": { "Line": 1, "Column": 19, "Offset": 18 }, "value": "|", "kind": "string" }]
                      },
                      {
                        "type": "type_expression",
                        "start": { "Line": 1, "Column": 19, "Offset": 18 },
                        "end": { "Line": 1, "Column": 31, "Offset": 30 },
                        "children": [
                          {
                            "type": "parameterized_type",
                            "start": { "Line": 1, "Column": 19, "Offset": 18 },
                            "end": { "Line": 1, "Column": 31, "Offset": 30 },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": { "Line": 1, "Column": 19, "Offset": 18 },
                                "end": { "Line": 1, "Column": 26, "Offset": 25 },
                                "children": [{ "type": "literal", "start": { "Line": 1, "Column": 19, "Offset": 18 }, "end": { "Line": 1, "Column": 26, "Offset": 25 }, "value": "HashRef", "kind": "string" }]
                              },
                              {
                                "type": "expression_stmt",
                                "start": { "Line": 1, "Column": 26, "Offset": 25 },
                                "end": { "Line": 1, "Column": 27, "Offset": 26 },
                                "children": [{ "type": "literal", "start": { "Line": 1, "Column": 26, "Offset": 25 }, "end": { "Line": 1, "Column": 27, "Offset": 26 }, "value": "[", "kind": "string" }]
                              },
                              {
                                "type": "type_parameter_list",
                                "start": { "Line": 1, "Column": 27, "Offset": 26 },
                                "end": { "Line": 1, "Column": 30, "Offset": 29 },
                                "children": [
                                  {
                                    "type": "type_expression",
                                    "start": { "Line": 1, "Column": 27, "Offset": 26 },
                                    "end": { "Line": 1, "Column": 30, "Offset": 29 },
                                    "children": [
                                      {
                                        "type": "expression_stmt",
                                        "start": { "Line": 1, "Column": 27, "Offset": 26 },
                                        "end": { "Line": 1, "Column": 30, "Offset": 29 },
                                        "children": [{ "type": "literal", "start": { "Line": 1, "Column": 27, "Offset": 26 }, "end": { "Line": 1, "Column": 30, "Offset": 29 }, "value": "Str", "kind": "string" }]
                                      }
                                    ]
                                  }
                                ]
                              },
                              {
                                "type": "expression_stmt",
                                "start": { "Line": 1, "Column": 30, "Offset": 29 },
                                "end": { "Line": 1, "Column": 31, "Offset": 30 },
                                "children": [{ "type": "literal", "start": { "Line": 1, "Column": 30, "Offset": 29 }, "end": { "Line": 1, "Column": 31, "Offset": 30 }, "value": "]", "kind": "string" }]
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  }
                ]
              },
              { "type": "token", "start": { "Line": 1, "Column": 31, "Offset": 30 }, "end": { "Line": 1, "Column": 32, "Offset": 31 }, "text": ")" },
              {
                "type": "scalar",
                "start": { "Line": 1, "Column": 33, "Offset": 32 },
                "end": { "Line": 1, "Column": 45, "Offset": 44 },
                "children": [
                  { "type": "token", "start": { "Line": 1, "Column": 33, "Offset": 32 }, "end": { "Line": 1, "Column": 34, "Offset": 33 }, "text": "$" },
                  { "type": "token", "start": { "Line": 1, "Column": 34, "Offset": 33 }, "end": { "Line": 1, "Column": 45, "Offset": 44 }, "text": "param_union" }
                ]
              }
            ]
          }
        ]
      },
      { "type": "token", "start": { "Line": 1, "Column": 45, "Offset": 44 }, "end": { "Line": 1, "Column": 46, "Offset": 45 }, "text": ";" }
    ]
  }
}
```

# Expected Type Errors

(none)
