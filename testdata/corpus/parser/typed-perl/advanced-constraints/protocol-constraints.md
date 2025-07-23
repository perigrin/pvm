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

#
# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
class Processor<T> where T:  { }
```

## Typed Perl Output

```perl
class Processor<T> where T: EventHandler { }
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

(none)
