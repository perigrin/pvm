---
category: untyped-perl
subcategory: expressions
tags:
    - word_form
    - or
    - logical
    - low_precedence
---

# Word Or

Word form of logical OR with lower precedence

```perl
$or_result = $x or $y;
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$or_result = $x or $y;
```

### Typed Perl Output

```perl
$or_result = $x or $y;
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
source_file
  expression_statement
    lowprec_logical_expression
      assignment_expression
        scalar
          token
          token
        token
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
          "type": "lowprec_logical_expression",
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
                  "type": "scalar",
                  "children": [
                    {"type": "token", "text": "$"},
                    {"type": "token", "text": "x"}
                  ]
                }
              ]
            },
            {
              "type": "expression_stmt",
              "children": [
                {"type": "literal", "value": "or", "kind": "string"}
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
    },
    {"type": "token", "text": ";"}
  ]
}
```
