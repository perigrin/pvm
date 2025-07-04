---
category: improved-parsing
subcategory: recovered-errors
tags: [error-recovery, improved-grammar]
---

# Missing Closing Bracket

<!-- should_error: false -->
<!-- Note: Parser now recovers from missing closing bracket in parameterized types -->

```perl
my ArrayRef[Int $var;
```

## Invalid Parameterized Space

<!-- should_error: false -->
<!-- Note: Parser now accepts spaces in parameterized types -->

```perl
my ArrayRef[ Int] $spaced;
```
