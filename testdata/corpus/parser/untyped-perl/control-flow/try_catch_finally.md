---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - try_catch
    - eval
    - error_handling
    - local
---

# Try Catch Finally

Try-catch-finally pattern with eval

```perl
{
    local $@;
    eval {
        dangerous_operation();
    };
    if ($@) {
        handle_exception($@);
    } else {
        success_handler();
    }
    # cleanup always runs
    cleanup();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
{
    local $@;
    eval {
        dangerous_operation();
    };
    if ($@) {
        handle_exception($@);
    } else {
        success_handler();
    }
    # cleanup always runs
    cleanup();
}
```

## Typed Perl Output

```perl
{
    local $@;
    eval {
        dangerous_operation();
    };
    if ($@) {
        handle_exception($@);
    } else {
        success_handler();
    }
    # cleanup always runs
    cleanup();
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
source_file
  block_statement
    token
    expression_statement
      variable_declaration
        expression_stmt
          literal
        scalar
          token
          token
    token
    expression_statement
      eval_expression
        expression_stmt
          literal
        block_stmt
          token
          expression_stmt
            literal
          token
          token
    token
    conditional_statement
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
        token
        token
      else
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
    expression_statement
      function_call_expression
        expression_stmt
          literal
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
      "type": "block_statement",
      "children": [
        {
          "type": "token",
          "text": "{"
        },
        {
          "type": "expression_statement",
          "children": [
            {
              "type": "variable_declaration",
              "children": [
                {
                  "type": "expression_stmt",
                  "children": [
                    {
                      "type": "literal",
                      "value": "local",
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
                      "text": "@"
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
              "type": "eval_expression",
              "children": [
                {
                  "type": "expression_stmt",
                  "children": [
                    {
                      "type": "literal",
                      "value": "eval",
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
                          "value": "dangerous_operation()",
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
        },
        {
          "type": "token",
          "text": ";"
        },
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
              "type": "scalar",
              "children": [
                {
                  "type": "token",
                  "text": "$"
                },
                {
                  "type": "token",
                  "text": "@"
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
                      "value": "handle_exception($@)",
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
              "type": "else",
              "children": [
                {
                  "type": "expression_stmt",
                  "children": [
                    {
                      "type": "literal",
                      "value": "else",
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
                          "value": "success_handler()",
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
        },
        {
          "type": "expression_stmt",
          "children": [
            {
              "type": "literal",
              "value": "# cleanup always runs",
              "kind": "string"
            }
          ]
        },
        {
          "type": "expression_statement",
          "children": [
            {
              "type": "function_call_expression",
              "children": [
                {
                  "type": "expression_stmt",
                  "children": [
                    {
                      "type": "literal",
                      "value": "cleanup",
                      "kind": "string"
                    }
                  ]
                },
                {
                  "type": "token",
                  "text": "("
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
        },
        {
          "type": "token",
          "text": "}"
        }
      ]
    }
  ]
}
```
