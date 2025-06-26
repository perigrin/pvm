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
```

#
# Expected Compilation Outcomes

## Clean Perl Output

```perl
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

# Expected Type Errors

(none)
