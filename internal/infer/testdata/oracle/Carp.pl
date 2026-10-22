# ABOUTME: Call cases for Carp's subs, which the return-type oracle calls to hold Carp.tpm to perl.
# ABOUTME: Each sub maps to a list of argument lists, one call per list.
(
    carp      => [ ['warned'], [ 'two', ' parts' ], [ { c => 1 } ] ],
    cluck     => [ ['clucked'], [ [1, 2] ] ],
    croak     => [ ['croaked'], [ { error => 1 } ] ],
    confess   => [ ['confessed'] ],
    longmess  => [ ['long'], [], [ { x => 1 }, 'b' ] ],
    shortmess => [ ['short'], [], [ { x => 1 }, 'b' ] ],
);
