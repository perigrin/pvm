# ABOUTME: Call cases for Cwd's subs, which the return-type oracle calls to hold Cwd.tpm to perl.
# ABOUTME: chdir(".") leaves the oracle's scratch directory unchanged for the calls after it.
(
    cwd           => [ [] ],
    getcwd        => [ [] ],
    fastcwd       => [ [] ],
    fastgetcwd    => [ [] ],
    chdir         => [ ['.'] ],
    abs_path      => [ ['.'] ],
    fast_abs_path => [ ['.'] ],
    realpath      => [ ['.'] ],
    fast_realpath => [ ['.'] ],
);
