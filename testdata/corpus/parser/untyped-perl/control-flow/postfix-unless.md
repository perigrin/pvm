---
category: untyped-perl
subcategory: control-flow
tags:
    - postfix
    - unless
    - conditional
---

# Postfix Unless

Postfix unless conditional statement

```perl
print "Error" unless $success;
$retry++ unless $done;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
print "Error" unless $success;
$retry++ unless $done;
```

## Typed Perl Output

```perl
print "Error" unless $success;
$retry++ unless $done;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```

## Text AST

```
source_file
  expression_statement
    postfix_conditional_expression
      ambiguous_function_call_expression
        expression_stmt
          literal
        interpolated_string_literal
          expression_stmt
            literal
          expression_stmt
            literal
          expression_stmt
            literal
      expression_stmt
        literal
      scalar
        token
        token
  token
  expression_statement
    postfix_conditional_expression
      postinc_expression
        scalar
          token
          token
        expression_stmt
          literal
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
          "type": "postfix_conditional_expression",
          "children": [
            {
              "type": "ambiguous_function_call_expression",
              "children": [
                {
                  "type": "expression_stmt",
                  "children": [
                    {
                      "type": "literal",
                      "value": "print",
                      "kind": "string"
                    }
                  ]
                },
                {
                  "type": "interpolated_string_literal",
                  "children": [
                    {
                      "type": "expression_stmt",
                      "children": [
                        {
                          "type": "literal",
                          "value": "\"",
                          "kind": "string"
                        }
                      ]
                    },
                    {
                      "type": "expression_stmt",
                      "children": [
                        {
                          "type": "literal",
                          "value": "Error",
                          "kind": "string"
                        }
                      ]
                    },
                    {
                      "type": "expression_stmt",
                      "children": [
                        {
                          "type": "literal",
                          "value": "\"",
                          "kind": "string"
                        }
                      ]
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
                  "value": "unless",
                  "kind": "string"
                }
              ]
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
            }
          ]
        }
      ]
    },
    {
      "type": "token",
      "text": ";"
    },
    {
      "type": "expression_statement",
      "children": [
        {
          "type": "postfix_conditional_expression",
          "children": [
            {
              "type": "postinc_expression",
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
                      "text": "retry"
                    }
                  ]
                },
                {
                  "type": "expression_stmt",
                  "children": [
                    {
                      "type": "literal",
                      "value": "++",
                      "kind": "string"
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
                  "value": "unless",
                  "kind": "string"
                }
              ]
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
                  "text": "done"
                }
              ]
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
