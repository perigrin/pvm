---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - foreach
    - default_variable
    - underscore
---

# Foreach No Variable

Foreach loop using default variable $_

```perl
foreach (@items) {
    process($_);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
foreach (@items) {
    process($_);
}
```

## Typed Perl Output

```perl
foreach (@items) {
    process($_);
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
  (token "(")
  (array
    (literal "@")
    (token "items"))
  (token ")")
  (block_stmt
    (token "{")
    (literal "process($_)")
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
    "Offset": 37
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
        "Column": 10,
        "Offset": 9
      },
      "text": "("
    },
    {
      "type": "array",
      "start": {
        "Line": 1,
        "Column": 10,
        "Offset": 9
      },
      "end": {
        "Line": 1,
        "Column": 16,
        "Offset": 15
      },
      "children": [
        {
          "type": "expression_stmt",
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
          "children": [
            {
              "type": "literal",
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
              "value": "@",
              "kind": "string"
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
            "Column": 16,
            "Offset": 15
          },
          "text": "items"
        }
      ]
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
        "Column": 17,
        "Offset": 16
      },
      "text": ")"
    },
    {
      "type": "block_stmt",
      "start": {
        "Line": 1,
        "Column": 18,
        "Offset": 17
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
            "Column": 18,
            "Offset": 17
          },
          "end": {
            "Line": 1,
            "Column": 19,
            "Offset": 18
          },
          "text": "{"
        },
        {
          "type": "expression_stmt",
          "start": {
            "Line": 2,
            "Column": 5,
            "Offset": 23
          },
          "end": {
            "Line": 2,
            "Column": 16,
            "Offset": 34
          },
          "children": [
            {
              "type": "literal",
              "start": {
                "Line": 2,
                "Column": 5,
                "Offset": 23
              },
              "end": {
                "Line": 2,
                "Column": 16,
                "Offset": 34
              },
              "value": "process($_)",
              "kind": "string"
            }
          ]
        },
        {
          "type": "token",
          "start": {
            "Line": 2,
            "Column": 16,
            "Offset": 34
          },
          "end": {
            "Line": 2,
            "Column": 17,
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
    }
  ]
}
```
