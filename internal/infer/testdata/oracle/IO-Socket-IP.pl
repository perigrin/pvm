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
);
