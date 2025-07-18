---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - given
    - when
    - regex
    - pattern
---

# Given When Regex

Given-when with regex matching

```perl
given ($input) {
    when (/^\d+$/) { print "number"; }
    when (/^[a-zA-Z]+$/) { print "letters"; }
    when (/^\s*$/) { print "empty"; }
    default { print "mixed"; }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
given ($input) {
    when (/^\d+$/) { print "number"; }
    when (/^[a-zA-Z]+$/) { print "letters"; }
    when (/^\s*$/) { print "empty"; }
    default { print "mixed"; }
}
```

## Typed Perl Output

```perl
given ($input) {
    when (/^\d+$/) { print "number"; }
    when (/^[a-zA-Z]+$/) { print "letters"; }
    when (/^\s*$/) { print "empty"; }
    default { print "mixed"; }
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
              "text": "input"
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
                  "value": "when (/^\\d+$/) { print \"number\"; }",
                  "kind": "string"
                }
              ]
            },
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "when (/^[a-zA-Z]+$/) { print \"letters\"; }",
                  "kind": "string"
                }
              ]
            },
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "when (/^\\s*$/) { print \"empty\"; }",
                  "kind": "string"
                }
              ]
            },
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "default { print \"mixed\"; }",
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
