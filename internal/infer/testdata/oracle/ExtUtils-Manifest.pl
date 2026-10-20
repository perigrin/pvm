# ABOUTME: Call cases for ExtUtils::Manifest's subs, which the return-type oracle calls to hold ExtUtils::Manifest.tpm to perl.
# ABOUTME: Every path is relative, so every read and write stays inside the oracle's scratch directory.
use ExtUtils::Manifest qw(mkmanifest manicheck filecheck fullcheck skipcheck manifind maniread manicopy maniadd maniskip);
open my $fh, '>', 'tracked.txt' or die $!;
print $fh "one\n";
close $fh;
mkmanifest();
open my $fh2, '>', 'untracked.txt' or die $!;
print $fh2 "two\n";
close $fh2;
(
    manifind   => [ [] ],
    manicheck  => [ [] ],
    filecheck  => [ [] ],
    fullcheck  => [ [] ],
    skipcheck  => [ [] ],
    maniread   => [ [] ],
    maniskip   => [ [] ],
    maniadd    => [ [ { 'untracked.txt' => 'a comment' } ] ],
    manicopy   => [ [ { 'tracked.txt' => '' }, 'copydst' ] ],
    mkmanifest => [ [] ],
);
