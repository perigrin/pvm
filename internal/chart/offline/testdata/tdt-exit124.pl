# ABOUTME: Exits 124 from BEGIN, the status timeout gives a killed perl, with no unit open.
# ABOUTME: tdt-capture.sh must call this refused, never a timeout.
BEGIN { exit 124 }
my $x = 1;
