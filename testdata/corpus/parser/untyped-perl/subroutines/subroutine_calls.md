---
category: untyped-perl
subcategory: subroutines
tags:
    - calls
    - arguments
    - parentheses
    - ampersand
    - nested
---

# Subroutine Calls

Test various subroutine call patterns and argument passing

```perl
my $result = simple_sub();
my $sum = with_params(10, 20);
my $doubled = $code_ref->(5);

# Calls without parentheses
my $no_parens = simple_sub;
my $with_args = function_call 'arg1', 'arg2';

# Calls with complex arguments
my $complex = process($var, @array, %hash);
my $nested = outer(inner(42), another());

# Ampersand calls
my $amp_call = &function_name;
my $amp_with_args = &other_function(@args);
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $result = simple_sub();
my $sum = with_params(10, 20);
my $doubled = $code_ref->(5);

# Calls without parentheses
my $no_parens = simple_sub;
my $with_args = function_call 'arg1', 'arg2';

# Calls with complex arguments
my $complex = process($var, @array, %hash);
my $nested = outer(inner(42), another());

# Ampersand calls
my $amp_call = &function_name;
my $amp_with_args = &other_function(@args);
```

## Typed Perl Output

```perl
my $result = simple_sub();
my $sum = with_params(10, 20);
my $doubled = $code_ref->(5);

# Calls without parentheses
my $no_parens = simple_sub;
my $with_args = function_call 'arg1', 'arg2';

# Calls with complex arguments
my $complex = process($var, @array, %hash);
my $nested = outer(inner(42), another());

# Ampersand calls
my $amp_call = &function_name;
my $amp_with_args = &other_function(@args);
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
