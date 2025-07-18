---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - given
    - when
    - array
    - qw
    - smartmatch
---

# Given When Arrays

Given-when with array reference matching

```perl
given ($day) {
    when ([qw(sat sun)]) { print "weekend"; }
    when ([qw(mon tue wed thu fri)]) { print "weekday"; }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
given ($day) {
    when ([qw(sat sun)]) { print "weekend"; }
    when ([qw(mon tue wed thu fri)]) { print "weekday"; }
}
```

## Typed Perl Output

```perl
given ($day) {
    when ([qw(sat sun)]) { print "weekend"; }
    when ([qw(mon tue wed thu fri)]) { print "weekday"; }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
(given_statement
  (literal "given")
  (token "(")
  (scalar
    (token "$")
    (token "day"))
  (token ")")
  (block_stmt
    (token "{")
    (literal "when ([qw(sat sun)]) { print \"weekend\"; }")
    (literal "when ([qw(mon tue wed thu fri)]) { print \"weekday\"; }")
    (token "}")))
```

## JSON AST

```json
{
  "type": "given_statement",
  "start": {
    "Line": 1,
    "Column": 1,
    "Offset": 0
  },
  "end": {
    "Line": 4,
    "Column": 2,
    "Offset": 120
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
          "value": "given",
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
        "Column": 12,
        "Offset": 11
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
            "Column": 12,
            "Offset": 11
          },
          "text": "day"
        }
      ]
    },
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
      "text": ")"
    },
    {
      "type": "block_stmt",
      "start": {
        "Line": 1,
        "Column": 14,
        "Offset": 13
      },
      "end": {
        "Line": 4,
        "Column": 2,
        "Offset": 120
      },
      "children": [
        {
          "type": "token",
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
          "text": "{"
        },
        {
          "type": "expression_stmt",
          "start": {
            "Line": 2,
            "Column": 5,
            "Offset": 19
          },
          "end": {
            "Line": 2,
            "Column": 46,
            "Offset": 60
          },
          "children": [
            {
              "type": "literal",
              "start": {
                "Line": 2,
                "Column": 5,
                "Offset": 19
              },
              "end": {
                "Line": 2,
                "Column": 46,
                "Offset": 60
              },
              "value": "when ([qw(sat sun)]) { print \"weekend\"; }",
              "kind": "string"
            }
          ]
        },
        {
          "type": "expression_stmt",
          "start": {
            "Line": 3,
            "Column": 5,
            "Offset": 65
          },
          "end": {
            "Line": 3,
            "Column": 58,
            "Offset": 118
          },
          "children": [
            {
              "type": "literal",
              "start": {
                "Line": 3,
                "Column": 5,
                "Offset": 65
              },
              "end": {
                "Line": 3,
                "Column": 58,
                "Offset": 118
              },
              "value": "when ([qw(mon tue wed thu fri)]) { print \"weekday\"; }",
              "kind": "string"
            }
          ]
        },
        {
          "type": "token",
          "start": {
            "Line": 4,
            "Column": 1,
            "Offset": 119
          },
          "end": {
            "Line": 4,
            "Column": 2,
            "Offset": 120
          },
          "text": "}"
        }
      ]
    }
  ]
}
```
