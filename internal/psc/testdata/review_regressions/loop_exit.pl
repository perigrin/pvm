# Past a last or next its guard sends out of the loop, the rest of the
# loop body knows the guard failed: IPC/Cmd.pm, File/Copy.pm, Net/FTP.pm,
# Test2/API/Instance.pm.
use strict;
use warnings;

sub drain {
    my ($h, $kid) = @_;
  OUTER: while (1) {
        my $len = sysread($h, my $buf, 4096);
        if (not defined $len) {
            warn "Error reading: $!";
            last OUTER;
        }
        print "done\n" if $h == $kid and $len == 0;
        last if $len == 0;
    }
}

sub copy_fh {
    my ($from_h, $to_h) = @_;
    for (;;) {
        my ($r, $w, $t);
        defined($r = sysread($from_h, my $buf, 1024)) or return 0;
        last unless $r;
        for ($w = 0; $w < $r; $w += $t) {
            $t = syswrite($to_h, $buf, $r - $w, $w) or return 0;
        }
    }
    return 1;
}

sub send_fh {
    my ($loc, $sock) = @_;
    my ($len, $buf);
    while (1) {
        last unless $len = read($loc, $buf, 1024);
        my $wlen;
        unless (defined($wlen = syswrite($sock, $buf, $len)) && $wlen == $len) {
            return 0;
        }
    }
    return 1;
}

sub errors {
    my $fail = 0;
    for my $t (@_) {
        my $err = $t ? lc($t) : undef;
        next unless $err;
        chomp($err);
        $fail++;
    }
    return $fail;
}

open my $in, '<', $0 or die;
drain($in, 0);
open my $null, '>', '/dev/null' or die;
seek $in, 0, 0;
print copy_fh($in, $null), "\n";
seek $in, 0, 0;
print send_fh($in, $null), "\n";
print errors("A\n", "", "B"), "\n";
