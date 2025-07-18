---
category: untyped-perl
subcategory: expressions
tags:
    - string
    - interpolation
    - variables
    - quoting
---

# Interpolated Strings

String interpolation in double quotes

```perl
$message = "Hello $name, your score is $score";
```

# Expected Compilation Outcomes

## Interpolated Strings

### Clean Perl Output

```perl
$message = "Hello $name, your score is $score";
```

### Typed Perl Output

```perl
$message = "Hello $name, your score is $score";
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```

# Expected AST

## Text Format

```
source_file
  expression_statement
    assignment_expression
      scalar
        token
        token
      token
      interpolated_string_literal
        expression_stmt
          literal
        string_content
          scalar
            token
            token
          scalar
            token
            token
        expression_stmt
          literal
  token
```

## JSON Format

```json
{
  "path": "/tmp/interpolated_strings.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 1,
      "Column": 48,
      "Offset": 47
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
          "Column": 47,
          "Offset": 46
        },
        "children": [
          {
            "type": "assignment_expression",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 1,
              "Column": 47,
              "Offset": 46
            },
            "children": [
              {
                "type": "scalar",
                "start": {
                  "Line": 1,
                  "Column": 1,
                  "Offset": 0
                },
                "end": {
                  "Line": 1,
                  "Column": 9,
                  "Offset": 8
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
                      "Column": 2,
                      "Offset": 1
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 2,
                      "Offset": 1
                    },
                    "end": {
                      "Line": 1,
                      "Column": 9,
                      "Offset": 8
                    },
                    "text": "message"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 10,
                  "Offset": 9
                },
                "end": {
                  "Line": 1,
                  "Column": 11,
                  "Offset": 10
                },
                "text": "="
              },
              {
                "type": "interpolated_string_literal",
                "start": {
                  "Line": 1,
                  "Column": 12,
                  "Offset": 11
                },
                "end": {
                  "Line": 1,
                  "Column": 47,
                  "Offset": 46
                },
                "children": [
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
                        "value": "\"",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "string_content",
                    "start": {
                      "Line": 1,
                      "Column": 13,
                      "Offset": 12
                    },
                    "end": {
                      "Line": 1,
                      "Column": 46,
                      "Offset": 45
                    },
                    "children": [
                      {
                        "type": "scalar",
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
                        "children": [
                          {
                            "type": "token",
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
                            "text": "$"
                          },
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 20,
                              "Offset": 19
                            },
                            "end": {
                              "Line": 1,
                              "Column": 24,
                              "Offset": 23
                            },
                            "text": "name"
                          }
                        ]
                      },
                      {
                        "type": "scalar",
                        "start": {
                          "Line": 1,
                          "Column": 40,
                          "Offset": 39
                        },
                        "end": {
                          "Line": 1,
                          "Column": 46,
                          "Offset": 45
                        },
                        "children": [
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 40,
                              "Offset": 39
                            },
                            "end": {
                              "Line": 1,
                              "Column": 41,
                              "Offset": 40
                            },
                            "text": "$"
                          },
                          {
                            "type": "token",
                            "start": {
                              "Line": 1,
                              "Column": 41,
                              "Offset": 40
                            },
                            "end": {
                              "Line": 1,
                              "Column": 46,
                              "Offset": 45
                            },
                            "text": "score"
                          }
                        ]
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 1,
                      "Column": 46,
                      "Offset": 45
                    },
                    "end": {
                      "Line": 1,
                      "Column": 47,
                      "Offset": 46
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 1,
                          "Column": 46,
                          "Offset": 45
                        },
                        "end": {
                          "Line": 1,
                          "Column": 47,
                          "Offset": 46
                        },
                        "value": "\"",
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
        "type": "token",
        "start": {
          "Line": 1,
          "Column": 47,
          "Offset": 46
        },
        "end": {
          "Line": 1,
          "Column": 48,
          "Offset": 47
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 47
}
```
