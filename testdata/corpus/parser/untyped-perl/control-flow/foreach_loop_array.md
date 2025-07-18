---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - foreach
    - array
---

# Foreach Loop Array

Foreach loop over array

```perl
foreach my $item (@list) {
    process($item);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
foreach my $item (@list) {
    process($item);
}
```

## Typed Perl Output

```perl
foreach my $item (@list) {
    process($item);
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
      expression_stmt
        literal
      token
      token
```

## JSON AST

```json
{
  "path": "/tmp/foreach_loop_array.pl",
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
  "source_length": 49
}
```
