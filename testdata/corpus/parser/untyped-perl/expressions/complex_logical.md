---
category: untyped-perl
subcategory: expressions
tags:
    - logical
    - complex
    - precedence
    - parentheses
---

# Complex Logical

Complex logical expression with precedence

```perl
$result = ($a && $b) || ($c && !$d) || $e;
```

# Expected Compilation Outcomes

## Complex Logical

### Clean Perl Output

```perl
$result = ($a && $b) || ($c && !$d) || $e;
```

### Typed Perl Output

```perl
$result = ($a && $b) || ($c && !$d) || $e;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
source_file
  expression_statement
    assignment_expression
      scalar
        token
        token
      token
      binary_expression
        binary_expression
          token
          binary_expression
            scalar
              token
              token
            expression_stmt
              literal
            scalar
              token
              token
          token
          expression_stmt
            literal
          token
          binary_expression
            scalar
              token
              token
            expression_stmt
              literal
            refgen_expression
              expression_stmt
                literal
              unary_expression
                expression_stmt
                  literal
                scalar
                  token
                  token
          token
        expression_stmt
          literal
        scalar
          token
          token
  token
```

## JSON AST

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
                    {"type": "token", "text": ")"},
                    {
                      "type": "expression_stmt",
                      "children": [
                        {"type": "literal", "value": "||", "kind": "string"}
                      ]
                    },
                    {"type": "token", "text": "("},
                    {
                      "type": "binary_expression",
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
                            {"type": "literal", "value": "&&", "kind": "string"}
                          ]
                        },
                        {
                          "type": "refgen_expression",
                          "children": [
                            {
                              "type": "expression_stmt",
                              "children": [
                                {"type": "literal", "value": "\\", "kind": "string"}
                              ]
                            },
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
                                    {"type": "token", "text": "d"}
                                  ]
                                }
                              ]
                            }
                          ]
                        }
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
                {
                  "type": "scalar",
                  "children": [
                    {"type": "token", "text": "$"},
                    {"type": "token", "text": "e"}
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
