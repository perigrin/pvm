---
category: untyped-perl
subcategory: packages
tags:
    - basic
    - declarations
    - namespaces
    - packages
---

# Basic Package Declarations

Test basic package declarations and namespace changes

```perl
package MyPackage;
package MyPackage::Subspace;
package MyPackage 1.23;
package Local::Test::Package;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
package MyPackage;
package MyPackage::Subspace;
package MyPackage 1.23;
package Local::Test::Package;
```

## Typed Perl Output

```perl
package MyPackage;
package MyPackage::Subspace;
package MyPackage 1.23;
package Local::Test::Package;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
source_file
  package_statement
    expression_stmt
      literal
    expression_stmt
      literal
    token
  package_statement
    expression_stmt
      literal
    expression_stmt
      literal
    token
  package_statement
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
    token
```

## JSON AST

```json
{
  "root": {
    "type": "source_file",
    "children": [
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
                "value": "MyPackage::Subspace",
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
            "type": "token",
            "text": "1.23"
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
                "value": "Local::Test::Package",
                "kind": "string"
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
