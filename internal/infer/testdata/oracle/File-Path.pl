# ABOUTME: Call cases for File::Path's subs, which the return-type oracle calls to hold File::Path.tpm to perl.
# ABOUTME: Every path is relative, so every create and remove stays inside the oracle's scratch directory.
mkdir 'rm_dir' or die $! unless -d 'rm_dir';
mkdir 'rt_dir' or die $! unless -d 'rt_dir';
(
    mkpath      => [ [ 'mk_dir/sub' ] ],
    make_path   => [ [ 'mp_dir/a', 'mp_dir/b' ] ],
    rmtree      => [ [ 'rm_dir' ], ['no_such_dir'] ],
    remove_tree => [ [ 'rt_dir' ] ],
);
