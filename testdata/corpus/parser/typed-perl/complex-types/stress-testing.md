---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - stress-testing
    - performance
    - deep-nesting
    - many-alternatives
    - complex-types
---

# Stress Testing

Stress testing with very deep nesting and many union alternatives

```perl
my Map[Str, ArrayRef[HashRef[Tuple[Int, Result[Data[UserInfo[Profile[Settings[Theme]]]], Error[ValidationFailure[FieldError[TypeError]]]]]]] %extremely_nested;
my (A|B|C|D|E|F|G|H|I|J|K|L|M|N|O|P|Q|R|S|T) $many_union_alternatives;
my Container[Wrapper[Inner[Core[Data[Value[Item[Element[Node[Leaf]]]]]]]]] $very_deep;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my %extremely_nested;
my $many_union_alternatives;
my $very_deep;
```

## Typed Perl Output

```perl
my Map[Str, ArrayRef[HashRef[Tuple[Int, Result[Data[UserInfo[Profile[Settings[Theme]]]], Error[ValidationFailure[FieldError[TypeError]]]]]]] %extremely_nested;
my (A|B|C|D|E|F|G|H|I|J|K|L|M|N|O|P|Q|R|S|T) $many_union_alternatives;
my Container[Wrapper[Inner[Core[Data[Value[Item[Element[Node[Leaf]]]]]]]]] $very_deep;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

(none)
