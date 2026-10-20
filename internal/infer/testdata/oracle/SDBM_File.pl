# ABOUTME: Call cases for SDBM_File's subs, which the return-type oracle calls to hold SDBM_File.tpm to perl.
# ABOUTME: The tie below names a relative file, so its .dir/.pag land under the oracle's scratch directory.
use Fcntl;
use SDBM_File;
my %h;
tie( %h, 'SDBM_File', 'sdbm_oracle', O_RDWR | O_CREAT, 0644 ) or die $!;
$h{key} = 'value';
untie %h;
(
    DIRFEXT => [ [] ],
    PAGFEXT => [ [] ],
    PAIRMAX => [ [] ],
);
