# ABOUTME: Sends perl -Dp into its endless loop: an :encoding layer on STDERR set in BEGIN.
# ABOUTME: tdt-capture.sh must end this with a timeout marker, never a whole stream.
BEGIN { binmode STDERR, ":encoding(UTF-8)" }
my $x = 1;
