---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - negation-types
    - complex-combinations
    - intersection-types
---

# Negation Combinations

Negation types combined with parameterized and intersection types

```perl
my ArrayRef[!Undef] @non_undef_array;
my HashRef[Str, !Empty] %non_empty_values;
my Optional[!Null&!Undef] $definitely_defined;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @non_undef_array;
my %non_empty_values;
my $definitely_defined;
```

## Typed Perl Output

```perl
my ArrayRef[!Undef] @non_undef_array;
my HashRef[Str, !Empty] %non_empty_values;
my Optional[!Null&!Undef] $definitely_defined;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
AST {
  Path: /tmp/negation-combinations.pl
  Source length: 127 characters
  Type Annotations:
    VarAnnotation: @non_undef_array :: ArrayRef[!Undef] at 1:1
    VarAnnotation: %non_empty_values :: HashRef[Str, !Empty] at 2:1
    VarAnnotation: $definitely_defined :: Optional[!Null&!Undef] at 3:1
  Root: source_file
  Tree Structure:
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
                negation_type
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
                expression_stmt
                  literal
              expression_stmt
                literal
              type_expression
                negation_type
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
                intersection_type
                  type_expression
                    negation_type
                      expression_stmt
                        literal
                      type_expression
                        expression_stmt
                          literal
                  expression_stmt
                    literal
                  type_expression
                    negation_type
                      expression_stmt
                        literal
                      type_expression
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
  "path": "/tmp/negation-combinations.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 3,
      "Column": 47,
      "Offset": 127
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
          "Column": 37,
          "Offset": 36
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 1,
              "Column": 37,
              "Offset": 36
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 1,
                  "Offset": 0
                },
                "end": {
                  "Line": 1,
                  "Column": 3,
                  "Offset": 2
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 1,
                  "Column": 4,
                  "Offset": 3
                },
                "end": {
                  "Line": 1,
                  "Column": 20,
                  "Offset": 19
                },
                "children": [
                  {
                    "type": "parameterized_type",
                    "start": {
                      "Line": 1,
                      "Column": 4,
                      "Offset": 3
                    },
                    "end": {
                      "Line": 1,
                      "Column": 20,
                      "Offset": 19
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 4,
                          "Offset": 3
                        },
                        "end": {
                          "Line": 1,
                          "Column": 12,
                          "Offset": 11
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 4,
                              "Offset": 3
                            },
                            "end": {
                              "Line": 1,
                              "Column": 12,
                              "Offset": 11
                            },
                            "value": "ArrayRef",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 12,
                          "Offset": 11
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
                              "Column": 12,
                              "Offset": 11
                            },
                            "end": {
                              "Line": 1,
                              "Column": 13,
                              "Offset": 12
                            },
                            "value": "[",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "type_parameter_list",
                        "start": {
                          "Line": 1,
                          "Column": 13,
                          "Offset": 12
                        },
                        "end": {
                          "Line": 1,
                          "Column": 19,
                          "Offset": 18
                        },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 1,
                              "Column": 13,
                              "Offset": 12
                            },
                            "end": {
                              "Line": 1,
                              "Column": 19,
                              "Offset": 18
                            },
                            "children": [
                              {
                                "type": "negation_type",
                                "start": {
                                  "Line": 1,
                                  "Column": 13,
                                  "Offset": 12
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 19,
                                  "Offset": 18
                                },
                                "children": [
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
                                        "value": "!",
                                        "kind": "string"
                                      }
                                    ]
                                  },
                                  {
                                    "type": "type_expression",
                                    "start": {
                                      "Line": 1,
                                      "Column": 14,
                                      "Offset": 13
                                    },
                                    "end": {
                                      "Line": 1,
                                      "Column": 19,
                                      "Offset": 18
                                    },
                                    "children": [
                                      {
                                        "type": "expression_stmt",
                                        "start": {
                                          "Line": 1,
                                          "Column": 14,
                                          "Offset": 13
                                        },
                                        "end": {
                                          "Line": 1,
                                          "Column": 19,
                                          "Offset": 18
                                        },
                                        "children": [
                                          {
                                            "type": "literal",
                                            "start": {
                                              "Line": 1,
                                              "Column": 14,
                                              "Offset": 13
                                            },
                                            "end": {
                                              "Line": 1,
                                              "Column": 19,
                                              "Offset": 18
                                            },
                                            "value": "Undef",
                                            "kind": "string"
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
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 19,
                          "Offset": 18
                        },
                        "end": {
                          "Line": 1,
                          "Column": 20,
                          "Offset": 19
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 19,
                              "Offset": 18
                            },
                            "end": {
                              "Line": 1,
                              "Column": 20,
                              "Offset": 19
                            },
                            "value": "]",
                            "kind": "string"
                          }
                        ]
                      }
                    ]
                  }
                ]
              },
              {
                "type": "array",
                "start": {
                  "Line": 1,
                  "Column": 21,
                  "Offset": 20
                },
                "end": {
                  "Line": 1,
                  "Column": 37,
                  "Offset": 36
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 1,
                      "Column": 21,
                      "Offset": 20
                    },
                    "end": {
                      "Line": 1,
                      "Column": 22,
                      "Offset": 21
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 1,
                          "Column": 21,
                          "Offset": 20
                        },
                        "end": {
                          "Line": 1,
                          "Column": 22,
                          "Offset": 21
                        },
                        "value": "@",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 22,
                      "Offset": 21
                    },
                    "end": {
                      "Line": 1,
                      "Column": 37,
                      "Offset": 36
                    },
                    "text": "non_undef_array"
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
          "Column": 37,
          "Offset": 36
        },
        "end": {
          "Line": 1,
          "Column": 38,
          "Offset": 37
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 38
        },
        "end": {
          "Line": 2,
          "Column": 42,
          "Offset": 79
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 38
            },
            "end": {
              "Line": 2,
              "Column": 42,
              "Offset": 79
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 38
                },
                "end": {
                  "Line": 2,
                  "Column": 3,
                  "Offset": 40
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 2,
                  "Column": 4,
                  "Offset": 41
                },
                "end": {
                  "Line": 2,
                  "Column": 24,
                  "Offset": 61
                },
                "children": [
                  {
                    "type": "parameterized_type",
                    "start": {
                      "Line": 2,
                      "Column": 4,
                      "Offset": 41
                    },
                    "end": {
                      "Line": 2,
                      "Column": 24,
                      "Offset": 61
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 2,
                          "Column": 4,
                          "Offset": 41
                        },
                        "end": {
                          "Line": 2,
                          "Column": 11,
                          "Offset": 48
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 4,
                              "Offset": 41
                            },
                            "end": {
                              "Line": 2,
                              "Column": 11,
                              "Offset": 48
                            },
                            "value": "HashRef",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 2,
                          "Column": 11,
                          "Offset": 48
                        },
                        "end": {
                          "Line": 2,
                          "Column": 12,
                          "Offset": 49
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 11,
                              "Offset": 48
                            },
                            "end": {
                              "Line": 2,
                              "Column": 12,
                              "Offset": 49
                            },
                            "value": "[",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "type_parameter_list",
                        "start": {
                          "Line": 2,
                          "Column": 12,
                          "Offset": 49
                        },
                        "end": {
                          "Line": 2,
                          "Column": 23,
                          "Offset": 60
                        },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 2,
                              "Column": 12,
                              "Offset": 49
                            },
                            "end": {
                              "Line": 2,
                              "Column": 15,
                              "Offset": 52
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": {
                                  "Line": 2,
                                  "Column": 12,
                                  "Offset": 49
                                },
                                "end": {
                                  "Line": 2,
                                  "Column": 15,
                                  "Offset": 52
                                },
                                "children": [
                                  {
                                    "type": "literal",
                                    "start": {
                                      "Line": 2,
                                      "Column": 12,
                                      "Offset": 49
                                    },
                                    "end": {
                                      "Line": 2,
                                      "Column": 15,
                                      "Offset": 52
                                    },
                                    "value": "Str",
                                    "kind": "string"
                                  }
                                ]
                              }
                            ]
                          },
                          {
                            "type": "expression_stmt",
                            "start": {
                              "Line": 2,
                              "Column": 15,
                              "Offset": 52
                            },
                            "end": {
                              "Line": 2,
                              "Column": 16,
                              "Offset": 53
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 2,
                                  "Column": 15,
                                  "Offset": 52
                                },
                                "end": {
                                  "Line": 2,
                                  "Column": 16,
                                  "Offset": 53
                                },
                                "value": ",",
                                "kind": "string"
                              }
                            ]
                          },
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 2,
                              "Column": 17,
                              "Offset": 54
                            },
                            "end": {
                              "Line": 2,
                              "Column": 23,
                              "Offset": 60
                            },
                            "children": [
                              {
                                "type": "negation_type",
                                "start": {
                                  "Line": 2,
                                  "Column": 17,
                                  "Offset": 54
                                },
                                "end": {
                                  "Line": 2,
                                  "Column": 23,
                                  "Offset": 60
                                },
                                "children": [
                                  {
                                    "type": "expression_stmt",
                                    "start": {
                                      "Line": 2,
                                      "Column": 17,
                                      "Offset": 54
                                    },
                                    "end": {
                                      "Line": 2,
                                      "Column": 18,
                                      "Offset": 55
                                    },
                                    "children": [
                                      {
                                        "type": "literal",
                                        "start": {
                                          "Line": 2,
                                          "Column": 17,
                                          "Offset": 54
                                        },
                                        "end": {
                                          "Line": 2,
                                          "Column": 18,
                                          "Offset": 55
                                        },
                                        "value": "!",
                                        "kind": "string"
                                      }
                                    ]
                                  },
                                  {
                                    "type": "type_expression",
                                    "start": {
                                      "Line": 2,
                                      "Column": 18,
                                      "Offset": 55
                                    },
                                    "end": {
                                      "Line": 2,
                                      "Column": 23,
                                      "Offset": 60
                                    },
                                    "children": [
                                      {
                                        "type": "expression_stmt",
                                        "start": {
                                          "Line": 2,
                                          "Column": 18,
                                          "Offset": 55
                                        },
                                        "end": {
                                          "Line": 2,
                                          "Column": 23,
                                          "Offset": 60
                                        },
                                        "children": [
                                          {
                                            "type": "literal",
                                            "start": {
                                              "Line": 2,
                                              "Column": 18,
                                              "Offset": 55
                                            },
                                            "end": {
                                              "Line": 2,
                                              "Column": 23,
                                              "Offset": 60
                                            },
                                            "value": "Empty",
                                            "kind": "string"
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
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 2,
                          "Column": 23,
                          "Offset": 60
                        },
                        "end": {
                          "Line": 2,
                          "Column": 24,
                          "Offset": 61
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 23,
                              "Offset": 60
                            },
                            "end": {
                              "Line": 2,
                              "Column": 24,
                              "Offset": 61
                            },
                            "value": "]",
                            "kind": "string"
                          }
                        ]
                      }
                    ]
                  }
                ]
              },
              {
                "type": "hash",
                "start": {
                  "Line": 2,
                  "Column": 25,
                  "Offset": 62
                },
                "end": {
                  "Line": 2,
                  "Column": 42,
                  "Offset": 79
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 2,
                      "Column": 25,
                      "Offset": 62
                    },
                    "end": {
                      "Line": 2,
                      "Column": 26,
                      "Offset": 63
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 2,
                          "Column": 25,
                          "Offset": 62
                        },
                        "end": {
                          "Line": 2,
                          "Column": 26,
                          "Offset": 63
                        },
                        "value": "%",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 26,
                      "Offset": 63
                    },
                    "end": {
                      "Line": 2,
                      "Column": 42,
                      "Offset": 79
                    },
                    "text": "non_empty_values"
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
          "Line": 2,
          "Column": 42,
          "Offset": 79
        },
        "end": {
          "Line": 2,
          "Column": 43,
          "Offset": 80
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 81
        },
        "end": {
          "Line": 3,
          "Column": 46,
          "Offset": 126
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 81
            },
            "end": {
              "Line": 3,
              "Column": 46,
              "Offset": 126
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 81
                },
                "end": {
                  "Line": 3,
                  "Column": 3,
                  "Offset": 83
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 3,
                  "Column": 4,
                  "Offset": 84
                },
                "end": {
                  "Line": 3,
                  "Column": 27,
                  "Offset": 107
                },
                "children": [
                  {
                    "type": "parameterized_type",
                    "start": {
                      "Line": 3,
                      "Column": 4,
                      "Offset": 84
                    },
                    "end": {
                      "Line": 3,
                      "Column": 27,
                      "Offset": 107
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 3,
                          "Column": 4,
                          "Offset": 84
                        },
                        "end": {
                          "Line": 3,
                          "Column": 12,
                          "Offset": 92
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 4,
                              "Offset": 84
                            },
                            "end": {
                              "Line": 3,
                              "Column": 12,
                              "Offset": 92
                            },
                            "value": "Optional",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 3,
                          "Column": 12,
                          "Offset": 92
                        },
                        "end": {
                          "Line": 3,
                          "Column": 13,
                          "Offset": 93
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 12,
                              "Offset": 92
                            },
                            "end": {
                              "Line": 3,
                              "Column": 13,
                              "Offset": 93
                            },
                            "value": "[",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "type_parameter_list",
                        "start": {
                          "Line": 3,
                          "Column": 13,
                          "Offset": 93
                        },
                        "end": {
                          "Line": 3,
                          "Column": 26,
                          "Offset": 106
                        },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 3,
                              "Column": 13,
                              "Offset": 93
                            },
                            "end": {
                              "Line": 3,
                              "Column": 26,
                              "Offset": 106
                            },
                            "children": [
                              {
                                "type": "intersection_type",
                                "start": {
                                  "Line": 3,
                                  "Column": 13,
                                  "Offset": 93
                                },
                                "end": {
                                  "Line": 3,
                                  "Column": 26,
                                  "Offset": 106
                                },
                                "children": [
                                  {
                                    "type": "type_expression",
                                    "start": {
                                      "Line": 3,
                                      "Column": 13,
                                      "Offset": 93
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 18,
                                      "Offset": 98
                                    },
                                    "children": [
                                      {
                                        "type": "negation_type",
                                        "start": {
                                          "Line": 3,
                                          "Column": 13,
                                          "Offset": 93
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 18,
                                          "Offset": 98
                                        },
                                        "children": [
                                          {
                                            "type": "expression_stmt",
                                            "start": {
                                              "Line": 3,
                                              "Column": 13,
                                              "Offset": 93
                                            },
                                            "end": {
                                              "Line": 3,
                                              "Column": 14,
                                              "Offset": 94
                                            },
                                            "children": [
                                              {
                                                "type": "literal",
                                                "start": {
                                                  "Line": 3,
                                                  "Column": 13,
                                                  "Offset": 93
                                                },
                                                "end": {
                                                  "Line": 3,
                                                  "Column": 14,
                                                  "Offset": 94
                                                },
                                                "value": "!",
                                                "kind": "string"
                                              }
                                            ]
                                          },
                                          {
                                            "type": "type_expression",
                                            "start": {
                                              "Line": 3,
                                              "Column": 14,
                                              "Offset": 94
                                            },
                                            "end": {
                                              "Line": 3,
                                              "Column": 18,
                                              "Offset": 98
                                            },
                                            "children": [
                                              {
                                                "type": "expression_stmt",
                                                "start": {
                                                  "Line": 3,
                                                  "Column": 14,
                                                  "Offset": 94
                                                },
                                                "end": {
                                                  "Line": 3,
                                                  "Column": 18,
                                                  "Offset": 98
                                                },
                                                "children": [
                                                  {
                                                    "type": "literal",
                                                    "start": {
                                                      "Line": 3,
                                                      "Column": 14,
                                                      "Offset": 94
                                                    },
                                                    "end": {
                                                      "Line": 3,
                                                      "Column": 18,
                                                      "Offset": 98
                                                    },
                                                    "value": "Null",
                                                    "kind": "string"
                                                  }
                                                ]
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
                                      "Line": 3,
                                      "Column": 18,
                                      "Offset": 98
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 19,
                                      "Offset": 99
                                    },
                                    "children": [
                                      {
                                        "type": "literal",
                                        "start": {
                                          "Line": 3,
                                          "Column": 18,
                                          "Offset": 98
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 19,
                                          "Offset": 99
                                        },
                                        "value": "&",
                                        "kind": "string"
                                      }
                                    ]
                                  },
                                  {
                                    "type": "type_expression",
                                    "start": {
                                      "Line": 3,
                                      "Column": 19,
                                      "Offset": 99
                                    },
                                    "end": {
                                      "Line": 3,
                                      "Column": 26,
                                      "Offset": 106
                                    },
                                    "children": [
                                      {
                                        "type": "negation_type",
                                        "start": {
                                          "Line": 3,
                                          "Column": 19,
                                          "Offset": 99
                                        },
                                        "end": {
                                          "Line": 3,
                                          "Column": 26,
                                          "Offset": 106
                                        },
                                        "children": [
                                          {
                                            "type": "expression_stmt",
                                            "start": {
                                              "Line": 3,
                                              "Column": 19,
                                              "Offset": 99
                                            },
                                            "end": {
                                              "Line": 3,
                                              "Column": 20,
                                              "Offset": 100
                                            },
                                            "children": [
                                              {
                                                "type": "literal",
                                                "start": {
                                                  "Line": 3,
                                                  "Column": 19,
                                                  "Offset": 99
                                                },
                                                "end": {
                                                  "Line": 3,
                                                  "Column": 20,
                                                  "Offset": 100
                                                },
                                                "value": "!",
                                                "kind": "string"
                                              }
                                            ]
                                          },
                                          {
                                            "type": "type_expression",
                                            "start": {
                                              "Line": 3,
                                              "Column": 20,
                                              "Offset": 100
                                            },
                                            "end": {
                                              "Line": 3,
                                              "Column": 26,
                                              "Offset": 106
                                            },
                                            "children": [
                                              {
                                                "type": "expression_stmt",
                                                "start": {
                                                  "Line": 3,
                                                  "Column": 20,
                                                  "Offset": 100
                                                },
                                                "end": {
                                                  "Line": 3,
                                                  "Column": 26,
                                                  "Offset": 106
                                                },
                                                "children": [
                                                  {
                                                    "type": "literal",
                                                    "start": {
                                                      "Line": 3,
                                                      "Column": 20,
                                                      "Offset": 100
                                                    },
                                                    "end": {
                                                      "Line": 3,
                                                      "Column": 26,
                                                      "Offset": 106
                                                    },
                                                    "value": "Undef",
                                                    "kind": "string"
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
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 3,
                          "Column": 26,
                          "Offset": 106
                        },
                        "end": {
                          "Line": 3,
                          "Column": 27,
                          "Offset": 107
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 26,
                              "Offset": 106
                            },
                            "end": {
                              "Line": 3,
                              "Column": 27,
                              "Offset": 107
                            },
                            "value": "]",
                            "kind": "string"
                          }
                        ]
                      }
                    ]
                  }
                ]
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 3,
                  "Column": 28,
                  "Offset": 108
                },
                "end": {
                  "Line": 3,
                  "Column": 46,
                  "Offset": 126
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 28,
                      "Offset": 108
                    },
                    "end": {
                      "Line": 3,
                      "Column": 29,
                      "Offset": 109
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 29,
                      "Offset": 109
                    },
                    "end": {
                      "Line": 3,
                      "Column": 46,
                      "Offset": 126
                    },
                    "text": "definitely_defined"
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
          "Line": 3,
          "Column": 46,
          "Offset": 126
        },
        "end": {
          "Line": 3,
          "Column": 47,
          "Offset": 127
        },
        "text": ";"
      }
    ]
  }
}
```

# Expected Type Errors

(none)
