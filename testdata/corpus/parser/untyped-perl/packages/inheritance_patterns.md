---
category: untyped-perl
subcategory: packages
tags:
    - inheritance
    - parent
    - base
    - isa
    - multiple
---

# Inheritance Patterns

Test inheritance and parent module patterns

```perl
use parent 'BaseClass';
use base qw(Base1 Base2);
use parent qw(Parent::Class Another::Parent);
our @ISA = qw(BaseClass);
push @ISA, 'Mixin::Class';
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
use parent 'BaseClass';
use base qw(Base1 Base2);
use parent qw(Parent::Class Another::Parent);
our @ISA = qw(BaseClass);
push @ISA, 'Mixin::Class';
```

## Typed Perl Output

```perl
use parent 'BaseClass';
use base qw(Base1 Base2);
use parent qw(Parent::Class Another::Parent);
our @ISA = qw(BaseClass);
push @ISA, 'Mixin::Class';
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
  expression_statement
    var_decl
      variable
  token
  expression_statement
    ambiguous_function_call_expression
      expression_stmt
        literal
      list_expression
        array
          expression_stmt
            literal
          token
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
                "value": "base",
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
                    "value": "Base1 Base2",
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
                "value": "parent",
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
                    "value": "Parent::Class Another::Parent",
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
        "type": "expression_statement",
        "children": [
          {
            "type": "var_decl",
            "decl_type": "our",
            "children": [
              {
                "type": "variable",
                "name": "ISA",
                "sigil": "@"
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
            "type": "ambiguous_function_call_expression",
            "children": [
              {
                "type": "expression_stmt",
                "children": [
                  {
                    "type": "literal",
                    "value": "push",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "list_expression",
                "children": [
                  {
                    "type": "array",
                    "children": [
                      {
                        "type": "expression_stmt",
                        "children": [
                          {
                            "type": "literal",
                            "value": "@",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "token",
                        "text": "ISA"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "children": [
                      {
                        "type": "literal",
                        "value": ",",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "text": "'Mixin::Class'"
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
      }
    ]
  },
  "type_annotations": [],
  "errors": []
}
```
