---
category: typed-perl
subcategory: methods-fields
tags:
    - method-definitions
    - return-types
    - complex-return-types
    - parameterized-types
type_check: true
---

# Method Return Types

Methods with various return type annotations including complex types

```perl
method Int get_number() {
    return 42;
}

method ArrayRef[Str] get_array() {
    return ["a", "b", "c"];
}

method HashRef[Int] get_hash() {
    return { count => 5, total => 100 };
}

method Void get_nothing() {
    # Side effects only
    print "Done\n";
}

method Optional[Str] get_optional() {
    return undef;
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 319 characters
  Type Annotations:
    MethodReturnAnnotation: get_number :: Int at 1:8
    MethodReturnAnnotation: get_array :: ArrayRef[Str] at 5:8
    MethodReturnAnnotation: get_hash :: HashRef[Int] at 9:8
    MethodReturnAnnotation: get_nothing :: Void at 13:8
    MethodReturnAnnotation: get_optional :: Optional[Str] at 18:8
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      type_expr
      block_stmt
        token
        expression_stmt
          literal
        token
        token
    method_decl
      type_expr
      block_stmt
        token
        expression_stmt
          literal
        token
        token
    method_decl
      type_expr
      block_stmt
        token
        expression_stmt
          literal
        token
        token
    method_decl
      type_expr
      block_stmt
        token
        expression_stmt
          literal
        expression_stmt
          literal
        token
        token
    method_decl
      type_expr
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
  "path": "/tmp/test_code_6.pl",
  "root": {
    "type": "source_file",
    "start": {"Line": 1, "Column": 1, "Offset": 0},
    "end": {"Line": 20, "Column": 2, "Offset": 319},
    "children": [/* 5 method declarations with various return types */]
  },
  "type_annotations": [
    {"annotated_item": "get_number", "type_expression": {"Kind": 0, "Name": "Int"}, "position": {"Line": 1, "Column": 8, "Offset": 0}, "kind": "MethodReturnAnnotation"},
    {"annotated_item": "get_array", "type_expression": {"Kind": 4, "Name": "ArrayRef[Str]"}, "position": {"Line": 5, "Column": 8, "Offset": 0}, "kind": "MethodReturnAnnotation"},
    {"annotated_item": "get_hash", "type_expression": {"Kind": 4, "Name": "HashRef[Int]"}, "position": {"Line": 9, "Column": 8, "Offset": 0}, "kind": "MethodReturnAnnotation"},
    {"annotated_item": "get_nothing", "type_expression": {"Kind": 0, "Name": "Void"}, "position": {"Line": 13, "Column": 8, "Offset": 0}, "kind": "MethodReturnAnnotation"},
    {"annotated_item": "get_optional", "type_expression": {"Kind": 4, "Name": "Optional[Str]"}, "position": {"Line": 18, "Column": 8, "Offset": 0}, "kind": "MethodReturnAnnotation"}
  ],
  "errors": [],
  "source_length": 319
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 319 characters
  Type Annotations:
    MethodReturnAnnotation: get_number :: Int at 1:8
    MethodReturnAnnotation: get_array :: ArrayRef[Str] at 5:8
    MethodReturnAnnotation: get_hash :: HashRef[Int] at 9:8
    MethodReturnAnnotation: get_nothing :: Void at 13:8
    MethodReturnAnnotation: get_optional :: Optional[Str] at 18:8
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      type_expr
      block_stmt
        token
        expression_stmt
          literal
        token
        token
    method_decl
      type_expr
      block_stmt
        token
        expression_stmt
          literal
        token
        token
    method_decl
      type_expr
      block_stmt
        token
        expression_stmt
          literal
        token
        token
    method_decl
      type_expr
      block_stmt
        token
        expression_stmt
          literal
        expression_stmt
          literal
        token
        token
    method_decl
      type_expr
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
method get_number() {
    return 42;
}

method get_array() {
    return ["a", "b", "c"];
}

method get_hash() {
    return { count => 5, total => 100 };
}

method get_nothing() {
    # Side effects only
    print "Done\n";
}

method get_optional() {
    return undef;
}
```

## Typed Perl Output

```perl
method Int get_number() {
    return 42;
}

method ArrayRef[Str] get_array() {
    return ["a", "b", "c"];
}

method HashRef[Int] get_hash() {
    return { count => 5, total => 100 };
}

method Void get_nothing() {
    # Side effects only
    print "Done\n";
}

method Optional[Str] get_optional() {
    return undef;
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
