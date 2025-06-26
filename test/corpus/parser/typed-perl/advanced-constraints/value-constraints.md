---
category: typed-perl
subcategory: advanced-constraints
type_check: true
tags:
    - value-constraints
    - runtime-validation
---

# Value Constraints

Test parsing of value constraints and runtime validation

```perl
class Array<T> where T: Any { }
```

## Expected AST

### Before Type Inference
```
AST {
  Path:
  Source length: 31 characters
  Type Annotations:
    VarAnnotation: Array :: class at 1:1
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
  Source length: 31 characters
  Type Annotations:
    VarAnnotation: Array :: class at 1:1
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
class Array<T> where T: Any { }
```

## Typed Perl Output

```perl
class Array<T> where T: Any { }
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

(none)
