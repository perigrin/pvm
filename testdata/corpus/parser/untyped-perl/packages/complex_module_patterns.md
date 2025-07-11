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
