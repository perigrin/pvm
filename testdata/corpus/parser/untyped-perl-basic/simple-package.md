---
category: untyped-perl
subcategory: packages
tags:
    - packages
    - modules
    - use-statements
---

# Simple Package

Basic package declaration and use statements

```perl
package MyModule;
use strict;
use warnings;

sub new {
    my $class = shift;
    return bless {}, $class;
}

1;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
package MyModule;
use strict;
use warnings;

sub new {
    my $class = shift;
    return bless {}, $class;
}

1;
```

## Typed Perl Output

```perl
package MyModule;
use strict;
use warnings;

sub new {
    my $class = shift;
    return bless {}, $class;
}

1;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
