---
category: untyped-perl
subcategory: expressions
tags:
    - defined_or
    - undef_handling
    - conditional
    - basic
---

# Defined Or

Defined-or operator for handling undef values

```perl
$result = $value // $default;
```

# Expected Compilation Outcomes

## Defined Or

### Clean Perl Output

```perl
$result = $value // $default;
```

### Typed Perl Output

```perl
$result = $value // $default;
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
    (token "result"))
  (token "=")
  (binary_expression
    (scalar
      (token "$")
      (token "value"))
    (literal "//")
    (scalar
      (token "$")
      (token "default"))))
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
    "Column": 29,
    "Offset": 28
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
        "Column": 8,
        "Offset": 7
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
            "Column": 8,
            "Offset": 7
          },
          "text": "result"
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
      "text": "="
    },
    {
      "type": "binary_expression",
      "start": {
        "Line": 1,
        "Column": 11,
        "Offset": 10
      },
      "end": {
        "Line": 1,
        "Column": 29,
        "Offset": 28
      },
      "children": [
        {
          "type": "scalar",
          "start": {
            "Line": 1,
            "Column": 11,
            "Offset": 10
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
                "Column": 11,
                "Offset": 10
              },
              "end": {
                "Line": 1,
                "Column": 12,
                "Offset": 11
              },
              "text": "$"
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
                "Column": 17,
                "Offset": 16
              },
              "text": "value"
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
            "Column": 20,
            "Offset": 19
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
                "Column": 20,
                "Offset": 19
              },
              "value": "//",
              "kind": "string"
            }
          ]
        },
        {
          "type": "scalar",
          "start": {
            "Line": 1,
            "Column": 21,
            "Offset": 20
          },
          "end": {
            "Line": 1,
            "Column": 29,
            "Offset": 28
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
              "text": "$"
            },
            {
              "type": "token",
              "start": {
                "Line": 1,
                "Column": 22,
                "Offset": 21
              },
              "end": {
                "Line": 1,
                "Column": 29,
                "Offset": 28
              },
              "text": "default"
            }
          ]
        }
      ]
    }
  ]
}
```
