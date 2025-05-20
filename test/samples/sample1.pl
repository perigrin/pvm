#!/usr/bin/env perl
use v5.40;
use experimental 'class';

# Simple variable type annotations
my Str $name = "John Doe";
my Int $age = 30;
our HashRef[Str] $config = { foo => "bar" };
state ArrayRef[Int] $visited_ids;

# Function with type annotations
sub add(Int $a, Int $b) -> Int {
    return $a + $b;
}

# Function with array parameter and return value
sub get_names(ArrayRef[User] $users) -> ArrayRef[Str] {
    my ArrayRef[Str] $names = [];

    foreach my User $user (@$users) {
        push @$names, $user->name;
    }

    return $names;
}

# Class with type annotations
class User {
    field Str $name :param;
    field Int $age :param;
    field EmailAddress $email :param;

    method display() -> Str {
        return sprintf("User %s (age: %d, email: %s)", $self->name, $self->age, $self->email);
    }

    method is_adult() -> Bool {
        return $self->age >= 18;
    }
}

# Usage of union types
sub maybe_get_user(Int $id) -> User|Undef {
    if ($id > 0 && $id < 100) {
        return User->new(
            name => "User $id",
            age => 20 + ($id % 10),
            email => "user$id@example.com"
        );
    }

    return undef;
}

# Main code
my User $user = User->new(
    name => $name,
    age => $age,
    email => "john.doe@example.com"
);

my User|Undef $maybe_user = maybe_get_user(42);

if ($maybe_user) {
    say $maybe_user->display();
} else {
    say "No user found";
}

my Int $result = add(5, 10);
say "The result is: $result";

my ArrayRef[User] $users = [
    User->new(name => "Alice", age => 25, email => "alice@example.com"),
    User->new(name => "Bob", age => 30, email => "bob@example.com"),
    User->new(name => "Charlie", age => 35, email => "charlie@example.com"),
];

my ArrayRef[Str] $user_names = get_names($users);
say "User names: " . join(", ", @$user_names);
