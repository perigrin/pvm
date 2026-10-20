# ABOUTME: Call cases for Symbol's subs, which the return-type oracle calls to hold Symbol.tpm to perl.
# ABOUTME: $Foo::Bar::x is set first so delete_package has a package to find and remove.
$Foo::Bar::x = 1;
(
    gensym         => [ [] ],
    geniosym       => [ [] ],
    ungensym       => [ [1] ],
    qualify        => [ ['x'], [ 'x', 'Foo' ], ['BAR::x'], [ \*STDOUT ] ],
    qualify_to_ref => [ ['x'], [ \*STDOUT ] ],
    delete_package => [ ['Foo::Bar'], ['No::Such'] ],
);
