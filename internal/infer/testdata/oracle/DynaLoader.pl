# ABOUTME: Return-type oracle call cases for DynaLoader's safe subs; the rest are exempt.
# ABOUTME: Each case is measured on perl 5.42.0 with nothing bootstrapped yet.
(
    croak                   => [ [ 'oracle test message' ], [ { x => 1 } ] ],
    dl_load_flags           => [ [] ],
    dl_findfile             => [ [ 'nosuchlib12345' ], [ '-lc' ] ],
    dl_expandspec           => [ [ '/no/such/file' ] ],
    dl_find_symbol_anywhere => [ [ 'nosuchsymbol12345' ] ],
    dl_undef_symbols        => [ [] ],
    dl_error                => [ [] ],
);
