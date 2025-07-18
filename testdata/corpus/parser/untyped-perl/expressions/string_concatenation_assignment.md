---
category: untyped-perl
subcategory: expressions
tags:
    - string
    - concatenation
    - assignment
---

# String Concatenation Assignment

String concatenation assignment operator

```perl
$message .= $suffix;
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$message .= $suffix;
```

### Typed Perl Output

```perl
$message .= $suffix;
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
    (token "message"))
  (literal ".=")
  (scalar
    (token "$")
    (token "suffix")))
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
    "Column": 20,
    "Offset": 19
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
      "type": "expression_stmt",
      "start": {
        "Line": 1,
        "Column": 10,
        "Offset": 9
      },
      "end": {
        "Line": 1,
        "Column": 12,
        "Offset": 11
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
            "Column": 12,
            "Offset": 11
          },
          "value": ".=",
          "kind": "string"
        }
      ]
    },
    {
      "type": "scalar",
      "start": {
        "Line": 1,
        "Column": 13,
        "Offset": 12
      },
      "end": {
        "Line": 1,
        "Column": 20,
        "Offset": 19
      },
      "children": [
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
          "text": "$"
        },
        {
          "type": "token",
          "start": {
            "Line": 1,
            "Column": 14,
            "Offset": 13
          },
          "end": {
            "Line": 1,
            "Column": 20,
            "Offset": 19
          },
          "text": "suffix"
        }
      ]
    }
  ]
}
```
