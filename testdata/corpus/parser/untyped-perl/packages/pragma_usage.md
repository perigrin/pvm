---
category: untyped-perl
subcategory: packages
tags:
    - pragmas
    - strict
    - warnings
    - features
    - utf8
    - constant
---

# Pragma Usage

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

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
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

## Typed Perl Output

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

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
