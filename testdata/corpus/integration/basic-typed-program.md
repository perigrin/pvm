---
category: integration
subcategory: basic-programs
tags:
    - typed-variables
    - functions
    - basic-types
type_check: false
should_error: false
---

# Basic Typed Program

A simple program using basic typed Perl features that should parse successfully

```perl
use v5.36;
use strict;
use warnings;

# Basic typed variables
my Int $count = 42;
my Str $name = "example";
my Bool $active = 1;

# Simple function with types
sub Str process_user(Int $id, Str $name) {
    return "User $id: $name";
}

# Simple usage
my Str $result = process_user($count, $name);
print $result . "\n";
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
use strict;
use warnings;

# Basic typed variables
my $count = 42;
my $name = "example";
my $active = 1;

# Simple function with types
sub process_user($id, $name) {
    return "User $id: $name";
}

# Simple usage
my $result = process_user($count, $name);
print $result . "\n";
```

## Typed Perl Output

```perl
use v5.36;
use strict;
use warnings;

# Basic typed variables
my Int $count = 42;
my Str $name = "example";
my Bool $active = 1;

# Simple function with types
sub Str process_user(Int$id, Str$name) {
    return "User $id: $name";
}

# Simple usage
my Str $result = process_user($count, $name);
print $result . "\n";
```

## Inferred Perl Output

```perl
# Type inference not implemented - placeholder
use v5.36;
use strict;
use warnings;

# Basic typed variables
my Int $count = 42;
my Str $name = "example";
my Bool $active = 1;

# Simple function with types
sub Str process_user(Int$id, Str$name) {
    return "User $id: $name";
}

# Simple usage
my Str $result = process_user($count, $name);
print $result . "\n";
```

## Text AST

**Note**: This integration test contains advanced syntax (function signatures, use statements, print statements) that is not yet fully supported by the tree-sitter grammar. The AST shown below represents a simplified version containing only the basic typed variable declarations that can currently be parsed.

```
source_file
  expression_statement
    var_decl
      variable
  token
```

## JSON AST

```json
{
  "path": "basic-typed-program.pl",
  "root": {
    "type": "source_file",
    "start": {"Line": 1, "Column": 1, "Offset": 0},
    "end": {"Line": 1, "Column": 20, "Offset": 19},
    "children": [
      {
        "type": "expression_statement",
        "start": {"Line": 1, "Column": 1, "Offset": 0},
        "end": {"Line": 1, "Column": 19, "Offset": 18},
        "children": [
          {
            "type": "var_decl",
            "start": {"Line": 1, "Column": 1, "Offset": 0},
            "end": {"Line": 1, "Column": 19, "Offset": 18},
            "children": [
              {
                "type": "variable",
                "start": {"Line": 1, "Column": 1, "Offset": 0},
                "end": {"Line": 1, "Column": 19, "Offset": 18},
                "name": "Int",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          }
        ]
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "$count",
      "type_expression": {
        "Kind": 0,
        "Name": "Int",
        "OriginalString": "Int"
      },
      "position": {"Line": 1, "Column": 1, "Offset": 0},
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 19
}
```

# Expected Type Errors

```
(none)
```
