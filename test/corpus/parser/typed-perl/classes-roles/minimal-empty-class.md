---
category: typed-perl
subcategory: classes-roles
tags:
    - class-declaration
    - minimal
    - empty-class
type_check: true
---

# Minimal Empty Class

Simplest possible class declaration

```perl
class Foo {}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 12 characters
  Type Annotations:
    VarAnnotation: Foo :: class at 1:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 12 characters
  Type Annotations:
    VarAnnotation: Foo :: class at 1:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
}
```

# Expected Type Errors

```
(none)
```
