---
category: untyped-perl
subcategory: packages
tags:
    - basic
    - declarations
    - namespaces
    - packages
---

# Basic Package Declarations

Test basic package declarations and namespace changes

```perl
package MyPackage;
package MyPackage::Subspace;
package MyPackage 1.23;
package Local::Test::Package;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
package MyPackage;
package MyPackage::Subspace;
package MyPackage 1.23;
package Local::Test::Package;
```

## Typed Perl Output

```perl
package MyPackage;
package MyPackage::Subspace;
package MyPackage 1.23;
package Local::Test::Package;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
