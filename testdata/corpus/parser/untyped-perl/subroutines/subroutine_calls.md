---
category: untyped-perl
subcategory: subroutines
tags:
    - calls
    - arguments
    - parentheses
    - ampersand
    - nested
---

# Subroutine Calls

Test various subroutine call patterns and argument passing

```perl
my $result = simple_sub();
my $sum = with_params(10, 20);
my $doubled = $code_ref->(5);

# Calls without parentheses
my $no_parens = simple_sub;
my $with_args = function_call 'arg1', 'arg2';

# Calls with complex arguments
my $complex = process($var, @array, %hash);
my $nested = outer(inner(42), another());

# Ampersand calls
my $amp_call = &function_name;
my $amp_with_args = &other_function(@args);
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $result = simple_sub();
my $sum = with_params(10, 20);
my $doubled = $code_ref->(5);

# Calls without parentheses
my $no_parens = simple_sub;
my $with_args = function_call 'arg1', 'arg2';

# Calls with complex arguments
my $complex = process($var, @array, %hash);
my $nested = outer(inner(42), another());

# Ampersand calls
my $amp_call = &function_name;
my $amp_with_args = &other_function(@args);
```

## Typed Perl Output

```perl
my $result = simple_sub();
my $sum = with_params(10, 20);
my $doubled = $code_ref->(5);

# Calls without parentheses
my $no_parens = simple_sub;
my $with_args = function_call 'arg1', 'arg2';

# Calls with complex arguments
my $complex = process($var, @array, %hash);
my $nested = outer(inner(42), another());

# Ampersand calls
my $amp_call = &function_name;
my $amp_with_args = &other_function(@args);
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
    var_decl
      variable
  token
  expression_stmt
    literal
  expression_statement
    var_decl
      variable
  token
  expression_statement
    var_decl
      variable
  token
  expression_stmt
    literal
  expression_statement
    var_decl
      variable
  token
  expression_statement
    var_decl
      variable
  token
  expression_stmt
    literal
  expression_statement
    var_decl
      variable
  token
  expression_statement
    var_decl
      variable
  token
```

## JSON Format

```json
{
  "path": "/tmp/subroutine_calls.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 15,
      "Column": 44,
      "Offset": 402
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
          "Column": 26,
          "Offset": 25
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
              "Line": 1,
              "Column": 26,
              "Offset": 25
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
                  "Line": 1,
                  "Column": 26,
                  "Offset": 25
                },
                "name": "result",
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
          "Line": 1,
          "Column": 26,
          "Offset": 25
        },
        "end": {
          "Line": 1,
          "Column": 27,
          "Offset": 26
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 27
        },
        "end": {
          "Line": 2,
          "Column": 30,
          "Offset": 56
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 27
            },
            "end": {
              "Line": 2,
              "Column": 30,
              "Offset": 56
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 27
                },
                "end": {
                  "Line": 2,
                  "Column": 30,
                  "Offset": 56
                },
                "name": "sum",
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
          "Line": 2,
          "Column": 30,
          "Offset": 56
        },
        "end": {
          "Line": 2,
          "Column": 31,
          "Offset": 57
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 58
        },
        "end": {
          "Line": 3,
          "Column": 29,
          "Offset": 86
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 58
            },
            "end": {
              "Line": 3,
              "Column": 29,
              "Offset": 86
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 58
                },
                "end": {
                  "Line": 3,
                  "Column": 29,
                  "Offset": 86
                },
                "name": "doubled",
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
          "Line": 3,
          "Column": 29,
          "Offset": 86
        },
        "end": {
          "Line": 3,
          "Column": 30,
          "Offset": 87
        },
        "text": ";"
      },
      {
        "type": "expression_stmt",
        "start": {
          "Line": 5,
          "Column": 1,
          "Offset": 89
        },
        "end": {
          "Line": 5,
          "Column": 28,
          "Offset": 116
        },
        "children": [
          {
            "type": "literal",
            "start": {
              "Line": 5,
              "Column": 1,
              "Offset": 89
            },
            "end": {
              "Line": 5,
              "Column": 28,
              "Offset": 116
            },
            "value": "# Calls without parentheses",
            "kind": "string"
          }
        ]
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 6,
          "Column": 1,
          "Offset": 117
        },
        "end": {
          "Line": 6,
          "Column": 27,
          "Offset": 143
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 6,
              "Column": 1,
              "Offset": 117
            },
            "end": {
              "Line": 6,
              "Column": 27,
              "Offset": 143
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 6,
                  "Column": 1,
                  "Offset": 117
                },
                "end": {
                  "Line": 6,
                  "Column": 27,
                  "Offset": 143
                },
                "name": "no_parens",
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
          "Line": 6,
          "Column": 27,
          "Offset": 143
        },
        "end": {
          "Line": 6,
          "Column": 28,
          "Offset": 144
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 7,
          "Column": 1,
          "Offset": 145
        },
        "end": {
          "Line": 7,
          "Column": 45,
          "Offset": 189
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 7,
              "Column": 1,
              "Offset": 145
            },
            "end": {
              "Line": 7,
              "Column": 45,
              "Offset": 189
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 7,
                  "Column": 1,
                  "Offset": 145
                },
                "end": {
                  "Line": 7,
                  "Column": 45,
                  "Offset": 189
                },
                "name": "with_args",
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
          "Column": 45,
          "Offset": 189
        },
        "end": {
          "Line": 7,
          "Column": 46,
          "Offset": 190
        },
        "text": ";"
      },
      {
        "type": "expression_stmt",
        "start": {
          "Line": 9,
          "Column": 1,
          "Offset": 192
        },
        "end": {
          "Line": 9,
          "Column": 31,
          "Offset": 222
        },
        "children": [
          {
            "type": "literal",
            "start": {
              "Line": 9,
              "Column": 1,
              "Offset": 192
            },
            "end": {
              "Line": 9,
              "Column": 31,
              "Offset": 222
            },
            "value": "# Calls with complex arguments",
            "kind": "string"
          }
        ]
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 10,
          "Column": 1,
          "Offset": 223
        },
        "end": {
          "Line": 10,
          "Column": 43,
          "Offset": 265
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 10,
              "Column": 1,
              "Offset": 223
            },
            "end": {
              "Line": 10,
              "Column": 43,
              "Offset": 265
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 10,
                  "Column": 1,
                  "Offset": 223
                },
                "end": {
                  "Line": 10,
                  "Column": 43,
                  "Offset": 265
                },
                "name": "complex",
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
          "Line": 10,
          "Column": 43,
          "Offset": 265
        },
        "end": {
          "Line": 10,
          "Column": 44,
          "Offset": 266
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 11,
          "Column": 1,
          "Offset": 267
        },
        "end": {
          "Line": 11,
          "Column": 41,
          "Offset": 307
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 11,
              "Column": 1,
              "Offset": 267
            },
            "end": {
              "Line": 11,
              "Column": 41,
              "Offset": 307
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 11,
                  "Column": 1,
                  "Offset": 267
                },
                "end": {
                  "Line": 11,
                  "Column": 41,
                  "Offset": 307
                },
                "name": "nested",
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
          "Line": 11,
          "Column": 41,
          "Offset": 307
        },
        "end": {
          "Line": 11,
          "Column": 42,
          "Offset": 308
        },
        "text": ";"
      },
      {
        "type": "expression_stmt",
        "start": {
          "Line": 13,
          "Column": 1,
          "Offset": 310
        },
        "end": {
          "Line": 13,
          "Column": 18,
          "Offset": 327
        },
        "children": [
          {
            "type": "literal",
            "start": {
              "Line": 13,
              "Column": 1,
              "Offset": 310
            },
            "end": {
              "Line": 13,
              "Column": 18,
              "Offset": 327
            },
            "value": "# Ampersand calls",
            "kind": "string"
          }
        ]
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 14,
          "Column": 1,
          "Offset": 328
        },
        "end": {
          "Line": 14,
          "Column": 30,
          "Offset": 357
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 14,
              "Column": 1,
              "Offset": 328
            },
            "end": {
              "Line": 14,
              "Column": 30,
              "Offset": 357
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 14,
                  "Column": 1,
                  "Offset": 328
                },
                "end": {
                  "Line": 14,
                  "Column": 30,
                  "Offset": 357
                },
                "name": "amp_call",
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
          "Line": 14,
          "Column": 30,
          "Offset": 357
        },
        "end": {
          "Line": 14,
          "Column": 31,
          "Offset": 358
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 15,
          "Column": 1,
          "Offset": 359
        },
        "end": {
          "Line": 15,
          "Column": 43,
          "Offset": 401
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 15,
              "Column": 1,
              "Offset": 359
            },
            "end": {
              "Line": 15,
              "Column": 43,
              "Offset": 401
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 15,
                  "Column": 1,
                  "Offset": 359
                },
                "end": {
                  "Line": 15,
                  "Column": 43,
                  "Offset": 401
                },
                "name": "amp_with_args",
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
          "Line": 15,
          "Column": 43,
          "Offset": 401
        },
        "end": {
          "Line": 15,
          "Column": 44,
          "Offset": 402
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 402
}
```
