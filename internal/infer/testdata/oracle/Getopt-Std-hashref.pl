# ABOUTME: An isolated call case for Getopt::Std::getopt's hashref form, kept
# ABOUTME: apart from Getopt-Std.pl so adding it cannot disturb that file's @ARGV staging.
@ARGV = ( '-b', 'v' );
(
    getopt => [ [ 'b', {} ] ],
);
