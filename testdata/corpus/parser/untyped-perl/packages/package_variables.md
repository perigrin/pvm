---
category: untyped-perl
subcategory: packages
tags:
    - variables
    - qualification
    - namespaces
    - references
    - global
---

# Package Variables

Test package-qualified variable declarations and usage

```perl
$MyPackage::variable = "value";
@Package::array = (1, 2, 3);
%Other::Package::hash = (key => 'value');
our $Package::qualified;
my $ref = \&MyPackage::function;
$main::global = "in main";
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
$MyPackage::variable = "value";
@Package::array = (1, 2, 3);
%Other::Package::hash = (key => 'value');
our $Package::qualified;
my $ref = \&MyPackage::function;
$main::global = "in main";
```

## Typed Perl Output

```perl
$MyPackage::variable = "value";
@Package::array = (1, 2, 3);
%Other::Package::hash = (key => 'value');
our $Package::qualified;
my $ref = \&MyPackage::function;
$main::global = "in main";
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
    assignment_expression
      scalar
        token
        token
      token
      interpolated_string_literal
        expression_stmt
          literal
        expression_stmt
          literal
        expression_stmt
          literal
  token
  expression_statement
    assignment_expression
      array
        expression_stmt
          literal
        token
      token
      token
      list_expression
        token
        expression_stmt
          literal
        token
        expression_stmt
          literal
        token
      token
  token
  expression_statement
    assignment_expression
      hash
        expression_stmt
          literal
        token
      token
      token
      list_expression
        expression_stmt
          literal
        expression_stmt
          literal
        token
      token
  token
  expression_statement
    variable_declaration
      expression_stmt
        literal
      scalar
        token
        token
  token
  expression_statement
    var_decl
      variable
  token
  expression_statement
    assignment_expression
      scalar
        token
        token
      token
      interpolated_string_literal
        expression_stmt
          literal
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
        "type": "expression_statement",
        "children": [
          {
            "type": "assignment_expression",
            "children": [
              {
                "type": "scalar",
                "children": [
                  {
                    "type": "token",
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "text": "MyPackage::variable"
                  }
                ]
              },
              {
                "type": "token",
                "text": "="
              },
              {
                "type": "interpolated_string_literal",
                "children": [
                  {
                    "type": "expression_stmt",
                    "children": [
                      {
                        "type": "literal",
                        "value": "\"",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "children": [
                      {
                        "type": "literal",
                        "value": "value",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "children": [
                      {
                        "type": "literal",
                        "value": "\"",
                        "kind": "string"
                      }
                    ]
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
            "type": "assignment_expression",
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
                    "text": "Package::array"
                  }
                ]
              },
              {
                "type": "token",
                "text": "="
              },
              {
                "type": "token",
                "text": "("
              },
              {
                "type": "list_expression",
                "children": [
                  {
                    "type": "token",
                    "text": "1"
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
                    "text": "2"
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
                    "text": "3"
                  }
                ]
              },
              {
                "type": "token",
                "text": ")"
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
            "type": "assignment_expression",
            "children": [
              {
                "type": "hash",
                "children": [
                  {
                    "type": "expression_stmt",
                    "children": [
                      {
                        "type": "literal",
                        "value": "%",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "text": "Other::Package::hash"
                  }
                ]
              },
              {
                "type": "token",
                "text": "="
              },
              {
                "type": "token",
                "text": "("
              },
              {
                "type": "list_expression",
                "children": [
                  {
                    "type": "expression_stmt",
                    "children": [
                      {
                        "type": "literal",
                        "value": "key",
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
                    "text": "'value'"
                  }
                ]
              },
              {
                "type": "token",
                "text": ")"
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
            "type": "variable_declaration",
            "children": [
              {
                "type": "expression_stmt",
                "children": [
                  {
                    "type": "literal",
                    "value": "our",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "scalar",
                "children": [
                  {
                    "type": "token",
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "text": "Package::qualified"
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
            "type": "var_decl",
            "decl_type": "my",
            "children": [
              {
                "type": "variable",
                "name": "ref",
                "sigil": "$"
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
            "type": "assignment_expression",
            "children": [
              {
                "type": "scalar",
                "children": [
                  {
                    "type": "token",
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "text": "main::global"
                  }
                ]
              },
              {
                "type": "token",
                "text": "="
              },
              {
                "type": "interpolated_string_literal",
                "children": [
                  {
                    "type": "expression_stmt",
                    "children": [
                      {
                        "type": "literal",
                        "value": "\"",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "children": [
                      {
                        "type": "literal",
                        "value": "in main",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "children": [
                      {
                        "type": "literal",
                        "value": "\"",
                        "kind": "string"
                      }
                    ]
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
