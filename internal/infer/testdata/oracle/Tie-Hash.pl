# ABOUTME: Call cases for Tie::Hash's methods, which the return-type oracle calls to hold
# ABOUTME: Tie/Hash.tpm to perl. Every one is called directly on Tie::Hash, which always dies.
(
    new     => [ [] ],
    TIEHASH => [ [] ],
    EXISTS  => [ [ 'k' ] ],
    CLEAR   => [ [] ],
);
