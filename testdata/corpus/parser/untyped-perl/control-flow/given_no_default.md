---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - given
    - when
    - switch
    - no_default
---

# Given No Default

Given-when without default clause

```perl
given ($option) {
    when ('verbose') { $verbose = 1; }
    when ('quiet') { $quiet = 1; }
    when ('debug') { $debug = 1; }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
given ($option) {
    when ('verbose') { $verbose = 1; }
    when ('quiet') { $quiet = 1; }
    when ('debug') { $debug = 1; }
}
```

## Typed Perl Output

```perl
given ($option) {
    when ('verbose') { $verbose = 1; }
    when ('quiet') { $quiet = 1; }
    when ('debug') { $debug = 1; }
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
              "text": "option"
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
                  "value": "when ('verbose') { $verbose = 1; }",
                  "kind": "string"
                }
              ]
            },
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "when ('quiet') { $quiet = 1; }",
                  "kind": "string"
                }
              ]
            },
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "when ('debug') { $debug = 1; }",
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
