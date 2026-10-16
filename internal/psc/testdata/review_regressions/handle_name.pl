# open takes a scalar holding a handle's name, without strict refs, and
# opens the handle it names: Devel/SelfStubber.pm.
use warnings;

sub stub {
    my ($module, $mod_file) = @_;
    my ($line, $fh);
    $fh = "${module}::DATA";
    open($fh, '<', $mod_file) || die "Unable to open $mod_file";
    $line = <$fh>;
    if (fileno($fh)) {
        print "open\n";
    }
    close($fh);
}

stub("main", $0);
