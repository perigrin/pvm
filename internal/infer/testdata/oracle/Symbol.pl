# ABOUTME: Call cases for Symbol's subs, which the return-type oracle calls to hold Symbol.tpm to perl.
# ABOUTME: $Foo::Bar::x is set first so delete_package has a package to find and remove.
use IO::Handle;
$Foo::Bar::x = 1;
my $io_handle = IO::Handle->new;
(
    gensym         => [ [] ],
    geniosym       => [ [] ],
    ungensym       => [ [1] ],
    qualify        => [ ['x'], [ 'x', 'Foo' ], ['BAR::x'], [ \*STDOUT ] ],
    # a blessed reference whose referent is a glob (IO::Handle) is
    # accepted and preserved unchanged, not only a bare GlobRef.
    qualify_to_ref => [ ['x'], [ \*STDOUT ], [$io_handle] ],
    delete_package => [ ['Foo::Bar'], ['No::Such'] ],
);
