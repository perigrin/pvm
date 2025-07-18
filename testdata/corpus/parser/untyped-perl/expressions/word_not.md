---
category: untyped-perl
subcategory: expressions
tags:
    - word_form
    - not
    - logical
---

# Word Not

Word form of logical NOT

```perl
$not_result = not $condition;
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$not_result = not $condition;
```

### Typed Perl Output

```perl
$not_result = not $condition;
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
              "type": "ambiguous_function_call_expression",
              "children": [
                {
                  "type": "expression_stmt",
                  "children": [
                    {"type": "literal", "value": "not", "kind": "string"}
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
