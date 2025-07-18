---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - foreach
    - hash
    - keys
---

# Foreach Loop Hash

Foreach loop over hash keys

```perl
foreach my $key (keys %hash) {
    process($key, $hash{$key});
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
foreach my $key (keys %hash) {
    process($key, $hash{$key});
}
```

## Typed Perl Output

```perl
foreach my $key (keys %hash) {
    process($key, $hash{$key});
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
    (token "key"))
  (token "(")
  (func1op_call_expression
    (literal "keys")
    (hash
      (literal "%")
      (token "hash")))
  (token ")")
  (block_stmt
    (token "{")
    (literal "process($key, $hash{$key})")
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
    "Offset": 64
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
        "Column": 16,
        "Offset": 15
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
            "Column": 16,
            "Offset": 15
          },
          "text": "key"
        }
      ]
    },
    {
      "type": "token",
      "start": {
        "Line": 1,
        "Column": 17,
        "Offset": 16
      },
      "end": {
        "Line": 1,
        "Column": 18,
        "Offset": 17
      },
      "text": "("
    },
    {
      "type": "func1op_call_expression",
      "start": {
        "Line": 1,
        "Column": 18,
        "Offset": 17
      },
      "end": {
        "Line": 1,
        "Column": 28,
        "Offset": 27
      },
      "children": [
        {
          "type": "expression_stmt",
          "start": {
            "Line": 1,
            "Column": 18,
            "Offset": 17
          },
          "end": {
            "Line": 1,
            "Column": 22,
            "Offset": 21
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
                "Column": 22,
                "Offset": 21
              },
              "value": "keys",
              "kind": "string"
            }
          ]
        },
        {
          "type": "hash",
          "start": {
            "Line": 1,
            "Column": 23,
            "Offset": 22
          },
          "end": {
            "Line": 1,
            "Column": 28,
            "Offset": 27
          },
          "children": [
            {
              "type": "expression_stmt",
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
              "children": [
                {
                  "type": "literal",
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
                  "value": "%",
                  "kind": "string"
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
                "Column": 28,
                "Offset": 27
              },
              "text": "hash"
            }
          ]
        }
      ]
    },
    {
      "type": "token",
      "start": {
        "Line": 1,
        "Column": 28,
        "Offset": 27
      },
      "end": {
        "Line": 1,
        "Column": 29,
        "Offset": 28
      },
      "text": ")"
    },
    {
      "type": "block_stmt",
      "start": {
        "Line": 1,
        "Column": 30,
        "Offset": 29
      },
      "end": {
        "Line": 3,
        "Column": 2,
        "Offset": 64
      },
      "children": [
        {
          "type": "token",
          "start": {
            "Line": 1,
            "Column": 30,
            "Offset": 29
          },
          "end": {
            "Line": 1,
            "Column": 31,
            "Offset": 30
          },
          "text": "{"
        },
        {
          "type": "expression_stmt",
          "start": {
            "Line": 2,
            "Column": 5,
            "Offset": 35
          },
          "end": {
            "Line": 2,
            "Column": 31,
            "Offset": 61
          },
          "children": [
            {
              "type": "literal",
              "start": {
                "Line": 2,
                "Column": 5,
                "Offset": 35
              },
              "end": {
                "Line": 2,
                "Column": 31,
                "Offset": 61
              },
              "value": "process($key, $hash{$key})",
              "kind": "string"
            }
          ]
        },
        {
          "type": "token",
          "start": {
            "Line": 2,
            "Column": 31,
            "Offset": 61
          },
          "end": {
            "Line": 2,
            "Column": 32,
            "Offset": 62
          },
          "text": ";"
        },
        {
          "type": "token",
          "start": {
            "Line": 3,
            "Column": 1,
            "Offset": 63
          },
          "end": {
            "Line": 3,
            "Column": 2,
            "Offset": 64
          },
          "text": "}"
        }
      ]
    }
  ]
}
```
