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
