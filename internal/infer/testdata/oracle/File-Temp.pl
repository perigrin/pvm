# ABOUTME: Call cases for File::Temp's subs, which the return-type oracle calls to hold File::Temp.tpm to perl.
# ABOUTME: Every template is relative, so every file and directory File::Temp creates stays inside the oracle's scratch directory.
open my $ufh, '>', 'u1' or die $!;
close $ufh;
open my $ufh2, '>', 'u1' or die $!;
(
    SEEK_SET  => [ [] ],
    SEEK_CUR  => [ [] ],
    SEEK_END  => [ [] ],
    cleanup   => [ [] ],
    mkdtemp   => [ [ 'mdXXXXXX' ] ],
    mkstemp   => [ [ 'msXXXXXX' ] ],
    mkstemps  => [ [ 'mtXXXXXX', 'suf' ] ],
    mktemp    => [ [ 'mkXXXXXX' ] ],
    tempdir   => [ [] ],
    tempfile  => [ [] ],
    tmpfile   => [ [] ],
    tmpnam    => [ [] ],
    unlink0   => [ [ $ufh2, 'u1' ] ],
);
