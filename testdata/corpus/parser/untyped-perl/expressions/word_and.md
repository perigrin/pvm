---
category: untyped-perl
subcategory: expressions
tags:
    - word_form
    - and
    - logical
    - low_precedence
---

# Word And

Word form of logical AND with lower precedence

```perl
$and_result = $a and $b;
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$and_result = $a and $b;
```

### Typed Perl Output

```perl
$and_result = $a and $b;
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
          "type": "lowprec_logical_expression",
          "children": [
            {
              "type": "assignment_expression",
              "children": [
                {
                  "type": "scalar",
                  "children": [
                    {"type": "token", "text": "$"},
                    {"type": "token", "text": "and_result"}
                  ]
                },
                {"type": "token", "text": "="},
                {
                  "type": "scalar",
                  "children": [
                    {"type": "token", "text": "$"},
                    {"type": "token", "text": "a"}
                  ]
                }
              ]
            },
            {
              "type": "expression_stmt",
              "children": [
                {"type": "literal", "value": "and", "kind": "string"}
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
        }
      ]
    },
    {"type": "token", "text": ";"}
  ]
}
```
