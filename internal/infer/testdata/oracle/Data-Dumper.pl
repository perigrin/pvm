# ABOUTME: Call cases for Data::Dumper's subs, which the return-type oracle calls to hold
# ABOUTME: Data/Dumper.tpm to perl. INSTANCE runs the OO methods on one constructed object.
(
    Dumper  => [ [ 1, 2 ] ],
    DumperX => [ [ 1, 2 ] ],

    new           => [ [ [ 1, 2 ], [ 'a', 'b' ] ] ],
    Dump          => [ [] ],
    Seen          => [ [], [ {} ] ],
    Values        => [ [], [ [ 9, 9 ] ] ],
    Names         => [ [], [ [ 'c', 'd' ] ] ],
    Reset         => [ [] ],
    Indent        => [ [], [1] ],
    Trailingcomma => [ [], [1] ],
    Pair          => [ [], [' - '] ],
    Pad           => [ [], ['  '] ],
    Varname       => [ [], ['X'] ],
    Purity        => [ [], [1] ],
    Useqq         => [ [], [1] ],
    Terse         => [ [], [1] ],
    Freezer       => [ [], [''] ],
    Toaster       => [ [], [''] ],
    Deepcopy      => [ [], [1] ],
    Quotekeys     => [ [], [1] ],
    Bless         => [ [], ['bless'] ],
    Maxdepth      => [ [], [0] ],
    Maxrecurse    => [ [], [100] ],
    Useperl       => [ [], [0] ],
    Sortkeys      => [ [], [1] ],
    Deparse       => [ [], [0] ],
    Sparseseen    => [ [], [0] ],

    INSTANCE => {
        ctor => 'new',
        args => [ [ 1, 2 ], [ 'a', 'b' ] ],
        subs => [
            qw(Dump Seen Values Names Reset
               Indent Trailingcomma Pair Pad Varname Purity Useqq Terse
               Freezer Toaster Deepcopy Quotekeys Bless Maxdepth
               Maxrecurse Useperl Sortkeys Deparse Sparseseen)
        ],
    },
);
