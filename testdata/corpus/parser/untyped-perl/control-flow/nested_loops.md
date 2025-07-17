---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - nested
    - for
    - foreach
    - complex
---

# Nested Loops

Nested loops with complex data structures

```perl
for my $outer (@outer_list) {
    foreach my $inner (@{$outer->{items}}) {
        if ($inner->{valid}) {
            process($inner);
        }
    }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
for my $outer (@outer_list) {
    foreach my $inner (@{$outer->{items}}) {
        if ($inner->{valid}) {
            process($inner);
        }
    }
}
```

## Typed Perl Output

```perl
for my $outer (@outer_list) {
    foreach my $inner (@{$outer->{items}}) {
        if ($inner->{valid}) {
            process($inner);
        }
    }
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
  for_statement
    expression_stmt
      literal
    token
    scalar
      token
      token
    token
    array
      expression_stmt
        literal
      token
    token
    block_stmt
      token
      for_statement
        expression_stmt
          literal
        token
        scalar
          token
          token
        token
        array
          expression_stmt
            literal
          dereference_expression
            scalar
              token
              token
            expression_stmt
              literal
            hashref_element
              expression_stmt
                literal
          token
        token
        block_stmt
          token
          conditional_statement
            expression_stmt
              literal
            token
            hashref_element
              scalar
                token
                token
              expression_stmt
                literal
              hashref_element
                expression_stmt
                  literal
            token
            block_stmt
              token
              expression_stmt
                literal
              token
              token
          token
      token
```

## JSON Format

```json
{
  "path": "/tmp/nested_loops.pl",
  "root": {
    "type": "source_file",
    "children": [
      {
        "type": "for_statement",
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
          }
        ]
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 126
}
```
