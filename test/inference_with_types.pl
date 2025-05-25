#!/usr/bin/env perl
use strict;
use warnings;

# Mix of explicit type annotations and inferred types

# Explicit type annotation
my Int $count = 42;

# Inferred from assignment and usage
my $price = 19.99;
my $total = $price * $count;  # Should infer Num for $total

# Function with partial type information
sub calculate_discount {
    my ($amount, $percentage) = @_;  # Types inferred from usage

    # $percentage used in numeric context
    my $discount = $amount * ($percentage / 100);

    return $discount;  # Return type inferred as Num
}

# Using the function
my $final_price = $total - calculate_discount($total, 10);

# Complex data structure with mixed inference
my $order = {
    items => [              # Inferred as ArrayRef
        { name => 'Widget', price => 9.99 },   # HashRef elements
        { name => 'Gadget', price => 19.99 },
    ],
    customer => 'John Doe',  # Inferred as Str
    paid => 0,              # Inferred as Int/Num
};

# Method calls help infer object types
sub process_order {
    my ($order_ref) = @_;  # Inferred as HashRef from usage

    my $total = 0;
    foreach my $item (@{$order_ref->{items}}) {  # Confirms HashRef with ArrayRef value
        $total += $item->{price};
    }

    $order_ref->{total} = $total;
    $order_ref->{paid} = 1 if $total > 0;

    return $order_ref;
}

# Context-sensitive inference
my @items = @{$order->{items}};  # List context
my $item_count = @items;          # Scalar context - infers Int

# Pattern matching for type refinement
sub validate_email {
    my ($email) = @_;  # Inferred as Str from regex usage

    if ($email =~ /^[\w\.\-]+@[\w\.\-]+\.\w+$/) {
        return 1;
    }
    return 0;
}

# Using typed variables in expressions
my Str $domain = 'example.com';
my $email = 'user@' . $domain;  # $email inferred as Str

print "Valid email\n" if validate_email($email);

1;
