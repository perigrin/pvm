---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - smartmatch
    - pattern
---

# Smartmatch Operator

Smartmatch operator usage

```perl
if ($value ~~ @valid_values) {
    process($value);
} elsif ($value ~~ /pattern/) {
    handle_pattern($value);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
if ($value ~~ @valid_values) {
    process($value);
} elsif ($value ~~ /pattern/) {
    handle_pattern($value);
}
```

## Typed Perl Output

```perl
if ($value ~~ @valid_values) {
    process($value);
} elsif ($value ~~ /pattern/) {
    handle_pattern($value);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
source_file
  conditional_statement
    expression_stmt
      literal
    token
    equality_expression
      scalar
        token
        token
      expression_stmt
        literal
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
    elsif
      expression_stmt
        literal
      token
      equality_expression
        scalar
          token
          token
        expression_stmt
          literal
        match_regexp
          expression_stmt
            literal
          expression_stmt
            literal
          expression_stmt
            literal
      token
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
      "type": "conditional_statement",
      "children": [
        {
          "type": "expression_stmt",
          "children": [
            {
              "type": "literal",
              "value": "if",
              "kind": "string"
            }
          ]
        },
        {
          "type": "token",
          "text": "("
        },
        {
          "type": "equality_expression",
          "children": [
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
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "~~",
                  "kind": "string"
                }
              ]
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
                  "text": "valid_values"
                }
              ]
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
                  "value": "process($value)",
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
          "type": "elsif",
          "children": [
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "elsif",
                  "kind": "string"
                }
              ]
            },
            {
              "type": "token",
              "text": "("
            },
            {
              "type": "equality_expression",
              "children": [
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
                  "type": "expression_stmt",
                  "children": [
                    {
                      "type": "literal",
                      "value": "~~",
                      "kind": "string"
                    }
                  ]
                },
                {
                  "type": "match_regexp",
                  "children": [
                    {
                      "type": "expression_stmt",
                      "children": [
                        {
                          "type": "literal",
                          "value": "/",
                          "kind": "string"
                        }
                      ]
                    },
                    {
                      "type": "expression_stmt",
                      "children": [
                        {
                          "type": "literal",
                          "value": "pattern",
                          "kind": "string"
                        }
                      ]
                    },
                    {
                      "type": "expression_stmt",
                      "children": [
                        {
                          "type": "literal",
                          "value": "/",
                          "kind": "string"
                        }
                      ]
                    }
                  ]
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
                      "value": "handle_pattern($value)",
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
  ]
}
```
