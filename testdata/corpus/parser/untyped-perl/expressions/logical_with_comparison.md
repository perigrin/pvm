---
category: untyped-perl
subcategory: expressions
tags:
    - logical
    - comparison
    - mixed_operators
    - complex
---

# Logical With Comparison

Logical operators combined with comparison operators

```perl
$result = ($a > 0) && ($b < 100) || ($c == 0);
```

# Expected Compilation Outcomes

## Logical With Comparison

### Clean Perl Output

```perl
$result = ($a > 0) && ($b < 100) || ($c == 0);
```

### Typed Perl Output

```perl
$result = ($a > 0) && ($b < 100) || ($c == 0);
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
                {"type": "token", "text": "result"}
              ]
            },
            {"type": "token", "text": "="},
            {
              "type": "binary_expression",
              "children": [
                {
                  "type": "binary_expression",
                  "children": [
                    {"type": "token", "text": "("},
                    {
                      "type": "relational_expression",
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
                            {"type": "literal", "value": ">", "kind": "string"}
                          ]
                        },
                        {"type": "token", "text": "0"}
                      ]
                    },
                    {"type": "token", "text": ")"},
                    {
                      "type": "expression_stmt",
                      "children": [
                        {"type": "literal", "value": "&&", "kind": "string"}
                      ]
                    },
                    {"type": "token", "text": "("},
                    {
                      "type": "relational_expression",
                      "children": [
                        {
                          "type": "scalar",
                          "children": [
                            {"type": "token", "text": "$"},
                            {"type": "token", "text": "b"}
                          ]
                        },
                        {
                          "type": "expression_stmt",
                          "children": [
                            {"type": "literal", "value": "<", "kind": "string"}
                          ]
                        },
                        {"type": "token", "text": "100"}
                      ]
                    },
                    {"type": "token", "text": ")"}
                  ]
                },
                {
                  "type": "expression_stmt",
                  "children": [
                    {"type": "literal", "value": "||", "kind": "string"}
                  ]
                },
                {"type": "token", "text": "("},
                {
                  "type": "equality_expression",
                  "children": [
                    {
                      "type": "scalar",
                      "children": [
                        {"type": "token", "text": "$"},
                        {"type": "token", "text": "c"}
                      ]
                    },
                    {
                      "type": "expression_stmt",
                      "children": [
                        {"type": "literal", "value": "==", "kind": "string"}
                      ]
                    },
                    {"type": "token", "text": "0"}
                  ]
                },
                {"type": "token", "text": ")"}
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
