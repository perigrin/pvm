---
category: untyped-perl
subcategory: packages
tags:
    - base
    - blocks
    - complex
    - comprehensive
    - declarations
    - dynamic
    - features
    - import
    - inheritance
    - isa
    - modules
    - namespaces
    - organization
    - packages
    - parent
    - perl
    - pragmas
    - qualification
    - require
    - strict
    - use
    - variables
    - versions
    - warnings
---

# Basic Package Declarations

Test basic package declarations and namespace changes

```perl
package MyPackage;
package MyPackage::Subspace;
package MyPackage 1.23;
package Local::Test::Package;
```

## Complex Module Patterns

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

## Comprehensive Package Test

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

## Inheritance Patterns

Test inheritance and parent module patterns

```perl
use parent 'BaseClass';
use base qw(Base1 Base2);
use parent qw(Parent::Class Another::Parent);
our @ISA = qw(BaseClass);
push @ISA, 'Mixin::Class';
```

## Package Variables

Test package-qualified variable declarations and usage

```perl
$MyPackage::variable = "value";
@Package::array = (1, 2, 3);
%Other::Package::hash = (key => 'value');
our $Package::qualified;
my $ref = \&MyPackage::function;
$main::global = "in main";
```

## Pragma Usage

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

## Require Statements

Test require statements and dynamic module loading

```perl
require DynamicModule;
require 'module.pl';
require v5.10;
require 5.010;
require Module::Name;
```

## Use Statements

Test use statements with various import patterns

```perl
use strict;
use warnings;
use Data::Dumper;
use MyModule qw(function1 function2);
use AnotherModule 1.5 qw(:all);
use Parent::Module ();
use feature 'say';
```

## Version Specifications

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
