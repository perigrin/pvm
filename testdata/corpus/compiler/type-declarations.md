---
category: compiler
subcategory: type-declarations
tags:
    - type-definitions
    - custom-types
    - type-removal
type_check: false
---

# Type Declarations

Custom type definitions

```perl
type UserId = Int;
type UserData = HashRef[Str];
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
type UserId = Int;
type UserData = ;
```

## Typed Perl Output

```perl
type UserId = Int;
type UserData = HashRef[Str];
```

## Text AST

**Note**: Type declarations are not yet supported by the tree-sitter grammar. This syntax would currently produce parse errors and cannot generate a meaningful AST.

```
(parse error - type declarations not supported)
```

## JSON AST

```json
{
  "error": "Type declarations not yet supported by grammar",
  "note": "This syntax requires grammar extension for 'type Name = Type' declarations"
}
```

# Expected Type Errors

```
(none)
```
