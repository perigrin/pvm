---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - conditional
    - assignment
    - variable_declaration
---

# Conditional With Assignment

Conditional with variable assignment in condition

```perl
if (my $result = function_call()) {
    process($result);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
if (my $result = function_call()) {
    process($result);
}
```

## Typed Perl Output

```perl
if (my $result = function_call()) {
    process($result);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
(conditional_statement
  (literal "if")
  (token "(")
  (var_decl
    (variable name="result" sigil="$"))
  (token ")")
  (block_stmt
    (token "{")
    (literal "process($result)")
    (token ";")
    (token "}")))
```

## JSON AST

```json
{
  "type": "conditional_statement",
  "start": {
    "Line": 1,
    "Column": 1,
    "Offset": 0
  },
  "end": {
    "Line": 3,
    "Column": 2,
    "Offset": 59
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
        "Column": 3,
        "Offset": 2
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
            "Column": 3,
            "Offset": 2
          },
          "value": "if",
          "kind": "string"
        }
      ]
    },
    {
      "type": "token",
      "start": {
        "Line": 1,
        "Column": 4,
        "Offset": 3
      },
      "end": {
        "Line": 1,
        "Column": 5,
        "Offset": 4
      },
      "text": "("
    },
    {
      "type": "var_decl",
      "start": {
        "Line": 1,
        "Column": 5,
        "Offset": 4
      },
      "end": {
        "Line": 1,
        "Column": 33,
        "Offset": 32
      },
      "children": [
        {
          "type": "variable",
          "start": {
            "Line": 1,
            "Column": 5,
            "Offset": 4
          },
          "end": {
            "Line": 1,
            "Column": 33,
            "Offset": 32
          },
          "name": "result",
          "sigil": "$"
        }
      ],
      "decl_type": "my"
    },
    {
      "type": "token",
      "start": {
        "Line": 1,
        "Column": 33,
        "Offset": 32
      },
      "end": {
        "Line": 1,
        "Column": 34,
        "Offset": 33
      },
      "text": ")"
    },
    {
      "type": "block_stmt",
      "start": {
        "Line": 1,
        "Column": 35,
        "Offset": 34
      },
      "end": {
        "Line": 3,
        "Column": 2,
        "Offset": 59
      },
      "children": [
        {
          "type": "token",
          "start": {
            "Line": 1,
            "Column": 35,
            "Offset": 34
          },
          "end": {
            "Line": 1,
            "Column": 36,
            "Offset": 35
          },
          "text": "{"
        },
        {
          "type": "expression_stmt",
          "start": {
            "Line": 2,
            "Column": 5,
            "Offset": 40
          },
          "end": {
            "Line": 2,
            "Column": 21,
            "Offset": 56
          },
          "children": [
            {
              "type": "literal",
              "start": {
                "Line": 2,
                "Column": 5,
                "Offset": 40
              },
              "end": {
                "Line": 2,
                "Column": 21,
                "Offset": 56
              },
              "value": "process($result)",
              "kind": "string"
            }
          ]
        },
        {
          "type": "token",
          "start": {
            "Line": 2,
            "Column": 21,
            "Offset": 56
          },
          "end": {
            "Line": 2,
            "Column": 22,
            "Offset": 57
          },
          "text": ";"
        },
        {
          "type": "token",
          "start": {
            "Line": 3,
            "Column": 1,
            "Offset": 58
          },
          "end": {
            "Line": 3,
            "Column": 2,
            "Offset": 59
          },
          "text": "}"
        }
      ]
    }
  ]
}
```
