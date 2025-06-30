---
category: compiler
subcategory: basic-functions
tags:
    - function-signatures
    - typed-parameters
    - clean-perl-output
type_check: false
---

# Simple Function Signature

Basic function with typed parameters

```perl
sub add (Int $a, Int $b) -> Int {
    return $a + $b;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
sub add ($a, $b)  {
    return $a + $b;
}
```
