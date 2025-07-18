---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - until
    - loop
---

# Basic Until Loop

Basic until loop

```perl
until ($done) {
    continue_processing();
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
until ($done) {
    continue_processing();
}
```

## Typed Perl Output

```perl
until ($done) {
    continue_processing();
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

## JSON AST

```json
{
  "path": "/tmp/basic_until_loop.pl",
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
                "value": "until",
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
  "source_length": 45
}
```
