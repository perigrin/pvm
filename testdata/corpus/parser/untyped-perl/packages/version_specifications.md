---
category: untyped-perl
subcategory: packages
tags:
    - versions
    - perl
    - modules
    - specifications
---

# Version Specifications

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
use v5.36;
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
