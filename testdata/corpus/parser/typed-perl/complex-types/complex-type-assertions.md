---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - type-assertions
    - method-calls
    - complex-types
    - parameterized-types
---

# Complex Type Assertions

Type assertions with complex type expressions

```perl
my $result = $data as ArrayRef[HashRef[Int|Bool]];
my $complex = ($input->process()) as Map[Str, ArrayRef[MyType]];
my $transformed = $obj->convert() as Result[Data[User], Error[String]];
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $result = $data;
my $complex = ($input->process());
my $transformed = $obj->convert();
```

## Typed Perl Output

```perl
my $result = $data as ArrayRef[HashRef[Int|Bool]];
my $complex = ($input->process()) as Map[Str, ArrayRef[MyType]];
my $transformed = $obj->convert() as Result[Data[User], Error[String]];
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
AST {
  Path: /tmp/complex-type-assertions.pl
  Source length: 187 characters
  Type Annotations:
    VarAnnotation: $data :: ArrayRef[HashRef[Int|Bool]] at 1:14
    VarAnnotation: ( :: Map[Str, ArrayRef[MyType]] at 2:15
    VarAnnotation: $obj->convert() :: Result[Data[User], Error[String]] at 3:19
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      var_decl
        variable
    token
    expression_statement
      var_decl
        variable
    token
    expression_statement
      var_decl
        variable
    token
}
```

## JSON AST

```json
{
  "path": "/tmp/complex-type-assertions.pl",
  "root": {
    "type": "source_file",
    "start": { "Line": 1, "Column": 1, "Offset": 0 },
    "end": { "Line": 3, "Column": 75, "Offset": 187 },
    "children": [
      {
        "type": "expression_statement",
        "start": { "Line": 1, "Column": 1, "Offset": 0 },
        "end": { "Line": 1, "Column": 50, "Offset": 49 },
        "children": [
          {
            "type": "var_decl",
            "start": { "Line": 1, "Column": 1, "Offset": 0 },
            "end": { "Line": 1, "Column": 50, "Offset": 49 },
            "children": [
              {
                "type": "variable",
                "start": { "Line": 1, "Column": 1, "Offset": 0 },
                "end": { "Line": 1, "Column": 50, "Offset": 49 },
                "children": [
                  { "type": "token", "text": "my" },
                  { "type": "token", "text": "$result" },
                  { "type": "token", "text": "=" },
                  { "type": "token", "text": "$data" },
                  { "type": "token", "text": "as" },
                  {
                    "type": "type_expression",
                    "children": [
                      {
                        "type": "parameterized_type",
                        "children": [
                          { "type": "literal", "value": "ArrayRef", "kind": "string" },
                          {
                            "type": "type_parameter_list",
                            "children": [
                              {
                                "type": "type_expression",
                                "children": [
                                  {
                                    "type": "parameterized_type",
                                    "children": [
                                      { "type": "literal", "value": "HashRef", "kind": "string" },
                                      {
                                        "type": "type_parameter_list",
                                        "children": [
                                          {
                                            "type": "type_expression",
                                            "children": [
                                              {
                                                "type": "union_type",
                                                "children": [
                                                  { "type": "literal", "value": "Int", "kind": "string" },
                                                  { "type": "literal", "value": "|", "kind": "string" },
                                                  { "type": "literal", "value": "Bool", "kind": "string" }
                                                ]
                                              }
                                            ]
                                          }
                                        ]
                                      }
                                    ]
                                  }
                                ]
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  }
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
