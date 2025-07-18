---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - given
    - when
    - switch
    - default
---

# Given When Basic

Basic given-when switch statement

```perl
given ($value) {
    when (1) { print "one"; }
    when (2) { print "two"; }
    default { print "other"; }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
given ($value) {
    when (1) { print "one"; }
    when (2) { print "two"; }
    default { print "other"; }
}
```

## Typed Perl Output

```perl
given ($value) {
    when (1) { print "one"; }
    when (2) { print "two"; }
    default { print "other"; }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
source_file
  given_statement
    expression_stmt
      literal
    token
    scalar
      token
      token
    token
    block_stmt
      token
      expression_stmt
        literal
      expression_stmt
        literal
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
      "type": "given_statement",
      "children": [
        {
          "type": "expression_stmt",
          "children": [
            {
              "type": "literal",
              "value": "given",
              "kind": "string"
            }
          ]
        },
        {
          "type": "token",
          "text": "("
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
              "text": "value"
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
                  "value": "when (1) { print \"one\"; }",
                  "kind": "string"
                }
              ]
            },
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "when (2) { print \"two\"; }",
                  "kind": "string"
                }
              ]
            },
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "default { print \"other\"; }",
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
```
