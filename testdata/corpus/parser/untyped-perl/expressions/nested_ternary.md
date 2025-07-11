---
category: untyped-perl
subcategory: expressions
tags:
    - nested
    - ternary
    - conditional
    - multiple
---

# Nested Ternary

Nested ternary operators for multiple conditions

```perl
$grade = $score >= 90 ? 'A' : $score >= 80 ? 'B' : $score >= 70 ? 'C' : 'F';
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
$grade = $score >= 90 ? 'A' : $score >= 80 ? 'B' : $score >= 70 ? 'C' : 'F';
```

### Typed Perl Output

```perl
$grade = $score >= 90 ? 'A' : $score >= 80 ? 'B' : $score >= 70 ? 'C' : 'F';
```

### Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
