---
category: untyped-perl
subcategory: subroutines
tags:
    - anonymous
    - code_references
    - closures
    - parameters
---

# Anonymous Subroutines

Test anonymous subroutine definitions and code references

```perl
my $code_ref = sub {
    my $param = shift;
    return $param * 2;
};

my $closure = sub {
    my $x = shift;
    return sub { return $x + shift };
};

my $anon_with_params = sub {
    my ($a, $b) = @_;
    return $a . $b;
};
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $code_ref = sub {
    my $param = shift;
    return $param * 2;
};

my $closure = sub {
    my $x = shift;
    return sub { return $x + shift };
};

my $anon_with_params = sub {
    my ($a, $b) = @_;
    return $a . $b;
};
```

## Typed Perl Output

```perl
my $code_ref = sub {
    my $param = shift;
    return $param * 2;
};

my $closure = sub {
    my $x = shift;
    return sub { return $x + shift };
};

my $anon_with_params = sub {
    my ($a, $b) = @_;
    return $a . $b;
};
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
```

## JSON Format

```json
{
  "path": "/tmp/anonymous_subroutines.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 14,
      "Column": 3,
      "Offset": 225
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
          "Line": 4,
          "Column": 2,
          "Offset": 68
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
              "Line": 4,
              "Column": 2,
              "Offset": 68
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
                  "Line": 4,
                  "Column": 2,
                  "Offset": 68
                },
                "name": "code_ref",
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
          "Line": 4,
          "Column": 2,
          "Offset": 68
        },
        "end": {
          "Line": 4,
          "Column": 3,
          "Offset": 69
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 6,
          "Column": 1,
          "Offset": 71
        },
        "end": {
          "Line": 9,
          "Column": 2,
          "Offset": 149
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 6,
              "Column": 1,
              "Offset": 71
            },
            "end": {
              "Line": 9,
              "Column": 2,
              "Offset": 149
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 6,
                  "Column": 1,
                  "Offset": 71
                },
                "end": {
                  "Line": 9,
                  "Column": 2,
                  "Offset": 149
                },
                "name": "closure",
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
          "Line": 9,
          "Column": 2,
          "Offset": 149
        },
        "end": {
          "Line": 9,
          "Column": 3,
          "Offset": 150
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 11,
          "Column": 1,
          "Offset": 152
        },
        "end": {
          "Line": 14,
          "Column": 2,
          "Offset": 224
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 11,
              "Column": 1,
              "Offset": 152
            },
            "end": {
              "Line": 14,
              "Column": 2,
              "Offset": 224
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 11,
                  "Column": 1,
                  "Offset": 152
                },
                "end": {
                  "Line": 14,
                  "Column": 2,
                  "Offset": 224
                },
                "name": "anon_with_params",
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
          "Column": 2,
          "Offset": 224
        },
        "end": {
          "Line": 14,
          "Column": 3,
          "Offset": 225
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 225
}
```
