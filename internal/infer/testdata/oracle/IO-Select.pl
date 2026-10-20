# ABOUTME: Call cases for IO::Select's methods, which the return-type oracle calls to hold
# ABOUTME: IO/Select.tpm to perl.
#
# Only new is callable here; the oracle's method call passes the class
# name as the invocant, never a constructed instance
# (returntypes_test.go's oracleExempt), so every other method, which
# needs one, is exempt.
(
    new => [ [] ],
);
