# ABOUTME: Call cases for Carp's subs, which the return-type oracle calls to hold Carp.tpm to perl.
# ABOUTME: Each sub maps to a list of argument lists, one call per list.
(
    carp      => [ ['warned'], [ 'two', ' parts' ] ],
    cluck     => [ ['clucked'] ],
    croak     => [ ['croaked'], [ { error => 1 } ] ],
    confess   => [ ['confessed'] ],
    longmess  => [ ['long'], [] ],
    shortmess => [ ['short'], [] ],
);
