---
category: typechecker
subcategory: variable-declarations
tags:
    - basic-types
    - symbol-binding
    - variable-declarations
type_check: true
---

# Simple Typed Variables

Basic typed variable declarations with proper symbol binding behavior.

```perl
my Int $count = 42;
my Str $name = "Alice";
my Bool $active = true;
```

# Expected Symbol Table

The binder should create symbols for the **variables**, not the **type names**.
Type names (Int, Str, Bool) are type references and should not appear as symbols.
The improved binder now properly extracts type annotations and flags typed variables.

```
=== SYMBOLS ===
scalar active :: Bool [lexical|typed] at 1:1
scalar count :: Int [lexical|typed] at 1:1
scalar name :: Str [lexical|typed] at 1:1
=== TYPE ERRORS ===
No type errors
```

# Expected Type Analysis

```
Variable $count: type Int (inferred from annotation)
Variable $name: type Str (inferred from annotation)
Variable $active: type Bool (inferred from annotation)
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $count = 42;
my $name = "Alice";
my $active = true;
```

## Typed Perl Output

```perl
use v5.36;
my Int $count = 42;
my Str $name = "Alice";
my Bool $active = true;
```

# Test Notes

This test verifies the core symbol binding fix where built-in type names
(Int, Str, Bool) are correctly treated as type references rather than
being incorrectly added to the symbol table as variable symbols.
