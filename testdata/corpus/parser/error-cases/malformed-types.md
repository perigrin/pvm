---
category: error-cases
subcategory: malformed-types
tags: [error-recovery, syntax-errors, type-parsing]
---

# Invalid Union Syntax

<!-- should_error: true -->
<!-- expected_error: error[TSP003] -->
<!-- expected_suggestion: Change '||' to '|' for union types -->

```perl
my Int||Str $bad_union;
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```

### Typed Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```

### Inferred Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```

# Incomplete Type Assertion

<!-- should_error: true -->
<!-- expected_error: error[TSP004] -->
<!-- expected_suggestion: Add the target type after 'as' keyword -->

```perl
my $val = $input as ;
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```

### Typed Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```

### Inferred Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```

# Invalid Intersection Syntax

<!-- should_error: true -->
<!-- expected_error: error[TSP008] -->
<!-- expected_suggestion: Change '&&' to '&' for intersection types -->

```perl
my Object&&Serializable $obj;
```

## Expected Compilation Outcomes

### Clean Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```

### Typed Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```

### Inferred Perl Output

```perl
# Error: Invalid type syntax - parsing should fail
```
