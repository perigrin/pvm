# ABOUTME: Call cases for File::Basename's subs, including the Windows-path cases its issue names.
# ABOUTME: Each maps to a list of argument lists, one call per list.
(
    fileparse             => [ [ 'a/b.txt', '.txt' ], [ 'C:\dir\f.txt' ], [ '/foo/bar/' ] ],
    basename              => [ [ '/foo/bar' ], [ 'C:\dir\f.txt' ] ],
    dirname               => [ [ '/foo/bar/baz' ], [ 'C:\dir\f.txt' ] ],
    fileparse_set_fstype  => [ [] ],
);
