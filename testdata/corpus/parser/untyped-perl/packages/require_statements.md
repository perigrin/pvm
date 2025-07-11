---
category: untyped-perl
subcategory: packages
tags:
    - require
    - dynamic
    - loading
    - versions
    - modules
---

# Require Statements

Test require statements and dynamic module loading

```perl
require DynamicModule;
require 'module.pl';
require v5.10;
require 5.010;
require Module::Name;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
require DynamicModule;
require 'module.pl';
require v5.10;
require 5.010;
require Module::Name;
```

## Typed Perl Output

```perl
require DynamicModule;
require 'module.pl';
require v5.10;
require 5.010;
require Module::Name;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
