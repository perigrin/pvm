# ABOUTME: Call cases for IO::Socket::IP's own methods, which the return-type oracle calls to hold
# ABOUTME: IO/Socket/IP.tpm to perl.
#
# new with no arguments only builds and blesses an unconnected socket, no
# network (measured); connect and the host/port readers open or read a
# real connection and are exempt (returntypes_test.go's oracleExempt), as
# are the 170 Socket names this module re-exports.
(
    new => [ [] ],
);
