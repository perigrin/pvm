---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - exception
    - error_handling
    - eval
    - do_block
---

# Exception Handling Eval

Exception handling with eval block

```perl
eval {
    risky_operation();
    1;
} or do {
    my $error = $@ || 'Unknown error';
    handle_error($error);
};
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
eval {
    risky_operation();
    1;
} or do {
    my $error = $@ || 'Unknown error';
    handle_error($error);
};
```

## Typed Perl Output

```perl
eval {
    risky_operation();
    1;
} or do {
    my $error = $@ || 'Unknown error';
    handle_error($error);
};
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
source_file
  expression_statement
    lowprec_logical_expression
      eval_expression
        expression_stmt
          literal
        block_stmt
          token
          expression_stmt
            literal
          token
          expression_stmt
            literal
          token
          token
      expression_stmt
        literal
      do_expression
        expression_stmt
          literal
        block_stmt
          token
          expression_stmt
            literal
          token
          expression_stmt
            literal
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
                          "value": "risky_operation()",
                          "kind": "string"
                        }
                      ]
                    },
                    {
                      "type": "token",
                      "text": ";"
                    },
                    {
                      "type": "expression_stmt",
                      "children": [
                        {
                          "type": "literal",
                          "value": "1",
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
                  "value": "or",
                  "kind": "string"
                }
              ]
            },
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
                          "value": "my $error = $@ || 'Unknown error'",
                          "kind": "string"
                        }
                      ]
                    },
                    {
                      "type": "token",
                      "text": ";"
                    },
                    {
                      "type": "expression_stmt",
                      "children": [
                        {
                          "type": "literal",
                          "value": "handle_error($error)",
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
    },
    {
      "type": "token",
      "text": ";"
    }
  ]
}
```
