# ABOUTME: Call cases for HTTP::Tiny's methods, which the return-type oracle calls to hold
# ABOUTME: HTTP/Tiny.tpm to perl.
#
# new, www_form_urlencode and can_ssl never open a connection or read
# instance state, so they are safe on the bare class name. agent,
# timeout and connected read $self->{...} (measured: each dies "Can't
# use string ... as a HASH ref" on the bare class name), so INSTANCE
# names the constructor return_oracle.pl calls once instead --
# HTTP::Tiny->new with no arguments, measured side-effect free: it only
# builds a hashref and reads %ENV for proxies, never the network. The
# rest open a real connection (returntypes_test.go's oracleExempt).
(
    new                => [ [] ],
    www_form_urlencode => [ [ { a => 1, b => 'x y' } ], [ [ a => 1, b => 2 ] ] ],
    can_ssl            => [ [] ],
    agent              => [ [], ['pvm-oracle/1'] ],
    timeout            => [ [], [5] ],
    connected          => [ [] ],
    INSTANCE           => { ctor => 'new', args => [], subs => [qw(agent timeout connected)] },
);
