---
category: untyped-perl
subcategory: variables
tags:
    - hashes
    - variables
    - assignments
---

# Hash Operations

Basic hash operations and assignments

```perl
my %person = (
    name => "John",
    age => 30,
    city => "Boston"
);

my $name = $person{name};
$person{age} = 31;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my %person = (
    name => "John",
    age => 30,
    city => "Boston"
);

my $name = $person{name};
$person{age} = 31;
```

## Typed Perl Output

```perl
my %person = (
    name => "John",
    age => 30,
    city => "Boston"
);

my $name = $person{name};
$person{age} = 31;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text Format

```
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
    assignment_expression
      hash_element_expression
        container_variable
          token
          token
        token
        expression_stmt
          literal
        token
      token
      token
  token
```

## JSON Format

```json
{
  "path": "/tmp/hash-operations.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 8,
      "Column": 19,
      "Offset": 119
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
          "Line": 5,
          "Column": 2,
          "Offset": 72
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 5,
              "Column": 2,
              "Offset": 72
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 1,
                  "Column": 1,
                  "Offset": 0
                },
                "end": {
                  "Line": 5,
                  "Column": 2,
                  "Offset": 72
                },
                "name": "person",
                "sigil": "%"
              }
            ],
            "decl_type": "my"
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 5,
          "Column": 2,
          "Offset": 72
        },
        "end": {
          "Line": 5,
          "Column": 3,
          "Offset": 73
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 7,
          "Column": 1,
          "Offset": 75
        },
        "end": {
          "Line": 7,
          "Column": 25,
          "Offset": 99
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 7,
              "Column": 1,
              "Offset": 75
            },
            "end": {
              "Line": 7,
              "Column": 25,
              "Offset": 99
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 7,
                  "Column": 1,
                  "Offset": 75
                },
                "end": {
                  "Line": 7,
                  "Column": 25,
                  "Offset": 99
                },
                "name": "name",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 7,
          "Column": 25,
          "Offset": 99
        },
        "end": {
          "Line": 7,
          "Column": 26,
          "Offset": 100
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 8,
          "Column": 1,
          "Offset": 101
        },
        "end": {
          "Line": 8,
          "Column": 18,
          "Offset": 118
        },
        "children": [
          {
            "type": "assignment_expression",
            "start": {
              "Line": 8,
              "Column": 1,
              "Offset": 101
            },
            "end": {
              "Line": 8,
              "Column": 18,
              "Offset": 118
            },
            "children": [
              {
                "type": "hash_element_expression",
                "start": {
                  "Line": 8,
                  "Column": 1,
                  "Offset": 101
                },
                "end": {
                  "Line": 8,
                  "Column": 13,
                  "Offset": 113
                },
                "children": [
                  {
                    "type": "container_variable",
                    "start": {
                      "Line": 8,
                      "Column": 1,
                      "Offset": 101
                    },
                    "end": {
                      "Line": 8,
                      "Column": 8,
                      "Offset": 108
                    },
                    "children": [
                      {
                        "type": "token",
                        "start": {
                          "Line": 8,
                          "Column": 1,
                          "Offset": 101
                        },
                        "end": {
                          "Line": 8,
                          "Column": 2,
                          "Offset": 102
                        },
                        "text": "$"
                      },
                      {
                        "type": "token",
                        "start": {
                          "Line": 8,
                          "Column": 2,
                          "Offset": 102
                        },
                        "end": {
                          "Line": 8,
                          "Column": 8,
                          "Offset": 108
                        },
                        "text": "person"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 8,
                      "Column": 8,
                      "Offset": 108
                    },
                    "end": {
                      "Line": 8,
                      "Column": 9,
                      "Offset": 109
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 8,
                      "Column": 9,
                      "Offset": 109
                    },
                    "end": {
                      "Line": 8,
                      "Column": 12,
                      "Offset": 112
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 8,
                          "Column": 9,
                          "Offset": 109
                        },
                        "end": {
                          "Line": 8,
                          "Column": 12,
                          "Offset": 112
                        },
                        "value": "age",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 8,
                      "Column": 12,
                      "Offset": 112
                    },
                    "end": {
                      "Line": 8,
                      "Column": 13,
                      "Offset": 113
                    },
                    "text": "}"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 8,
                  "Column": 14,
                  "Offset": 114
                },
                "end": {
                  "Line": 8,
                  "Column": 15,
                  "Offset": 115
                },
                "text": "="
              },
              {
                "type": "token",
                "start": {
                  "Line": 8,
                  "Column": 16,
                  "Offset": 116
                },
                "end": {
                  "Line": 8,
                  "Column": 18,
                  "Offset": 118
                },
                "text": "31"
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 8,
          "Column": 18,
          "Offset": 118
        },
        "end": {
          "Line": 8,
          "Column": 19,
          "Offset": 119
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 119
}
```
