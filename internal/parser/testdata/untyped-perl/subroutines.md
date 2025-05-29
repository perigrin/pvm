---
category: untyped-perl
subcategory: subroutines
tags:
    - anonymous
    - arguments
    - arrow_notation
    - attributes
    - basic
    - calls
    - code_references
    - complex
    - definitions
    - dereferencing
    - edge_cases
    - methods
    - objects
    - prototypes
    - references
    - subroutines
---

# Anonymous Subroutines

Test anonymous subroutine definitions and code references

```perl
my $code_ref = sub {
    my $param = shift;
    return $param * 2;
};

my $closure = sub {
    my $x = shift;
    return sub { return $x + shift };
};

my $anon_with_params = sub {
    my ($a, $b) = @_;
    return $a . $b;
};
```

## Basic Subroutine Definitions

Test basic subroutine definitions and simple subroutines

```perl
sub simple_sub {
    return "result";
}

sub with_params {
    my ($first, $second) = @_;
    return $first + $second;
}

sub no_body;

sub Package::qualified_sub {
    return 42;
}
```

## Complex Subroutines

Test complex subroutine patterns and edge cases

```perl
# Nested subroutine definitions
sub outer_function {
    my $param = shift;
    
    my $inner = sub {
        return $param * 2;
    };
    
    return $inner;
}

# Recursive subroutines
sub factorial {
    my $n = shift;
    return $n <= 1 ? 1 : $n * factorial($n - 1);
}

# Subroutines with complex parameter handling
sub complex_params {
    my ($first, @rest) = @_;
    my %options = @rest;
    return process($first, \%options);
}

# Subroutines returning different types
sub context_sensitive {
    return wantarray ? ('list', 'result') : 'scalar result';
}

# Forward declarations
sub forward_declared;

sub uses_forward {
    return forward_declared(42);
}

sub forward_declared {
    my $value = shift;
    return $value * 3;
}
```

## Method Calls

Test method calls and arrow notation for object-oriented patterns

```perl
# Object method calls
my $object = Package->new();
$object->method($arg1, $arg2);
my $result = $obj->process()->transform();

# Class method calls
Package::function($args);
My::Module->class_method(@parameters);

# Chained method calls
my $chained = $obj->first()->second()->third();

# Method calls on complex expressions
my $complex = get_object()->method();
$hash->{key}->process();
$array->[0]->handle();
```

## Prototypes And Attributes

Test subroutine prototypes and attributes

```perl
# Prototypes
sub prototype_sub ($$) {
    my ($a, $b) = @_;
    return $a + $b;
}

sub arrayref_proto (\@) {
    my $arrayref = shift;
    return @$arrayref;
}

sub optional_proto ($;$) {
    my ($required, $optional) = @_;
    return defined $optional ? $required + $optional : $required;
}

# Attributes
sub attributed_sub : lvalue {
    return $global_var;
}

sub multi_attr : method lvalue {
    return $_[0]->{value};
}

sub const_sub : const {
    return 42;
}
```

## Subroutine Calls

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

## Subroutine References

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
my $selected = $code_refs[1];
$selected->();
```
