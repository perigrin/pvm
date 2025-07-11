---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - postfix
    - do_while
---

# Postfix While

Do-while loop (postfix while)

```perl
do {
    process();
} while ($continue);
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
do {
    process();
} while ($continue);
```

## Typed Perl Output

```perl
do {
    process();
} while ($continue);
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
