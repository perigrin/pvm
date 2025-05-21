#!/usr/bin/perl

my Str $name = "World";
my Int $count = 42;

sub greet(Str $name) -> Str {
    return "Hello, $name!";
}

print greet($name) . "\n";
