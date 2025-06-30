---
category: typed-perl
subcategory: methods-fields
tags:
    - mixed-typing
    - gradual-typing
    - backward-compatibility
type_check: true
---

# Mixed Typed Untyped

Mixed typed and untyped methods and fields in the same context

```perl
# Mixed typed and untyped methods and fields
field Int $typed_field = 42;
field $untyped_field = "hello";

method typed_method(Str $input) returns Str {
    return uc($input);
}

sub untyped_sub {
    my ($param) = @_;
    return $param * 2;
}

method partially_typed($untyped, Int $typed) returns Str {
    return "$untyped: $typed";
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 336 characters
  Type Annotations:
    VarAnnotation: $typed_field :: Int at 2:1
    MethodReturnAnnotation: typed_method :: Str at 5:41
    MethodReturnAnnotation: partially_typed :: Str at 14:54
    MethodParamAnnotation: $input :: Str at 5:1
    MethodParamAnnotation: $typed :: Int at 14:1
  Root: source_file
  Tree Structure:
  source_file
    expression_stmt
      literal
    expression_statement
      assignment_expression
        variable_declaration
          token
          type_expression
            expression_stmt
              literal
          scalar
            token
            token
        token
        token
    token
    expression_statement
      assignment_expression
        variable_declaration
          token
          scalar
            token
            token
        token
        interpolated_string_literal
          expression_stmt
            literal
          expression_stmt
            literal
          expression_stmt
            literal
    token
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
    sub_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        expression_stmt
          literal
        token
        token
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 336 characters
  Type Annotations:
    VarAnnotation: $typed_field :: Int at 2:1
    MethodReturnAnnotation: typed_method :: Str at 5:41
    MethodReturnAnnotation: partially_typed :: Str at 14:54
    MethodParamAnnotation: $input :: Str at 5:1
    MethodParamAnnotation: $typed :: Int at 14:1
  Root: source_file
  Tree Structure:
  source_file
    expression_stmt
      literal
    expression_statement
      assignment_expression
        variable_declaration
          token
          type_expression
            expression_stmt
              literal
          scalar
            token
            token
        token
        token
    token
    expression_statement
      assignment_expression
        variable_declaration
          token
          scalar
            token
            token
        token
        interpolated_string_literal
          expression_stmt
            literal
          expression_stmt
            literal
          expression_stmt
            literal
    token
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
    sub_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        expression_stmt
          literal
        token
        token
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
# Mixed typed and untyped methods and fields
field $typed_field = 42;
field $untyped_field = "hello";{ return uc($input); }







sub untyped_sub { my ($param) = @_; return $param * 2; }{ return "$untyped: $typed"; }
```

## Typed Perl Output

```perl
# Mixed typed and untyped methods and fields
field Int $typed_field = 42;
field $untyped_field = "hello";

method typed_method(Str $input) returns Str {
    return uc($input);
}

sub untyped_sub {
    my ($param) = @_;
    return $param * 2;
}

method partially_typed($untyped, Int $typed) returns Str {
    return "$untyped: $typed";
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
