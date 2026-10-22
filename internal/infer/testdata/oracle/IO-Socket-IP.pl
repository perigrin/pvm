# ABOUTME: Call cases for IO::Socket::IP's own methods, which the return-type oracle calls to hold
# ABOUTME: IO/Socket/IP.tpm to perl.
#
# new with no arguments only builds and blesses an unconnected socket, no
# network (measured); connect and the host/port readers open or read a
# real connection and are exempt (returntypes_test.go's oracleExempt), as
# are 161 of the 170 Socket names this module re-exports. The other
# nine -- AF_INET, sockaddr_family, SO_REUSEADDR, SO_REUSEPORT,
# SO_BROADCAST, SO_ERROR, SOCK_DGRAM, SOCK_STREAM and SOL_SOCKET -- are
# also imported directly into IO::Socket::IP's own package by its own
# file-scope `use Socket ... qw(...)` (measured), so the oracle's
# sub-form call reaches them here directly, with Socket.tpm's own call
# cases for the same names.
(
    new              => [ [] ],
    AF_INET          => [ [] ],
    SO_REUSEADDR     => [ [] ],
    SO_REUSEPORT     => [ [] ],
    SO_BROADCAST     => [ [] ],
    SO_ERROR         => [ [] ],
    SOCK_DGRAM       => [ [] ],
    SOCK_STREAM      => [ [] ],
    SOL_SOCKET       => [ [] ],
    sockaddr_family  => [
        [ Socket::pack_sockaddr_in( 80, Socket::INADDR_ANY() ) ],
        [ Socket::pack_sockaddr_un('/tmp/x') ],
    ],

    # Documented methods and overrides a prior pass of this .tpm left
    # out (IO/Socket/IP.tpm's own comment explains each), called on
    # INSTANCE: a single UDP socket built once by new(Domain=>AF_INET,
    # Type=>SOCK_DGRAM, Proto=>'udp', LocalAddr=>'127.0.0.1',
    # LocalPort=>0), which only does socket(2) and a local bind(2), no
    # connect, so it never touches the network. join_addr and
    # split_addr are class methods (the invocant is discarded), called
    # without an instance.
    as_inet => [ [] ],
    connected => [ [] ],
    join_addr => [ [ '::1', 80 ], [ 'h', undef ] ],
    peerhost_service => [ [1] ],
    peerhostname => [ [] ],
    peerservice => [ [] ],
    sockaddr => [ [] ],
    socket => [ [ Socket::AF_INET(), Socket::SOCK_DGRAM(), 0 ] ],
    sockhost_service => [ [1] ],
    sockhostname => [ [] ],
    sockservice => [ [] ],
    split_addr => [ [ '[::1]:80' ], [ 'x' ] ],
    INSTANCE => {
        ctor => 'new',
        args => [ Domain => Socket::AF_INET(), Type => Socket::SOCK_DGRAM(), Proto => 'udp',
            LocalAddr => '127.0.0.1', LocalPort => 0 ],
        subs => [qw(as_inet connected peerhost_service peerhostname peerservice
            sockaddr socket sockhost_service sockhostname sockservice)],
    },
);
