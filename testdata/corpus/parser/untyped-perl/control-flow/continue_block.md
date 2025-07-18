---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - while
    - continue
    - cleanup
---

# Continue Block

While loop with continue block

```perl
while ($condition) {
    process();
} continue {
    cleanup();
    update_condition();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
while ($condition) {
    process();
} continue {
    cleanup();
    update_condition();
}
```

## Typed Perl Output

```perl
while ($condition) {
    process();
} continue {
    cleanup();
    update_condition();
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
(loop_statement
  (literal "while")
  (token "(")
  (scalar
    (token "$")
    (token "condition"))
  (token ")")
  (block_stmt
    (token "{")
    (literal "process()")
    (token ";")
    (token "}"))
  (literal "continue")
  (block_stmt
    (token "{")
    (literal "cleanup()")
    (token ";")
    (literal "update_condition()")
    (token ";")
    (token "}")))
```

## JSON AST

```json
{
  "type": "loop_statement",
  "start": {
    "Line": 1,
    "Column": 1,
    "Offset": 0
  },
  "end": {
    "Line": 6,
    "Column": 2,
    "Offset": 89
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
          "value": "while",
          "kind": "string"
        }
      ]
    },
    {
      "type": "token",
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
      "text": "("
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
        "Column": 18,
        "Offset": 17
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
            "Column": 18,
            "Offset": 17
          },
          "text": "condition"
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
      "text": ")"
    },
    {
      "type": "block_stmt",
      "start": {
        "Line": 1,
        "Column": 20,
        "Offset": 19
      },
      "end": {
        "Line": 3,
        "Column": 2,
        "Offset": 37
      },
      "children": [
        {
          "type": "token",
          "start": {
            "Line": 1,
            "Column": 20,
            "Offset": 19
          },
          "end": {
            "Line": 1,
            "Column": 21,
            "Offset": 20
          },
          "text": "{"
        },
        {
          "type": "expression_stmt",
          "start": {
            "Line": 2,
            "Column": 5,
            "Offset": 25
          },
          "end": {
            "Line": 2,
            "Column": 14,
            "Offset": 34
          },
          "children": [
            {
              "type": "literal",
              "start": {
                "Line": 2,
                "Column": 5,
                "Offset": 25
              },
              "end": {
                "Line": 2,
                "Column": 14,
                "Offset": 34
              },
              "value": "process()",
              "kind": "string"
            }
          ]
        },
        {
          "type": "token",
          "start": {
            "Line": 2,
            "Column": 14,
            "Offset": 34
          },
          "end": {
            "Line": 2,
            "Column": 15,
            "Offset": 35
          },
          "text": ";"
        },
        {
          "type": "token",
          "start": {
            "Line": 3,
            "Column": 1,
            "Offset": 36
          },
          "end": {
            "Line": 3,
            "Column": 2,
            "Offset": 37
          },
          "text": "}"
        }
      ]
    },
    {
      "type": "expression_stmt",
      "start": {
        "Line": 3,
        "Column": 3,
        "Offset": 38
      },
      "end": {
        "Line": 3,
        "Column": 11,
        "Offset": 46
      },
      "children": [
        {
          "type": "literal",
          "start": {
            "Line": 3,
            "Column": 3,
            "Offset": 38
          },
          "end": {
            "Line": 3,
            "Column": 11,
            "Offset": 46
          },
          "value": "continue",
          "kind": "string"
        }
      ]
    },
    {
      "type": "block_stmt",
      "start": {
        "Line": 3,
        "Column": 12,
        "Offset": 47
      },
      "end": {
        "Line": 6,
        "Column": 2,
        "Offset": 89
      },
      "children": [
        {
          "type": "token",
          "start": {
            "Line": 3,
            "Column": 12,
            "Offset": 47
          },
          "end": {
            "Line": 3,
            "Column": 13,
            "Offset": 48
          },
          "text": "{"
        },
        {
          "type": "expression_stmt",
          "start": {
            "Line": 4,
            "Column": 5,
            "Offset": 53
          },
          "end": {
            "Line": 4,
            "Column": 14,
            "Offset": 62
          },
          "children": [
            {
              "type": "literal",
              "start": {
                "Line": 4,
                "Column": 5,
                "Offset": 53
              },
              "end": {
                "Line": 4,
                "Column": 14,
                "Offset": 62
              },
              "value": "cleanup()",
              "kind": "string"
            }
          ]
        },
        {
          "type": "token",
          "start": {
            "Line": 4,
            "Column": 14,
            "Offset": 62
          },
          "end": {
            "Line": 4,
            "Column": 15,
            "Offset": 63
          },
          "text": ";"
        },
        {
          "type": "expression_stmt",
          "start": {
            "Line": 5,
            "Column": 5,
            "Offset": 68
          },
          "end": {
            "Line": 5,
            "Column": 23,
            "Offset": 86
          },
          "children": [
            {
              "type": "literal",
              "start": {
                "Line": 5,
                "Column": 5,
                "Offset": 68
              },
              "end": {
                "Line": 5,
                "Column": 23,
                "Offset": 86
              },
              "value": "update_condition()",
              "kind": "string"
            }
          ]
        },
        {
          "type": "token",
          "start": {
            "Line": 5,
            "Column": 23,
            "Offset": 86
          },
          "end": {
            "Line": 5,
            "Column": 24,
            "Offset": 87
          },
          "text": ";"
        },
        {
          "type": "token",
          "start": {
            "Line": 6,
            "Column": 1,
            "Offset": 88
          },
          "end": {
            "Line": 6,
            "Column": 2,
            "Offset": 89
          },
          "text": "}"
        }
      ]
    }
  ]
}
```
