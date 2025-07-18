---
category: untyped-perl
subcategory: packages
tags:
    - packages
    - modules
    - use-statements
---

# Simple Package

Basic package declaration and use statements

```perl
package MyModule;
use strict;
use warnings;

sub new {
    my $class = shift;
    return bless {}, $class;
}

1;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
package MyModule;
use strict;
use warnings;

sub new {
    my $class = shift;
    return bless {}, $class;
}

1;
```

## Typed Perl Output

```perl
package MyModule;
use strict;
use warnings;

sub new {
    my $class = shift;
    return bless {}, $class;
}

1;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text Format

```
source_file
  package_statement
    expression_stmt
      literal
    expression_stmt
      literal
    token
  use_statement
    expression_stmt
      literal
    expression_stmt
      literal
    token
  use_statement
    expression_stmt
      literal
    expression_stmt
      literal
    token
  sub_decl
    block_stmt
      token
      expression_stmt
        literal
      token
      expression_stmt
        literal
      token
      token
  expression_statement
    token
  token
```

## JSON Format

```json
{
  "path": "/tmp/simple-package.pl",
  "root": {
    "type": "source_file",
    "start": { "Line": 1, "Column": 1, "Offset": 0 },
    "end": { "Line": 10, "Column": 3, "Offset": 112 },
    "children": [
      {
        "type": "package_statement",
        "start": { "Line": 1, "Column": 1, "Offset": 0 },
        "end": { "Line": 1, "Column": 18, "Offset": 17 },
        "children": [
          {
            "type": "expression_stmt",
            "start": { "Line": 1, "Column": 1, "Offset": 0 },
            "end": { "Line": 1, "Column": 8, "Offset": 7 },
            "children": [
              {
                "type": "literal",
                "start": { "Line": 1, "Column": 1, "Offset": 0 },
                "end": { "Line": 1, "Column": 8, "Offset": 7 },
                "value": "package",
                "kind": "string"
              }
            ]
          },
          {
            "type": "expression_stmt",
            "start": { "Line": 1, "Column": 9, "Offset": 8 },
            "end": { "Line": 1, "Column": 17, "Offset": 16 },
            "children": [
              {
                "type": "literal",
                "start": { "Line": 1, "Column": 9, "Offset": 8 },
                "end": { "Line": 1, "Column": 17, "Offset": 16 },
                "value": "MyModule",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "start": { "Line": 1, "Column": 17, "Offset": 16 },
            "end": { "Line": 1, "Column": 18, "Offset": 17 },
            "text": ";"
          }
        ]
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 112
}
```
