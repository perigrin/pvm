# ABOUTME: Fixture for the psc check e2e test covering undef in arithmetic.
# ABOUTME: An EXPLICIT undef assignment - a bare `my $x;` cannot be proven undef.

# perl itself warns here: "Use of uninitialized value $x in addition (+)".
my $x = undef;
my $y = $x + 1;
