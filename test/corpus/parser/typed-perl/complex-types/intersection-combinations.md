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

## Expected AST

### Before Type Inference

```
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
```

### After Type Inference

```
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
```

## Expected Type Errors

(none)
