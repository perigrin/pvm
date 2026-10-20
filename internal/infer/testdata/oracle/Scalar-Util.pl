# ABOUTME: Call cases for Scalar::Util's subs, which the return-type oracle calls to hold
# ABOUTME: Scalar/Util.tpm to perl. Each sub maps to a list of argument lists, one call per list.
use Scalar::Util qw(weaken dualvar);
use IO::File;

my $obj = bless {}, 'Scalar::Util::OracleFixture';
my $plain = { a => 1 };
my $weak = $plain;
weaken($weak);
my $dualvar = dualvar( 5, 'five' );
my $coderef = sub { 42 };
open my $fh, '<', __FILE__ or die "cannot open @{[__FILE__]}: $!";
my $iofile = IO::File->new( __FILE__, 'r' ) or die "cannot open @{[__FILE__]}: $!";

(
    blessed           => [ [$obj], [ {} ], [5] ],
    dualvar           => [ [ 5, 'five' ] ],
    isdual            => [ [$dualvar], [5] ],
    isvstring         => [ [ v1.2.3 ], [5] ],
    isweak            => [ [$weak], [$plain] ],
    looks_like_number => [ ['5'], ['abc'] ],
    # openhandle returns the handle it was given unchanged (measured);
    # a blessed IO::File comes back as itself, an Object. (A bare
    # typeglob also comes back unchanged, measured `ref
    # \openhandle(*STDIN)` is GLOB, but this oracle's observer has no
    # Glob case of its own -- ref() of a non-reference is always "" --
    # so that case is not run here; see Scalar/Util.tpm's comment.)
    openhandle        => [ [$fh], ['x'], [$iofile] ],
    readonly          => [ [5] ],
    refaddr           => [ [$obj], [5] ],
    reftype           => [ [$obj], [5] ],
    # undef is documented to remove a sub's prototype ("If $prototype
    # is undef, then the function's prototype is removed"), measured
    # to work rather than die.
    set_prototype     => [ [ $coderef, '$$' ], [ $coderef, undef ] ],
    tainted           => [ [5] ],
    unweaken          => [ [$weak] ],
    weaken            => [ [$plain] ],
);
