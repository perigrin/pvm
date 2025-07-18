---
category: untyped-perl
subcategory: expressions
tags:
    - logical
    - or
    - short_circuit
    - binary_operator
---

# Logical Or

Logical OR operator with short-circuit evaluation

```perl
$or_result = $x || $y;
```

# Expected Compilation Outcomes

## Logical Or

### Clean Perl Output

```perl
$or_result = $x || $y;
```

### Typed Perl Output

```perl
$or_result = $x || $y;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Expected AST

```json
{
  "type": "source_file",
  "children": [
    {
      "type": "expression_statement",
      "children": [
        {
          "type": "assignment_expression",
          "children": [
            {
              "type": "scalar",
              "children": [
                {"type": "token", "text": "$"},
                {"type": "token", "text": "or_result"}
              ]
            },
            {"type": "token", "text": "="},
            {
              "type": "binary_expression",
              "children": [
                {
                  "type": "scalar",
                  "children": [
                    {"type": "token", "text": "$"},
                    {"type": "token", "text": "x"}
                  ]
                },
                {
                  "type": "expression_stmt",
                  "children": [
                    {"type": "literal", "value": "||", "kind": "string"}
                  ]
                },
                {
                  "type": "scalar",
                  "children": [
                    {"type": "token", "text": "$"},
                    {"type": "token", "text": "y"}
                  ]
                }
              ]
            }
          ]
        }
      ]
    },
    {"type": "token", "text": ";"}
  ]
}
```
