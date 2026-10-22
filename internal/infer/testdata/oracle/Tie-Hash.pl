# ABOUTME: Call cases for Tie::Hash's methods, which the return-type oracle calls to hold
# ABOUTME: Tie/Hash.tpm to perl. EXISTS is called directly on Tie::Hash, which always dies; new, TIEHASH and CLEAR are exempt (ooExempt) since the bare class name always dies too, but that is not the only path they are declared for.
(
    EXISTS  => [ [ 'k' ] ],
);
