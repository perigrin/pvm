# ABOUTME: Call cases for Term::ReadLine's methods, which the return-type oracle calls to hold
# ABOUTME: Term/ReadLine.tpm to perl.
#
# readline is not here: it blocks on a real terminal, so it is exempt
# (returntypes_test.go's oracleExempt) and never called.
(
    new         => [ [ 'test' ] ],
    ReadLine    => [ [] ],
    findConsole => [ [] ],
    addhistory  => [ [ 'test line' ] ],
    MinLine     => [ [ 5 ] ],
    Attribs     => [ [] ],
    Features    => [ [] ],
);
