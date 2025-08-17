---
category: typed-perl
subcategory: methods-fields
tags:
    - complex-fields
    - parameterized-types
    - custom-types
    - field-declarations
type_check: true
---

# Complex Field Types

Field declarations with complex parameterized types and custom types

```perl
field ArrayRef[Int] $numbers = [];
field HashRef[Str] $config = {};
field CodeRef[Int, Str] $formatter;
field ArrayRef[MyType] $items;
field HashRef[ArrayRef[Str]] $grouped_data = {};
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 183 characters
  Type Annotations:
    VarAnnotation: $numbers :: ArrayRef[Int] at 1:1
    VarAnnotation: $config :: HashRef[Str] at 2:1
    VarAnnotation: $formatter :: CodeRef[Int, Str] at 3:1
    VarAnnotation: $items :: ArrayRef[MyType] at 4:1
    VarAnnotation: $grouped_data :: HashRef[ArrayRef[Str]] at 5:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      assignment_expression
        token
        anonymous_array_expression
          expression_stmt
            literal
          expression_stmt
            literal
    token
    expression_statement
      assignment_expression
        token
        anonymous_hash_expression
          token
          token
    token
    expression_stmt
      literal
    token
    expression_stmt
      literal
    token
    expression_statement
      assignment_expression
        token
        anonymous_hash_expression
          token
          token
    token
}
```

## JSON AST

```json
{
  "path": "/tmp/test_code_4.pl",
  "root": {
    "type": "source_file",
    "start": {"Line": 1, "Column": 1, "Offset": 0},
    "end": {"Line": 5, "Column": 49, "Offset": 183},
    "children": [/* Detailed AST structure with 5 field declarations with complex parameterized types */]
  },
  "type_annotations": [
    {"annotated_item": "$numbers", "type_expression": {"Kind": 4, "Name": "ArrayRef[Int]", "Parameters": [{"Kind": 0, "Name": "Int"}]}, "position": {"Line": 1, "Column": 1, "Offset": 0}, "kind": "VarAnnotation"},
    {"annotated_item": "$config", "type_expression": {"Kind": 4, "Name": "HashRef[Str]", "Parameters": [{"Kind": 0, "Name": "Str"}]}, "position": {"Line": 2, "Column": 1, "Offset": 0}, "kind": "VarAnnotation"},
    {"annotated_item": "$formatter", "type_expression": {"Kind": 4, "Name": "CodeRef[Int,Str]", "Parameters": [{"Kind": 0, "Name": "Int"}, {"Kind": 0, "Name": "Str"}]}, "position": {"Line": 3, "Column": 1, "Offset": 0}, "kind": "VarAnnotation"},
    {"annotated_item": "$items", "type_expression": {"Kind": 4, "Name": "ArrayRef[MyType]", "Parameters": [{"Kind": 0, "Name": "MyType"}]}, "position": {"Line": 4, "Column": 1, "Offset": 0}, "kind": "VarAnnotation"},
    {"annotated_item": "$grouped_data", "type_expression": {"Kind": 4, "Name": "HashRef[ArrayRef[Str]]", "Parameters": [{"Kind": 4, "Name": "ArrayRef[Str]", "Parameters": [{"Kind": 0, "Name": "Str"}]}]}, "position": {"Line": 5, "Column": 1, "Offset": 0}, "kind": "VarAnnotation"}
  ],
  "errors": [],
  "source_length": 183
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 183 characters
  Type Annotations:
    VarAnnotation: $numbers :: ArrayRef[Int] at 1:1
    VarAnnotation: $config :: HashRef[Str] at 2:1
    VarAnnotation: $formatter :: CodeRef[Int, Str] at 3:1
    VarAnnotation: $items :: ArrayRef[MyType] at 4:1
    VarAnnotation: $grouped_data :: HashRef[ArrayRef[Str]] at 5:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      assignment_expression
        token
        anonymous_array_expression
          expression_stmt
            literal
          expression_stmt
            literal
    token
    expression_statement
      assignment_expression
        token
        anonymous_hash_expression
          token
          token
    token
    expression_stmt
      literal
    token
    expression_stmt
      literal
    token
    expression_statement
      assignment_expression
        token
        anonymous_hash_expression
          token
          token
    token
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
field $numbers = [];
field $config = {};
field $formatter;
field $items;
field $grouped_data = {};
```

## Typed Perl Output

```perl
field ArrayRef[Int] $numbers = [];
field HashRef[Str] $config = {};
field CodeRef[Int, Str] $formatter;
field ArrayRef[MyType] $items;
field HashRef[ArrayRef[Str]] $grouped_data = {};
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
