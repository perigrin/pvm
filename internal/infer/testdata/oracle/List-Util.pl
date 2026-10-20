# ABOUTME: Call cases for List::Util's subs, which the return-type oracle calls to hold
# ABOUTME: List/Util.tpm to perl. A block parameter's argument is the coderef the driver calls
# ABOUTME: it with, since the driver invokes every sub through a coderef rather than its own syntax.

my $gt1  = sub { $_ > 1 };
my $sum  = sub { $a + $b };
my $pmap = sub { ( $a, $b * 2 ) };
my $pgt1 = sub { $b > 1 };

(
    first     => [ [ $gt1, 1, 2, 3 ] ],
    any       => [ [ $gt1, 1, 2, 3 ] ],
    all       => [ [ $gt1, 1, 2, 3 ] ],
    none      => [ [ $gt1, 1, 2, 3 ] ],
    notall    => [ [ $gt1, 1, 2, 3 ] ],
    reduce    => [ [ $sum, 1, 2, 3 ] ],
    reductions => [ [ $sum, 1, 2, 3 ] ],
    pairmap   => [ [ $pmap, a => 1, b => 2 ] ],
    pairgrep  => [ [ $pgt1, a => 1, b => 2 ] ],
    pairfirst => [ [ $pgt1, a => 1, b => 2 ] ],
    sum       => [ [ 1, 2, 3 ], [], [ 1e308, 1e308 ], [ 'nan' ] ],
    sum0      => [ [ 1, 2, 3 ], [ 'inf' ] ],
    max       => [ [ 1, 2, 3 ], [ 'abc' ], [ [1] ] ],
    min       => [ [ 1.5, 2, 3 ], [ 'abc' ] ],
    maxstr    => [ [ 'a', 'b' ], [ [1] ] ],
    minstr    => [ [ 'a', 'b' ], [ [1] ] ],
    product   => [ [ 2, 3, 4 ], [ 'inf' ] ],
    shuffle   => [ [ 1, 2, 3 ] ],
    uniq      => [ [ 1, 1, 2 ] ],
    uniqnum   => [ [ 1, '1', 2 ] ],
    uniqstr   => [ [ 'a', 'a', 'b' ] ],
    uniqint   => [ [ 1, 1, 2 ] ],
    head      => [ [ 2, 1, 2, 3, 4 ] ],
    tail      => [ [ 2, 1, 2, 3, 4 ] ],
    sample    => [ [ 2, 1, 2, 3, 4 ] ],
    mesh         => [ [ [ 1, 2 ], [ 3, 4 ] ] ],
    zip          => [ [ [ 1, 2 ], [ 3, 4 ] ] ],
    zip_longest  => [ [ [ 1, 2 ], [3] ] ],
    zip_shortest => [ [ [ 1, 2 ], [3] ] ],
    mesh_longest  => [ [ [ 1, 2 ], [3] ] ],
    mesh_shortest => [ [ [ 1, 2 ], [3] ] ],
    pairs      => [ [ a => 1, b => 2 ] ],
    pairkeys   => [ [ a => 1, b => 2 ] ],
    pairvalues => [ [ a => 1, b => 2 ] ],
    unpairs    => [ [ [ a => 1 ], [ b => 2 ] ] ],
);
