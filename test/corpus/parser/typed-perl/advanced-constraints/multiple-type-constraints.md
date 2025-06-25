---
category: typed-perl
subcategory: advanced-constraints
type_check: true
tags:
    - multiple-constraints
    - type-parameters
---

# Multiple Type Constraints

Test parsing of multiple type constraints on type parameters

```perl
class Handler<T> where T: Serializable&Defined { }
```

## Expected AST

### Before Type Inference
```
AST {
  Path:
  Source length: 50 characters
  Type Annotations:
    VarAnnotation: Handler :: class at 1:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
}
```

### After Type Inference
```
AST {
  Path:
  Source length: 50 characters
  Type Annotations:
    VarAnnotation: Handler :: class at 1:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
}
```

## Expected Type Errors

(none)
