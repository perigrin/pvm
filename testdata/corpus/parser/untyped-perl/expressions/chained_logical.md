---
category: untyped-perl
subcategory: expressions
tags:
    - logical
    - chained
    - and
    - multiple
---

# Chained Logical

Chained logical AND operations

```perl
$valid = $a && $b && $c && $d;
```

# Expected Compilation Outcomes

## Chained Logical

### Clean Perl Output

```perl
$valid = $a && $b && $c && $d;
```

### Typed Perl Output

```perl
$valid = $a && $b && $c && $d;
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
                {"type": "token", "text": "valid"}
              ]
            },
            {"type": "token", "text": "="},
            {
              "type": "binary_expression",
              "children": [
                {
                  "type": "binary_expression",
                  "children": [
                    {
                      "type": "binary_expression",
                      "children": [
                        {
                          "type": "scalar",
                          "children": [
                            {"type": "token", "text": "$"},
                            {"type": "token", "text": "a"}
                          ]
                        },
                        {
                          "type": "expression_stmt",
                          "children": [
                            {"type": "literal", "value": "&&", "kind": "string"}
                          ]
                        },
                        {
                          "type": "scalar",
                          "children": [
                            {"type": "token", "text": "$"},
                            {"type": "token", "text": "b"}
                          ]
                        }
                      ]
                    },
                    {
                      "type": "expression_stmt",
                      "children": [
                        {"type": "literal", "value": "&&", "kind": "string"}
                      ]
                    },
                    {
                      "type": "scalar",
                      "children": [
                        {"type": "token", "text": "$"},
                        {"type": "token", "text": "c"}
                      ]
                    }
                  ]
                },
                {
                  "type": "expression_stmt",
                  "children": [
                    {"type": "literal", "value": "&&", "kind": "string"}
                  ]
                },
                {
                  "type": "scalar",
                  "children": [
                    {"type": "token", "text": "$"},
                    {"type": "token", "text": "d"}
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
