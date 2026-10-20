# ABOUTME: Call cases for Scalar::Util's subs, which the return-type oracle calls to hold
# ABOUTME: Scalar/Util.tpm to perl. Each sub maps to a list of argument lists, one call per list.
use Scalar::Util qw(weaken dualvar);

my $obj = bless {}, 'Scalar::Util::OracleFixture';
my $plain = { a => 1 };
my $weak = $plain;
weaken($weak);
my $dualvar = dualvar( 5, 'five' );
my $coderef = sub { 42 };
open my $fh, '<', $0 or die "cannot open $0: $!";

(
    blessed           => [ [$obj], [ {} ], [5] ],
    dualvar           => [ [ 5, 'five' ] ],
    isdual            => [ [$dualvar], [5] ],
    isvstring         => [ [ v1.2.3 ], [5] ],
    isweak            => [ [$weak], [$plain] ],
    looks_like_number => [ ['5'], ['abc'] ],
    openhandle        => [ [$fh], ['x'] ],
    readonly          => [ [5] ],
    refaddr           => [ [$obj], [5] ],
    reftype           => [ [$obj], [5] ],
    set_prototype     => [ [ $coderef, '$$' ] ],
    tainted           => [ [5] ],
    unweaken          => [ [$weak] ],
    weaken            => [ [$plain] ],
);
