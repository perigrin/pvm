# ABOUTME: An isolated call case for Getopt::Long::HelpMessage and VersionMessage
# ABOUTME: with -exitval => "NOEXIT", the one form that returns rather than exiting.
(
    HelpMessage    => [ [ '-exitval', 'NOEXIT', '-message', 'm' ] ],
    VersionMessage => [ [ '-exitval', 'NOEXIT', '-message', 'm' ] ],
);
