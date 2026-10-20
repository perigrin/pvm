# ABOUTME: Call cases for Getopt::Long's subs, which the return-type oracle calls to hold
# ABOUTME: Getopt::Long.tpm to perl. HelpMessage and VersionMessage are exempt; see there.
my $foo;
@ARGV = ( '--foo', 'bar' );
my $x;
my @fromArray = ( '--x', '1' );
(
    GetOptions           => [ [ 'foo=s', \$foo ] ],
    GetOptionsFromArray   => [ [ \@fromArray, 'x=s', \$x ] ],
    GetOptionsFromString => [ [ '--y 2', 'y=s', \my $y ] ],
    Configure             => [ ['bundling'] ],
);
