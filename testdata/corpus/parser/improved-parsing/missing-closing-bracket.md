---
category: improved-parsing
subcategory: recovered-errors
tags: [error-recovery, improved-grammar, missing-bracket]
---

# Missing Closing Bracket

<!-- should_error: false -->
<!-- Note: Parser now recovers from missing closing bracket in parameterized types -->

```perl
my ArrayRef[Int $var;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $var;
```

## Typed Perl Output

```perl
use v5.36;
my ArrayRef[Int $var;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
