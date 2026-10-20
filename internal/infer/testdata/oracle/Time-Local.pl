# ABOUTME: Call cases for Time::Local's subs, which the return-type oracle calls to hold Time::Local.tpm
# ABOUTME: to perl. Every date given is in range, so none of these dies.
(
    timegm           => [ [ 0, 0, 0, 1, 0, 2000 ], [ 1.5, 0, 0, 1, 0, 100 ] ],
    timelocal        => [ [ 0, 0, 0, 1, 0, 70 ] ],
    timegm_modern    => [ [ 0, 0, 0, 1, 0, 1 ] ],
    timelocal_modern => [ [ 0, 0, 0, 1, 0, 1 ] ],
    timegm_nocheck   => [ [ 0, 0, 0, 40, 0, 2000 ] ],
    timelocal_nocheck => [ [ 0, 0, 0, 40, 0, 2000 ] ],
    timegm_posix     => [ [ 0, 0, 0, 1, 0, 100 ] ],
    timelocal_posix  => [ [ 0, 0, 0, 1, 0, 100 ] ],
);
