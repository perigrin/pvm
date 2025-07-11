---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - postfix
    - do_until
---

# Postfix Until

Do-until loop (postfix until)

```perl
do {
    attempt();
} until ($success);
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
do {
    attempt();
} until ($success);
```

## Typed Perl Output

```perl
do {
    attempt();
} until ($success);
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
