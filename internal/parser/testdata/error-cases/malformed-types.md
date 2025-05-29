---
category: error-cases
subcategory: malformed-types
tags: [error-recovery, syntax-errors, type-parsing]
---

# Missing Closing Bracket

<!-- should_error: true -->
<!-- expected_error: MissingClosingBracketError -->
<!-- expected_suggestion: Add closing ']' to complete the parameterized type -->

```perl
my ArrayRef[Int $var;
```

## Invalid Union Syntax

<!-- should_error: true -->
<!-- expected_error: InvalidUnionSyntaxError -->
<!-- expected_suggestion: Change '||' to '|' for union types -->

```perl
my Int||Str $bad_union;
```

## Incomplete Type Assertion

<!-- should_error: true -->
<!-- expected_error: IncompleteTypeAssertionError -->
<!-- expected_suggestion: Add the target type after 'as' keyword -->

```perl
my $val = $input as ;
```

## Invalid Parameterized Space

<!-- should_error: true -->
<!-- expected_error: InvalidParameterizedTypeError -->
<!-- expected_suggestion: Remove space after '[' in parameterized type -->

```perl
my ArrayRef[ Int] $spaced;
```

## Invalid Intersection Syntax

<!-- should_error: true -->
<!-- expected_error: InvalidIntersectionSyntaxError -->
<!-- expected_suggestion: Change '&&' to '&' for intersection types -->

```perl
my Object&&Serializable $obj;
```