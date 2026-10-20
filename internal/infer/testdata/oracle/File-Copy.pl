# ABOUTME: Call cases for File::Copy's subs, which the return-type oracle calls to hold File::Copy.tpm to perl.
# ABOUTME: Every path is relative, so every copy and move stays inside the oracle's scratch directory.
open my $fh, '>', 'src' or die $!;
close $fh;
#
# Sub names run in alphabetical order (copy, cp, move, mv), so each
# later sub's source is one an earlier sub already created.
(
    copy => [ [ 'src', 'copy-dst' ], [ 'no-such-src', 'copy-fail-dst' ] ],
    cp   => [ [ 'src', 'cp-dst' ] ],
    move => [ [ 'copy-dst', 'move-dst' ], [ 'no-such-src', 'move-fail-dst' ] ],
    mv   => [ [ 'cp-dst', 'mv-dst' ] ],
);
