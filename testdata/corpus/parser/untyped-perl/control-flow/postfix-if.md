---
category: untyped-perl
subcategory: control-flow
tags:
    - postfix
    - if
    - conditional
---

# Postfix If

Postfix if conditional statement

```perl
print "Hello" if $debug;
$count++ if $item;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
print "Hello" if $debug;
$count++ if $item;
```

## Typed Perl Output

```perl
print "Hello" if $debug;
$count++ if $item;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```

# Expected AST

## Text AST

```
AST {
  Path: /tmp/simple_postfix_if.pl
  Source length: 25 characters
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      postfix_conditional_expression
        ambiguous_function_call_expression
          expression_stmt
            literal
          interpolated_string_literal
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
}
```

## JSON AST

```json
{
  "path": "/tmp/simple_postfix_if.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 2,
      "Column": 1,
      "Offset": 25
    },
    "children": [
      {
        "type": "expression_statement",
        "start": {
          "Line": 1,
          "Column": 1,
          "Offset": 0
        },
        "end": {
          "Line": 1,
          "Column": 24,
          "Offset": 23
        },
        "children": [
          {
            "type": "postfix_conditional_expression",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 1,
              "Column": 24,
              "Offset": 23
            },
            "children": [
              {
                "type": "ambiguous_function_call_expression",
                "start": {
                  "Line": 1,
                  "Column": 1,
                  "Offset": 0
                },
                "end": {
                  "Line": 1,
                  "Column": 14,
                  "Offset": 13
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 1,
                      "Column": 1,
                      "Offset": 0
                    },
                    "end": {
                      "Line": 1,
                      "Column": 6,
                      "Offset": 5
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 1,
                          "Column": 1,
                          "Offset": 0
                        },
                        "end": {
                          "Line": 1,
                          "Column": 6,
                          "Offset": 5
                        },
                        "value": "print",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "interpolated_string_literal",
                    "start": {
                      "Line": 1,
                      "Column": 7,
                      "Offset": 6
                    },
                    "end": {
                      "Line": 1,
                      "Column": 14,
                      "Offset": 13
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 7,
                          "Offset": 6
                        },
                        "end": {
                          "Line": 1,
                          "Column": 8,
                          "Offset": 7
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 7,
                              "Offset": 6
                            },
                            "end": {
                              "Line": 1,
                              "Column": 8,
                              "Offset": 7
                            },
                            "value": "\"",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 8,
                          "Offset": 7
                        },
                        "end": {
                          "Line": 1,
                          "Column": 13,
                          "Offset": 12
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 8,
                              "Offset": 7
                            },
                            "end": {
                              "Line": 1,
                              "Column": 13,
                              "Offset": 12
                            },
                            "value": "Hello",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 13,
                          "Offset": 12
                        },
                        "end": {
                          "Line": 1,
                          "Column": 14,
                          "Offset": 13
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 13,
                              "Offset": 12
                            },
                            "end": {
                              "Line": 1,
                              "Column": 14,
                              "Offset": 13
                            },
                            "value": "\"",
                            "kind": "string"
                          }
                        ]
                      }
                    ]
                  }
                ]
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 1,
                  "Column": 15,
                  "Offset": 14
                },
                "end": {
                  "Line": 1,
                  "Column": 17,
                  "Offset": 16
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 1,
                      "Column": 15,
                      "Offset": 14
                    },
                    "end": {
                      "Line": 1,
                      "Column": 17,
                      "Offset": 16
                    },
                    "value": "if",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 1,
                  "Column": 18,
                  "Offset": 17
                },
                "end": {
                  "Line": 1,
                  "Column": 24,
                  "Offset": 23
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 18,
                      "Offset": 17
                    },
                    "end": {
                      "Line": 1,
                      "Column": 19,
                      "Offset": 18
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 19,
                      "Offset": 18
                    },
                    "end": {
                      "Line": 1,
                      "Column": 24,
                      "Offset": 23
                    },
                    "text": "debug"
                  }
                ]
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 1,
          "Column": 24,
          "Offset": 23
        },
        "end": {
          "Line": 1,
          "Column": 25,
          "Offset": 24
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 25
}
```
