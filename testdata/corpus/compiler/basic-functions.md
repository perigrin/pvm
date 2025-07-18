---
category: compiler
subcategory: basic-functions
tags:
    - function-signatures
    - typed-parameters
    - clean-perl-output
type_check: false
---

# Simple Function Signature

Basic function with typed parameters

```perl
sub Int add (Int $a, Int $b) {
    return $a + $b;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
sub add ($a, $b) {
    return $a + $b;
}
```

## Text AST

**Note**: Function signatures with type annotations are not yet supported by the tree-sitter grammar. This syntax would currently produce parse errors and cannot generate a meaningful AST.

```
(parse error - function signatures not supported)
```

## JSON AST

```json
{
  "error": "Function signatures with type annotations not yet supported by grammar",
  "note": "This syntax requires grammar extension for sub return types and typed parameters"
}
```
