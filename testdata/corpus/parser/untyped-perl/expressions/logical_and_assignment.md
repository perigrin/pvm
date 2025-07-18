---
category: untyped-perl
subcategory: expressions
tags:
    - logical
    - and
    - assignment
    - conditional
---

# Logical And Assignment

Logical AND assignment operator

```perl
$flag &&= $condition;
```

# Expected Compilation Outcomes

## Logical And Assignment

### Clean Perl Output

```perl
$flag &&= $condition;
```

### Typed Perl Output

```perl
$flag &&= $condition;
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
                {"type": "token", "text": "flag"}
              ]
            },
            {
              "type": "expression_stmt",
              "children": [
                {"type": "literal", "value": "&&=", "kind": "string"}
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
    },
    {"type": "token", "text": ";"}
  ]
}
```
