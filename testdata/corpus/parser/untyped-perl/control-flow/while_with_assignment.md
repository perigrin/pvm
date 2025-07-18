---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - while
    - assignment
    - file_handle
---

# While With Assignment

While loop with assignment in condition

```perl
while (my $line = <$fh>) {
    chomp $line;
    process($line);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
while (my $line = <$fh>) {
    chomp $line;
    process($line);
}
```

## Typed Perl Output

```perl
while (my $line = <$fh>) {
    chomp $line;
    process($line);
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
  loop_statement
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
    block_stmt
      token
      expression_stmt
        literal
      token
      expression_stmt
        literal
      token
      token
```

## JSON AST

```json
{
  "path": "/tmp/while_with_assignment.pl",
  "root": {
    "type": "source_file",
    "children": [
      {
        "type": "loop_statement",
        "children": [
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "while",
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
  "source_length": 67
}
```
