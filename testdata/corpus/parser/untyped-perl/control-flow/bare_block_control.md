---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - bare_block
    - loop_control
    - last
---

# Bare Block Control

Loop control in bare block

```perl
{
    my $x = calculate();
    last if $x > 100;
    process($x);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
{
    my $x = calculate();
    last if $x > 100;
    process($x);
}
```

## Typed Perl Output

```perl
{
    my $x = calculate();
    last if $x > 100;
    process($x);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text Format

```
source_file
  block_stmt
    token
    assignment_expression
      var_decl
        variable
      expression_stmt
        literal
      expression_stmt
        literal
    token
    conditional_statement
      expression_stmt
        literal
      relational_expression
        scalar
          token
          token
        expression_stmt
          literal
        numeric_literal
      token
    expression_stmt
      literal
    token
    token
```

## JSON Format

```json
{
  "path": "/tmp/bare_block_control.pl",
  "root": {
    "type": "source_file",
    "children": [
      {
        "type": "block_stmt",
        "children": [
          {
            "type": "assignment_expression",
            "children": [
              {
                "type": "var_decl",
                "children": [
                  {
                    "type": "variable",
                    "name": "x",
                    "sigil": "$"
                  }
                ],
                "decl_type": "my"
              }
            ]
          }
        ]
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 63
}
```
