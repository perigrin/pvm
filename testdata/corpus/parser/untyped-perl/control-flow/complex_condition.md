---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - conditional
    - complex_expression
    - logical
---

# Complex Condition

Conditional with complex boolean expression

```perl
if (($a > 0) && ($b < 100) || ($c == $d)) {
    complex_action();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
if (($a > 0) && ($b < 100) || ($c == $d)) {
    complex_action();
}
```

## Typed Perl Output

```perl
if (($a > 0) && ($b < 100) || ($c == $d)) {
    complex_action();
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
  conditional_statement
    expression_stmt
      literal
    token
    logical_expression
      logical_expression
        relational_expression
          scalar
            token
            token
          expression_stmt
            literal
          numeric_literal
        expression_stmt
          literal
        relational_expression
          scalar
            token
            token
          expression_stmt
            literal
          numeric_literal
      expression_stmt
        literal
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
  "path": "/tmp/complex_condition.pl",
  "root": {
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
          }
        ]
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 60
}
```
