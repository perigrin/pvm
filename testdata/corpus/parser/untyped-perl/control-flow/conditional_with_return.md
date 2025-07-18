---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - conditional
    - return
---

# Conditional With Return

Conditional with return statements

```perl
if ($error) {
    return $error_value;
}
return $success_value;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
if ($error) {
    return $error_value;
}
return $success_value;
```

## Typed Perl Output

```perl
if ($error) {
    return $error_value;
}
return $success_value;
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
  expression_statement
    return_expression
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
              "text": "error"
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
                  "value": "return $error_value",
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
      "type": "expression_statement",
      "children": [
        {
          "type": "return_expression",
          "children": [
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "return",
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
                  "text": "success_value"
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
