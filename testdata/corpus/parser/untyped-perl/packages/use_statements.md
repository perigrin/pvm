---
category: untyped-perl
subcategory: packages
tags:
    - use
    - import
    - modules
    - features
    - versions
---

# Use Statements

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

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
use strict;
use warnings;
use Data::Dumper;
use MyModule qw(function1 function2);
use AnotherModule 1.5 qw(:all);
use Parent::Module ();
use feature 'say';
```

## Typed Perl Output

```perl
use strict;
use warnings;
use Data::Dumper;
use MyModule qw(function1 function2);
use AnotherModule 1.5 qw(:all);
use Parent::Module ();
use feature 'say';
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
