---
category: untyped-perl
subcategory: packages
tags:
    - comprehensive
    - declarations
    - imports
    - qualification
    - versions
    - complex
---

# Comprehensive Package Test

Comprehensive test covering all package and module system features

```perl
#!/usr/bin/perl
use v5.20;
use strict;
use warnings;

# Package declarations
package MyPackage;
package MyPackage::Subspace;
package MyPackage 1.23;

# Module imports
use Data::Dumper;
use MyModule qw(function1 function2);
use AnotherModule 1.5 qw(:all);
require DynamicModule;

# Package qualification
$MyPackage::variable = "value";
MyPackage::function();
my $ref = \&MyPackage::function;

# Complex patterns
{
    package LocalPackage;
    use parent 'BaseClass';

    sub new {
        my $class = shift;
        return bless {}, $class;
    }
}

# Version specifications
use 5.010;
use MyModule v1.2.3;
package TestPackage v2.0.0;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
#!/usr/bin/perl
use v5.20;
use strict;
use warnings;

# Package declarations
package MyPackage;
package MyPackage::Subspace;
package MyPackage 1.23;

# Module imports
use Data::Dumper;
use MyModule qw(function1 function2);
use AnotherModule 1.5 qw(:all);
require DynamicModule;

# Package qualification
$MyPackage::variable = "value";
MyPackage::function();
my $ref = \&MyPackage::function;

# Complex patterns
{
    package LocalPackage;
    use parent 'BaseClass';

    sub new {
        my $class = shift;
        return bless {}, $class;
    }
}

# Version specifications
use 5.010;
use MyModule v1.2.3;
package TestPackage v2.0.0;
```

## Typed Perl Output

```perl
#!/usr/bin/perl
use v5.20;
use strict;
use warnings;

# Package declarations
package MyPackage;
package MyPackage::Subspace;
package MyPackage 1.23;

# Module imports
use Data::Dumper;
use MyModule qw(function1 function2);
use AnotherModule 1.5 qw(:all);
require DynamicModule;

# Package qualification
$MyPackage::variable = "value";
MyPackage::function();
my $ref = \&MyPackage::function;

# Complex patterns
{
    package LocalPackage;
    use parent 'BaseClass';

    sub new {
        my $class = shift;
        return bless {}, $class;
    }
}

# Version specifications
use 5.010;
use MyModule v1.2.3;
package TestPackage v2.0.0;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
source_file
  expression_stmt
    literal
  use_version_statement
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
  expression_stmt
    literal
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
  expression_stmt
    literal
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
    require_expression
      expression_stmt
        literal
      token
  token
  expression_stmt
    literal
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
  [... Additional nodes continue ...]
```

## JSON AST

```json
{
  "root": {
    "type": "source_file",
    "children": [
      {
        "type": "expression_stmt",
        "children": [
          {
            "type": "literal",
            "value": "#!/usr/bin/perl",
            "kind": "string"
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
                "value": "v5.20",
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
        "type": "expression_stmt",
        "children": [
          {
            "type": "literal",
            "value": "# Package declarations",
            "kind": "string"
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
        "type": "expression_stmt",
        "children": [
          {
            "type": "literal",
            "value": "# Module imports",
            "kind": "string"
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
                "value": "Data::Dumper",
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
                "value": "MyModule",
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
                    "value": "function1 function2",
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
      }
    ]
  },
  "type_annotations": [],
  "errors": []
}
```
