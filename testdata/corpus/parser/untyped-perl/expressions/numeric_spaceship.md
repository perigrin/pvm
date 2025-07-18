---
category: untyped-perl
subcategory: expressions
tags:
    - numeric
    - spaceship
    - three_way
    - comparison
---

# Numeric Spaceship

Numeric three-way comparison (spaceship operator)

```perl
$cmp_result = $a <=> $b;
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$cmp_result = $a <=> $b;
```

### Typed Perl Output

```perl
$cmp_result = $a <=> $b;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
(assignment_expression
  (scalar
    (token "$")
    (token "cmp_result"))
  (token "=")
  (equality_expression
    (scalar
      (token "$")
      (token "a"))
    (literal "<=>")
    (scalar
      (token "$")
      (token "b"))))
```

## JSON AST

```json
{
  "type": "assignment_expression",
  "start": {
    "Line": 1,
    "Column": 1,
    "Offset": 0
  },
  "end": {
    "Line": 1,
    "Column": 24,
    "Offset": 23
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
        "Column": 12,
        "Offset": 11
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
            "Column": 12,
            "Offset": 11
          },
          "text": "cmp_result"
        }
      ]
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
        "Column": 14,
        "Offset": 13
      },
      "text": "="
    },
    {
      "type": "equality_expression",
      "start": {
        "Line": 1,
        "Column": 15,
        "Offset": 14
      },
      "end": {
        "Line": 1,
        "Column": 24,
        "Offset": 23
      },
      "children": [
        {
          "type": "scalar",
          "start": {
            "Line": 1,
            "Column": 15,
            "Offset": 14
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
                "Column": 17,
                "Offset": 16
              },
              "text": "a"
            }
          ]
        },
        {
          "type": "expression_stmt",
          "start": {
            "Line": 1,
            "Column": 18,
            "Offset": 17
          },
          "end": {
            "Line": 1,
            "Column": 21,
            "Offset": 20
          },
          "children": [
            {
              "type": "literal",
              "start": {
                "Line": 1,
                "Column": 18,
                "Offset": 17
              },
              "end": {
                "Line": 1,
                "Column": 21,
                "Offset": 20
              },
              "value": "<=>",
              "kind": "string"
            }
          ]
        },
        {
          "type": "scalar",
          "start": {
            "Line": 1,
            "Column": 22,
            "Offset": 21
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
                "Column": 22,
                "Offset": 21
              },
              "end": {
                "Line": 1,
                "Column": 23,
                "Offset": 22
              },
              "text": "$"
            },
            {
              "type": "token",
              "start": {
                "Line": 1,
                "Column": 23,
                "Offset": 22
              },
              "end": {
                "Line": 1,
                "Column": 24,
                "Offset": 23
              },
              "text": "b"
            }
          ]
        }
      ]
    }
  ]
}
```
