---
category: typed-perl
subcategory: methods-fields
tags:
    - complex-methods
    - method-signatures
    - optional-parameters
    - parameterized-types
    - complex-return-types
type_check: true
---

# Complex Method Signatures

Complex method signatures with parameterized types, optional parameters, and multiple parameter types

```perl
method ArrayRef[Str] process(ArrayRef[Str] $data, Bool $validate = 1) {
    my @result = @{$data};
    return \@result;
}

method HashRef[Int] transform(HashRef[Int] $input, CodeRef $callback) {
    my %result;
    for my $key (keys %{$input}) {
        $result{$key} = $callback->($input->{$key});
    }
    return \%result;
}

method Bool complex_method(ArrayRef[HashRef[Int]] $data, Optional[CodeRef] $processor) {
    return 1;
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 433 characters
  Type Annotations:
    MethodReturnAnnotation: process :: ArrayRef[Str] at 1:8
    MethodParamAnnotation: $data :: ArrayRef[Str] at 1:30
    MethodParamAnnotation: $validate :: Bool at 1:51
    MethodReturnAnnotation: transform :: HashRef[Int] at 6:8
    MethodParamAnnotation: $input :: HashRef[Int] at 6:31
    MethodParamAnnotation: $callback :: CodeRef at 6:52
    MethodReturnAnnotation: complex_method :: Bool at 14:8
    MethodParamAnnotation: $data :: ArrayRef[HashRef[Int]] at 14:28
    MethodParamAnnotation: $processor :: Optional[CodeRef] at 14:58
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      type_expr
      block_stmt
        token
        token
        return_stmt
          literal
        token
        token
    method_decl
      type_expr
      block_stmt
        token
        token
        for_stmt
        return_stmt
          literal
        token
        token
    method_decl
      type_expr
      block_stmt
        token
        return_stmt
          literal
        token
        token
}
```

## JSON AST

```json
{
  "path": "/tmp/test_code_5.pl",
  "root": {
    "type": "source_file",
    "start": {"Line": 1, "Column": 1, "Offset": 0},
    "end": {"Line": 16, "Column": 2, "Offset": 433},
    "children": [/* 3 method declarations with complex signatures */]
  },
  "type_annotations": [
    {"annotated_item": "process", "type_expression": {"Kind": 4, "Name": "ArrayRef[Str]"}, "position": {"Line": 1, "Column": 8, "Offset": 0}, "kind": "MethodReturnAnnotation"},
    {"annotated_item": "$data", "type_expression": {"Kind": 4, "Name": "ArrayRef[Str]"}, "position": {"Line": 1, "Column": 30, "Offset": 0}, "kind": "MethodParamAnnotation"},
    {"annotated_item": "$validate", "type_expression": {"Kind": 0, "Name": "Bool"}, "position": {"Line": 1, "Column": 51, "Offset": 0}, "kind": "MethodParamAnnotation"},
    {"annotated_item": "transform", "type_expression": {"Kind": 4, "Name": "HashRef[Int]"}, "position": {"Line": 6, "Column": 8, "Offset": 0}, "kind": "MethodReturnAnnotation"},
    {"annotated_item": "$input", "type_expression": {"Kind": 4, "Name": "HashRef[Int]"}, "position": {"Line": 6, "Column": 31, "Offset": 0}, "kind": "MethodParamAnnotation"},
    {"annotated_item": "$callback", "type_expression": {"Kind": 0, "Name": "CodeRef"}, "position": {"Line": 6, "Column": 52, "Offset": 0}, "kind": "MethodParamAnnotation"},
    {"annotated_item": "complex_method", "type_expression": {"Kind": 0, "Name": "Bool"}, "position": {"Line": 14, "Column": 8, "Offset": 0}, "kind": "MethodReturnAnnotation"},
    {"annotated_item": "$data", "type_expression": {"Kind": 4, "Name": "ArrayRef[HashRef[Int]]"}, "position": {"Line": 14, "Column": 28, "Offset": 0}, "kind": "MethodParamAnnotation"},
    {"annotated_item": "$processor", "type_expression": {"Kind": 4, "Name": "Optional[CodeRef]"}, "position": {"Line": 14, "Column": 58, "Offset": 0}, "kind": "MethodParamAnnotation"}
  ],
  "errors": [],
  "source_length": 433
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 433 characters
  Type Annotations:
    MethodReturnAnnotation: process :: ArrayRef[Str] at 1:8
    MethodParamAnnotation: $data :: ArrayRef[Str] at 1:30
    MethodParamAnnotation: $validate :: Bool at 1:51
    MethodReturnAnnotation: transform :: HashRef[Int] at 6:8
    MethodParamAnnotation: $input :: HashRef[Int] at 6:31
    MethodParamAnnotation: $callback :: CodeRef at 6:52
    MethodReturnAnnotation: complex_method :: Bool at 14:8
    MethodParamAnnotation: $data :: ArrayRef[HashRef[Int]] at 14:28
    MethodParamAnnotation: $processor :: Optional[CodeRef] at 14:58
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      type_expr
      block_stmt
        token
        token
        return_stmt
          literal
        token
        token
    method_decl
      type_expr
      block_stmt
        token
        token
        for_stmt
        return_stmt
          literal
        token
        token
    method_decl
      type_expr
      block_stmt
        token
        return_stmt
          literal
        token
        token
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
method process($data, $validate = 1) {
    my @result = @{$data};
    return \@result;
}

method transform($input, $callback) {
    my %result;
    for my $key (keys %{$input}) {
        $result{$key} = $callback->($input->{$key});
    }
    return \%result;
}

method complex_method($data, $processor) {
    return 1;
}
```

## Typed Perl Output

```perl
method ArrayRef[Str] process(ArrayRef[Str] $data, Bool $validate = 1) {
    my @result = @{$data};
    return \@result;
}

method HashRef[Int] transform(HashRef[Int] $input, CodeRef $callback) {
    my %result;
    for my $key (keys %{$input}) {
        $result{$key} = $callback->($input->{$key});
    }
    return \%result;
}

method Bool complex_method(ArrayRef[HashRef[Int]] $data, Optional[CodeRef] $processor) {
    return 1;
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
