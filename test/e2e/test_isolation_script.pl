#!/usr/bin/env perl
use strict;
use warnings;

# Print environment variables
print "PERL5LIB: $ENV{PERL5LIB}\n";
print "HOME: $ENV{HOME}\n";
print "Working directory: " . qx(pwd) . "\n";

# Create a test file to check filesystem isolation
open(my $fh, '>', 'test_file.txt') or die "Could not create file: $!";
print $fh "This is a test file\n";
close($fh);

print "Created test file: test_file.txt\n";

# Try to read a file from parent directory
if (open(my $fh, '<', '../test_isolation_script.pl')) {
    print "Could read parent directory file\n";
    close($fh);
} else {
    print "Could not read parent directory file: $!\n";
}

print "Script completed successfully\n";
