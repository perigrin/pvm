# ABOUTME: Fixture with genuine coercion mismatches for the psc check e2e test.
# ABOUTME: References in numeric and string positions - legal perl, almost never intended.

my $href = {};
my $cref = sub { 1 };

# A hashref numifies to its address. This runs, and is essentially never
# what was meant - unlike `@arr + 1`, which is the array's count and IS
# the ordinary idiom.
my $a = $href + 1;

# A coderef in a numeric comparison, likewise an address.
my $b = ($cref > 0);
