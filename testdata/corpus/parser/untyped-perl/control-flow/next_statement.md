---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - next
    - loop_control
---

# Next Statement

Next statement to skip to next iteration

```perl
foreach my $item (@list) {
    next if skip_condition($item);
    process($item);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
foreach my $item (@list) {
    next if skip_condition($item);
    process($item);
}
```

## Typed Perl Output

```perl
foreach my $item (@list) {
    next if skip_condition($item);
    process($item);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
(for_statement
  (literal "foreach")
  (token "my")
  (scalar
    (token "$")
    (token "item"))
  (token "(")
  (array
    (literal "@")
    (token "list"))
  (token ")")
  (block_stmt
    (token "{")
    (literal "next if skip_condition($item)")
    (token ";")
    (literal "process($item)")
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
    "Line": 4,
    "Column": 2,
    "Offset": 83
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
        "Column": 8,
        "Offset": 7
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
            "Column": 8,
            "Offset": 7
          },
          "value": "foreach",
          "kind": "string"
        }
      ]
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
        "Column": 11,
        "Offset": 10
      },
      "text": "my"
    },
    {
      "type": "scalar",
      "start": {
        "Line": 1,
        "Column": 12,
        "Offset": 11
      },
      "end": {
        "Line": 1,
        "Column": 17,
        "Offset": 16
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
            "Column": 13,
            "Offset": 12
          },
          "text": "$"
        },
        {
          "type": "token",
          "start": {
            "Line": 1,
            "Column": 13,
            "Offset": 12
          },
          "end": {
            "Line": 1,
            "Column": 17,
            "Offset": 16
          },
          "text": "item"
        }
      ]
    },
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
      "text": "("
    },
    {
      "type": "array",
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
              "value": "@",
              "kind": "string"
            }
          ]
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
          "text": "list"
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
      "text": ")"
    },
    {
      "type": "block_stmt",
      "start": {
        "Line": 1,
        "Column": 26,
        "Offset": 25
      },
      "end": {
        "Line": 4,
        "Column": 2,
        "Offset": 83
      },
      "children": [
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
          "text": "{"
        },
        {
          "type": "expression_stmt",
          "start": {
            "Line": 2,
            "Column": 5,
            "Offset": 31
          },
          "end": {
            "Line": 2,
            "Column": 34,
            "Offset": 60
          },
          "children": [
            {
              "type": "literal",
              "start": {
                "Line": 2,
                "Column": 5,
                "Offset": 31
              },
              "end": {
                "Line": 2,
                "Column": 34,
                "Offset": 60
              },
              "value": "next if skip_condition($item)",
              "kind": "string"
            }
          ]
        },
        {
          "type": "token",
          "start": {
            "Line": 2,
            "Column": 34,
            "Offset": 60
          },
          "end": {
            "Line": 2,
            "Column": 35,
            "Offset": 61
          },
          "text": ";"
        },
        {
          "type": "expression_stmt",
          "start": {
            "Line": 3,
            "Column": 5,
            "Offset": 66
          },
          "end": {
            "Line": 3,
            "Column": 19,
            "Offset": 80
          },
          "children": [
            {
              "type": "literal",
              "start": {
                "Line": 3,
                "Column": 5,
                "Offset": 66
              },
              "end": {
                "Line": 3,
                "Column": 19,
                "Offset": 80
              },
              "value": "process($item)",
              "kind": "string"
            }
          ]
        },
        {
          "type": "token",
          "start": {
            "Line": 3,
            "Column": 19,
            "Offset": 80
          },
          "end": {
            "Line": 3,
            "Column": 20,
            "Offset": 81
          },
          "text": ";"
        },
        {
          "type": "token",
          "start": {
            "Line": 4,
            "Column": 1,
            "Offset": 82
          },
          "end": {
            "Line": 4,
            "Column": 2,
            "Offset": 83
          },
          "text": "}"
        }
      ]
    }
  ]
}
```
