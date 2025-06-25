---
category: typed-perl
subcategory: advanced-constraints
type_check: true
tags:
    - protocol-constraints
    - capabilities
---

# Protocol Constraints

Test parsing of protocol and capability constraints

```perl
class Processor<T> where T: EventHandler { }
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
