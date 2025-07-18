---
category: untyped-perl
subcategory: packages
tags:
    - versions
    - perl
    - modules
    - specifications
---

# Version Specifications

Test version specifications in use statements and package declarations

```perl
use 5.010;
use v5.12;
use 5.020;
use MyModule v1.2.3;
use AnotherModule 2.5;
package MyPackage v1.0.0;
use perl 5.032;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
use 5.010;
use v5.12;
use 5.020;
use MyModule v1.2.3;
use AnotherModule 2.5;
package MyPackage v1.0.0;
use perl 5.032;
```

## Typed Perl Output

```perl
use 5.010;
use v5.12;
use 5.020;
use MyModule v1.2.3;
use AnotherModule 2.5;
package MyPackage v1.0.0;
use perl 5.032;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
source_file
  use_version_statement
    expression_stmt
      literal
    token
    token
  use_version_statement
    expression_stmt
      literal
    expression_stmt
      literal
    token
  use_version_statement
    expression_stmt
      literal
    token
    token
  use_statement
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
    token
  package_statement
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
    token
```

## JSON AST

```json
{
  "root": {
    "type": "source_file",
    "children": [
      {
        "type": "use_version_statement",
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
            "type": "token",
            "text": "5.010"
          },
          {
            "type": "token",
            "text": ";"
          }
        ]
      },
      {
        "type": "use_version_statement",
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
                "value": "v5.12",
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
        "type": "use_version_statement",
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
            "type": "token",
            "text": "5.020"
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
                "value": "MyModule",
                "kind": "string"
              }
            ]
          },
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "v1.2.3",
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
                "value": "AnotherModule",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "text": "2.5"
          },
          {
            "type": "token",
            "text": ";"
          }
        ]
      },
      {
        "type": "package_statement",
        "children": [
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "package",
                "kind": "string"
              }
            ]
          },
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "MyPackage",
                "kind": "string"
              }
            ]
          },
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "v1.0.0",
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
                "value": "perl",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "text": "5.032"
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
