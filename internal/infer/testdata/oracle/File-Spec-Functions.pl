# ABOUTME: Call cases for File::Spec::Functions's subs, which the return-type oracle calls to hold File::Spec::Functions.tpm to perl.
# ABOUTME: Every sub is a pure path computation, touching no file; the paths given are illustrative, not read from disk.
(
    abs2rel               => [ [ 'a/b/c', 'a' ] ],
    canonpath             => [ [ 'a/./b' ] ],
    case_tolerant         => [ [] ],
    catdir                => [ [ 'a', 'b' ] ],
    catfile               => [ [ 'a', 'b' ] ],
    catpath               => [ [ '', 'a/b', 'c.txt' ] ],
    curdir                => [ [] ],
    devnull               => [ [] ],
    file_name_is_absolute => [ [ 'a/b' ] ],
    no_upwards            => [ [ 'a', '..', 'b' ] ],
    path                  => [ [] ],
    rel2abs               => [ [ 'b/c', 'a' ] ],
    rootdir               => [ [] ],
    splitdir              => [ [ 'a/b/c' ] ],
    splitpath             => [ [ 'a/b/c.txt' ] ],
    tmpdir                => [ [] ],
    updir                 => [ [] ],
);
