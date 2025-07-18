---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - for
    - range
---

# For Loop Range

For loop with range operator

```perl
for my $i (0..$max) {
    handle($i);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
for my $i (0..$max) {
    handle($i);
}
```

## Typed Perl Output

```perl
for my $i (0..$max) {
    handle($i);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
(for_statement
  (literal "for")
  (token "my")
  (scalar
    (token "$")
    (token "i"))
  (token "(")
  (binary_expression
    (token "0.")
    (literal ".")
    (scalar
      (token "$")
      (token "max")))
  (token ")")
  (block_stmt
    (token "{")
    (literal "handle($i)")
    (token ";")
    (token "}")))
```

## JSON AST

```json
{
  "type": "for_statement",
  "start": {
    "Line": 1,
    "Column": 1,
    "Offset": 0
  },
  "end": {
    "Line": 3,
    "Column": 2,
    "Offset": 39
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
        "Column": 4,
        "Offset": 3
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
            "Column": 4,
            "Offset": 3
          },
          "value": "for",
          "kind": "string"
        }
      ]
    },
    {
      "type": "token",
      "start": {
        "Line": 1,
        "Column": 5,
        "Offset": 4
      },
      "end": {
        "Line": 1,
        "Column": 7,
        "Offset": 6
      },
      "text": "my"
    },
    {
      "type": "scalar",
      "start": {
        "Line": 1,
        "Column": 8,
        "Offset": 7
      },
      "end": {
        "Line": 1,
        "Column": 10,
        "Offset": 9
      },
      "children": [
        {
          "type": "token",
          "start": {
            "Line": 1,
            "Column": 8,
            "Offset": 7
          },
          "end": {
            "Line": 1,
            "Column": 9,
            "Offset": 8
          },
          "text": "$"
        },
        {
          "type": "token",
          "start": {
            "Line": 1,
            "Column": 9,
            "Offset": 8
          },
          "end": {
            "Line": 1,
            "Column": 10,
            "Offset": 9
          },
          "text": "i"
        }
      ]
    },
    {
      "type": "token",
      "start": {
        "Line": 1,
        "Column": 11,
        "Offset": 10
      },
      "end": {
        "Line": 1,
        "Column": 12,
        "Offset": 11
      },
      "text": "("
    },
    {
      "type": "binary_expression",
      "start": {
        "Line": 1,
        "Column": 12,
        "Offset": 11
      },
      "end": {
        "Line": 1,
        "Column": 19,
        "Offset": 18
      },
      "children": [
        {
          "type": "token",
          "start": {
            "Line": 1,
            "Column": 12,
            "Offset": 11
          },
          "end": {
            "Line": 1,
            "Column": 14,
            "Offset": 13
          },
          "text": "0."
        },
        {
          "type": "expression_stmt",
          "start": {
            "Line": 1,
            "Column": 14,
            "Offset": 13
          },
          "end": {
            "Line": 1,
            "Column": 15,
            "Offset": 14
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
                "Column": 15,
                "Offset": 14
              },
              "value": ".",
              "kind": "string"
            }
          ]
        },
        {
          "type": "scalar",
          "start": {
            "Line": 1,
            "Column": 15,
            "Offset": 14
          },
          "end": {
            "Line": 1,
            "Column": 19,
            "Offset": 18
          },
          "children": [
            {
              "type": "token",
              "start": {
                "Line": 1,
                "Column": 15,
                "Offset": 14
              },
              "end": {
                "Line": 1,
                "Column": 16,
                "Offset": 15
              },
              "text": "$"
            },
            {
              "type": "token",
              "start": {
                "Line": 1,
                "Column": 16,
                "Offset": 15
              },
              "end": {
                "Line": 1,
                "Column": 19,
                "Offset": 18
              },
              "text": "max"
            }
          ]
        }
      ]
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
        "Column": 20,
        "Offset": 19
      },
      "text": ")"
    },
    {
      "type": "block_stmt",
      "start": {
        "Line": 1,
        "Column": 21,
        "Offset": 20
      },
      "end": {
        "Line": 3,
        "Column": 2,
        "Offset": 39
      },
      "children": [
        {
          "type": "token",
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
          "text": "{"
        },
        {
          "type": "expression_stmt",
          "start": {
            "Line": 2,
            "Column": 5,
            "Offset": 26
          },
          "end": {
            "Line": 2,
            "Column": 15,
            "Offset": 36
          },
          "children": [
            {
              "type": "literal",
              "start": {
                "Line": 2,
                "Column": 5,
                "Offset": 26
              },
              "end": {
                "Line": 2,
                "Column": 15,
                "Offset": 36
              },
              "value": "handle($i)",
              "kind": "string"
            }
          ]
        },
        {
          "type": "token",
          "start": {
            "Line": 2,
            "Column": 15,
            "Offset": 36
          },
          "end": {
            "Line": 2,
            "Column": 16,
            "Offset": 37
          },
          "text": ";"
        },
        {
          "type": "token",
          "start": {
            "Line": 3,
            "Column": 1,
            "Offset": 38
          },
          "end": {
            "Line": 3,
            "Column": 2,
            "Offset": 39
          },
          "text": "}"
        }
      ]
    }
  ]
}
```
