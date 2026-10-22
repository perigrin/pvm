# ABOUTME: Call cases for Term::ReadLine's methods, which the return-type oracle calls to hold
# ABOUTME: Term/ReadLine.tpm to perl.
#
# readline is not here: it blocks on a real terminal, so it is exempt
# (returntypes_test.go's oracleExempt) and never called. IN, OUT,
# ornaments, tkRunning and event_loop each read instance state the
# bare class name has none of, so INSTANCE builds one from
# new('test', \*STDIN, \*STDOUT) for return_oracle.pl to call them on
# instead (measured side-effect free: Term::ReadLine::Stub's new only
# stores the handles).
(
    new         => [ [ 'test' ] ],
    ReadLine    => [ [] ],
    findConsole => [ [] ],
    addhistory  => [ [ 'test line' ] ],
    MinLine     => [ [ 5 ] ],
    Attribs     => [ [] ],
    Features    => [ [] ],
    IN          => [ [] ],
    OUT         => [ [] ],
    ornaments   => [ [] ],
    newTTY      => [ [ \*STDIN, \*STDOUT ] ],
    tkRunning   => [ [] ],
    event_loop  => [ [] ],
    INSTANCE    => { ctor => 'new', args => [ 'test', \*STDIN, \*STDOUT ], subs => [qw(IN OUT ornaments newTTY tkRunning event_loop)] },
);
