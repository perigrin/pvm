---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - foreach
    - nested
    - loop_control
    - next
    - last
---

# Complex Loop Control

Complex loop control with multiple conditions

```perl
foreach my $batch (@batches) {
    foreach my $item (@{$batch}) {
        next if $item->{skip};
        my $result = process($item);
        if ($result->{error}) {
            log_error($result->{error});
            next;
        }
        last if $result->{complete};
    }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
foreach my $batch (@batches) {
    foreach my $item (@{$batch}) {
        next if $item->{skip};
        my $result = process($item);
        if ($result->{error}) {
            log_error($result->{error});
            next;
        }
        last if $result->{complete};
    }
}
```

## Typed Perl Output

```perl
foreach my $batch (@batches) {
    foreach my $item (@{$batch}) {
        next if $item->{skip};
        my $result = process($item);
        if ($result->{error}) {
            log_error($result->{error});
            next;
        }
        last if $result->{complete};
    }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

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
          token
        token
        block_stmt
          token
          conditional_statement
            expression_stmt
              literal
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
              expression_stmt
                literal
              token
              token
          conditional_statement
            expression_stmt
              literal
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
          token
      token
```

## JSON AST

```json
{
  "path": "/tmp/complex_loop_control.pl",
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
                "value": "foreach",
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
  "source_length": 243
}
```
