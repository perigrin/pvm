---
category: untyped-perl
subcategory: subroutines
tags:
    - references
    - code_references
    - dereferencing
    - anonymous
    - builtin
---

# Subroutine References

Test subroutine references and code reference handling

```perl
# Subroutine references
my $sub_ref = \&function_name;
my $qualified_ref = \&Package::function;

# Code reference calls
my $result1 = $sub_ref->();
my $result2 = $sub_ref->(1, 2, 3);
my $result3 = &$sub_ref;
my $result4 = &{$sub_ref}(42);

# Anonymous subroutine references
my $anon_ref = sub { return shift * 2; };
my $doubled = $anon_ref->(21);

# References to built-in functions
my $print_ref = \&print;
my $map_ref = \&map;

# Complex reference operations
my @code_refs = (\&func1, \&func2, sub { 'anon' });
my $selected = $;
$selected->();
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
# Subroutine references
my $sub_ref = \&function_name;
my $qualified_ref = \&Package::function;

# Code reference calls
my $result1 = $sub_ref->();
my $result2 = $sub_ref->(1, 2, 3);
my $result3 = &$sub_ref;
my $result4 = &{$sub_ref}(42);

# Anonymous subroutine references
my $anon_ref = sub { return shift * 2; };
my $doubled = $anon_ref->(21);

# References to built-in functions
my $print_ref = \&print;
my $map_ref = \&map;

# Complex reference operations
my @code_refs = (\&func1, \&func2, sub { 'anon' });
my $selected = $;
$selected->();
```

## Typed Perl Output

```perl
# Subroutine references
my $sub_ref = \&function_name;
my $qualified_ref = \&Package::function;

# Code reference calls
my $result1 = $sub_ref->();
my $result2 = $sub_ref->(1, 2, 3);
my $result3 = &$sub_ref;
my $result4 = &{$sub_ref}(42);

# Anonymous subroutine references
my $anon_ref = sub { return shift * 2; };
my $doubled = $anon_ref->(21);

# References to built-in functions
my $print_ref = \&print;
my $map_ref = \&map;

# Complex reference operations
my @code_refs = (\&func1, \&func2, sub { 'anon' });
my $selected = $;
$selected->();
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
