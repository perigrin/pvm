---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - foreach
    - continue
    - progress
---

# Foreach Continue

Foreach loop with continue block

```perl
foreach my $item (@items) {
    process($item);
} continue {
    log_progress();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
foreach my $item (@items) {
    process($item);
} continue {
    log_progress();
}
```

## Typed Perl Output

```perl
foreach my $item (@items) {
    process($item);
} continue {
    log_progress();
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
source_file
  for_statement
    expression_stmt
      literal
    token
    scalar
      token
      token
    token
    array
      expression_stmt
        literal
      token
    token
    block_stmt
      token
      expression_stmt
        literal
      token
      token
    expression_stmt
      literal
    block_stmt
      token
      expression_stmt
        literal
      token
      token
```

## JSON AST

```json
{
  "type": "source_file",
  "children": [
    {
      "type": "for_statement",
      "children": [
        {
          "type": "expression_stmt",
          "children": [
            {
              "type": "literal",
              "value": "foreach",
              "kind": "string"
            }
          ]
        },
        {
          "type": "token",
          "text": "my"
        },
        {
          "type": "scalar",
          "children": [
            {
              "type": "token",
              "text": "$"
            },
            {
              "type": "token",
              "text": "item"
            }
          ]
        },
        {
          "type": "token",
          "text": "("
        },
        {
          "type": "array",
          "children": [
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "@",
                  "kind": "string"
                }
              ]
            },
            {
              "type": "token",
              "text": "items"
            }
          ]
        },
        {
          "type": "token",
          "text": ")"
        },
        {
          "type": "block_stmt",
          "children": [
            {
              "type": "token",
              "text": "{"
            },
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "process($item)",
                  "kind": "string"
                }
              ]
            },
            {
              "type": "token",
              "text": ";"
            },
            {
              "type": "token",
              "text": "}"
            }
          ]
        },
        {
          "type": "expression_stmt",
          "children": [
            {
              "type": "literal",
              "value": "continue",
              "kind": "string"
            }
          ]
        },
        {
          "type": "block_stmt",
          "children": [
            {
              "type": "token",
              "text": "{"
            },
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "log_progress()",
                  "kind": "string"
                }
              ]
            },
            {
              "type": "token",
              "text": ";"
            },
            {
              "type": "token",
              "text": "}"
            }
          ]
        }
      ]
    }
  ]
}
```
