# ABOUTME: Call cases for File::Find's subs, which the return-type oracle calls to hold File::Find.tpm to perl.
# ABOUTME: The walked directory is relative, so the walk stays inside the oracle's scratch directory.
mkdir 'd' unless -d 'd';
open my $fh, '>', 'd/f' or die $!;
close $fh;
(
    find      => [ [ sub { }, 'd' ], [ sub { } ] ],
    finddepth => [ [ sub { }, 'd' ], [ sub { } ] ],
);
