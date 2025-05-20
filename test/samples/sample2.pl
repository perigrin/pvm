#!/usr/bin/env perl
use v5.40;
use experimental 'class';

# This sample has a type error (assigning a string to an Int variable)

# Simple variable type annotations
my Str $name = "John Doe";
my Int $age = "thirty"; # Type error: assigning a string to an Int variable

# Function with type annotations
sub add(Int $a, Int $b) -> Int {
    return $a + $b;
}

# Function with array parameter but return type mismatch
sub get_names(ArrayRef[User] $users) -> ArrayRef[Str] {
    # Return incorrect type
    return $users; # Type error: returning ArrayRef[User] instead of ArrayRef[Str]
}

# Class with type annotations
class User {
    field Str $name :param;
    field Int $age :param;

    method display() -> Str {
        return sprintf("User %s (age: %d)", $self->name, $self->age);
    }

    method update_age(Str $new_age) { # Type error: using Str for age instead of Int
        $self->{age} = $new_age;
    }
}

# Usage with errors
my User $user = User->new(
    name => $name,
    age => $age,
);

say $user->display();

# Invalid parameters to add
my Str $str_value = "5";
my Int $result = add($str_value, 10); # Type error: passing a Str to an Int parameter

say "The result is: $result";
