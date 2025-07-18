---
category: untyped-perl
subcategory: expressions
tags:
    - logical
    - not
    - unary_operator
    - basic
---

# Logical Not

Logical NOT operator

```perl
$not_result = !$condition;
```

# Expected Compilation Outcomes

## Logical Not

### Clean Perl Output

```perl
$not_result = !$condition;
```

### Typed Perl Output

```perl
$not_result = !$condition;
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
                {"type": "token", "text": "not_result"}
              ]
            },
            {"type": "token", "text": "="},
            {
              "type": "unary_expression",
              "children": [
                {
                  "type": "expression_stmt",
                  "children": [
                    {"type": "literal", "value": "!", "kind": "string"}
                  ]
                },
                {
                  "type": "scalar",
                  "children": [
                    {"type": "token", "text": "$"},
                    {"type": "token", "text": "condition"}
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
