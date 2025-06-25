---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - negation-types
    - complex-combinations
    - intersection-types
---

# Negation Combinations

Negation types combined with parameterized and intersection types

```perl
my ArrayRef[!Undef] @non_undef_array;
my HashRef[Str, !Empty] %non_empty_values;
my Optional[!Null&!Undef] $definitely_defined;
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
              negation_type
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
              expression_stmt
                literal
            expression_stmt
              literal
            type_expression
              negation_type
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
                  negation_type
                    expression_stmt
                      literal
                    type_expression
                      expression_stmt
                        literal
                expression_stmt
                  literal
                type_expression
                  negation_type
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
              negation_type
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
              expression_stmt
                literal
            expression_stmt
              literal
            type_expression
              negation_type
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
                  negation_type
                    expression_stmt
                      literal
                    type_expression
                      expression_stmt
                        literal
                expression_stmt
                  literal
                type_expression
                  negation_type
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
