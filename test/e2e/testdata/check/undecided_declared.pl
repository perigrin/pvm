# ABOUTME: Fixture whose call to My::Lib's context multi when_ is a sub's return, a context the call site cannot decide.
# ABOUTME: Its candidates are joined, and perl -w runs it silently in both calling contexts.
use strict;
use warnings;
use My::Lib;

sub pieces { return when_() }

my @parts = pieces();
my $one = pieces();
print scalar(@parts), " $one\n";
