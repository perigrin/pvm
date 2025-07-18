---
category: untyped-perl
subcategory: expressions
tags:
    - conditional
    - assignment
    - logical
    - or
---

# Conditional Assignment

Using logical OR for conditional assignment

```perl
$value = $input || $default_value;
```

# Expected Compilation Outcomes

## Conditional Assignment

### Clean Perl Output

```perl
$value = $input || $default_value;
```

### Typed Perl Output

```perl
$value = $input || $default_value;
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
        scalar
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
                {"type": "token", "text": "value"}
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
                    {"type": "token", "text": "input"}
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
                    {"type": "token", "text": "default_value"}
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
