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
use v5.38.2;
```

## Typed Perl Output

```perl
type UserId = Int;
type UserData = HashRef[Str];
```

## Text AST

```
(source
  (type_declaration
    name: (identifier)
    type: (simple_type name: (identifier)))
  (type_declaration
    name: (identifier)
    type: (parameterized_type
      base: (identifier)
      args: (simple_type name: (identifier)))))
```

## JSON AST

```json
{
  "type": "source",
  "children": [
    {
      "type": "type_declaration",
      "name": "UserId",
      "type_definition": {
        "type": "simple_type",
        "name": "Int"
      }
    },
    {
      "type": "type_declaration",
      "name": "UserData",
      "type_definition": {
        "type": "parameterized_type",
        "base": "HashRef",
        "args": [
          {
            "type": "simple_type",
            "name": "Str"
          }
        ]
      }
    }
  ]
}
```

# Expected Type Errors

```
(none)
```
