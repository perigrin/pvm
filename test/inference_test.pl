#!/usr/bin/env perl
use strict;
use warnings;

# Test type inference from assignments
my $count = 42;                      # Should infer Int
my $price = 19.99;                   # Should infer Num
my $name = "Alice";                  # Should infer Str
my $data = { key => 'value' };       # Should infer HashRef
my $list = [1, 2, 3];                # Should infer ArrayRef

# Test type inference from usage patterns
push @$list, 4;                      # Confirms ArrayRef
my $size = @$list;                   # $size should be Int (scalar context)

my $key_count = keys %$data;         # Confirms HashRef, $key_count is Int
$data->{another} = 'test';           # Hash access

# Test type inference from operators
my $sum = $count + 10;               # Numeric operation, $sum is Int/Num
my $product = $price * 2;            # Numeric operation
my $greeting = "Hello, " . $name;    # String concatenation

# Test type inference from method calls
my $obj = SomeClass->new();          # Should infer SomeClass
$obj->do_something();                # Method call confirms object type

# Test type inference from built-in functions
my $length = length($name);          # String function, $length is Int
my $upper = uc($name);               # String function, $upper is Str
my $rounded = int($price);           # Numeric function, $rounded is Int

# Test type inference from regular expressions
if ($greeting =~ /Hello/) {          # Regex match confirms Str
    print "Match found\n";
}

# Test type inference from file operations
open my $fh, '<', 'test.txt';       # $fh should be FileHandle
close $fh;

# Test contextual inference
sub process_array {
    my ($array_ref) = @_;            # Parameter should infer ArrayRef from usage
    foreach my $item (@$array_ref) {
        print "$item\n";
    }
    return scalar @$array_ref;       # Return type should be Int
}

sub get_config {
    return {                         # Return type should be HashRef
        host => 'localhost',
        port => 8080,
    };
}

# Package for testing object inference
package SomeClass;

sub new {
    my $class = shift;
    return bless {}, $class;
}

sub do_something {
    my $self = shift;
    return "done";
}

1;
