---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - postfix
    - do_until
---

# Postfix Until

Do-until loop (postfix until)

```perl
do {
    attempt();
} until ($success);
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
do {
    attempt();
} until ($success);
```

## Typed Perl Output

```perl
do {
    attempt();
} until ($success);
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
source_file
  expression_statement
    postfix_loop_expression
      do_expression
        expression_stmt
          literal
        block_stmt
          token
          expression_stmt
            literal
          token
          token
      expression_stmt
        literal
      token
      scalar
        token
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
          "type": "postfix_loop_expression",
          "children": [
            {
              "type": "do_expression",
              "children": [
                {
                  "type": "expression_stmt",
                  "children": [
                    {
                      "type": "literal",
                      "value": "do",
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
                          "value": "attempt()",
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
            },
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "until",
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
                  "text": "success"
                }
              ]
            },
            {
              "type": "token",
              "text": ")"
            }
          ]
        }
      ]
    },
    {
      "type": "token",
      "text": ";"
    }
  ]
}
```
