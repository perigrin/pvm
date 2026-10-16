# A guard on an assignment or a declaration guards the variable it
# assigns: ExtUtils/Install.pm, File/Compare.pm, IPC/Cmd.pm,
# Memoize/Expire.pm, Net/FTP/A.pm, Sys/Syslog.pm.
use strict;
use warnings;

sub copy_fh {
    my ($from, $to) = @_;
    while (my $len = sysread($from, my $buf, 1024)) {
        syswrite($to, $buf, $len);
    }
}

sub same_fh {
    my ($from, $to) = @_;
    my ($fr, $tr, $fbuf, $tbuf);
    while (defined($fr = read($from, $fbuf, 1024)) && $fr > 0) {
        unless (defined($tr = read($to, $tbuf, $fr)) && $tbuf eq $fbuf) {
            return 0;
        }
    }
    return 0 if defined($tr = read($to, $tbuf, 1024)) && $tr > 0;
    return 1;
}

sub reap {
    my $pid;
    if ($pid = fork) {
        waitpid($pid, 0);
    } elsif (defined $pid) {
        exit 0;
    }
    if (my $kid = fork) {
        waitpid($kid, 0);
    } elsif (defined $kid) {
        exit 0;
    }
}

sub first_v {
    my %h = @_;
    while (defined(my $key = each %h)) {
        return substr $key, 1 if 'V' eq substr $key, 0, 1;
    }
    return;
}

sub chunk {
    my ($fh) = @_;
    my $n;
    if ($n = sysread($fh, my $buf, 10)) {
        print "got $n\n";
    } else {
        return unless defined $n;
    }
    return $n > 0 ? 1 : 0;
}

open my $in, '<', $0 or die;
open my $in2, '<', $0 or die;
copy_fh($in, \*STDOUT) if @ARGV;
seek $in, 0, 0;
print same_fh($in, $in2), "\n";
reap();
print first_v(a => 1, Vb => 2) // "none", "\n";
seek $in, 0, 0;
print chunk($in) // "error", "\n";
