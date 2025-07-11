---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - given
    - when
    - switch
    - no_default
---

# Given No Default

Given-when without default clause

```perl
given ($option) {
    when ('verbose') { $verbose = 1; }
    when ('quiet') { $quiet = 1; }
    when ('debug') { $debug = 1; }
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
given ($option) {
    when ('verbose') { $verbose = 1; }
    when ('quiet') { $quiet = 1; }
    when ('debug') { $debug = 1; }
}
```

## Typed Perl Output

```perl
given ($option) {
    when ('verbose') { $verbose = 1; }
    when ('quiet') { $quiet = 1; }
    when ('debug') { $debug = 1; }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
