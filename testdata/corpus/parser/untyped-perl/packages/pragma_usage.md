---
category: untyped-perl
subcategory: packages
tags:
    - pragmas
    - strict
    - warnings
    - features
    - utf8
    - constant
---

# Pragma Usage

Test pragma usage patterns (strict, warnings, etc.)

```perl
use strict;
use warnings;
use strict 'vars';
use warnings 'all';
no strict 'refs';
no warnings 'uninitialized';
use feature qw(say switch);
use utf8;
use constant PI => 3.14159;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
use strict;
use warnings;
use strict 'vars';
use warnings 'all';
no strict 'refs';
no warnings 'uninitialized';
use feature qw(say switch);
use utf8;
use constant PI => 3.14159;
```

## Typed Perl Output

```perl
use strict;
use warnings;
use strict 'vars';
use warnings 'all';
no strict 'refs';
no warnings 'uninitialized';
use feature qw(say switch);
use utf8;
use constant PI => 3.14159;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
source_file
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
  use_statement
    expression_stmt
      literal
    expression_stmt
      literal
    token
    token
  use_statement
    expression_stmt
      literal
    expression_stmt
      literal
    token
    token
  use_statement
    expression_stmt
      literal
    expression_stmt
      literal
    token
    token
  use_statement
    expression_stmt
      literal
    expression_stmt
      literal
    token
    token
  use_statement
    expression_stmt
      literal
    expression_stmt
      literal
    quoted_word_list
      expression_stmt
        literal
      expression_stmt
        literal
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
    list_expression
      expression_stmt
        literal
      expression_stmt
        literal
      token
    token
```

## JSON AST

```json
{
  "root": {
    "type": "source_file",
    "children": [
      {
        "type": "use_statement",
        "children": [
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "use",
                "kind": "string"
              }
            ]
          },
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "strict",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "text": ";"
          }
        ]
      },
      {
        "type": "use_statement",
        "children": [
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "use",
                "kind": "string"
              }
            ]
          },
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "warnings",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "text": ";"
          }
        ]
      },
      {
        "type": "use_statement",
        "children": [
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "use",
                "kind": "string"
              }
            ]
          },
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "strict",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "text": "'vars'"
          },
          {
            "type": "token",
            "text": ";"
          }
        ]
      },
      {
        "type": "use_statement",
        "children": [
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "use",
                "kind": "string"
              }
            ]
          },
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "warnings",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "text": "'all'"
          },
          {
            "type": "token",
            "text": ";"
          }
        ]
      },
      {
        "type": "use_statement",
        "children": [
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "no",
                "kind": "string"
              }
            ]
          },
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "strict",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "text": "'refs'"
          },
          {
            "type": "token",
            "text": ";"
          }
        ]
      },
      {
        "type": "use_statement",
        "children": [
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "no",
                "kind": "string"
              }
            ]
          },
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "warnings",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "text": "'uninitialized'"
          },
          {
            "type": "token",
            "text": ";"
          }
        ]
      },
      {
        "type": "use_statement",
        "children": [
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "use",
                "kind": "string"
              }
            ]
          },
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "feature",
                "kind": "string"
              }
            ]
          },
          {
            "type": "quoted_word_list",
            "children": [
              {
                "type": "expression_stmt",
                "children": [
                  {
                    "type": "literal",
                    "value": "qw",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "expression_stmt",
                "children": [
                  {
                    "type": "literal",
                    "value": "(",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "expression_stmt",
                "children": [
                  {
                    "type": "literal",
                    "value": "say switch",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "expression_stmt",
                "children": [
                  {
                    "type": "literal",
                    "value": ")",
                    "kind": "string"
                  }
                ]
              }
            ]
          },
          {
            "type": "token",
            "text": ";"
          }
        ]
      },
      {
        "type": "use_statement",
        "children": [
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "use",
                "kind": "string"
              }
            ]
          },
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "utf8",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "text": ";"
          }
        ]
      },
      {
        "type": "use_statement",
        "children": [
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "use",
                "kind": "string"
              }
            ]
          },
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "constant",
                "kind": "string"
              }
            ]
          },
          {
            "type": "list_expression",
            "children": [
              {
                "type": "expression_stmt",
                "children": [
                  {
                    "type": "literal",
                    "value": "PI",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "expression_stmt",
                "children": [
                  {
                    "type": "literal",
                    "value": "=>",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "text": "3.14159"
              }
            ]
          },
          {
            "type": "token",
            "text": ";"
          }
        ]
      }
    ]
  },
  "type_annotations": [],
  "errors": []
}
```
