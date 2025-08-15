---
category: untyped-perl
subcategory: packages
tags:
    - complex
    - modules
    - inheritance
    - parent
    - conditional
    - loading
---

# Complex Module Patterns

Test complex module loading and package organization patterns

```perl
{
    package LocalPackage;
    use parent 'BaseClass';

    sub new {
        my $class = shift;
        return bless {}, $class;
    }
}

package main;
use LocalPackage;

# Module with conditional loading
BEGIN {
    if ($ENV{DEBUG}) {
        require Data::Dumper;
        Data::Dumper->import('Dumper');
    }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
{
    package LocalPackage;
    use parent 'BaseClass';

    sub new {
        my $class = shift;
        return bless {}, $class;
    }
}

package main;
use LocalPackage;

# Module with conditional loading
BEGIN {
    if ($ENV{DEBUG}) {
        require Data::Dumper;
        Data::Dumper->import('Dumper');
    }
}
```

## Typed Perl Output

```perl
{
    package LocalPackage;
    use parent 'BaseClass';

    sub new {
        my $class = shift;
        return bless {}, $class;
    }
}

package main;
use LocalPackage;

# Module with conditional loading
BEGIN {
    if ($ENV{DEBUG}) {
        require Data::Dumper;
        Data::Dumper->import('Dumper');
    }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
source_file
  block_statement
    token
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
      token
    sub_decl
      block_stmt
        token
        return_stmt
          literal
        token
        return_stmt
          literal
        token
        token
    token
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
  expression_stmt
    literal
  phaser_statement
    expression_stmt
      literal
    block_stmt
      token
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
        "type": "block_statement",
        "children": [
          {
            "type": "token",
            "text": "{"
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
                    "value": "LocalPackage",
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
                    "value": "parent",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "text": "'BaseClass'"
              },
              {
                "type": "token",
                "text": ";"
              }
            ]
          },
          {
            "type": "sub_decl",
            "name": "new",
            "children": [
              {
                "type": "block_stmt",
                "children": [
                  {
                    "type": "token",
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "children": [
                      {
                        "type": "literal",
                        "value": "my $class = shift",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "text": ";"
                  },
                  {
                    "type": "expression_stmt",
                    "children": [
                      {
                        "type": "literal",
                        "value": "return bless {}, $class",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "text": "}"
                  }
                ]
              }
            ]
          },
          {
            "type": "token",
            "text": "}"
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
                "value": "main",
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
                "value": "LocalPackage",
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
        "type": "expression_stmt",
        "children": [
          {
            "type": "literal",
            "value": "# Module with conditional loading",
            "kind": "string"
          }
        ]
      },
      {
        "type": "phaser_statement",
        "children": [
          {
            "type": "expression_stmt",
            "children": [
              {
                "type": "literal",
                "value": "BEGIN",
                "kind": "string"
              }
            ]
          },
          {
            "type": "block_stmt",
            "children": [
              {
                "type": "token",
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "children": [
                  {
                    "type": "literal",
                    "value": "if ($ENV{DEBUG}) {\n        require Data::Dumper;\n        Data::Dumper->import('Dumper');\n    }",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "text": "}"
              }
            ]
          }
        ]
      }
    ]
  },
  "type_annotations": [],
  "errors": []
}
```
