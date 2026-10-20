# ABOUTME: Call cases for Text::ParseWords's subs, which the return-type oracle calls to hold
# ABOUTME: Text::ParseWords.tpm to perl. One case per sub is an unbalanced quote, its issue's own case.
(
    shellwords        => [ [ 'a "b c" d' ], [ 'a "b c' ] ],
    quotewords        => [ [ '\s+', 0, 'a "b c" d' ], [ ',', 1, 'a,,"b"' ] ],
    nested_quotewords => [ [ '\s+', 0, 'a b', 'c' ] ],
    parse_line        => [ [ ',', 0, 'a,,b' ], [ '(,)', 'delimiters', 'a,b' ] ],
    old_shellwords    => [ [ q{a 'b c' d} ], [ q{a 'b c} ] ],
);
