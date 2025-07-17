---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - for
    - c_style
    - increment
    - initialization
---

# For Loop C Style

C-style for loop with initialization, condition, and increment

```perl
for (my $i = 0; $i < $max; $i++) {
    handle($i);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
for (my $i = 0; $i < $max; $i++) {
    handle($i);
}
```

## Typed Perl Output

```perl
for (my $i = 0; $i < $max; $i++) {
    handle($i);
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
  cstyle_for_statement
    expression_stmt
      literal
    token
    expression_statement
      var_decl
        variable
    token
    expression_statement
      relational_expression
        scalar
          token
          token
        expression_stmt
          literal
        scalar
          token
          token
    token
    postinc_expression
      scalar
        token
        token
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

## JSON Format

```json
{
  "path": "/tmp/for_loop_c_style.pl",
  "root": {
    "type": "source_file",
    "children": [
      {
        "type": "cstyle_for_statement",
        "children": [
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "for",
                "kind": "string"
              }
            ]
          },
          {
            "type": "expression_statement",
            "children": [
              {
                "type": "var_decl",
                "children": [
                  {
                    "type": "variable",
                    "name": "i",
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
  "source_length": 53
}
```
