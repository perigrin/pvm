---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - stress-testing
    - performance
    - deep-nesting
    - many-alternatives
    - complex-types
---

# Stress Testing

Stress testing with very deep nesting and many union alternatives

```perl
my Map[Str, ArrayRef[HashRef[Tuple[Int, Result[Data[UserInfo[Profile[Settings[Theme]]]], Error[ValidationFailure[FieldError[TypeError]]]]]]] %extremely_nested;
my (A|B|C|D|E|F|G|H|I|J|K|L|M|N|O|P|Q|R|S|T) $many_union_alternatives;
my Container[Wrapper[Inner[Core[Data[Value[Item[Element[Node[Leaf]]]]]]]]] $very_deep;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my %extremely_nested;
my $many_union_alternatives;
my $very_deep;
```

## Typed Perl Output

```perl
my Map[Str, ArrayRef[HashRef[Tuple[Int, Result[Data[UserInfo[Profile[Settings[Theme]]]], Error[ValidationFailure[FieldError[TypeError]]]]]]] %extremely_nested;
my (A|B|C|D|E|F|G|H|I|J|K|L|M|N|O|P|Q|R|S|T) $many_union_alternatives;
my Container[Wrapper[Inner[Core[Data[Value[Item[Element[Node[Leaf]]]]]]]]] $very_deep;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
AST {
  Path: /tmp/stress-testing.pl
  Source length: 317 characters
  Type Annotations:
    VarAnnotation: %extremely_nested :: Map[Str, ArrayRef[HashRef[Tuple[Int, Result[Data[UserInfo[Profile[Settings[Theme]]]], Error[ValidationFailure[FieldError[TypeError]]]]]]] at 1:1
    VarAnnotation: $many_union_alternatives :: A|B|C|D|E|F|G|H|I|J|K|L|M|N|O|P|Q|R|S|T at 2:1
    VarAnnotation: $very_deep :: Container[Wrapper[Inner[Core[Data[Value[Item[Element[Node[Leaf]]]]]]]]] at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      variable_declaration
        token
        type_expression
          parameterized_type
            [extremely deep nesting structure omitted for brevity]
        hash
          expression_stmt
            literal
          token
    token
    expression_statement
      variable_declaration
        token
        token
        type_expression
          union_type
            [20 union alternatives omitted for brevity]
        token
        scalar
          token
          token
    token
    expression_statement
      variable_declaration
        token
        type_expression
          parameterized_type
            [10 levels of deep nesting omitted for brevity]
        scalar
          token
          token
    token
}
```

## JSON AST

```json
{
  "path": "/tmp/stress-testing.pl",
  "root": {
    "type": "source_file",
    "start": { "Line": 1, "Column": 1, "Offset": 0 },
    "end": { "Line": 3, "Column": 78, "Offset": 317 },
    "children": [
      {
        "type": "expression_statement",
        "start": { "Line": 1, "Column": 1, "Offset": 0 },
        "end": { "Line": 1, "Column": 143, "Offset": 142 },
        "children": [
          {
            "type": "variable_declaration",
            "children": [
              { "type": "token", "text": "my" },
              {
                "type": "type_expression",
                "children": [
                  {
                    "type": "parameterized_type",
                    "note": "Extremely nested type structure with Map[Str, ArrayRef[HashRef[Tuple[...]]]]"
                  }
                ]
              },
              {
                "type": "hash",
                "children": [
                  { "type": "literal", "value": "%", "kind": "string" },
                  { "type": "token", "text": "extremely_nested" }
                ]
              }
            ]
          }
        ]
      },
      {
        "type": "expression_statement",
        "start": { "Line": 2, "Column": 1, "Offset": 143 },
        "end": { "Line": 2, "Column": 65, "Offset": 207 },
        "children": [
          {
            "type": "variable_declaration",
            "children": [
              { "type": "token", "text": "my" },
              { "type": "token", "text": "(" },
              {
                "type": "type_expression",
                "children": [
                  {
                    "type": "union_type",
                    "note": "Union of 20 alternatives: A|B|C|D|E|F|G|H|I|J|K|L|M|N|O|P|Q|R|S|T"
                  }
                ]
              },
              { "type": "token", "text": ")" },
              {
                "type": "scalar",
                "children": [
                  { "type": "token", "text": "$many_union_alternatives" }
                ]
              }
            ]
          }
        ]
      },
      {
        "type": "expression_statement",
        "start": { "Line": 3, "Column": 1, "Offset": 208 },
        "end": { "Line": 3, "Column": 78, "Offset": 317 },
        "children": [
          {
            "type": "variable_declaration",
            "children": [
              { "type": "token", "text": "my" },
              {
                "type": "type_expression",
                "children": [
                  {
                    "type": "parameterized_type",
                    "note": "Very deeply nested: Container[Wrapper[Inner[Core[Data[Value[Item[Element[Node[Leaf]]]]]]]]]"
                  }
                ]
              },
              {
                "type": "scalar",
                "children": [
                  { "type": "token", "text": "$very_deep" }
                ]
              }
            ]
          }
        ]
      }
    ]
  }
}
```

# Expected Type Errors

(none)
