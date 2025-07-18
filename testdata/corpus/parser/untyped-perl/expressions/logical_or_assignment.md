---
category: untyped-perl
subcategory: expressions
tags:
    - logical
    - or
    - assignment
    - conditional
---

# Logical Or Assignment

Logical OR assignment operator

```perl
$value ||= $default;
```

# Expected Compilation Outcomes

## Logical Or Assignment

### Clean Perl Output

```perl
$value ||= $default;
```

### Typed Perl Output

```perl
$value ||= $default;
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
            {
              "type": "expression_stmt",
              "children": [
                {"type": "literal", "value": "||=", "kind": "string"}
              ]
            },
            {
              "type": "scalar",
              "children": [
                {"type": "token", "text": "$"},
                {"type": "token", "text": "default"}
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
