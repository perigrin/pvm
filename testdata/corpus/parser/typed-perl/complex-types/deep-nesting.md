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
```

#
# Expected Compilation Outcomes

## Clean Perl Output

```perl
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

# Expected Type Errors

(none)
