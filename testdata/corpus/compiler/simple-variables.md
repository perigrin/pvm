---
category: compiler
subcategory: simple-variables
tags:
    - typed-variables
    - clean-perl-output
type_check: false
---

# Simple Typed Variable

Basic typed variable declaration compilation

```perl
my Int $count = 42;
print "Count: $count\n";
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $count = 42;
print "Count: $count\n";
```

# Simple String Variable

String typed variable compilation

```perl
my Str $name = "hello";
print "Name: $name\n";
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $name = "hello";
print "Name: $name\n";
```
