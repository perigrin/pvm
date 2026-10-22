# ABOUTME: Call cases for Getopt::Std's subs, which the return-type oracle calls to hold Getopt::Std.tpm
# ABOUTME: to perl. getopt runs first (alphabetically), so @ARGV is staged for both calls in sequence.
#
# After getopt consumes '-b v --', getopts('a') is left '-a -z': -z is
# not in its spec, so this one call covers both getopts' true return
# (the known -a) and its false one (the unknown -z warns and counts an
# error, its issue's own case).
@ARGV = ( '-b', 'v', '--', '-a', '-z' );
(
    getopt  => [ ['b'] ],
    getopts => [ ['a'] ],
);
