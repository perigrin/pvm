---
category: untyped-perl
subcategory: subroutines
tags:
    - methods
    - objects
    - arrow_notation
    - chained
    - calls
---

# Method Calls

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

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
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

## Typed Perl Output

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

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```
