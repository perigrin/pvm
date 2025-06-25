---
category: typed-perl
subcategory: parameterized-types
tags:
    - complex-combinations
    - complex-nesting
    - deep-nesting
    - unions
    - Map
    - parameterized-types
type_check: true
---

# Complex Combinations

Complex combinations of parameterized types with unions and deep nesting

```perl
my Map[Str, ArrayRef[HashRef[Int|Bool]]] $complex;
my Container[ArrayRef[MyType]|HashRef[OtherType]] $flexible;
my Result[Data[UserInfo], Error[ValidationFailure]] $nested_result;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 179 characters
  Type Annotations:
    VarAnnotation: $complex :: Map[Str, ArrayRef[HashRef[Int|Bool]]] at 1:1
    VarAnnotation: $flexible :: Container[ArrayRef[MyType]|HashRef[OtherType]] at 2:1
    VarAnnotation: $nested_result :: Result[Data[UserInfo], Error[ValidationFailure]] at 3:1
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
        scalar
          token
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
        scalar
          token
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
        scalar
          token
          token
    token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 179 characters
  Type Annotations:
    VarAnnotation: $complex :: Map[Str, ArrayRef[HashRef[Int|Bool]]] at 1:1
    VarAnnotation: $flexible :: Container[ArrayRef[MyType]|HashRef[OtherType]] at 2:1
    VarAnnotation: $nested_result :: Result[Data[UserInfo], Error[ValidationFailure]] at 3:1
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
        scalar
          token
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
        scalar
          token
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
        scalar
          token
          token
    token
}
```

# Expected Type Errors

```
(none)
```
