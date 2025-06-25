---
category: typed-perl
subcategory: advanced-constraints
type_check: true
tags:
    - constraint-inheritance
    - roles
    - inheritance
---

# Constraint Inheritance

Test constraint inheritance from roles and parent classes

```perl
class Base<T> where T: Defined { }
```

## Expected AST

### Before Type Inference

```
AST {
  Path:
  Source length: 44 characters
  Type Annotations:
    VarAnnotation: Processor :: class at 1:1
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
  Source length: 44 characters
  Type Annotations:
    VarAnnotation: Processor :: class at 1:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
}
```

## Expected Type Errors

(none)
