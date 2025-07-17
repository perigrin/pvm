---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - unless
---

# Unless Statement

Unless conditional statement

```perl
unless ($negative_condition) {
    execute();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
unless ($negative_condition) {
    execute();
}
```

## Typed Perl Output

```perl
unless ($negative_condition) {
    execute();
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
  "path": "/tmp/unless_statement.pl",
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
                "value": "unless",
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
  "source_length": 46
}
```
