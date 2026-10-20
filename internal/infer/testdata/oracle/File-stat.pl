# ABOUTME: Call cases for File::stat's subs, which the return-type oracle calls to hold
# ABOUTME: File::stat.tpm to perl. Every path is relative, staying inside the oracle's scratch cwd.
open my $fh, '>', 'oracle-file-stat.tmp' or die $!;
close $fh;
my $st = File::stat::stat('oracle-file-stat.tmp');
(
    stat        => [ ['oracle-file-stat.tmp'], ['no-such-file'] ],
    lstat       => [ ['oracle-file-stat.tmp'], ['no-such-file'] ],
    stat_cando  => [ [ $st, 0400, 1 ] ],
);
