#!perl
# `!~` is not a match op with a flag: it is `match` followed by `not`, a
# second op stacked on top of the first.
#
# TIER 09 regex
# INTRODUCES the `!~` negated binding operator
# USES not, from tier 04
#
# `not` is tier 04's op, claimed there with the logical operators, and its
# appearance here is the negated binding REUSING it rather than this tier
# introducing anything. The measurement is still this tier's: reading the
# source would suggest a negated match op, and there is none.
#
# The pattern deliberately does not match, so the printed value is the
# empty string rather than 0 -- Perl's false is "" -- which is why the
# brackets are there. Without them the expected output would be a line
# with nothing on it and a mis-measured match would print the same.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $s = "abc"; my $m = $s !~ /zzz/; print "[$m]\n"'
#   [1]

--- source
my $s = "abc";
my $m = $s !~ /zzz/;
print "[$m]\n";

--- expect output
[1]

--- expect parses

--- expect tokens
one operator whose text is "!~"
no operator whose text is "!"
no operator whose text is "/"
