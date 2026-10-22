# ABOUTME: Call cases for File::Temp's subs, which the return-type oracle calls to hold File::Temp.tpm to perl.
# ABOUTME: Every template is relative, so every file and directory File::Temp creates stays inside the oracle's scratch directory.
open my $ufh, '>', 'u1' or die $!;
close $ufh;
open my $ufh2, '>', 'u1' or die $!;
# cmpstat and unlink1 each stat() the handle itself, which only answers
# while it is still open (measured: closed, stat returns empty and
# cmpstat answers undef, not 0), so $ufh3 and $ufh4 stay open; unlink1
# closes $ufh4 itself once it is done with it.
open my $ufh3, '>', 'u2' or die $!;
open my $ufh4, '>', 'u3' or die $!;
(
    SEEK_SET           => [ [] ],
    SEEK_CUR           => [ [] ],
    SEEK_END           => [ [] ],
    cleanup            => [ [] ],
    mkdtemp            => [ [ 'mdXXXXXX' ] ],
    mkstemp            => [ [ 'msXXXXXX' ] ],
    mkstemps           => [ [ 'mtXXXXXX', 'suf' ] ],
    mktemp             => [ [ 'mkXXXXXX' ] ],
    tempdir            => [ [] ],
    tempfile           => [ [] ],
    tmpfile            => [ [] ],
    tmpnam             => [ [] ],
    unlink0            => [ [ $ufh2, 'u1' ] ],
    tempnam            => [ [ '.', 'pre' ] ],
    cmpstat            => [ [ $ufh3, 'u2' ] ],
    unlink1            => [ [ $ufh4, 'u3' ] ],
    new                => [ [] ],
    newdir             => [ [] ],
    filename           => [ [] ],
    unlink_on_destroy  => [ [] ],
    safe_level         => [ [] ],
    top_system_uid     => [ [] ],
    INSTANCE           => { ctor => 'new', args => [], subs => [qw(filename unlink_on_destroy)] },
);
