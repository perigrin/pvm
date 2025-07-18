use v5.36;
use strict;
use warnings;

# Basic typed variables
my Int $count = 42;
my Str $name = "example";
my Bool $active = 1;

# Simple function with types
sub Str process_user(Int $id, Str $name) {
    return "User $id: $name";
}

# Simple usage
my Str $result = process_user($count, $name);
print $result . "\n";
EOF < /dev/null
