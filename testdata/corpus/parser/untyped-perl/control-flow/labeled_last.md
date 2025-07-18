---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - labeled
    - last
    - search
    - nested
---

# Labeled Last

Labeled last to break out of outer loop

```perl
SEARCH: foreach my $item (@items) {
    foreach my $prop (@properties) {
        if ($item->{$prop} eq $target) {
            $found = $item;
            last SEARCH;
        }
    }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
SEARCH: foreach my $item (@items) {
    foreach my $prop (@properties) {
        if ($item->{$prop} eq $target) {
            $found = $item;
            last SEARCH;
        }
    }
}
```

## Typed Perl Output

```perl
SEARCH: foreach my $item (@items) {
    foreach my $prop (@properties) {
        if ($item->{$prop} eq $target) {
            $found = $item;
            last SEARCH;
        }
    }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
source_file
  statement_label
    token
    expression_stmt
      literal
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
```

## JSON AST

```json
{
  "type": "source_file",
  "children": [
    {
      "type": "statement_label",
      "children": [
        {
          "type": "token",
          "text": "SEARCH:"
        },
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
                      "value": "foreach my $prop (@properties) {\n        if ($item->{$prop} eq $target) {\n            $found = $item;\n            last SEARCH;\n        }\n    }",
                      "kind": "string"
                    }
                  ]
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
  ]
}
```
