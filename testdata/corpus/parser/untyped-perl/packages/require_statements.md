---
category: untyped-perl
subcategory: packages
tags:
    - require
    - dynamic
    - loading
    - versions
    - modules
---

# Require Statements

Test require statements and dynamic module loading

```perl
require DynamicModule;
require 'module.pl';
require v5.10;
require 5.010;
require Module::Name;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
require DynamicModule;
require 'module.pl';
require v5.10;
require 5.010;
require Module::Name;
```

## Typed Perl Output

```perl
require DynamicModule;
require 'module.pl';
require v5.10;
require 5.010;
require Module::Name;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
source_file
  expression_statement
    require_expression
      expression_stmt
        literal
      token
  token
  expression_statement
    require_expression
      expression_stmt
        literal
      token
  token
  expression_statement
    require_version_expression
      expression_stmt
        literal
      expression_stmt
        literal
  token
  expression_statement
    require_version_expression
      expression_stmt
        literal
      token
  token
  expression_statement
    require_expression
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
        "type": "expression_statement",
        "children": [
          {
            "type": "require_expression",
            "children": [
              {
                "type": "expression_stmt",
                "children": [
                  {
                    "type": "literal",
                    "value": "require",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "text": "DynamicModule"
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "text": ";"
      },
      {
        "type": "expression_statement",
        "children": [
          {
            "type": "require_expression",
            "children": [
              {
                "type": "expression_stmt",
                "children": [
                  {
                    "type": "literal",
                    "value": "require",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "text": "'module.pl'"
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "text": ";"
      },
      {
        "type": "expression_statement",
        "children": [
          {
            "type": "require_version_expression",
            "children": [
              {
                "type": "expression_stmt",
                "children": [
                  {
                    "type": "literal",
                    "value": "require",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "expression_stmt",
                "children": [
                  {
                    "type": "literal",
                    "value": "v5.10",
                    "kind": "string"
                  }
                ]
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "text": ";"
      },
      {
        "type": "expression_statement",
        "children": [
          {
            "type": "require_version_expression",
            "children": [
              {
                "type": "expression_stmt",
                "children": [
                  {
                    "type": "literal",
                    "value": "require",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "text": "5.010"
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "text": ";"
      },
      {
        "type": "expression_statement",
        "children": [
          {
            "type": "require_expression",
            "children": [
              {
                "type": "expression_stmt",
                "children": [
                  {
                    "type": "literal",
                    "value": "require",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "text": "Module::Name"
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
  "type_annotations": [],
  "errors": []
}
```
