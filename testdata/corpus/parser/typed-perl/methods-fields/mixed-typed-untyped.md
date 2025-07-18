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

method Str typed_method(Str $input) {
    return uc($input);
}

sub untyped_sub {
    my ($param) = @_;
    return $param * 2;
}

method Str partially_typed($untyped, Int $typed) {
    return "$untyped: $typed";
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 320 characters
  Type Annotations:
    VarAnnotation: $typed_field :: Int at 2:1
    MethodReturnAnnotation: typed_method :: Str at 5:8
    MethodParamAnnotation: $input :: Str at 5:25
    MethodReturnAnnotation: partially_typed :: Str at 14:8
    MethodParamAnnotation: $typed :: Int at 14:38
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

## JSON AST

```json
{
  "path": "/tmp/test_code_7.pl",
  "root": {
    "type": "source_file",
    "start": {"Line": 1, "Column": 1, "Offset": 0},
    "end": {"Line": 16, "Column": 2, "Offset": 320},
    "children": [/* Mixed typed and untyped field/method declarations */]
  },
  "type_annotations": [
    {"annotated_item": "$typed_field", "type_expression": {"Kind": 0, "Name": "Int"}, "position": {"Line": 2, "Column": 1, "Offset": 0}, "kind": "VarAnnotation"},
    {"annotated_item": "typed_method", "type_expression": {"Kind": 0, "Name": "Str"}, "position": {"Line": 5, "Column": 8, "Offset": 0}, "kind": "MethodReturnAnnotation"},
    {"annotated_item": "$input", "type_expression": {"Kind": 0, "Name": "Str"}, "position": {"Line": 5, "Column": 25, "Offset": 0}, "kind": "MethodParamAnnotation"},
    {"annotated_item": "partially_typed", "type_expression": {"Kind": 0, "Name": "Str"}, "position": {"Line": 14, "Column": 8, "Offset": 0}, "kind": "MethodReturnAnnotation"},
    {"annotated_item": "$typed", "type_expression": {"Kind": 0, "Name": "Int"}, "position": {"Line": 14, "Column": 38, "Offset": 0}, "kind": "MethodParamAnnotation"}
  ],
  "errors": [],
  "source_length": 320
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 320 characters
  Type Annotations:
    VarAnnotation: $typed_field :: Int at 2:1
    MethodReturnAnnotation: typed_method :: Str at 5:8
    MethodParamAnnotation: $input :: Str at 5:25
    MethodReturnAnnotation: partially_typed :: Str at 14:8
    MethodParamAnnotation: $typed :: Int at 14:38
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
use v5.36;
# Mixed typed and untyped methods and fields
field $typed_field = 42;
field $untyped_field = "hello";

method typed_method($input) {
    return uc($input);
}







sub untyped_sub {
    my ($param) = @_;
    return $param * 2;
}

method partially_typed($untyped, $typed) {
    return "$untyped: $typed";
}
```

## Typed Perl Output

```perl
# Mixed typed and untyped methods and fields
field Int $typed_field = 42;
field $untyped_field = "hello";

method Str typed_method(Str $input) {
    return uc($input);
}

sub untyped_sub {
    my ($param) = @_;
    return $param * 2;
}

method Str partially_typed($untyped, Int $typed) {
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
