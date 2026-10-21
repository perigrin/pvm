# ABOUTME: Call cases for File::Spec::Unix's methods, which the return-type oracle calls to hold
# ABOUTME: File/Spec/Unix.tpm to perl. Every path given is a defined Str, as each method requires.
(
    canonpath => [ [ 'a/b/../c' ], [ '' ], [ '/a//b/./c/' ] ],
    catdir    => [ [ 'a', 'b' ], [ 'a' ], [] ],
    catfile   => [ [ 'a', 'b' ], [ 'f.txt' ], [] ],
    join      => [ [ 'a', 'b' ], [ 'f.txt' ], [] ],
    catpath   => [ [ '', 'a/b', 'c.txt' ], [ '', '', '' ] ],
    curdir    => [ [] ],
    devnull   => [ [] ],
    rootdir   => [ [] ],
    updir     => [ [] ],
    tmpdir    => [ [] ],
    case_tolerant => [ [] ],
    file_name_is_absolute => [ [ '/a/b' ], [ 'a/b' ] ],
    path      => [ [] ],
    no_upwards => [ [ '.', '..', 'a', 'b' ], [ 'a' ] ],
    splitpath => [ [ '/a/b/c.txt' ], [ '/a/b/', 1 ] ],
    splitdir  => [ [ '/a/b//c/' ], [ 'a' ] ],
    abs2rel   => [ [ '/a/b', '/a' ], [ '/a/b' ] ],
    rel2abs   => [ [ 'b', '/a' ], [ '/a/b' ] ],
);
